#pragma once
#include <string>
#include <functional>
#include <peconv.h>

#include <squirrel.h>


namespace patchs {
    namespace BattleBrothers {
        void InstallBattleBrothersInitHooks(BYTE* loaded_pe, std::string digest, std::function<void()> on_scripts_binding_inited_hook = nullptr);
    }
}
