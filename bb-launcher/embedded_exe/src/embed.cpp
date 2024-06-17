#include "embed.h"
#include "BattleBrothers.EXE.h"

unsigned int GetBattleBrothersExeSize() {
    return bin2cpp::getBattleBrothersExeFile().getSize();
};

const char* GetBattleBrothersExeContent() {
    return bin2cpp::getBattleBrothersExeFile().getBuffer();
};

#ifdef GOG_EXE
std::string GetBattleBrothersExeDigest() {
    return "e2ea659ad0afb221c964db918853af687762452c41d7fa9d4b90ecf496605b8e";
}
#else
std::string GetBattleBrothersExeDigest() {
    return "9a3328d7224c0856689234a605321be486fde8b31190e400f8e247cf89b031b4";
}
#endif
