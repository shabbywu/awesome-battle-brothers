#pragma once
#include <string>
#include <functional>
#include <peconv.h>


namespace patchs { namespace ICore {
        void InstallICoreHook(BYTE* loaded_pe, std::string digest);

        void CallInMainThread(std::function<void(void)> callback, bool once);
    }
}
