set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS ON)

set(CMAKE_C_FLAGS_RELEASE "${CMAKE_C_FLAGS_RELEASE} /MT")
set(CMAKE_CXX_FLAGS_RELEASE "${CMAKE_CXX_FLAGS_RELEASE} /MT")

include_directories(${external_headers_dir})
include_directories(${squirrel_dir}/include ${squirrel_dir}/squirrel)
link_directories(${external_libs_dir})
add_subdirectory(src/testcase)

# add_library(bb-embedded-steam STATIC src/BattleBrothers.STEAM.cpp
# src/embed.cpp) target_compile_definitions(bb-embedded-steam PRIVATE STEAM_EXE)

# add_library(bb-embedded-gog STATIC src/BattleBrothers.GOG.cpp src/embed.cpp)
# target_compile_definitions(bb-embedded-gog PRIVATE GOG_EXE)
