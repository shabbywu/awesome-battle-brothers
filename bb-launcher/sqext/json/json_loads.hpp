#pragma once
#include <nlohmann/json.hpp>
#include <sqbind17/sqbind17.hpp>
#include <squirrel.h>
#include <string>

using namespace sqbind17;
using json = nlohmann::json;

namespace sqext_json {
static SQObjectPtr json_loads_impl(SQInteger integer) {
    return SQObjectPtr(integer);
}

static SQObjectPtr json_loads_impl(SQFloat floating) {
    return SQObjectPtr(floating);
}

static SQObjectPtr json_loads_impl(bool boolean) {
    return SQObjectPtr(boolean);
}

static SQObjectPtr json_loads_impl(std::string str) {
    return SQObjectPtr(SQString::Create(_ss(**vm), str.c_str(), str.size()));
}

static SQObjectPtr json_loads_impl(json obj);

static SQObjectPtr json_loads_impl(json::object_t table) {
    auto sq_table = sqbind17::detail::Table(*vm);
    for (auto [k, v] : table) {
        auto key = json_loads_impl(k);
        auto val = json_loads_impl(v);
        sq_table.set(key, val);
    }
    return sq_table.pTable();
}

static SQObjectPtr json_loads_impl(json::array_t array) {
    auto sq_array = sqbind17::detail::Array(*vm);
    for (json element : array) {
        auto item = json_loads_impl(element);
        sq_array.append(item);
    }
    return sq_array.pArray();
}

static SQObjectPtr json_loads_impl(json obj) {
    switch (obj.type()) {
    case json::value_t::string: {
        return json_loads_impl(obj.get<std::string>());
    }
    case json::value_t::array: {
        return json_loads_impl(obj.get<json::array_t>());
    }
    case json::value_t::object: {
        return json_loads_impl((json::object_t)obj);
    }
    case json::value_t::boolean: {
        return json_loads_impl(obj.get<bool>());
    }
    case json::value_t::null: {
        return SQObjectPtr();
    }
    case json::value_t::number_unsigned:
    case json::value_t::number_integer: {
        return json_loads_impl(obj.get<SQInteger>());
    }
    case json::value_t::number_float: {
        return json_loads_impl(obj.get<SQFloat>());
    }
    case json::value_t::binary: {
        // TODO: load as closure?
        return SQObjectPtr();
    }
    }
    return SQObjectPtr(SQString::Create(_ss(**vm), "<unknown object>", 17));
}

static SQObjectPtr json_loads(SQObjectPtr obj) {
    json j = json::parse(detail::generic_cast<SQObjectPtr, std::string>(*vm, std::forward<SQObjectPtr>(obj)));
    return json_loads_impl(j);
}

static SQObjectPtr json_loads(std::string text) {
    json j = json::parse(text);
    return json_loads_impl(j);
}
} // namespace sqext_json

namespace sqbind17 {
namespace detail {

// cast json to SQObjectPtr/HSQOBJECT
template <typename FromType, typename ToType,
          typename std::enable_if_t<std::is_same_v<json, std::decay_t<FromType>>> * = nullptr,
          typename std::enable_if_t<std::is_same_v<std::decay_t<ToType>, SQObjectPtr> ||
                                    std::is_same_v<std::decay_t<ToType>, HSQOBJECT>> * = nullptr>
static ToType generic_cast(detail::VM vm, FromType &&from) {
#ifdef TRACE_OBJECT_CAST
    std::cout << "[TRACING] cast " << typeid(decltype(from)).name() << " to " << typeid(ToType).name() << std::endl;
#endif
    return sqext_json::json_loads_impl(from);
}
} // namespace detail
} // namespace sqbind17
