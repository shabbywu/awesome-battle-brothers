#pragma once
// GLM
#include <glm/glm.hpp>
#include <glm/gtc/matrix_transform.hpp>
#include <glm/gtc/type_ptr.hpp>
// FreeType
#include <ft2build.h>
#include FT_FREETYPE_H
// GL includes
#include "Shader.h"
#include <iostream>
#include <map>
/// Holds all state information relevant to a character as loaded using FreeType
struct Character
{
    GLuint TextureID;   // ID handle of the glyph texture
    glm::ivec2 Size;    // Size of glyph
    glm::ivec2 Bearing;  // Offset from baseline to left/top of glyph
    GLuint Advance;    // Horizontal offset to advance to next glyph
};


class TureTypeRender
{
private:
    Shader shader;
    FT_Library ft;
    FT_Face face;

    GLuint VAO, VBO, EBO;
    std::map<wchar_t, Character> Characters;

    // VBO 定点顺序
    constexpr static GLuint VBOIndices[6] = {
        0, 1, 2, // 第一个三角形
        0, 2, 3  // 第二个三角形
    };

    TureTypeRender(int windowWidth, int windowHeight);
public:
    TureTypeRender(int window_width, int window_height, const unsigned char* fontBuffer, int bufferSize, int fontSize);
    ~TureTypeRender()
    {
        FT_Done_Face(face);
        FT_Done_FreeType(ft);
    }

    void RenderText(std::string text, GLfloat x, GLfloat y, GLfloat scale, glm::vec3 color);
    void fillChar(wchar_t c);
};
