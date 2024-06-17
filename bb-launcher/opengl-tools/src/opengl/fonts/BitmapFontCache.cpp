#include "BitmapFontCache.h"
#include "../../utils/converterX.h"


BitmapFontHolder::BitmapFontHolder(std::string text, BitmapFontBuilder& builder) {
    std::wstring wide = wstringFromUTF8(text);
    for (auto c = wide.begin(); c != wide.end(); c++) {
        set.insert(*c);
    }
    this->fnt = builder.Build(text);
};


bool BitmapFontHolder::Contains(std::string text) {
    return Contains(wstringFromUTF8(text));
};


bool BitmapFontHolder::Contains(std::wstring text) {
    for (auto c = text.begin(); c != text.end(); c++) {
        if (set.find(*c) == set.end()) {
            return false;
        }
    }
    return true;
}


BitmapFontCache::BitmapFontCache(BitmapFontBuilder builder) : builder{ builder } {
};


BitmapFont* BitmapFontCache::GetOrCreateBitmapFont(std::string text) {
    std::wstring wide = wstringFromUTF8(text);

    for (auto font = fonts.begin(); font != fonts.end(); font++) {
        if (font->Contains(wide)) {
            return font->fnt;
        }
    }
    fonts.emplace_back(text, builder);
    return fonts.back().fnt;
}
