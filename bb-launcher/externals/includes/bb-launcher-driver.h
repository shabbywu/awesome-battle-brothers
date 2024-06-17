#pragma once
#include <memory>

struct LauncherMetadata {
    bool valid();
    std::filesystem::path& GetGamePath();

    // BattleBrothersExe API
    unsigned int GetBattleBrothersExeSize();
    const char* GetBattleBrothersExeContent();
    std::string GetBattleBrothersExeDigest();
};

LauncherMetadata* DetectLauncherMetadata();

bool setup_launcher(LauncherMetadata* metadata, char* self);
