#include <chrono>
#include <thread>
#include <array>
#include <iostream>

using namespace std::chrono_literals;

/*
A program which will dynamic allocate more than 2gb memory.
For 32-bit application, this program will never run succeed unless it was patch with 4gb-patch.
*/
int main() {
    std::array<char*, 22> charPtrs;
    for (size_t i = 0; i < 22; i++)
    {
        charPtrs[i] = (char*)malloc(1024 * 1024 * 96);
        if (charPtrs[i] == NULL) {
            std::cerr << "[x] Failed to allocated memory." << std::endl;
            std::cin.get();
            return 1;
        } else {
            std::cout << "Allocated memory at least: " << 96 * i << " MB.\n";
            std::this_thread::sleep_for(200ms);
        }

    }
    std::cout << "[✓] Allocated more than 2GB memory." << std::endl;
    std::cin.get();
    return 0;
}
