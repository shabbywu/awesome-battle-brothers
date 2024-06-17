#include "BitmapRender.h"
#include "../../utils/converterX.h"


static std::string vertexCode = R"(
#version 330 core
layout (location = 0) in vec2 pos;
layout (location = 1) in vec2 tex;
out vec2 TexCoord;

uniform mat4 projection;

void main()
{
    gl_Position = projection * vec4(pos, 0.0, 1.0);
    TexCoord = tex;
}
)";

static std::string fragmentCode = R"(
#version 330 core
in vec2 TexCoord;
out vec4 color;

uniform sampler2D text;
uniform vec3 textColor;

void main()
{
    vec4 sampled = vec4(1.0, 1.0, 1.0, texture(text, TexCoord).r);
    color = vec4(textColor, 1.0) * sampled;
}

)";


BitmapRender::BitmapRender(int windowWidth, int windowHeight, BitmapFont* fnt) :shader(vertexCode, fragmentCode)
{
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
    // opengl 中 分配 4 * 4 个 float 的内存(通过 EBO 实现定点复用)
    glBufferData(GL_ARRAY_BUFFER, sizeof(GLfloat) * 4 * 4, NULL, GL_DYNAMIC_DRAW);
    glGenBuffers(1, &EBO);
    glBindBuffer(GL_ELEMENT_ARRAY_BUFFER, EBO);
    glBufferData(GL_ELEMENT_ARRAY_BUFFER, sizeof(VBOIndices), VBOIndices, GL_STATIC_DRAW);
    // 设置 location=0 pointer (position)
    glVertexAttribPointer(0, 2, GL_FLOAT, GL_FALSE, 4 * sizeof(GLfloat), (void*)0);
    glEnableVertexAttribArray(0);
    // 设置 location=1 pointer (texture)
    glVertexAttribPointer(1, 2, GL_FLOAT, GL_FALSE, 4 * sizeof(GLfloat), (void*)(2 * sizeof(GL_FLOAT)));
    glEnableVertexAttribArray(1);

    glm::mat4 projection = glm::ortho(0.0f, static_cast<GLfloat>(windowWidth), 0.0f, static_cast<GLfloat>(windowHeight));
    shader.use();
    shader.setMat4("projection", projection);
    // Disable byte-alignment restriction
    glPixelStorei(GL_UNPACK_ALIGNMENT, 1);
    this->fnt = fnt;
    initTextureID();
}


void BitmapRender::initTextureID() {
    glGenTextures(1, &TextureID);

    glBindTexture(GL_TEXTURE_2D, TextureID);
    glTexImage2D(
        GL_TEXTURE_2D,
        0,
        GL_RED,
        fnt->width,
        fnt->height,
        0,
        GL_RED,
        GL_UNSIGNED_BYTE,
        fnt->buffer
    );
    // Set texture options
    glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, GL_CLAMP_TO_EDGE);
    glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, GL_CLAMP_TO_EDGE);
    glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, GL_LINEAR);
    glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_LINEAR);
    glBindTexture(GL_TEXTURE_2D, 0);
}


void BitmapRender::RenderText(std::string text, GLfloat x, GLfloat y, GLfloat scale, glm::vec3 color)
{
    // Activate corresponding render state
    shader.use();
    shader.setVec3("textColor", color);
    // 激活纹理序号 0
    glActiveTexture(GL_TEXTURE0);
    // 绑定 VAO 和 EBO
    glBindVertexArray(VAO);
    glBindBuffer(GL_ELEMENT_ARRAY_BUFFER, EBO);

    std::wstring wide = wstringFromUTF8(text);
    // Iterate through all characters
    std::wstring::const_iterator c;

    for (c = wide.begin(); c != wide.end(); c++)
    {
        if (fnt->Characters.find(*c) == fnt->Characters.end()) {
            // 不支持渲染该字符
            continue;
        }
        BitmapCharacter ch = fnt->Characters[*c];

        GLfloat xpos = x + ch.Bearing.x * scale;
        GLfloat ypos = y - (ch.Size.y - ch.Bearing.y) * scale;

        float w = ch.Advance * scale;
        float h = ch.Size.y * scale;

        GLfloat spriteLeft = (ch.TopXY.x) / fnt->width;
        GLfloat spriteRight = (ch.TopXY.x + ch.Advance) / fnt->width;
        GLfloat spriteBottom = (ch.TopXY.y) / fnt->height;
        GLfloat spriteTop = (ch.TopXY.y + ch.Size.y) / fnt->height;

        // Update VBO for each character
        GLfloat vertices[4][4] =
        {
            // 顶点坐标 (x, y)       纹理坐标 (s, t)
            { xpos,     ypos + h,   spriteLeft, spriteBottom },
            { xpos,     ypos,       spriteLeft, spriteTop },
            { xpos + w, ypos,       spriteRight, spriteTop },

            // { xpos,     ypos + h,   spriteLeft, spriteBottom },
            // { xpos + w, ypos,       spriteRight, spriteTop },
            { xpos + w, ypos + h,   spriteRight, spriteBottom }
        };
        // Render glyph texture over quad
        glBindTexture(GL_TEXTURE_2D, TextureID);
        // 更新 VBO 缓冲区的内容
        glBindBuffer(GL_ARRAY_BUFFER, VBO);
        glBufferSubData(GL_ARRAY_BUFFER, 0, sizeof(vertices), vertices);
        // 解绑 VBO
        glBindBuffer(GL_ARRAY_BUFFER, 0);
        // Render quad
        glDrawElements(GL_TRIANGLES, 6, GL_UNSIGNED_INT, 0);
        x += w;
    }
    // 解绑 VAO/EBO 和纹理
    glBindVertexArray(0);
    glBindBuffer(GL_ELEMENT_ARRAY_BUFFER, 0);
    glBindTexture(GL_TEXTURE_2D, 0);
}
