set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS ON)

if(WIN32)
  set(CMAKE_C_FLAGS_RELEASE "${CMAKE_C_FLAGS_RELEASE} /MT")
  set(CMAKE_CXX_FLAGS_RELEASE "${CMAKE_CXX_FLAGS_RELEASE} /MT")
endif()

if(NOT WIN32)
  find_package(OpenGL REQUIRED)
  find_package(glfw3 CONFIG REQUIRED)
  find_package(Freetype REQUIRED)
endif()

include_directories(${external_headers_dir})
link_directories(${external_libs_dir})

set(LIBS $<IF:$<TARGET_EXISTS:OpenGL::GL>,OpenGL::GL,opengl32>
         $<IF:$<TARGET_EXISTS:Freetype::Freetype>,Freetype::Freetype,freetype>)

# first create relevant static libraries required for other projects
add_library(GLAD "src/opengl/glad.c")
set(LIBS_MT ${LIBS} GLAD)
set(LIBS_MD ${LIBS} GLAD)

file(
  GLOB
  TOOLS_SOURCE
  "src/embedded_resources/*.h"
  "src/embedded_resources/*.cpp"
  "src/opengl/fonts/*.h"
  "src/opengl/fonts/*.cpp"
  "src/opengl/renders/*.h"
  "src/opengl/renders/*.cpp"
  "src/utils/*.h"
  "src/utils/*.cpp")

add_library(opengl-tools ${TOOLS_SOURCE})
target_link_libraries(opengl-tools PUBLIC ${LIBS_MT})
set_property(TARGET opengl-tools PROPERTY CXX_STANDARD 20)

file(GLOB BB_SOURCE "src/utils/*.h" "src/utils/*.cpp" "src/bb-text-render/*.h"
     "src/bb-text-render/*.cpp")

add_library(bb-text-render ${BB_SOURCE})
target_link_libraries(bb-text-render PUBLIC ${LIBS_MT} opengl-tools)
set_property(TARGET bb-text-render PROPERTY CXX_STANDARD 20)

add_executable(opengl-test src/testcase/main.cpp)
if(WIN32)
  add_library(opengl-tools-M ${TOOLS_SOURCE})
  target_link_libraries(opengl-tools-M PUBLIC ${LIBS_MD})
  target_link_libraries(opengl-test PUBLIC opengl-tools-M glfw)
  target_compile_options(opengl-tools-M PRIVATE $<$<CONFIG:Release>:-MD>)
  target_compile_options(opengl-test PRIVATE $<$<CONFIG:Release>:-MD>)
else()
  target_link_libraries(opengl-test PUBLIC opengl-tools glfw)
endif()
