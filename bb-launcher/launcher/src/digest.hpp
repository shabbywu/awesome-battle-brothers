#pragma once
#include <string>
#include <iostream>

void validate_program_hash(std::string& digest, bool& isSteam) {
    if (digest == "f45547d8afaebb02cdda0d7db7a0cda225ec04c30e7237bec12675b173d3eb61") {
        // GOG 1.5.0.14
        isSteam = false;
    }
    else if (digest == "6a3495d14634fe004c0fef626738636f791a5e24570d1c3b9467703c3f5bae4f") {
        // GOG 1.5.0.14 with 4gb patch
        isSteam = false;
    }
    else if (digest == "9be89dbe2d4f893a8e2a82ff9b96b9f6a40377a3c3fad3ae9021746ead8bbc53") {
        // GOG 1.5.0.15
        isSteam = false;
    }
    else if (digest == "e2ea659ad0afb221c964db918853af687762452c41d7fa9d4b90ecf496605b8e") {
        // GOG 1.5.0.15 with 4gb patch
        isSteam = false;
    }
    else if (digest == "0def5c8bd93bdb39af2b28195f06f34a5bdad7381d3e5401522d668780612105") {
        // Steam 1.5.0.14
        isSteam = true;
    }
    else if (digest == "e9900d7fe38d9dc0a8438ac9ca79aba24403136c6c415028c223f13d97fb844d" || digest == "9a3328d7224c0856689234a605321be486fde8b31190e400f8e247cf89b031b4") {
        // Steam 1.5.0.15
        isSteam = true;
    }
    else {
        std::cerr << "unknown digest: " << digest << std::endl;
        std::cerr << "Unsupport version, only version 1.5.0.14/1.5.0.15(GOG) or version 1.5.0.14/1.5.0.15(STEAM) are supported." << std::endl;
        exit(1);
    }
}
