#pragma once
#include "../vm.hpp"
#include "nakama/client.hpp"
#include "nakama/matches.hpp"
#include "nakama/storage.hpp"


namespace online_battle {
    void static sqext_register_online_battle_impl(sqbind17::detail::VM& v) {
        vm = &v;
        auto lib_table = sqbind17::detail::Table(*vm);
        auto api = sqbind17::detail::Table(*vm);
        lib_table.set(std::string("api"), api);

        api.bindFunc("login", &nakama::login);
        api.bindFunc("getAccount", &nakama::getAccount);
        api.bindFunc("isAuthed", &nakama::isAuthed);
        // matches
        api.bindFunc("createMatch", &nakama::matches::createMatch);
        api.bindFunc("listMatches", &nakama::matches::listMatches);
        api.bindFunc("joinMatch", &nakama::matches::joinMatch);
        api.bindFunc("leaveMatch", &nakama::matches::leaveMatch);
        api.bindFunc("setReady", &nakama::matches::setReady);
        api.bindFunc("emitEvent", &nakama::matches::emitEvent);
        api.bindFunc("getPlayerFaction", &nakama::matches::getPlayerFaction);
        // storages
        api.bindFunc("loadPlayerRoster", &nakama::storages::loadPlayerRoster);
        api.bindFunc("savePlayerRoster", &nakama::storages::savePlayerRoster);

        api.bindFunc("setupsqbind17", [](){
            auto roottable = detail::Table(_table(vm->roottable()), *vm);
            auto ToastifyScreen = roottable.get<std::string, detail::Table>("ToastifyScreen");
            auto showToast = ToastifyScreen.get<std::string, detail::Closure<void(json)>>("showToast");
            nakama::showToast = showToast.to_ptr();
            nakama::matches::setupMatchBinding();
        });
        api.bindFunc("nextTicket", [](detail::Closure<void()> closure){
            auto ptr = closure.to_ptr();
            patchs::ICore::CallInMainThreadHighPriority([ptr](){
                ptr->operator()(); }); });

        auto MatchOpCode = sqbind17::detail::Table(*vm);
        MatchOpCode.set(std::string("TACTICAL__SKILL_USE"), (int)nakama::matches::MatchOpCode::TACTICAL__SKILL_USE);
        MatchOpCode.set(std::string("TACTICAL__TRAVEL"), (int)nakama::matches::MatchOpCode::TACTICAL__TRAVEL);
        MatchOpCode.set(std::string("TACTICAL__ENTITY_WAIT_TURN"), (int)nakama::matches::MatchOpCode::TACTICAL__ENTITY_WAIT_TURN);
        MatchOpCode.set(std::string("TACTICAL__ENTITY_TURN_END"), (int)nakama::matches::MatchOpCode::TACTICAL__ENTITY_TURN_END);
        MatchOpCode.set(std::string("SYNC_SEED"), (int)nakama::matches::MatchOpCode::SYNC_SEED);
        MatchOpCode.set(std::string("BATTLE_END"), (int)nakama::matches::MatchOpCode::BATTLE_END);
        lib_table.set(std::string("MatchOpCode"), MatchOpCode);

        auto callbacks = sqbind17::detail::Table(*vm);
        lib_table.set(std::string("callbacks"), callbacks);

        auto roottable = sqbind17::detail::Table(_table(vm->roottable()), *vm);
        roottable.set(std::string("online_battle_module"), lib_table);
    }
}
