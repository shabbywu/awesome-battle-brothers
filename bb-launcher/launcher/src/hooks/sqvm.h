#pragma once
#include <string>
#include <functional>
#include <peconv.h>

#include <squirrel.h>


namespace patchs {
    namespace vm {
        void PatchSQVM(BYTE* loaded_pe, std::string digest, std::function<void(HSQUIRRELVM)> on_squirrel_vm_opened);
        HSQUIRRELVM open_sqvm(int size, unsigned char libsToLoad);

        static const unsigned char LIB_IO   = 0x01;                                              ///< Input/Output library
        static const unsigned char LIB_BLOB = 0x02;                                              ///< Blob library
        static const unsigned char LIB_MATH = 0x04;                                              ///< Math library
        static const unsigned char LIB_SYST = 0x08;                                              ///< System library
        static const unsigned char LIB_STR  = 0x10;                                              ///< String library
        static const unsigned char LIB_ALL  = LIB_IO | LIB_BLOB | LIB_MATH | LIB_SYST | LIB_STR; ///< All libraries

    }
}
