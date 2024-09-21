#include <sqbind17/sqbind17.hpp>
#include "buffer/register.hpp"
#include "online-battle/register.hpp"
#ifdef BUILD_WITH_ONLINE_MOD
#include "online-battle/setup_embedded_mod.hpp"
#endif
#include "random/register.hpp"
#include "vm.hpp"
#include "json/register.hpp"

void sqext_override_isDevmode(HSQUIRRELVM &v) {
    if (vm == nullptr) {
        vm = new detail::VM(v);
    }
    auto roottable = sqbind17::detail::Table(_table(vm->roottable()), *vm);
    roottable.bindFunc("isDevmode", []() { return false; });
}

void sqext_register_jsonlib(HSQUIRRELVM &v) {
    if (vm == nullptr) {
        vm = new detail::VM(v);
    }
    sqext_json::sqext_register_jsonlib_impl(*vm);
}

void sqext_register_msgpacklib(HSQUIRRELVM &v) {
    if (vm == nullptr) {
        vm = new detail::VM(v);
    }
    sqext_msgpack::sqext_register_msgpacklib_impl(*vm);
}

void sqext_register_bufferio(HSQUIRRELVM &v) {
    if (vm == nullptr) {
        vm = new detail::VM(v);
    }
    sqext_buffer::sqext_register_bufferio_impl(*vm);
}

void sqext_register_online_battle(HSQUIRRELVM &v) {
    if (vm == nullptr) {
        vm = new detail::VM(v);
    }
    online_battle::sqext_register_online_battle_impl(*vm);
}

void sqext_register_random(HSQUIRRELVM &v) {
    if (vm == nullptr) {
        vm = new detail::VM(v);
    }
    sqext_random::sqext_register_random_impl(*vm);
}

void sqext_register_print(HSQUIRRELVM &v) {
    if (vm == nullptr) {
        vm = new detail::VM(v);
    }
    auto roottable = sqbind17::detail::Table(_table(vm->roottable()), *vm);
    roottable.bindFunc("print", [](SQObjectPtr obj) {
        if (sq_type(obj) == tagSQObjectType::OT_STRING) {
            std::cout << _stringval(obj) << std::endl;
        } else {
            std::cout << detail::sqobject_to_string(obj) << std::endl;
        }
    });
}
