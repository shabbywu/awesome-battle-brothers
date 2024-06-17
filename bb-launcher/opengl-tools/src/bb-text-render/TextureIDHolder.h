#include <glm/glm.hpp>
#include <glm/gtc/matrix_transform.hpp>
#include <glm/gtc/type_ptr.hpp>
#include <glad/glad.h>
#include "../opengl/fonts/BitmapFontBuilder.h"


struct TextureIDHolder {
    BitmapFont* fnt;
    GLuint TextureID;   // ID handle of the bitmap texture

    void initTextureID() {
        glGenTextures(1, &TextureID);
        glBindTexture(GL_TEXTURE_2D, TextureID);

        int width = (int)fnt->width;
        int height = (int)fnt->height;
        GLubyte* textureData = new GLubyte[width * height * 4];
        // 创建 FreeImage 位图
        for (int y = 0; y < height; ++y) {
            for (int x = 0; x < width; ++x) {
                textureData[x * 4 + 0 + y * width * 4] = 255;                        // Alpha
                textureData[x * 4 + 1 + y * width * 4] = fnt->buffer[y * width + x]; // Red
                textureData[x * 4 + 2 + y * width * 4] = fnt->buffer[y * width + x]; // Green
                textureData[x * 4 + 3 + y * width * 4] = fnt->buffer[y * width + x]; // Blue
            }
        }

        glTexImage2D(
            GL_TEXTURE_2D,
            0,
            GL_RGBA,
            fnt->width,
            fnt->height,
            0,
            GL_BGRA,
            GL_UNSIGNED_BYTE,
            textureData
        );
        // Set texture options
        glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, GL_CLAMP_TO_EDGE);
        glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, GL_CLAMP_TO_EDGE);
        glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, GL_LINEAR);
        glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_LINEAR);
        glBindTexture(GL_TEXTURE_2D, 0);

        // 释放FreeImage资源
        delete textureData;
    }
};
