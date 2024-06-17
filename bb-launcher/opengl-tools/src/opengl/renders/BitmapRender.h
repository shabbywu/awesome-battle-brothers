#pragma once
// GLM
#include <glm/glm.hpp>
#include <glm/gtc/matrix_transform.hpp>
#include <glm/gtc/type_ptr.hpp>
// GL includes
#include "Shader.h"
#include "../fonts/BitmapFontBuilder.h"

class BitmapRender
{
private:
    Shader shader;
    BitmapFont* fnt;

    GLuint VAO, VBO, EBO;
    GLuint TextureID;   // ID handle of the bitmap texture

    // VBO 定点顺序
    constexpr static GLuint VBOIndices[6] = {
        0, 1, 2, // 第一个三角形
        0, 2, 3  // 第二个三角形
    };

public:
    BitmapRender(int window_width, int window_height, BitmapFont* fnt);
    ~BitmapRender() {
        glDeleteTextures(1, &TextureID);
    }
    void RenderText(std::string text, GLfloat x, GLfloat y, GLfloat scale, glm::vec3 color);
    void initTextureID();

};
