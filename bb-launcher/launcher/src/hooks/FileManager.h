#pragma once
#include <string>
#include <functional>
#include <peconv.h>


namespace patchs { namespace FileManager {
        void InstallFileManagerHook(BYTE* loaded_pe, std::string digest, std::function<void()> callback);
    }
}
