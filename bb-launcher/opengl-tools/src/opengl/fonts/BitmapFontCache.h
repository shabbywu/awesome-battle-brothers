#pragma once
#include <set>
#include <vector>
#include "BitmapFontBuilder.h"


class BitmapFontHolder {
public:
    BitmapFontHolder(std::string text, BitmapFontBuilder& builder);
    bool Contains(std::string text);
    bool Contains(std::wstring text);

    std::set<wchar_t> set;
    BitmapFont* fnt;
};



class BitmapFontCache {
public:
    BitmapFontCache(BitmapFontBuilder builder);
    BitmapFont* GetOrCreateBitmapFont(std::string text);
    int Size() {
        return fonts.size();
    }

private:
    BitmapFontBuilder builder;
    std::vector<BitmapFontHolder> fonts;
};
