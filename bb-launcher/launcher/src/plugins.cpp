#include "plugins.h"
#include <format>
#include <iostream>
#include <map>
#include <windows.h>

namespace plugins {
static std::map<std::string, HMODULE> plugins;
static bool validate_plugin(HMODULE plugin) {
    bool validated = false;

    for (auto func : std::vector<std::string>{
             "on_squirrel_vm_init",
             "on_squirrel_vm_init_ext",
         }) {
        FARPROC nOffset = GetProcAddress(plugin, (LPCSTR)func.c_str());
        if (nOffset == NULL) {
            continue;
        }
        validated = true;
        break;
    }
    return validated;
}

std::map<std::string, HMODULE> &get_plugins() {
    return plugins;
}

void load_plugins(LauncherContext *ctx) {
    std::filesystem::path plugin_dir = ctx->plugin_dir;
    if (!std::filesystem::exists(plugin_dir) || !std::filesystem::is_directory(plugin_dir)) {
        return;
    }
    AddDllDirectory((PCWSTR)plugin_dir.wstring().c_str());

    for (const auto &entry : std::filesystem::directory_iterator(plugin_dir)) {
        if (entry.path().extension() != ".dll") {
            continue;
        }

        std::string plugin_name = entry.path().filename().string();
        HMODULE dll = LoadLibraryW(entry.path().c_str());
        if (!dll) {
            std::cerr << std::format("[!] Loading File<{}> failed!", plugin_name) << std::endl;
            continue;
        } else if (!validate_plugin(dll)) {
            std::cerr << std::format("[!] File<{}> is not an validated plugin!", plugin_name) << std::endl;

            continue;
        } else {
            std::cerr << std::format("[!] Loaded plugin {}", plugin_name) << std::endl;
        }

        // store dll
        plugins[plugin_name] = dll;
    }
}

void on_squirrel_vm_init(LauncherContext *ctx) {
    for (const auto [name, entry] : plugins) {
        FARPROC nOffset;
        if (nOffset = GetProcAddress(entry, (LPCSTR) "on_squirrel_vm_init"); nOffset != NULL) {
            std::cout << std::format("[*] call '{}'.on_squirrel_vm_init", name) << std::endl;
            auto call = (void(__cdecl *)(HSQUIRRELVM))nOffset;
            call(ctx->root_vm);
        }

        if (nOffset = GetProcAddress(entry, (LPCSTR) "on_squirrel_vm_init_ext"); nOffset != NULL) {
            std::cout << std::format("[*] call '{}'.on_squirrel_vm_init", name) << std::endl;
            auto call = (void(__cdecl *)(LauncherContext *))nOffset;
            call(ctx);
        }
    }
}
} // namespace plugins
