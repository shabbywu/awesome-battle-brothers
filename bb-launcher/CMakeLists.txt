# CMakeList.txt: 顶层 CMake 项目文件，在此处执行全局配置 并包含子项目。
#
cmake_minimum_required(VERSION 3.26)

if(NOT WIN32)
  list(APPEND VCPKG_MANIFEST_FEATURES "with-opengl-tools-dependencies")
endif()

project("bb-launcher")

# 设置 x86 架构
if(WIN32)
  set(CMAKE_GENERATOR_PLATFORM x86)
endif()

set(CMAKE_CXX_STANDARD 20)

set(BIN2CPP_COMMAND ${CMAKE_CURRENT_LIST_DIR}/externals/bin/bin2cpp.exe)

set(nlohmann_json_dir ${CMAKE_CURRENT_LIST_DIR}/externals/nlohmann_json)
set(external_headers_dir ${CMAKE_CURRENT_LIST_DIR}/externals/includes)
set(embedded_assets_dir ${CMAKE_CURRENT_LIST_DIR}/embedded_assets)
set(opengl-tools-src-dir ${CMAKE_CURRENT_LIST_DIR}/opengl-tools/src)

file(
  DOWNLOAD
  https://github.com/shabbywu/awesome-battle-brothers/archive/refs/heads/external-libs.zip
  ${CMAKE_CURRENT_BINARY_DIR}/external-libs.zip
  SHOW_PROGRESS
  EXPECTED_MD5 1fe6e6d8bc542c17af7e0f1fa64e0ae4)
file(ARCHIVE_EXTRACT INPUT ${CMAKE_CURRENT_BINARY_DIR}/external-libs.zip
     DESTINATION ${CMAKE_CURRENT_BINARY_DIR})
set(external_libs_dir
    ${CMAKE_CURRENT_BINARY_DIR}/awesome-battle-brothers-external-libs/libs)

add_compile_options("$<$<C_COMPILER_ID:MSVC>:/utf-8>")
add_compile_options("$<$<CXX_COMPILER_ID:MSVC>:/utf-8>")

if((DEFINED ENV{BB_DEBUG}) AND ($ENV{BB_DEBUG}))
  message(NOTICE "ENABLE DEBUG MESSAGE")
  add_compile_definitions(BB_DEBUG_MESSAGE=1)
endif()

# 包含子项目 If you only include this third party in PRIVATE source files, you do not
# need to install it when your main project gets installed. Usage:
# target_link_libraries(example PRIVATE nlohmann_json::nlohmann_json)
set(JSON_Install
    OFF
    CACHE INTERNAL "")
add_subdirectory(${nlohmann_json_dir})

add_subdirectory("opengl-tools")

if(WIN32)
  add_subdirectory("${CMAKE_CURRENT_LIST_DIR}/externals/sqbind17")

  add_subdirectory("embedded_assets/embedded_exe")
  add_subdirectory("launcher")
  add_subdirectory("embedded_assets/online_battle")
endif()

if(DEFINED ENV{BB_LAUNCHER_VERSION})
  message(NOTICE BB_LAUNCHER_VERSION="$ENV{BB_LAUNCHER_VERSION}")
  target_compile_definitions(
    bb-launcher-gog PRIVATE BB_LAUNCHER_VERSION="$ENV{BB_LAUNCHER_VERSION}")
  target_compile_definitions(
    bb-launcher-steam PRIVATE BB_LAUNCHER_VERSION="$ENV{BB_LAUNCHER_VERSION}")
endif()

if(DEFINED ENV{FILEVER_VSINFO})
  message(NOTICE FILEVER_VSINFO="$ENV{FILEVER_VSINFO}")
  target_compile_definitions(bb-launcher-gog
                             PRIVATE FILEVER_VSINFO=$ENV{FILEVER_VSINFO})
  target_compile_definitions(bb-launcher-steam
                             PRIVATE FILEVER_VSINFO=$ENV{FILEVER_VSINFO})
endif()
