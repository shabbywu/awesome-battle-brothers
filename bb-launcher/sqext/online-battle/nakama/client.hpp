#pragma once
#include <memory>
#include <nlohmann/json.hpp>
#define NLOGS_ENABLED
#include "nakama-cpp/Nakama.h"

using namespace sqbind17;
using json = nlohmann::json;

namespace patchs {
namespace ICore {
void CallInMainThread(std::function<void(void)> callback, bool once = true);
void CallInMainThreadHighPriority(std::function<void(void)> callback, bool once = true);
} // namespace ICore
} // namespace patchs

namespace nakama {
static Nakama::NClientPtr client = nullptr;
static Nakama::NRtClientPtr rtClient = nullptr;
static Nakama::NRtTransportPtr transport = nullptr;
static Nakama::NSessionPtr session = nullptr;
static Nakama::NRtDefaultClientListener listener;
static bool done = false;
static bool connectionClosed = false;

static std::shared_ptr<detail::Closure<void(json)>> showToast = nullptr;

namespace matches {
void setup_listener();
}

void setup_client() {
    Nakama::NLogger::initWithConsoleSink(Nakama::NLogLevel::Info);
    Nakama::NClientParameters parameters;
    parameters.serverKey = NAKAMA_SERVERKEY;
    parameters.host = NAKAMA_HOST;
    parameters.port = Nakama::DEFAULT_PORT;
    client = Nakama::createDefaultClient(parameters);
}

void fetch_tick() {
    while (!done) {
        if (client)
            client->tick();
        if (rtClient) {
            // if (!rtClient->isConnected() && !rtClient->isConnecting()) {
            //     rtClient->connectAsync(session, true,
            //     Nakama::NRtClientProtocol::Json);
            // }
            rtClient->tick();
        }

        std::this_thread::sleep_for(std::chrono::milliseconds(50));
    }
}

bool restoreSession() {
    // TODO: try restore session from disk
    return false;
}

static void dumpSession() {
    // TODO: dump session to disk
}

void login(std::string username, std::string password, std::string nickname = "") {
    NLOG_INFO("Authenticating...");
    bool create = nickname != "";
    auto email = username + "@nobody.com";
    session = client->authenticateEmailAsync(email, password, username, create, {}).get();
    if (nickname != "")
        client->updateAccountAsync(session, username, nickname);
    rtClient = client->createRtClient();
    rtClient->setListener(&listener);
    matches::setup_listener();
    rtClient->connectAsync(session, true, Nakama::NRtClientProtocol::Json).get();

    dumpSession();
}

json getAccount() {
    auto account = client->getAccountAsync(session).get();
    json::object_t json_object = json::object();
    auto email = account.email;
    auto username = email.substr(0, email.find_last_of("@"));
    json_object["username"] = username;
    json_object["nickname"] = account.user.displayName;
    return json_object;
}

bool isAuthed() {
    return session != nullptr;
}
} // namespace nakama
