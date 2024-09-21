#include <windows.h>

#ifndef BB_LAUNCHER_VERSION
#define BB_LAUNCHER_VERSION "develop"
#endif
#ifndef FILEVER_VSINFO
#define FILEVER_VSINFO      1,0,0,0
#endif

#define FILEVER_TEXT        BB_LAUNCHER_VERSION
#define COMPANY_NAME        "shabbywu\0"
#define FILE_DESCRIPTION    "Awesome Battle Brothers Launcher\0"
#define PRODUCT_NAME        "Awesome Battle Brothers Launcher\0"
#define PRODVER_VSINFO      1,5,0,15
#define PRODVER_TEXT        "1.5.0.15\0"
#ifdef GOG_EXE
#define INTERNAL_NAME        "bb-launcher-gog.exe\0"
#else
#define INTERNAL_NAME        "bb-launcher-steam.exe\0"
#endif
#define LEGAL_COPYRIGHT      "shabbywu (C) 2023-2024"
