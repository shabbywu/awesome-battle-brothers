#pragma once
#include "../vm.hpp"
#include <effolkronium/random.hpp>
#include <sqbind17/sqbind17.hpp>

namespace sqext_random {

using Random = effolkronium::random_static;

static unsigned int totalCall = 0;
static bool stackTrace = false;

void dump_caller_stack() {
    if (!stackTrace)
        return;
    auto v = vm->vm();
    auto level = 1;
    SQStackInfos info;
    if (sq_stackinfos(v, level, &info) == SQ_OK) {
        std::cout << "  stack[" << level << "] " << info.source << " > " << info.funcname << std::endl;
        if (std::string(info.funcname) == std::string("update")) {
            for (; level < v->_callsstacksize;) {
                if (sq_stackinfos(v, ++level, &info) == SQ_OK) {
                    std::cout << "  stack[" << level << "] " << info.source << " > " << info.funcname << std::endl;
                }
            }
        }
    }
}


void setStackTraceState(bool toggle) {
    stackTrace = toggle;
}

static int rand() {
    #ifndef RELEASE_BUILD
    std::cout << "call rand() with call time: " << ++totalCall << std::endl;
    dump_caller_stack();
    #endif
    return Random::get<int>();
}

static int randBetween(int left, int right) {
    #ifndef RELEASE_BUILD
    std::cout << "call randBetween() with call time: " << ++totalCall << std::endl;
    dump_caller_stack();
    #endif
    return Random::get<int>(left, right);
}

static void seedRandom(unsigned int seed) {
    if (seed == 0) {
        seed = time(NULL);
    }
    std::cout << "seed to: " << seed << std::endl;
    totalCall = 0;
    Random::seed(seed);
}

static void seedRandomString(std::string seedString) {
    unsigned int seed = 0;
    for (auto c : seedString) {
        seed += (unsigned int)c;
    }
    totalCall = 0;
    Random::seed(seed);
}

void static sqext_register_random_impl(sqbind17::detail::VM &v) {
    vm = &v;
    Random::seed(time(NULL));

    auto random = sqbind17::detail::Table(*vm);
    random.bindFunc("rand", &rand);
    random.bindFunc("rand", &randBetween);
    random.bindFunc("seedRandom", &seedRandom);
    random.bindFunc("seedRandomString", &seedRandomString);
    random.bindFunc("setStackTraceState", &setStackTraceState);

    auto roottable = sqbind17::detail::Table(_table(vm->roottable()), *vm);
    roottable.set(std::string("Random"), random);
}
} // namespace sqext_random
