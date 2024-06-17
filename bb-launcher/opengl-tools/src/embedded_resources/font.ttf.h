#pragma once
#include <filesystem>

namespace bb { namespace font {
    unsigned int GetFontSize();

    unsigned char* GetFontContent();

    bool loadFont(std::filesystem::path fontPath);
}};
