#pragma once
#include "json_loads.hpp"

namespace sqext_msgpack {
static SQObjectPtr msgpack_loads(SQObjectPtr obj) {
    std::string string = detail::generic_cast<SQObjectPtr, std::string>(*vm, std::forward<SQObjectPtr>(obj));
    json j = json::from_msgpack(string);
    return sqext_json::json_loads_impl(j);
}
} // namespace sqext_msgpack
