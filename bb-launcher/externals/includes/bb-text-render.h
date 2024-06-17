#pragma once
#include <string>
#include <map>
#include <vector>
#include <bb.h>


struct PatchContext {
    Font* originFnt;
    Font* hackedFnt;
    std::string originText;
    std::string hackedText;
};


class BBTextCache {
public:
    std::map<Text*, PatchContext> contexts;
    std::map<Text*, Text*> textCache;

    // gc
    std::map<Text*, int> TextGarbageCollector1;
    std::map<Text*, int> TextGarbageCollector2;
    std::vector<Font*> FntGarbageCollector1;
    std::vector<Font*> FntGarbageCollector2;
    uint32_t GcTimer1 = 0;
    uint32_t GcTimer2 = 0;
    uint32_t deleted = 0;

    void gc();
    void buildTextCtx(Text* text);
    Text* createFakeText(Text* text);
};
