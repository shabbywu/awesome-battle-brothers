#pragma once
#include <functional>
#include <peconv.h>
#include <string>

namespace patchs {
namespace atexit {
void InstallAtExit(BYTE *loaded_pe, std::string digest, void (*_Function)(void));
}
} // namespace patchs
