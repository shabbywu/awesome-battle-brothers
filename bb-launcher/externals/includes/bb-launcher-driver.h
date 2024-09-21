#pragma once
#include <filesystem>
#include <memory>
#include <string>

namespace launcher {
struct LauncherMetadata {
    bool Valid();
    std::filesystem::path &GetGamePath();

    // BattleBrothersExe API
    unsigned int GetBattleBrothersExeSize();
    const char *GetBattleBrothersExeContent();
    std::string GetBattleBrothersExeDigest();
};

LauncherMetadata *DetectLauncherMetadata();

bool setup_launcher(LauncherMetadata *metadata, char *self);
void atexit(bool pause = true);

} // namespace launcher
