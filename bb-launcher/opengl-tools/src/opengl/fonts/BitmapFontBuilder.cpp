
#include "BitmapFontBuilder.h"
#include "string"

BitmapFontBuilder::BitmapFontBuilder(FontLoader* loader) {
    this->loader = loader;
};


BitmapFont* BitmapFontBuilder::Build(std::string text) {
    BitmapFont* fnt = new BitmapFont();
    int width = 0;
    int height = loader->fontsize * 2;
    for (auto iter = loader->FindGlyphs(text); !iter.isEnd(); iter.next()) {
        auto c = (*iter).first;
        auto ttFnt = (*iter).second;
        if (fnt->Characters.find(c) != fnt->Characters.end()) {
            continue;
        }
        fnt->Characters.insert(std::pair<wchar_t, BitmapCharacter>(c, BitmapCharacter()));
        width += ttFnt->advance_x;
    }
    // 创建位图
    fnt->buffer = new unsigned char[width * height];
    // 初始化位图
    std::memset(fnt->buffer, 0, width * height);

    // 渲染位图
    int x = 0;
    // 因为 TrueType 的 bitmap 数据的行顺序是从下到上, 因此需要将 y(baseline) 设置成字体大小
    int y = 0;
    for (auto iter = loader->FindGlyphs(text); !iter.isEnd(); iter.next()) {
        auto c = (*iter).first;
        auto ttFnt = (*iter).second;
        BitmapCharacter& ch = fnt->Characters[c];

        if (ch.isInited) {
            continue;
        }
        for (int row = 0; row < ttFnt->bitmap.rows; ++row) {
            for (int col = 0; col < ttFnt->bitmap.width; ++col) {
                int index = (x + col + ttFnt->bitmap_left) + (row)*width;
                fnt->buffer[index] = ttFnt->bitmap.buffer[col + row * ttFnt->bitmap.pitch];
            }
        }
        ch.TopXY = glm::ivec2(x, 0);
        ch.Size = glm::ivec2(ttFnt->bitmap.pitch, ttFnt->bitmap.rows);
        ch.Bearing = glm::ivec2(ttFnt->bitmap_left, ttFnt->bitmap_top);
        ch.Advance = ttFnt->advance_x;
        ch.isInited = true;
        x += ttFnt->advance_x;
    }
    fnt->width = width;
    fnt->height = height;
    fnt->fontsize = loader->fontsize;
    return fnt;
};
