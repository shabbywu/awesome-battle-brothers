#pragma once

#include <filesystem>
#include <iostream>
#include <physfs.h>
#include <set>
static std::filesystem::path GamePath;
static std::filesystem::path GameDataPath;

namespace patchs {
namespace physfs {
namespace impl {

void setGamePath(std::filesystem::path gamepath) {
    GamePath = gamepath;
    GameDataPath = gamepath / "data";
}

static std::set<std::string> vanillaModList = {"data_001.dat", "data_003.dat", "data_004.dat", "data_006.dat",
                                               "data_008.dat", "data_010.dat", "data_160.dat"};

static int physfs_mount_impl_for_online_mod(const char *newDir, const char *mountPoint, int appendToPath) {
    std::filesystem::path absoluteNewDir = std::filesystem::absolute(newDir);
    auto relativePath = absoluteNewDir.lexically_relative(GameDataPath);

    if (!relativePath.empty() && std::filesystem::is_regular_file(absoluteNewDir)) {
        if (!vanillaModList.contains(relativePath.string())) {
            std::cerr << "online battle mod doesn't support any mod now! mod<" << absoluteNewDir << "> will be ignore!"
                      << std::endl;
            return true;
        }
    }
    auto result = PHYSFS_mount(newDir, mountPoint, appendToPath);
    return result;
}

static int physfs_mount_impl_for_default(const char *newDir, const char *mountPoint, int appendToPath) {
    auto result = PHYSFS_mount(newDir, mountPoint, appendToPath);
    return result;
}
static int physfs_mount_impl_for_mod_mounting_feature(const char *newDir, const char *mountPoint, int appendToPath) {
    std::filesystem::path absoluteNewDir = std::filesystem::absolute(newDir);
    if (absoluteNewDir.string().starts_with(GamePath.string())) {
        return true;
    }

#ifdef BB_DEBUG_MESSAGE
    if (mountPoint != nullptr) {
        std::cout << "PHYSFS_mount: '" << absoluteNewDir << "' to '" << mountPoint << "'" << std::endl;
    } else {
        std::cout << "PHYSFS_mount: '" << absoluteNewDir << "' to '/'" << std::endl;
    }
#endif

    auto result = PHYSFS_mount(newDir, mountPoint, appendToPath);
#ifdef BB_DEBUG_MESSAGE
    auto error_message = PHYSFS_getLastError();
    if (error_message != NULL) {
        std::cerr << "PHYSFS_mount is failure. Reason: " << error_message << std::endl;
    }
#endif
    return result;
}

} // namespace impl

namespace stub {

// patch PHYSFS_mount to ignore mount mods.
int _PHYSFS_mount(const char *newDir, const char *mountPoint, int appendToPath) {
#ifdef BUILD_WITH_ONLINE_MOD
    return impl::physfs_mount_impl_for_online_mod(newDir, mountPoint, appendToPath);
#endif

#ifdef DISABLE_MOD_MOUNTING_FEATURE
    return impl::physfs_mount_impl_for_default(newDir, mountPoint, appendToPath);
#else
    return impl::physfs_mount_impl_for_mod_mounting_feature(newDir, mountPoint, appendToPath);
#endif
}

int _PHYSFS_init(const char *argv0) {
    auto result = PHYSFS_init(argv0);
#ifdef BB_DEBUG_MESSAGE
    auto error_message = PHYSFS_getLastError();
    if (error_message != NULL) {
        std::cerr << "PHYSFS_init is failure. Reason: " << error_message << std::endl;
    }
#endif
    return result;
}

int _PHYSFS_deinit() {
    auto result = PHYSFS_deinit();
#ifdef BB_DEBUG_MESSAGE
    auto error_message = PHYSFS_getLastError();
    if (error_message != NULL) {
        std::cerr << "PHYSFS_deinit is failure. Reason: " << error_message << std::endl;
    }
#endif
    return result;
}

const char *_PHYSFS_getMountPoint(const char *dir) {
    auto result = PHYSFS_getMountPoint(dir);
#ifdef BB_DEBUG_MESSAGE
    auto error_message = PHYSFS_getLastError();
    if (error_message != NULL) {
        std::cerr << "PHYSFS_getMountPoint is failure. Reason: " << error_message << std::endl;
    }
#endif
    return result;
}

} // namespace stub
} // namespace physfs
} // namespace patchs
