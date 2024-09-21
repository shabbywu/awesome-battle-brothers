#include <iostream>
#include <peconv.h>
#include <tchar.h>


using namespace peconv;

LPCTSTR executable_filename = TEXT("allocate_2gb_memory.exe");
static void(_cdecl* ep_func)();

int main(int argc, LPTSTR argv[]) {
    size_t v_size = 0;
    std::cout << "[*] Loading PE..." << std::endl;
    BYTE* self_pe = load_pe_executable(argv[0], v_size);
    std::cout << "[*] Loaded, size: " << v_size << std::endl;

    std::cout << "[*] Verifying 4GB Patch" << std::endl;
    std::cout << "IMAGE_FILE_LARGE_ADDRESS_AWARE: " << bool(get_file_characteristics(self_pe) & IMAGE_FILE_LARGE_ADDRESS_AWARE) << std::endl;

    std::cout << "[*] Loading PE..." << std::endl;
    if (argc != 1) {
        executable_filename = argv[1];
    }
    BYTE* loaded_pe = load_pe_executable(executable_filename, v_size);

    std::cout << "[*] Loaded, size: " << v_size << std::endl;
    std::cout << "[*] Running..." << std::endl;
    // if the loaded PE needs to access resources, you may need to connect it to the PEB:
    // peconv::set_main_module_in_peb((HMODULE)loaded_pe);

    // calculate the Entry Point of the manually loaded module
    DWORD ep_rva = peconv::get_entry_point_rva(loaded_pe);
    if (!ep_rva) {
        return -2;
    }

    ULONGLONG ep_exp_offset = (ULONGLONG)loaded_pe + ep_rva;
    ep_func = (void(_cdecl*)()) (ep_exp_offset);

    // call the Entry Point of the manually loaded PE:
    ep_func();

    peconv::free_pe_buffer(loaded_pe, v_size);
    std::cout << "[*] Exited" << std::endl;
    return 0;
}
