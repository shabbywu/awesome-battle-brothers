#pragma once
// FreeType
#include <ft2build.h>
#include FT_FREETYPE_H
#include <iostream>
#include <map>
#include "./GlyphIterator.h"


/// Holds all state information relevant to a character as loaded using FreeType
class FreeTypeGlyph {
public:
    FreeTypeGlyph(FT_Face face) {
        auto glyph = face->glyph;
        FT_Bitmap& g_bitmap = glyph->bitmap;

        bitmap = g_bitmap;
        // 复制 buffer 区域
        bitmap.buffer = new unsigned char[g_bitmap.rows * g_bitmap.width];
        memcpy(bitmap.buffer, g_bitmap.buffer, g_bitmap.rows * g_bitmap.width);

        bitmap_left = glyph->bitmap_left;
        bitmap_top = glyph->bitmap_top;

        // Now advance cursors for next glyph (note that advance is number of 1/64 pixels)
        // Bitshift by 6 to get value in pixels (2^6 = 64 (divide amount of 1/64th pixels by 64 to get amount of pixels))
        advance_x = glyph->advance.x >> 6;
        advance_y = glyph->advance.y >> 6;
    }

    virtual ~FreeTypeGlyph() {
        delete bitmap.buffer;
    }

    FT_Bitmap bitmap;
    // Offset from baseline to left/top of glyph
    FT_Int    bitmap_left;
    FT_Int    bitmap_top;
    // Horizontal(x)/Vertical(y) offset to advance to next glyph
    FT_UInt advance_x;
    FT_UInt advance_y;
};


class FontLoader {
public:
    FontLoader(const char* path, int fontsize);
    FontLoader(const unsigned char* fnt, int size, int fontsize);
    virtual ~FontLoader();
    void LoadGlyphs(std::string text);
    void LoadGlyph(wchar_t c);
    GlyphIterator FindGlyphs(std::string text);

    std::map<wchar_t, FreeTypeGlyph*> Glyphs;
    int fontsize;

private:
    FT_Library ft;
    FT_Face face;
};
