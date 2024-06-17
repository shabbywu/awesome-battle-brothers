#include "GlyphIterator.h"
#include "FontLoader.h"


GlyphIterator::GlyphIterator(std::wstring text, FontLoader* loader) {
    this->text = text;
    this->c = this->text.begin();
    this->loader = loader;
};


// 迭代器操作符：解引用获取字符与字形对
std::pair<wchar_t, FreeTypeGlyph*> GlyphIterator::operator*() {
    loader->LoadGlyph(*c);
    return std::pair<wchar_t, FreeTypeGlyph*>(*c, loader->Glyphs[*c]);
};


// 迭代器操作符：下一个对字符/字形
GlyphIterator& GlyphIterator::next() {
    c++;
    return *this;
};


bool GlyphIterator::isEnd() {
    return c == text.end();
};
