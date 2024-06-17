#ifndef BB_TEXT_CACHE_H
#define BB_TEXT_CACHE_H
#include <map>
#include <vector>
#include "bb.h"
#include "PatchContext.h"


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
#endif
