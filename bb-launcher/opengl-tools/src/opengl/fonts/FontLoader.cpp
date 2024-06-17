#include "FontLoader.h"
#include "../../utils/converterX.h"


// 构造函数: 从 path 读取 .ttf, 初始化 ft 和 face
FontLoader::FontLoader(const char* path, int fontsize) {
    if (FT_Init_FreeType(&ft)) {
        throw std::runtime_error("ERROR::FREETYPE: Could not init FreeType Library");
    }

    if (FT_New_Face(ft, path, 0, &face)) {
        throw std::runtime_error("ERROR::FREETYPE: Failed to load font");
    }
    // Set size to load glyphs as
    FT_Set_Pixel_Sizes(face, 0, fontsize);
    this->fontsize = fontsize;
};


FontLoader::FontLoader(const unsigned char* fnt, int size, int fontsize) {
    if (FT_Init_FreeType(&ft)) {
        throw std::runtime_error("ERROR::FREETYPE: Could not init FreeType Library");
    }

    if (FT_New_Memory_Face(ft, fnt, size, 0, &face)) {
        throw std::runtime_error("ERROR::FREETYPE: Failed to load font");
    }
    // Set size to load glyphs as
    FT_Set_Pixel_Sizes(face, 0, fontsize);
    this->fontsize = fontsize;
};


// 析构函数: 释放 face 和 ft
FontLoader::~FontLoader()
{
    //dtor
    FT_Done_Face(face);
    FT_Done_FreeType(ft);
    for (auto iter = Glyphs.begin(); iter != Glyphs.end(); iter++) {
        FreeTypeGlyph* c = iter->second;
        delete c;
    }
};


GlyphIterator FontLoader::FindGlyphs(std::string text) {
    std::wstring wide = wstringFromUTF8(text);
    return GlyphIterator(wide, this);
};


void FontLoader::LoadGlyphs(std::string text) {
    std::wstring wide = wstringFromUTF8(text);
    // Iterate through all characters
    std::wstring::const_iterator c;
    for (c = wide.begin(); c != wide.end(); c++) {
        LoadGlyph(*c);
    }
}


void FontLoader::LoadGlyph(wchar_t c)
{
    if (Glyphs.find(c) != Glyphs.end()) {
        return;
    }

    //没在map里
    if (FT_Load_Char(face, c, FT_LOAD_RENDER))
    {
        std::cout << "ERROR::FREETYTPE: Failed to load Glyph" << std::endl;
        return;
    }
    Glyphs.insert(std::pair<wchar_t, FreeTypeGlyph*>(c, new FreeTypeGlyph(face)));
    return;
}
