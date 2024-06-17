# CMakeList.txt: 顶层 CMake 项目文件，在此处执行全局配置 并包含子项目。
#
cmake_minimum_required(VERSION 3.12)
if(NOT WIN32)
  list(APPEND VCPKG_MANIFEST_FEATURES "with-opengl-tools-dependencies")
endif()

set(VCPKG_OVERLAY_TRIPLETS "${CMAKE_CURRENT_LIST_DIR}/.vcpkg/triplets")
# Enable VCPKG if VCPKG_ROOT defined.
set(vcpkg "$ENV{VCPKG_ROOT}/scripts/buildsystems/vcpkg.cmake")
if(NOT CMAKE_TOOLCHAIN_FILE AND EXISTS "${vcpkg}")
  set(CMAKE_TOOLCHAIN_FILE
      "${vcpkg}"
      CACHE FILEPATH "CMake toolchain file")
  message(STATUS "vcpkg toolchain found: ${CMAKE_TOOLCHAIN_FILE}")
endif()

project("awesome-battle-brothers")

# 包含子项目
add_subdirectory("bb-launcher")
