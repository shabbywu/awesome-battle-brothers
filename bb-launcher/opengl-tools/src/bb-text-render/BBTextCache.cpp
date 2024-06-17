#include "BBTextCache.h"
#include "TextureIDHolder.h"
#include "../embedded_resources/font.ttf.h"
#include "../opengl/fonts/FontLoader.h"
#include "../opengl/fonts/BitmapFontBuilder.h"
#include "../opengl/fonts/BitmapFontCache.h"
#include "../utils/converterX.h"

std::map<BitmapFont*, TextureIDHolder> textureCache;;
std::map<float, BitmapFontCache*> loaders;


TextureIDHolder getOrCreateTextureIDHolder(std::string text, float fontsize) {
    if (loaders.find(fontsize) == loaders.end()) {
        BitmapFontCache* contexts = new BitmapFontCache(BitmapFontBuilder(new FontLoader(bb::font::GetFontContent(), bb::font::GetFontSize(), fontsize)));
        loaders.insert(std::pair<float, BitmapFontCache*>(fontsize, contexts));
    }

    BitmapFontCache* contexts = loaders[fontsize];
    BitmapFont* fnt = contexts->GetOrCreateBitmapFont(text + "ABCDEFGHIJKLMNOPQRSTUVWXYZ");

    if (textureCache.find(fnt) == textureCache.end()) {
        TextureIDHolder h = TextureIDHolder{ fnt, 0 };
        h.initTextureID();
        textureCache[fnt] = h;
    }
    return textureCache[fnt];
};



void BBTextCache::gc() {
    GcTimer1 += 1;
    GcTimer2 += 1;
    // 65536 = 2**16
    // 由于 gc2 会重置 gc1 的计数器
    // 实际上 gc1 的间隔是 2**15 + 2**17 次计数
    // gc1 与 gc2 的最少间隔是 2**15 次计数
    if (GcTimer1 >= 65536) {
        GcTimer1 = 0;
        for (auto p = FntGarbageCollector1.begin(); p != FntGarbageCollector1.end(); p++) {
            FntGarbageCollector2.push_back(*p);
        }
        FntGarbageCollector1.clear();
        TextGarbageCollector2 = TextGarbageCollector1;
    }
    // 98304 = 2**15 + 2**16
    if (GcTimer2 >= 98304) {
        GcTimer1 = GcTimer2 = 0;
        for (auto p = FntGarbageCollector2.begin(); p != FntGarbageCollector2.end(); p++) {
            delete* p;
            deleted++;
        }
        FntGarbageCollector2.clear();
        for (auto p = TextGarbageCollector2.begin(); p != TextGarbageCollector2.end(); p++) {
            if (TextGarbageCollector1[p->first] == p->second) {
                PatchContext& ctx = contexts[p->first];
                delete textCache[p->first];
                delete ctx.hackedFnt;
                textCache.erase(p->first);
                contexts.erase(p->first);
                TextGarbageCollector1.erase(p->first);
                deleted += 2;
            }
        }
    }
};


void BBTextCache::buildTextCtx(Text* text) {
    gc();
    if (contexts.find(text) != contexts.end()) {
        PatchContext& ctx = contexts[text];
        if (ctx.originText != text->m_sText) {
            // Text 对象被重复利用, 上下文已经失效了
            FntGarbageCollector1.push_back(ctx.hackedFnt);
            contexts.erase(text);
        }
    }
    if (contexts.find(text) == contexts.end()) {
        TextureIDHolder holder = getOrCreateTextureIDHolder(text->m_sText, text->m_pFont.m_pPointee->m_fCharacterHeight);
        // 构造 Font 对象
        Font* originFnt = text->m_pFont.m_pPointee;
        // 仅初始化 Entity 的字段
        Font* hackedFnt = new Font{};
        *hackedFnt = *originFnt;
        hackedFnt->m_uTextureID = holder.TextureID;
        hackedFnt->m_iWidth = (int32_t)holder.fnt->width;
        hackedFnt->m_iHeight = (int32_t)holder.fnt->height;
        hackedFnt->m_uSizeInMemory = (uint32_t)(holder.fnt->width * holder.fnt->height);
        // 接下来初始化 Font 特有的字段
        // m_Characters
        // m_fCharacterWidth
        // m_fCharacterHeight
        for (int idx = 0; idx <= 0x100; idx++) {
            hackedFnt->m_fCharacterWidth[idx] = 0.0f;
        }
        hackedFnt->m_fCharacterHeight = text->m_pFont.m_pPointee->m_fCharacterHeight;

        // 压缩编码
        char start = 'A';
        std::map<wchar_t, char> projection;
        std::wstring wide = wstringFromUTF8(text->m_sText);
        for (auto cp = wide.begin(); cp != wide.end(); cp++) {
            wchar_t c = *cp;
            // alias a-z to A-Z
            if (c >= 97 && c <= 122) {
                c = c - 32;
            }
            if (projection.find(c) != projection.end()) {
                continue;
            }
            projection[c] = start++;
        }
        std::string new_sText = "";
        for (auto cp = wide.begin(); cp != wide.end(); cp++) {
            wchar_t c = *cp;
            // alias a-z to A-Z
            if (c >= 97 && c <= 122) {
                c = c - 32;
            }
            new_sText += projection[c];
        }

        // 设置字符大小
        for (auto pair = projection.begin(); pair != projection.end(); pair++) {
            wchar_t c = pair->first;
            wchar_t e = pair->second;
            auto ch = holder.fnt->Characters[c];
            // 设置 m_Characters 和 m_fCharacterWidth

            hackedFnt->m_Characters[e] = Rect{
                (ch.TopXY.x) / holder.fnt->width,
                (ch.TopXY.x + ch.Advance) / holder.fnt->width,
                (ch.TopXY.y + ch.Size.y) / holder.fnt->height,
                (ch.TopXY.y) / holder.fnt->height,
            };
            hackedFnt->m_fCharacterWidth[e] = ch.Advance;
        }

        contexts[text] = PatchContext{
            originFnt,
            hackedFnt,
            text->m_sText,
            new_sText
        };
        // 初始化计数器
        TextGarbageCollector1[text] = 0;
    }
    TextGarbageCollector1[text] += 1;
};


Text* BBTextCache::createFakeText(Text* text) {
    buildTextCtx(text);
    // 判断 contexts 是否已失效
    // 如果无缓存, 即缓存 text
    if (textCache.find(text) == textCache.end()) {
        PatchContext& ctx = contexts[text];
        Text* fake = new Text{};
        *fake = *text;
        fake->m_sText = ctx.hackedText;
        fake->m_pFont.m_pPointee = ctx.hackedFnt;
        textCache[text] = fake;
    }
    else {
        PatchContext& ctx = contexts[text];
        Text* fake = textCache[text];
        *fake = *text;
        fake->m_sText = ctx.hackedText;
        fake->m_pFont.m_pPointee = ctx.hackedFnt;
    }
    return textCache[text];
};
