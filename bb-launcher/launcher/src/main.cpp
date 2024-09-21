#include <bb-launcher-driver.h>
#include <cstdlib>
#include <filesystem>
#include <fstream>
#include <peconv.h>
#include <physfs.h>
#include <stdlib.h>
#include <string.h>
#include <string>
#include <thread>

#include "hooks/CoherentGT.h"
#include "hooks/FileManager.h"
#include "hooks/HtmlDevice.h"
#include "hooks/ICore.h"
#include "hooks/SteamAPI.h"
#include "hooks/atexit.h"
#include "hooks/bbBridge.h"
#include "hooks/physfs.h"
#include "hooks/sqvm.h"
#include "hooks/textRender.h"
#include "hooks/trace.hpp"

#include "digest.hpp"
#include "plugins.h"
#include "sqext.h"

using namespace std;
using namespace peconv;

namespace launcher {
int launch(LauncherMetadata *);
} // namespace launcher

namespace setup {
enum class Priority {
    p0,
    p1,
    p2,
    none,
};
void register_setup_task(Priority p, std::string name, std::function<bool(launcher::LauncherMetadata *)> task);
} // namespace setup

int main(int argc, char *argv[]) {
    auto metadata = launcher::DetectLauncherMetadata();
#ifdef BUILD_WITH_ONLINE_MOD
    setup::register_setup_task(setup::Priority::p2, "setup_online_battle_mod", online_battle::mount_mod_to_physfs);
#endif
    try {
        if (launcher::setup_launcher(metadata, argv[0])) {
            launcher::launch(metadata);
        }
    }
    catch (const std::exception &e) {
        std::cerr << e.what() << '\n';
    }
}

namespace launcher {
static size_t v_size = 0;
static BYTE *loaded_pe;
static void(_cdecl *ep_func)();
static std::thread th;
static std::thread nakama_thread;
static std::filesystem::path GamePath;

bool setup_pe(LauncherMetadata *metadata) {
    GamePath = metadata->GetGamePath();
    bool isSteam;
    std::cout << "[*] Validating..." << std::endl;
    std::string digest = metadata->GetBattleBrothersExeDigest();
    validate_program_hash(digest, isSteam);
    std::cout << "[*] Validated" << std::endl;

    // 注入 steam_appid.txt
    if (isSteam) {
        std::cout << "[*] creating win32/steam_appid.tx" << std::endl;
        auto steam_appid_file = metadata->GetGamePath() / "win32" / "steam_appid.txt";
        std::fstream fs;
        fs.open(steam_appid_file, std::ios::out);
        fs << "365360" << std::endl;
        fs.flush();
        fs.close();
        if (!std::filesystem::exists(steam_appid_file)) {
            std::cout << "failed to create " << steam_appid_file << std::endl;
            system("pause");
            return false;
        }
    }

    std::cout << "[*] Loading PE..." << std::endl;
    hooking_func_resolver resolver;
    resolver.add_hook("InitializeUIGTSystem", (FARPROC)&patchs::CoherentGT::InitializeUIGTSystem::patch);
    resolver.add_hook("SteamAPI_Init", (FARPROC)&patchs::SteamAPI::SteamAPI_Init::patch);

    loaded_pe =
        load_pe_executable((BYTE *)metadata->GetBattleBrothersExeContent(), metadata->GetBattleBrothersExeSize(),
                           v_size, (peconv::t_function_resolver *)&resolver);

    std::cout << "[*] Loaded, size: " << v_size << std::endl;

    std::cout << "[*] Initializing..." << std::endl;
    LauncherContext launcherCtx = {digest,
                                   loaded_pe,
                                   metadata->GetGamePath(),
                                   metadata->GetGamePath() / "awesome-battle-brothers" / "plugins",
                                   nullptr,
                                   patchs::physfs::GetPhysfsLibrary()};

    patchs::physfs::UpgrateToPhysfs3(loaded_pe, metadata->GetGamePath());

    patchs::bigmap::PatchBigMapTextRender(loaded_pe, digest);
    patchs::ICore::InstallICoreHook(loaded_pe, digest);
    patchs::atexit::InstallAtExit(loaded_pe, digest, []() { launcher::atexit(); });

#ifdef BUILD_WITH_ONLINE_MOD
    patchs::RedirectLogs::InstallRedirectLogsHook(loaded_pe, digest);
#endif

    patchs::FileManager::InstallFileManagerHook(loaded_pe, digest, [metadata]() {
#ifdef BUILD_WITH_ONLINE_MOD
        auto mp = (metadata->GetGamePath() / "win32\\..\\data\\").string();
        PHYSFS_unmount(mp.data());
#endif
    });

    patchs::vm::PatchSQVM(loaded_pe, digest, [&launcherCtx](HSQUIRRELVM root_vm) {
        launcherCtx.root_vm = root_vm;
        // 触发插件回调
        plugins::on_squirrel_vm_init(&launcherCtx);

#ifdef BUILD_WITH_ONLINE_MOD
        sqext_register_jsonlib(root_vm);
        sqext_register_msgpacklib(root_vm);
        sqext_register_bufferio(root_vm);
        sqext_register_random(root_vm);
        sqext_register_print(root_vm);
        sqext_register_online_battle(root_vm);
        nakama::setup_client();
        nakama_thread = std::thread(nakama::fetch_tick);
#endif
    });

    patchs::BattleBrothers::InstallBattleBrothersInitHooks(loaded_pe, digest, [&launcherCtx, digest]() {
        if (launcherCtx.root_vm != nullptr) {
            sqext_override_isDevmode(launcherCtx.root_vm);
        }
    });
    patchs::CoherentGT::LoadCoherentGTLib(metadata->GetGamePath());

    std::cout << "[*] Loading Plugins..." << std::endl;
    plugins::load_plugins(&launcherCtx);
    std::cout << "[*] Initialized" << std::endl;
    return true;
}

int launch(LauncherMetadata *metadata) {
    std::cout << "[*] Running..." << std::endl;
    // if the loaded PE needs to access resources, you may need to connect it to
    // the PEB: peconv::set_main_module_in_peb((HMODULE)loaded_pe);

    // calculate the Entry Point of the manually loaded module
    DWORD ep_rva = peconv::get_entry_point_rva(loaded_pe);
    if (!ep_rva) {
        return -2;
    }

    ULONGLONG ep_exp_offset = (ULONGLONG)loaded_pe + ep_rva;
    ep_func = (void(_cdecl *)())(ep_exp_offset);

    // call the Entry Point of the manually loaded PE:
    ep_func();

    peconv::free_pe_buffer(loaded_pe, v_size);
    return 0;
}

} // namespace launcher
