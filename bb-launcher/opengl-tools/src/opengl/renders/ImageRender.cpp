#define STB_IMAGE_IMPLEMENTATION
#define STB_IMAGE_STATIC
#define STBI_ONLY_JPEG
#define STBI_ONLY_PNG
#define STBI_NO_LINEAR
#define STBI_NO_HDR
#define STBI_NO_STDIO
#include <stb_image.h>
#include "ImageRender.h"


static std::string vertexCode = R"(
#version 330 core
layout (location = 0) in vec2 pos;
layout (location = 1) in vec2 tex;
layout (location = 2) in float alpha;
out vec2 TexCoord;
out float Alpha;

uniform mat4 projection;

void main()
{
    gl_Position = projection * vec4(pos, 0.0, 1.0);
    TexCoord = tex;
    Alpha = alpha;
}
)";

static std::string fragmentCode = R"(
#version 330 core
in vec2 TexCoord;
in float Alpha;
out vec4 color;

uniform sampler2D image;

void main()
{
    color = texture(image, TexCoord) * vec4(1.0, 1.0, 1.0, Alpha);
}
)";


ImageRender::ImageRender(int windowWidth, int windowHeight) :shader(vertexCode, fragmentCode) {
    //ctor
    // Set OpenGL options
    glEnable(GL_CULL_FACE);
    glEnable(GL_BLEND);
    glBlendFunc(GL_SRC_ALPHA, GL_ONE_MINUS_SRC_ALPHA);

    // Configure VAO/VBO/EBO for texture quads
    glGenVertexArrays(1, &VAO);
    glBindVertexArray(VAO);
    glGenBuffers(1, &VBO);
    glBindBuffer(GL_ARRAY_BUFFER, VBO);
    // opengl 中 分配 4 * 5 个 float 的内存(通过 EBO 实现定点复用)
    glBufferData(GL_ARRAY_BUFFER, sizeof(GLfloat) * 4 * 5, NULL, GL_DYNAMIC_DRAW);
    glGenBuffers(1, &EBO);
    glBindBuffer(GL_ELEMENT_ARRAY_BUFFER, EBO);
    glBufferData(GL_ELEMENT_ARRAY_BUFFER, sizeof(VBOIndices), VBOIndices, GL_STATIC_DRAW);
    // 设置 location=0 pointer (position)
    glVertexAttribPointer(0, 2, GL_FLOAT, GL_FALSE, 5 * sizeof(GLfloat), (void*)0);
    glEnableVertexAttribArray(0);
    // 设置 location=1 pointer (texture)
    glVertexAttribPointer(1, 2, GL_FLOAT, GL_FALSE, 5 * sizeof(GLfloat), (void*)(2 * sizeof(GL_FLOAT)));
    glEnableVertexAttribArray(1);
    // 设置 location=2 pointer (alpha)
    glVertexAttribPointer(2, 1,  GL_FLOAT, GL_FALSE, 5 * sizeof(GLfloat), (void*)(4 * sizeof(GL_FLOAT)));
    glEnableVertexAttribArray(2);

    glm::mat4 projection = glm::ortho(0.0f, static_cast<GLfloat>(windowWidth), 0.0f, static_cast<GLfloat>(windowHeight));
    shader.use();
    shader.setMat4("projection", projection);
    // Disable byte-alignment restriction
    glPixelStorei(GL_UNPACK_ALIGNMENT, 1);
}


ImageRender::ImageRender(int windowWidth, int windowHeight, const unsigned char *buffer, int size) : ImageRender(windowWidth, windowHeight) {
    initTextureID(buffer, size);
}


void ImageRender::initTextureID(const unsigned char *buffer, int size) {
    glGenTextures(1, &TextureID);
    glBindTexture(GL_TEXTURE_2D, TextureID);
    // 为当前绑定的纹理对象设置环绕、过滤方式
    glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, GL_REPEAT);
    glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, GL_REPEAT);
    glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, GL_LINEAR);
    glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_LINEAR);

    // 加载并生成纹理
    int width, height, nrChannels;
    unsigned char *data = stbi_load_from_memory(buffer, size, &width, &height, &nrChannels, 0);

    if (data)
    {
        auto texChannel = GL_RGB;
        if (nrChannels == 4) {
            texChannel = GL_RGBA;
        }
        glTexImage2D(GL_TEXTURE_2D, 0, texChannel, width, height, 0, texChannel, GL_UNSIGNED_BYTE, data);
        glGenerateMipmap(GL_TEXTURE_2D);
    }
    else
    {
        throw std::runtime_error("Failed to load texture");
    }
    stbi_image_free(data);
}


void ImageRender::RenderImage(glm::vec2 leftBottom, glm::vec2 rightTop, float alpha) {
    // Activate corresponding render state
    shader.use();
    // 激活纹理序号 0
    glActiveTexture(GL_TEXTURE0);
    // 绑定 VAO 和 EBO
    glBindVertexArray(VAO);
    glBindBuffer(GL_ELEMENT_ARRAY_BUFFER, EBO);

    // Update VBO for each character
    GLfloat vertices[4][5] =
    {
        // 顶点坐标 (x, y)       纹理坐标 (s, t)
        { leftBottom.x,     rightTop.y,   0.0, 0.0, alpha },
        { leftBottom.x,     leftBottom.y,       0.0, 1.0, alpha },
        { rightTop.x, leftBottom.y,       1.0, 1.0, alpha },
        { rightTop.x, rightTop.y,   1.0, 0.0, alpha }
    };

    // Render glyph texture over quad
    glBindTexture(GL_TEXTURE_2D, TextureID);
    // 更新 VBO 缓冲区的内容
    glBindBuffer(GL_ARRAY_BUFFER, VBO);
    glBufferSubData(GL_ARRAY_BUFFER, 0, sizeof(vertices), vertices);

    // Render quad
    glDrawElements(GL_TRIANGLES, 6, GL_UNSIGNED_INT, 0);

    // 解绑 VAO/VBO/EBO 和纹理
    glBindVertexArray(0);
    glBindBuffer(GL_ARRAY_BUFFER, 0);
    glBindBuffer(GL_ELEMENT_ARRAY_BUFFER, 0);
    glBindTexture(GL_TEXTURE_2D, 0);
}
