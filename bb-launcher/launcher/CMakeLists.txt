set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS ON)

set(CMAKE_C_FLAGS_RELEASE "${CMAKE_C_FLAGS_RELEASE} /MT")
set(CMAKE_CXX_FLAGS_RELEASE "${CMAKE_CXX_FLAGS_RELEASE} /MT")

include_directories(${external_headers_dir})
include_directories(${embedded_assets_dir})
link_directories(${external_libs_dir})

file(GLOB SOURCE "src/*.h" "src/*.cpp" "src/sqext/*.cpp" "src/version.rc")

set(LIBS physfs-static libpeconv_mt bb-text-render bb-launcher-hooks sqbind17::header)

add_executable(bb-launcher-gog ${SOURCE})
target_link_libraries(bb-launcher-gog PRIVATE ${LIBS} bb-launcher-driver-gog)
target_compile_definitions(bb-launcher-gog PRIVATE GOG_EXE
                                                   DISABLE_MOD_MOUNTING_FEATURE)

add_executable(bb-launcher-steam ${SOURCE})
target_link_libraries(bb-launcher-steam PRIVATE ${LIBS}
                                                bb-launcher-driver-steam)
target_compile_definitions(bb-launcher-steam
                           PRIVATE STEAM_EXE DISABLE_MOD_MOUNTING_FEATURE)

# SET LARGEADDRESSAWARE to allow App to handle >2gb addresses
if(MSVC)
  set_target_properties(bb-launcher-steam PROPERTIES LINK_FLAGS
                                                     "/LARGEADDRESSAWARE")
  set_target_properties(bb-launcher-gog PROPERTIES LINK_FLAGS
                                                   "/LARGEADDRESSAWARE")
endif()

if((DEFINED ENV{BB_RELEASE_BUILD}) AND ($ENV{BB_RELEASE_BUILD}))
  add_compile_definitions(RELEASE_BUILD=1)
endif()

if((DEFINED ENV{BB_BUILD_WITH_ONLINE_MOD}) AND ($ENV{BB_BUILD_WITH_ONLINE_MOD}))
  target_compile_definitions(bb-launcher-gog PUBLIC -DBUILD_WITH_ONLINE_MOD)
  target_link_libraries(bb-launcher-gog PRIVATE embbed_online_battle)
  set_target_properties(bb-launcher-gog PROPERTIES OUTPUT_NAME
                                                   online-championships-gog)

  target_compile_definitions(bb-launcher-steam PUBLIC -DBUILD_WITH_ONLINE_MOD)
  target_link_libraries(bb-launcher-steam PRIVATE embbed_online_battle)
  set_target_properties(bb-launcher-steam PROPERTIES OUTPUT_NAME
                                                     online-championships-steam)
endif()
