#ifndef PATCH_CONTEXT_H
#define PATCH_CONTEXT_H
#include <string>
#include "bb.h"


struct PatchContext {
    Font* originFnt;
    Font* hackedFnt;
    std::string originText;
    std::string hackedText;
};

#endif
