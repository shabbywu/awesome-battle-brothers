#pragma once
#include "json_dumps.hpp"
#include <base64.hpp>

namespace sqext_msgpack {
static std::string msgpack_dumps(HSQOBJECT obj) {
  auto ptr = SQObjectPtr(obj);
  auto j = sqext_json::json_dumps_impl(ptr);
  auto msgpack = json::to_msgpack(j);
  return base64::encode_into<std::string>(std::begin(msgpack),
                                          std::end(msgpack));
}
} // namespace sqext_msgpack
