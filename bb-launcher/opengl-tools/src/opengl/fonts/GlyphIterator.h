#pragma once

#include <iostream>
#include <map>

class FontLoader;
class FreeTypeGlyph;


class GlyphIterator {
public:
    GlyphIterator(std::wstring text, FontLoader* loader);
    // 迭代器操作符：解引用获取字符与字形对
    std::pair<wchar_t, FreeTypeGlyph*> operator*();
    // 迭代器操作符：下一个对字符/字形
    GlyphIterator& next();
    bool isEnd();

private:
    std::wstring text;
    std::wstring::const_iterator c;
    FontLoader* loader;
};
