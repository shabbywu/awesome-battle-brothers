#pragma once
#include <peconv.h>
#include <filesystem>
#include <string>
#include <map>


namespace patchs {
    namespace CoherentGT {
        struct SystemSettings {
            int32_t DebuggerPort;
            char const* InspectorResourcesFolder;
            void* ProxyData;
            void* DiskCache;
            void* MemoryAllocator;
            void* LocalizationManagerInstance;
            uint8_t EnableWebSecurity;
            uint8_t AllowPerformanceWarnings;
            uint8_t RunAsynchronous;
            int JavaScriptGCMode;
            uint8_t ControlTime;
            uint8_t IgnoreSSLCertificateErrors;
            char const* CertificateAuthorities;
            uint32_t CertificateAuthoritiesLength;
            float AnimationUpdateDefer;
            uint8_t AllowMultipleRenderingThreads;
            char const* CookiesResource;
            char const* LocalStorageFolder;
            uint8_t SetRenderingResourcesDebugNames;
            uint8_t LoadSystemFonts;
            int32_t TestDriverPort;
            char const* TestDriverResourcesFolder;
            char const* SharedLibraryLocation;
            void* FileManipulatorInstance;
            void* EnabledFeatures;
        };

        namespace InitializeUIGTSystem {
            void* patch(const char *licenseKey, SystemSettings* settings, int severity, void* logHandler, void* systemListener);
        }

        void LoadCoherentGTLib(std::filesystem::path gamepat);
    }
}
