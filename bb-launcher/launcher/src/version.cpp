#ifndef BB_LAUNCHER_VERSION
#define BB_LAUNCHER_VERSION "develop"
#endif
#include <iostream>
#include <string>

namespace launcher {
std::string get_version() {
    return BB_LAUNCHER_VERSION;
}
} // namespace launcher
