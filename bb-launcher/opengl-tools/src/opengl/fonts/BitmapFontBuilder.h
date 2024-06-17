#pragma once
#include <iostream>
#include <map>
// GLM
#include <glm/glm.hpp>
#include <glm/gtc/matrix_transform.hpp>
#include <glm/gtc/type_ptr.hpp>
#include "FontLoader.h"
/// Holds all state information relevant to a character as loaded using FreeType
struct BitmapCharacter {
    // Character 在 BitmapFont 中的坐标
    float left;
    float right;
    float bottom;
    float top;
    glm::ivec2   TopXY;      // 端点坐标
    glm::ivec2   Size;       // Size of glyph
    glm::ivec2   Bearing;    // Offset from baseline to left/top of glyph
    unsigned int Advance;    // Offset to advance to next glyph
    bool isInited;

    BitmapCharacter() {
        left = right = bottom = top = 0.0f;
        isInited = false;
    }
};


struct BitmapFont
{
    float width;
    float height;
    unsigned char* buffer;
    std::map<wchar_t, BitmapCharacter> Characters;
    int fontsize;

    virtual ~BitmapFont() {
        delete buffer;
    }
};


class BitmapFontBuilder
{
public:
    BitmapFontBuilder(FontLoader* loader);
    BitmapFont* Build(std::string text);
protected:

private:
    FontLoader* loader;
};
