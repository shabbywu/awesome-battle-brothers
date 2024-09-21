#pragma once
#include "../../vm.hpp"
#include "client.hpp"
#include <nlohmann/json.hpp>
#include <set>
#include <sqbind17/sqbind17.hpp>

using namespace sqbind17;
using json = nlohmann::json;

namespace nakama {
namespace storages {
json loadPlayerRoster() {
    Nakama::NReadStorageObjectId fetchRequest = {
        .collection = "roster",
        .key = "inFormation",
        .userId = session->getUserId(),
    };
    auto objs = client->readStorageObjectsAsync(session, {fetchRequest}).get();
    if (objs.size() == 0) {
        return json{{"entities", json::array_t()}};
    }
    auto o = objs[0];
    return json::parse(o.value);
}

void savePlayerRoster(detail::Table roster) {
    auto data = sqext_json::json_dumps((SQObjectPtr)roster.pTable());
    Nakama::NStorageObjectWrite saveRequest = {
        .collection = "roster",
        .key = "inFormation",
        .value = data,
        .permissionRead = Nakama::NStoragePermissionRead::OWNER_READ,
        .permissionWrite = Nakama::NStoragePermissionWrite::OWNER_WRITE,
    };
    client->writeStorageObjectsAsync(session, {saveRequest}).get();
}
} // namespace storages
} // namespace nakama
