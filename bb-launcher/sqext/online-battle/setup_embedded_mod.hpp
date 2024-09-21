#pragma once
#include <bb-launcher-driver.h>
#include <online_battle/generated_files/mod_modern_hooks.h>
#include <online_battle/generated_files/online_battle.h>
#include <physfs.h>

namespace online_battle {
bool mount_mod_to_physfs(launcher::LauncherMetadata *metadata) {
    // 需要覆盖前面的, 所以 appendToPath = false (即添加到搜索路径最前面)
#ifdef BB_DEBUG_MESSAGE
    std::cout << "[TRACE] before PHYSFS_mountMemory" << std::endl;
#endif
    {
        auto &file = bin2cpp::getOnline_battleZipFile();
        if (!PHYSFS_mountMemory(file.getBuffer(), file.getSize(), NULL, "OnlineBattleMod.zip", "/", 0)) {
            throw std::runtime_error(
                std::format("failed to mount OnlineBattleMod.zip, Reason: [{}]", PHYSFS_getLastError()));
        }
    }
    {
        auto &file = bin2cpp::getMod_modern_hooksZipFile();
        if (!PHYSFS_mountMemory(file.getBuffer(), file.getSize(), NULL, "mod_modern_hooks.zip", "/", 0)) {
            throw std::runtime_error(
                std::format("failed to mount mod_modern_hooks.zip, Reason: [{}]", PHYSFS_getLastError()));
        }
    }
#ifdef BB_DEBUG_MESSAGE
    std::cout << "[TRACE] after PHYSFS_mountMemory" << std::endl;
#endif
    return true;
}
} // namespace online_battle
