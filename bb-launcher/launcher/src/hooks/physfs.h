#pragma once
#include <peconv.h>
#include <filesystem>
#include <string>
#include <map>


namespace patchs {
    namespace physfs {
        // patch all physfs func to physfs3.0
        void UpgrateToPhysfs3(BYTE* pe, std::filesystem::path gamepath);
        std::map<std::string, ULONGLONG> GetPhysfsLibrary();
    }
}
