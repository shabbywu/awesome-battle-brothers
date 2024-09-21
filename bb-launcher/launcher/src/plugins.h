#pragma once
#include <filesystem>
#include <peconv.h>

#include <map>
#include <squirrel.h>

class LauncherContext {
  public:
    std::string digest;
    BYTE *loaded_pe;
    std::filesystem::path gamepath;
    std::filesystem::path plugin_dir;

    HSQUIRRELVM root_vm;
    std::map<std::string, ULONGLONG> physfs_library;

    LauncherContext(std::string digest, BYTE *loaded_pe, std::filesystem::path gamepath,
                    std::filesystem::path plugin_dir, HSQUIRRELVM root_vm,
                    std::map<std::string, ULONGLONG> physfs_library)
        : digest(digest), loaded_pe(loaded_pe), gamepath(gamepath), plugin_dir(plugin_dir), root_vm(root_vm),
          physfs_library(physfs_library) {
    }
};

namespace plugins {
void load_plugins(LauncherContext *ctx);
std::map<std::string, HMODULE> &get_plugins();

void on_squirrel_vm_init(LauncherContext *ctx);
} // namespace plugins
