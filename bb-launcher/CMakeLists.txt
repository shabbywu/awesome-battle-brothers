# 设置 x86 架构
if(WIN32)
  set(CMAKE_GENERATOR_PLATFORM x86)
endif()
set(CMAKE_CXX_STANDARD 20)

set(squirrel_dir ${CMAKE_CURRENT_LIST_DIR}/externals/SQUIRREL305)
set(external_headers_dir ${CMAKE_CURRENT_LIST_DIR}/externals/includes)
set(external_libs_dir ${CMAKE_CURRENT_LIST_DIR}/externals/libs)
set(opengl-tools-src-dir ${CMAKE_CURRENT_LIST_DIR}/opengl-tools/src)

add_compile_options("$<$<C_COMPILER_ID:MSVC>:/utf-8>")
add_compile_options("$<$<CXX_COMPILER_ID:MSVC>:/utf-8>")

if(DEFINED ENV{BB_DEBUG})
  if("$ENV{BB_DEBUG}" GREATER 0)
    message(NOTICE "ENABLE DEBUG MESSAGE")
    add_compile_definitions(BB_DEBUG_MESSAGE=1)
  endif()
endif()

if(DEFINED ENV{BB_LAUNCHER_VERSION})
  message(NOTICE BB_LAUNCHER_VERSION="$ENV{BB_LAUNCHER_VERSION}")
  add_compile_definitions(BB_LAUNCHER_VERSION="$ENV{BB_LAUNCHER_VERSION}")
endif()

# 包含子项目
add_subdirectory("opengl-tools")
if(WIN32)
  add_subdirectory("embedded_exe")
  # add_subdirectory("launcher-driver")
  add_subdirectory("launcher")
endif()
