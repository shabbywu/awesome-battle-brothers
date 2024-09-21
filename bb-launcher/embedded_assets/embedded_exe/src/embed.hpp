#pragma once
#ifdef GOG_EXE
#include "BattleBrothers.GOG.hpp"
#else
#include "BattleBrothers.STEAM.hpp"
#endif
#include <cstring>
#include <string>

unsigned int GetBattleBrothersExeSize();
const char* GetBattleBrothersExeContent();
std::string GetBattleBrothersExeDigest();

namespace BattleBrothers {
    unsigned int GetExeSize() {
        return bin2cpp::getBattleBrothersExeFile().getSize();
    };

    const char* GetExeContent() {
        return bin2cpp::getBattleBrothersExeFile().getBuffer();
    };

    std::string GetExeDigest() {
        #ifdef GOG_EXE
        return "e2ea659ad0afb221c964db918853af687762452c41d7fa9d4b90ecf496605b8e";
        #else
        return "9a3328d7224c0856689234a605321be486fde8b31190e400f8e247cf89b031b4";
        #endif
    }
}
