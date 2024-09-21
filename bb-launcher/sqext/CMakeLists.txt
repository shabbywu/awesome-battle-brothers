set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS ON)

set(CMAKE_C_FLAGS_RELEASE "${CMAKE_C_FLAGS_RELEASE} /MT")
set(CMAKE_CXX_FLAGS_RELEASE "${CMAKE_CXX_FLAGS_RELEASE} /MT")

include_directories(${external_headers_dir})
include_directories(${embedded_assets_dir})
link_directories(${external_libs_dir})

add_library(squirrel-extensions ext.cpp)
target_link_libraries(squirrel-extensions PUBLIC nlohmann_json::nlohmann_json
                                                 sqbind17::sqbind17 nakama-sdk)

if((DEFINED ENV{BB_BUILD_WITH_ONLINE_MOD}) AND ($ENV{BB_BUILD_WITH_ONLINE_MOD}))
  target_compile_definitions(squirrel-extensions PUBLIC -DBUILD_WITH_ONLINE_MOD)
endif()
target_compile_definitions(squirrel-extensions PUBLIC NAKAMA_SERVERKEY
                                                      ${NAKAMA_SERVERKEY})
target_compile_definitions(squirrel-extensions PUBLIC NAKAMA_HOST
                                                      ${NAKAMA_HOST})
