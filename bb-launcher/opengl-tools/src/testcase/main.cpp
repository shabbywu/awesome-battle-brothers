#include <iostream>
#include <map>
#include <string>
#include <iomanip>
#include <sstream>

#include <glad/glad.h>
#include <GLFW/glfw3.h>

#include <glm/glm.hpp>
#include <glm/gtc/matrix_transform.hpp>
#include <glm/gtc/type_ptr.hpp>
#include <codecvt>

#include <ft2build.h>
#include FT_FREETYPE_H
#define STB_IMAGE_IMPLEMENTATION
#include <stb_image.h>
#define STBI_MSC_SECURE_CRT
#define STB_IMAGE_WRITE_IMPLEMENTATION
#include <stb_image_write.h>

#include "../embedded_resources/retirement_01.h"
#include "../embedded_resources/gh-avatar.h"
#include "../embedded_resources/font.ttf.h"
#include "../opengl/fonts/FontLoader.h"
#include "../opengl/fonts/BitmapFontBuilder.h"
#include "../opengl/fonts/BitmapFontCache.h"
#include "../opengl/renders/TureTypeRender.h"
#include "../opengl/renders/BitmapRender.h"
#include "../opengl/renders/ImageRender.h"

void framebuffer_size_callback(GLFWwindow* window, int width, int height);
void processInput(GLFWwindow* window);
// settings
const unsigned int SCR_WIDTH = 800;
const unsigned int SCR_HEIGHT = 450;


void saveBitmapAsPNG(int width, int height, unsigned char* buffer, const std::string& filename) {
    auto CHANNEL_NUM = 4;
    uint8_t* pixels = new uint8_t[width * height * CHANNEL_NUM];
    for (int y = 0; y < height; ++y)  {
        for (int x = 0; x < width; ++x) {
            auto from_pixel = (height - y) * width + x;
            auto to_pixel = y * width + x;

            pixels[CHANNEL_NUM * to_pixel + 0] = buffer[from_pixel]; // Red
            pixels[CHANNEL_NUM * to_pixel + 1] = buffer[from_pixel]; // Green
            pixels[CHANNEL_NUM * to_pixel + 2] = buffer[from_pixel]; // Blue
            pixels[CHANNEL_NUM * to_pixel + 3] = buffer[from_pixel] > 0 ? 255 : 0; // Blue
        }
    }
    stbi_write_png(filename.c_str(), width, height, CHANNEL_NUM, pixels, width * CHANNEL_NUM);
    delete[] pixels;
}


int main()
{
    // glfw: initialize and configure
    // ------------------------------
    glfwInit();
    glfwWindowHint(GLFW_CONTEXT_VERSION_MAJOR, 3);
    glfwWindowHint(GLFW_CONTEXT_VERSION_MINOR, 3);
    glfwWindowHint(GLFW_OPENGL_PROFILE, GLFW_OPENGL_CORE_PROFILE);
    glfwWindowHint(GLFW_DECORATED, GL_FALSE);

#ifdef __APPLE__
    glfwWindowHint(GLFW_OPENGL_FORWARD_COMPAT, GL_TRUE);
#endif

    // glfw window creation
    // --------------------
    GLFWwindow* window = glfwCreateWindow(SCR_WIDTH, SCR_HEIGHT, "LearnOpenGL", NULL, NULL);
    if (window == NULL)
    {
        std::cout << "Failed to create GLFW window" << std::endl;
        glfwTerminate();
        return -1;
    }
    glfwMakeContextCurrent(window);

    GLFWmonitor* monitor = glfwGetPrimaryMonitor();
    const GLFWvidmode* mode = glfwGetVideoMode(monitor);
    glfwSetWindowPos(window, (mode->width - SCR_WIDTH) / 2, (mode->height - SCR_HEIGHT) / 2);
    glfwSetFramebufferSizeCallback(window, framebuffer_size_callback);

    // glad: load all OpenGL function pointers
    // ---------------------------------------
    if (!gladLoadGLLoader((GLADloadproc)glfwGetProcAddress))
    {
        std::cout << "Failed to initialize GLAD" << std::endl;
        return -1;
    }

    // OpenGL state
    // ------------
    glEnable(GL_CULL_FACE);
    glEnable(GL_BLEND);
    glBlendFunc(GL_SRC_ALPHA, GL_ONE_MINUS_SRC_ALPHA);

    // 测试 ttf 生成 bitmap font
    FontLoader loader(bb::font::GetFontContent(), bb::font::GetFontSize(), 24);
    BitmapFontBuilder builder(&loader);
    BitmapFontCache* cache = new BitmapFontCache(builder);
    auto fnt = cache->GetOrCreateBitmapFont("This(这) is sample text (C) LearnOpenGL.com");
    cache->GetOrCreateBitmapFont("This(这)");
    std::cout << std::to_string(cache->Size()) << std::endl;
    // 保存位图数据到文件（可选）
    saveBitmapAsPNG(fnt->width, fnt->height, fnt->buffer, "merged_bitmap.png");

    TureTypeRender textRender(SCR_WIDTH, SCR_HEIGHT, bb::font::GetFontContent(), bb::font::GetFontSize(), 20);
    // BitmapRender textRender(SCR_WIDTH, SCR_HEIGHT, fnt);

    const auto& backgroundFile = bin2cpp::getRetirement_01JpegFile();
    ImageRender backgroundRender (SCR_WIDTH, SCR_HEIGHT, (const unsigned char *)backgroundFile.getBuffer(), backgroundFile.getSize());

    const auto& avatarFile = bin2cpp::getGhavatarPngFile();
    ImageRender avatarRender (SCR_WIDTH, SCR_HEIGHT, (const unsigned char *)avatarFile.getBuffer(), avatarFile.getSize());

    auto brandColor = glm::vec3(0.6875f, 0.621f, 0.539f);
    auto waitTime = 3;
    auto end = glfwGetTime() + waitTime;
    // render loop
    // -----------
    while (!glfwWindowShouldClose(window))
    {
        // input
        // -----
        processInput(window);

        // render
        // ------
        glClearColor(0.2f, 0.3f, 0.3f, 1.0f);
        glClear(GL_COLOR_BUFFER_BIT);

        backgroundRender.RenderImage(glm::vec2(0,0), glm::vec2(SCR_WIDTH, SCR_HEIGHT));

        textRender.RenderText("Awesome", 630.f, SCR_HEIGHT - 50.f, 1.0f, brandColor);
        textRender.RenderText("Battle Brothers", 600.f, SCR_HEIGHT - 72.f, 1.0f, brandColor);

        auto timer = end - glfwGetTime();
        auto percentage = (waitTime - timer) / waitTime * 100;
        std::stringstream stream;
        stream << std::fixed << std::setprecision(2) << percentage;
        textRender.RenderText("loading: " + stream.str() + "%", 620.0f, SCR_HEIGHT - 144.f, 0.8f, brandColor);

        avatarRender.RenderImage(glm::vec2(600.f, 92.f), glm::vec2(760.f, 252.f), timer / waitTime);
        textRender.RenderText("shabbywu", 625.f, 72.f, 1.0f, glm::vec3(0.5, 0.8f, 0.2f) * glm::vec3(timer / waitTime, timer / waitTime, timer / waitTime));

        // glfw: swap buffers and poll IO events (keys pressed/released, mouse moved etc.)
        // -------------------------------------------------------------------------------
        glfwSwapBuffers(window);
        glfwPollEvents();

        if (timer < 0) {
            glfwSetWindowShouldClose(window, true);
        }
    }

    glfwTerminate();
    delete fnt;
    return 0;
}

// process all input: query GLFW whether relevant keys are pressed/released this frame and react accordingly
// ---------------------------------------------------------------------------------------------------------
void processInput(GLFWwindow* window)
{
    if (glfwGetKey(window, GLFW_KEY_ESCAPE) == GLFW_PRESS)
        glfwSetWindowShouldClose(window, true);
}

// glfw: whenever the window size changed (by OS or user resize) this callback function executes
// ---------------------------------------------------------------------------------------------
void framebuffer_size_callback(GLFWwindow* window, int width, int height)
{
    // make sure the viewport matches the new window dimensions; note that width and
    // height will be significantly larger than specified on retina displays.
    glViewport(0, 0, width, height);
}
