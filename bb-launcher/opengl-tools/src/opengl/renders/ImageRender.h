#pragma once
// GLM
#include <glm/glm.hpp>
#include <glm/gtc/matrix_transform.hpp>
#include <glm/gtc/type_ptr.hpp>
// GL includes
#include "Shader.h"

class ImageRender
{
private:
    int imageWidth, imageHeight, imageChannels;

    Shader shader;
    GLuint VAO, VBO, EBO;
    GLuint TextureID;   // ID handle of the texture

    // VBO 定点顺序
    constexpr static GLuint VBOIndices[6] = {
        0, 1, 2, // 第一个三角形
        0, 2, 3  // 第二个三角形
    };

    ImageRender(int windowWidth, int windowHeight);
public:
    ImageRender(int windowWidth, int windowHeight, const unsigned char *buffer, int size);
    void initTextureID(const unsigned char *buffer, int size);
    void RenderImage(glm::vec2 leftBottom, glm::vec2 rightTop, float alpha = 1.0f);


    ~ImageRender() {
        glDeleteTextures(1, &TextureID);
    }
};
