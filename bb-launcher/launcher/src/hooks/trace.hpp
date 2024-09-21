#pragma once
#include <functional>
#include <peconv.h>
using namespace peconv;

namespace patchs {
namespace trace {
namespace thiscall {
template <ULONGLONG address, typename Return, typename... Args> class Tracer {
  public:
    static inline PatchBackup backup;
    static inline ULONGLONG g_offset;
    static inline Return(__thiscall *original)(void*, Args...);
    static inline std::string scope;

    static Return __fastcall patch(void * that, void *edx, Args... args) {
        std::cout << "[TRACE] before call " << scope << " -> " << typeid(decltype(original)).name() << std::endl;
        // un-patch hook to call original call
        backup.applyBackup();
        // call original base register
        Return result = original(that, args...);
        std::cout << "[TRACE] after call " << scope << std::endl;
        // re-patch hook
        redirect_to_local32((BYTE *)g_offset, (ULONGLONG)&patch);
        return result;
    }

    Tracer(BYTE *loaded_pe, std::string scope) {
        this->scope = scope;
        size_t v_size = 0;
        g_offset = (ULONGLONG)loaded_pe + (address - 0x400000);
        original = decltype(original)(g_offset);
        if (!redirect_to_local32((BYTE *)original, (ULONGLONG)&patch, &backup)) {
            std::cout << "Failed replacing target!" << std::endl;
            std::cout << (ULONGLONG)loaded_pe << std::endl;
            free_pe_buffer(loaded_pe, v_size);
            exit(1);
        }
    }
};
} // namespace thiscall

template <ULONGLONG address, typename Signature, typename... Args> class VoidTracer {
  public:
    static inline PatchBackup backup;
    static inline ULONGLONG g_offset;
    static inline Signature original;
    static inline std::string scope;

    static void patch(Args... args) {
        std::cout << "[TRACE] before call " << scope << " -> " << typeid(decltype(original)).name() << std::endl;
        // un-patch hook to call original call
        backup.applyBackup();
        // call original base register
        original(args...);
        std::cout << "[TRACE] after call " << scope << std::endl;
        // re-patch hook
        redirect_to_local32((BYTE *)g_offset, (ULONGLONG)&patch);
        return;
    }

    VoidTracer(BYTE *loaded_pe, std::string scope) {
        this->scope = scope;
        size_t v_size = 0;
        g_offset = (ULONGLONG)loaded_pe + (address - 0x400000);
        original = decltype(original)(g_offset);
        if (!redirect_to_local32((BYTE *)original, (ULONGLONG)&patch, &backup)) {
            std::cout << "Failed replacing target!" << std::endl;
            std::cout << (ULONGLONG)loaded_pe << std::endl;
            free_pe_buffer(loaded_pe, v_size);
            exit(1);
        }
    }
};

} // namespace trace
} // namespace patchs
