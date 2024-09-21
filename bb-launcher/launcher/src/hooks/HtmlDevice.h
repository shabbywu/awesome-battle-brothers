#pragma once
#include <string>
#include <functional>
#include <peconv.h>


namespace patchs { namespace RedirectLogs {
        void InstallRedirectLogsHook(BYTE* loaded_pe, std::string digest);
    }
}
