#pragma once
#include <nlohmann/json.hpp>
#include <sqbind17/sqbind17.hpp>
#include <squirrel.h>
#include <string>

using namespace sqbind17;
using json = nlohmann::json;

namespace sqext_json {
static json json_dumps_impl(SQInteger integer) {
    return integer;
}

static json json_dumps_impl(SQFloat floating) {
    return floating;
}

static json json_dumps_impl(bool boolean) {
    return boolean;
}

static json json_dumps_impl(SQChar *str) {
    return str;
}

static json json_dumps_impl(SQObjectPtr obj);

static json json_dumps_impl(detail::Table table) {
    json json_object = json::object();
    for (auto [k, v] : table.iterator()) {
        json_object[detail::generic_cast<SQObjectPtr, std::string>(*vm, std::forward<SQObjectPtr>(k))] =
            json_dumps_impl(v);
    }
    return json_object;
}

static json json_dumps_impl(detail::Array array) {
    json json_array = json::array();
    for (SQObjectPtr element : array.iterator()) {
        json_array.push_back(json_dumps_impl(element));
    }
    return json_array;
}

static json json_dumps_impl(SQObjectPtr obj) {

    switch (obj._type) {
    case OT_ARRAY:
        return json_dumps_impl(detail::generic_cast<SQObjectPtr, detail::Array>(*vm, std::forward<SQObjectPtr>(obj)));
    case OT_TABLE: {
      return json_dumps_impl(detail::generic_cast<SQObjectPtr, detail::Table>(*vm, std::forward<SQObjectPtr>(obj)));
    }
    case OT_INTEGER: {
        return json_dumps_impl(_integer(obj));
    }
    case OT_FLOAT: {
        return json_dumps_impl(_float(obj));
    }
    case OT_BOOL: {
        return json_dumps_impl((bool)_integer(obj));
    }
    case OT_STRING: {
        return json_dumps_impl(_stringval(obj));
    }
    case OT_NULL:
        return nullptr;
    default:
        // OT_CLOSURE
        // OT_USERDATA
        // OT_NATIVECLOSURE
        return json(detail::sqobject_to_string(obj));
    }
}

static std::string json_dumps(SQObjectPtr obj, const int indent = 2, const char indent_char = ' ',
                              const bool ensure_ascii = false) {
    auto j = json_dumps_impl(obj);
    return j.dump(indent, indent_char, ensure_ascii);
}
} // namespace sqext_json

namespace sqbind17 {
namespace detail {
// cast SQObjectPtr/HSQOBJECT to json
template <typename FromType, typename ToType,
          typename std::enable_if_t<std::is_same_v<std::decay_t<FromType>, SQObjectPtr> ||
                                    std::is_same_v<std::decay_t<FromType>, HSQOBJECT>> * = nullptr,
          typename std::enable_if_t<std::is_same_v<json, std::decay_t<ToType>>> * = nullptr>
static ToType generic_cast(detail::VM vm, FromType &&from) {
#ifdef TRACE_OBJECT_CAST
    std::cout << "[TRACING] cast " << typeid(decltype(from)).name() << " to " << typeid(ToType).name() << std::endl;
#endif
    return sqext_json::json_dumps_impl(from);
}
} // namespace detail
} // namespace sqbind17
