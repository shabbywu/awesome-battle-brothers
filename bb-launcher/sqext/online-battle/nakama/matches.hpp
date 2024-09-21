#pragma once
#include "../../json/register.hpp"
#include "../../random/register.hpp"
#include "../../vm.hpp"
#include "client.hpp"
#include <nlohmann/json.hpp>
#include <set>
#include <sqbind17/sqbind17.hpp>

using namespace sqbind17;
using json = nlohmann::json;

namespace nakama {
namespace matches {
enum class State {
    IN_HALL, // 在大厅
    IN_MATCHING,
    IN_BATTLE,
    WAITING_OPPONENT,
    WAITING_START,
};

enum class Faction {
    None = 0,
    RedSide = 1,
    BlueSide = 2,
};

enum class MatchOpCode {
    // 对局关键事件
    TACTICAL__SKILL_USE,
    TACTICAL__TRAVEL,
    TACTICAL__ENTITY_WAIT_TURN,
    TACTICAL__ENTITY_TURN_END,
    // 服务器事件
    // - 服务器关闭
    MATCH_TERMINATE = 100,
    // 派系分配
    MATCH_FACTION_DISPATCH,
    // 战斗结束
    BATTLE_END,

    // 对局关键信息
    // - 对手准备就绪
    OPPONENT_READY = 200,
    // - 开启比赛 -> 进入 Tactical State -> 由 host 发送该事件
    MATCH_START,
    // - 对手强制退出
    OPPONENT_FORCE_QUITED,
    // - 同步随机数种子
    SYNC_SEED,
};

class GameObject {
  public:
    Faction faction;
    State state;
    bool isVisitor = false;
    Nakama::NUserPresence presence;
    // update roster when player ready
    json roster;

  public:
    GameObject(Nakama::NUserPresence presence, bool isVisitor = false)
        : state(State::IN_BATTLE), isVisitor(isVisitor), presence(presence), roster(nullptr), faction(Faction::None) {
    }
};

std::map<std::string, std::shared_ptr<GameObject>> players = {};
std::weak_ptr<GameObject> currentUser;
static Nakama::NMatch match;
static Nakama::NMatchmakerTicket matchTicket{.ticket = ""};

std::shared_ptr<detail::Table> MatchManager = nullptr;
std::shared_ptr<detail::Table> Callbacks = nullptr;
static std::shared_ptr<GameObject> spawn_player(Nakama::NUserPresence presence);

void removeMatchMaker() {
    if (matchTicket.ticket != "") {
        rtClient->removeMatchmakerAsync(matchTicket.ticket).get();
    }
    matchTicket.ticket = "";
}

void createMatchMaker() {
    removeMatchMaker();
    matchTicket = rtClient->addMatchmakerAsync(2, 2, "*").get();
}

// 分割线
int getPlayerFaction() {
    if (auto p = currentUser.lock(); p != nullptr) {
        return (int)p->faction;
    }
    return (int)Faction::None;
}

json listMatches() {
    json::array_t matches;
    Nakama::NMatchListPtr data = client->listMatchesAsync(session).get();
    for (Nakama::NMatch match : data->matches) {
        auto data = json::parse(match.label);
        data["id"] = match.matchId;
        matches.push_back(data);
    }
    return matches;
}

bool joinMatch(std::string matchId) {
    try {
        match = rtClient->joinMatchAsync(matchId, {}).get();
    }
    catch (const std::exception &e) {
        patchs::ICore::CallInMainThread([e]() {
            showToast->operator()({
                {"text", std::format("failed to join match: {}", e.what())},
                {"type", "error"},
            });
        });
        return false;
    }

    players.clear();
    for (auto &presence : match.presences) {
        auto gameObject = spawn_player(presence);
        players[presence.sessionId] = gameObject;
    }

    if (auto p = players.find(match.self.sessionId); p == players.end()) {
        auto gameObject = spawn_player(match.self);
        players[match.self.sessionId] = gameObject;
    }
    std::cout << "players.size: " << players.size() << std::endl;
    std::cout << "self: " << match.self.sessionId << std::endl;
    return true;
}

json createMatch(json options) {
    if (players.size() > 0) {
        players.clear();
    }

    Nakama::NRpc rpc;
    try {
        rpc = rtClient->rpcAsync("createMatch", options.dump(-1)).get();
    }
    catch (const std::exception &e) {
        sq_throwerror(vm->vm(), e.what());
        return nullptr;
    }

    auto data = json::parse(rpc.payload);
    NLOG_INFO(std::format("create match rpc result: {}", data.dump(-1)));
    if (joinMatch(data["id"])) {
        return data;
    } else {
        return nullptr;
    }
}

void leaveMatch(bool forceQuit = false) {
    if (forceQuit) {
        rtClient->sendMatchDataAsync(match.matchId, (int64_t)MatchOpCode::OPPONENT_FORCE_QUITED, "{}");
    }

    rtClient->leaveMatchAsync(match.matchId);
}

void setReady(json roster) {
    rtClient->sendMatchDataAsync(match.matchId, (int64_t)MatchOpCode::OPPONENT_READY, roster.dump(-1)).get();
    if (auto p = currentUser.lock(); p != nullptr) {
        p->state = State::WAITING_START;
        p->roster = roster;
    } else {
        NLOG_ERROR("current user is not set!!!!!!!!!!!");
    }
}

void emitEvent(int opCode, json args) {
    try {
        if (auto p = currentUser.lock(); p != nullptr) {
            rtClient->sendMatchDataAsync(match.matchId, opCode, args.dump(-1)).get();
        }
    }
    catch (const std::exception &e) {
        NLOG_ERROR(std::format("emit event<{}> but an exception<{}> raised!", opCode, e.what()));
    }
}

void setupMatchBinding() {
    auto roottable = detail::Table(_table(vm->roottable()), *vm);
    auto module = roottable.get<std::string, detail::Table>(std::string("online_battle_module"));
    auto matchManager = module.get<std::string, detail::Table>(std::string("MatchManager"));
    MatchManager = std::make_shared<detail::Table>(matchManager.pTable(), *vm);

    auto callbacks = module.get<std::string, detail::Table>(std::string("callbacks"));
    Callbacks = std::make_shared<detail::Table>(callbacks.pTable(), *vm);
}

static std::shared_ptr<GameObject> spawn_player(Nakama::NUserPresence presence) {
    auto player = std::make_shared<GameObject>(presence);
    if (match.self.sessionId == presence.sessionId) {
        player->state = State::IN_BATTLE;
        currentUser = player;
    }
    return player;
}

// match op handler
static void onMatchMatchTerminate() {
    // players.clear();
    // currentUser = nullptr;
}

static void onMatchFactionDispatch(json message) {
    NLOG_INFO(std::format("server dispatch a faction: {}", message.dump(-1)));
    auto presence = std::string(message["presence"]);
    for (auto &[sessionid, p] : players) {
        if (sessionid == presence) {
            p->faction = (Faction)(message["faction"]);
            return;
        }
    }
    NLOG_ERROR("no player found!!!!");
}

static void onBattleEnd(json battleEndEvent) {
    Faction faction = battleEndEvent["faction"];
    bool factionVictory = battleEndEvent["isVictory"];
    if (auto p = currentUser.lock(); p != nullptr) {
        Faction userFaction = p->faction;
        patchs::ICore::CallInMainThread([=]() {
            bool isVictory = faction == userFaction ? factionVictory : false;
            NLOG_INFO(std::format("server determine that the game is over: isVictory = {}", isVictory));
            if (isVictory) {
                showToast->operator()(json{{"text", "you win the battle!"}, {"type", "success"}});
            } else {
                showToast->operator()(json{{"text", "you loose the battle!"}});
            }
            MatchManager->get<std::string, detail::Closure<void(bool)>>("onBattleEnded")(isVictory);
        });
    } else {
        NLOG_INFO(std::format("server determine that the game is over: {}", battleEndEvent.dump(-1)));
    }
}

static void onOpponentJoined(Nakama::NUserPresence presence) {
    if (auto p = currentUser.lock(); p != nullptr) {
        p->state = State::IN_BATTLE;
    }
    auto message = std::format("new player<{}> joined", presence.username);
    NLOG_INFO(message);

    patchs::ICore::CallInMainThread([message]() { showToast->operator()(json{{"text", message}}); });
}

static void onOpponentReady(Nakama::NUserPresence presence) {
    auto p = players.find(presence.sessionId);
    if (p == players.end()) {
        NLOG_ERROR(std::format("unknown player<{}> set state to ready", presence.username));
        return;
    }

    std::string message = std::format("player<{}> set state to ready", presence.username);
    patchs::ICore::CallInMainThread([message]() {
        NLOG_INFO(message);
        showToast->operator()(json{{"text", message}, {"type", "success"}});
    });
}

static void onMatchStart(Nakama::NUserPresence presence, json payload) {
    patchs::ICore::CallInMainThread([payload]() {
        NLOG_INFO("Start Match!");
        NLOG_INFO(payload.dump(2));
        json::object_t presenceFactions = payload["presenceFactions"];
        for (auto &[sessionId, p] : players) {
            auto item = presenceFactions.find(sessionId);
            if (item != presenceFactions.end()) {
                p->faction = item->second;
            }
        }

        if (auto p = currentUser.lock(); p != nullptr) {
            std::cout << std::format("currentUser<{}>.faction={}", p->presence.sessionId, (int)p->faction) << std::endl;
        }

        ;
        MatchManager->get<std::string, detail::Closure<void(SQObjectPtr)>>(std::string("onMatchStart"))(
            sqext_json::json_loads_impl(payload));
    });
}

static void onTactical__skill_use(Nakama::NUserPresence presence, json payload) {
    patchs::ICore::CallInMainThread([=]() {
        NLOG_INFO(std::format("{} use skill!", presence.username));
        Callbacks->get<std::string, detail::Closure<void(json, json, json, json)>>("remote_on_use")(
            payload[0], payload[1], payload[2], payload[3]);
    });
}

static void onTactical__travel(Nakama::NUserPresence presence, json payload) {
    patchs::ICore::CallInMainThread([=]() {
        NLOG_INFO(std::format("{} move entity!", presence.username));
        Callbacks->get<std::string, detail::Closure<void(json, json)>>("remote_on_travel")(payload[0], payload[1]);
    });
}

static void onTactical__wait_turn(Nakama::NUserPresence presence, json payload) {
    patchs::ICore::CallInMainThread([=]() {
        NLOG_INFO(std::format("{} wait for next turn!", presence.username));
        Callbacks->get<std::string, detail::Closure<void(json)>>("remote_on_entityWaitTurn")(payload[0]);
    });
}

static void onTactical__end_turn(Nakama::NUserPresence presence, json payload) {
    patchs::ICore::CallInMainThread([=]() {
        NLOG_INFO(std::format("{} end this turn!", presence.username));
        Callbacks->get<std::string, detail::Closure<void(json, json)>>("remote_on_onTurnEnd")(payload[0], payload[1]);
    });
}

static void onSyncSeed(Nakama::NUserPresence presence, json payload) {
    auto seed = (int)payload[0];
    patchs::ICore::CallInMainThread([=]() {
        NLOG_INFO(std::format("{} broadcast an random seed {}!", presence.username, seed));
        sqext_random::seedRandom(seed);
    });
}

void setup_listener() {
    // 注册处理玩家进出的回调
    listener.setMatchPresenceCallback([](const Nakama::NMatchPresenceEvent &matchPresence) {
        for (Nakama::NUserPresence presence : matchPresence.joins) {
            // Spawn a player for this presence and store it in the dictionary by
            // session id.
            if (players.find(presence.sessionId) == players.end()) {
                auto gameObject = spawn_player(presence);
                players.insert({presence.sessionId, gameObject});
                onOpponentJoined(presence);
            }
        }

        // For each player that has left in this event...
        for (Nakama::NUserPresence presence : matchPresence.leaves) {
            // Remove the player from the game if they've been spawned
            if (auto p = players.find(presence.sessionId); p != players.end()) {
                players.erase(p);
            }
        }
    });
    listener.setMatchmakerMatchedCallback([](Nakama::NMatchmakerMatchedPtr matched) {
        // joinMatchByTokenAsync(matched->token).get();
        patchs::ICore::CallInMainThread(
            []() { Callbacks->get<std::string, detail::Closure<void()>>("onMatchmakerMatched")(); });
    });

    // 处理比赛事件的回调
    listener.setMatchDataCallback([](const Nakama::NMatchData &matchData) {
        switch ((MatchOpCode)matchData.opCode) {
        case MatchOpCode::MATCH_TERMINATE:
            onMatchMatchTerminate();
            break;
        // 服务器事件
        case MatchOpCode::MATCH_FACTION_DISPATCH:
            onMatchFactionDispatch(json::parse(matchData.data));
            break;
        case MatchOpCode::BATTLE_END:
            onBattleEnd(json::parse(matchData.data));
            break;

        case MatchOpCode::OPPONENT_READY:
            onOpponentReady(matchData.presence);
            break;
        case MatchOpCode::MATCH_START:
            onMatchStart(matchData.presence, json::parse(matchData.data));
            break;

        // 对局关键帧
        case MatchOpCode::TACTICAL__SKILL_USE:
            onTactical__skill_use(matchData.presence, json::parse(matchData.data));
            break;
        case MatchOpCode::TACTICAL__TRAVEL:
            onTactical__travel(matchData.presence, json::parse(matchData.data));
            break;
        case MatchOpCode::TACTICAL__ENTITY_WAIT_TURN:
            onTactical__wait_turn(matchData.presence, json::parse(matchData.data));
            break;
        case MatchOpCode::TACTICAL__ENTITY_TURN_END:
            onTactical__end_turn(matchData.presence, json::parse(matchData.data));
            break;
        case MatchOpCode::SYNC_SEED:
            onSyncSeed(matchData.presence, json::parse(matchData.data));
            break;

        default:
            break;
        }
    });
}
} // namespace matches
} // namespace nakama
