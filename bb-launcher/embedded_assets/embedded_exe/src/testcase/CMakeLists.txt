add_executable(allocate_2gb_memory main.cpp)
add_executable(allocate_2gb_memory.patched 4gb.cpp)
target_link_libraries(allocate_2gb_memory.patched PRIVATE libpeconv_mt)
target_link_directories(allocate_2gb_memory.patched PRIVATE
                        ${external_libs_dir})
if(MSVC)
  set_target_properties(allocate_2gb_memory.patched
                        PROPERTIES LINK_FLAGS "/LARGEADDRESSAWARE")
endif()
