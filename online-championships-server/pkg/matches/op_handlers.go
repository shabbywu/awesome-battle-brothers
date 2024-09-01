package matches

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/samber/lo"
)

var handlers = map[int64]func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, state interface{}, message runtime.MatchData){}

func broadcastMessageToAllOtherPresences(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, state interface{}, message runtime.MatchData) {
	mState, _ := state.(*MatchState)

	var receivers []runtime.Presence

	for _, presence := range mState.presences {
		if presence.GetSessionId() != message.GetSessionId() {
			receivers = append(receivers, presence)
		}
	}

	if err := dispatcher.BroadcastMessage(message.GetOpCode(), message.GetData(), receivers, message, message.GetReliable()); err != nil {
		logger.Error("erro in broadcasting message: %s", err.Error())
	}
}

func onOpponentReady(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, state interface{}, message runtime.MatchData) {
	mState, _ := state.(*MatchState)
	if p, ok := mState.presences[message.GetSessionId()]; ok {

		payload := message.GetData()
		var roster map[string]interface{}
		if err := json.Unmarshal([]byte(payload), &roster); err != nil {
			logger.Error("unable to unmarshal roster: %v", err)
			return
		}

		p.Roster = roster
		p.IsReady = true
		if err := dispatcher.BroadcastMessage(message.GetOpCode(), []byte{}, nil, message, true); err != nil {
			logger.Error("err in broadcasting message: %", err)
		}
	}

	// check if all player ready
	if len(mState.factions) == 2 {
		allReady := lo.EveryBy(lo.ToPairs(mState.factions), func(pair lo.Entry[int, *Presence]) bool {
			return pair.Value.IsReady && pair.Value.Roster != nil
		})
		if allReady {
			mState.battleStarted = true
			startBattle(logger, dispatcher, state)
		}
	}
}

func onBattleEnd(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, state interface{}, message runtime.MatchData) {
	mState, _ := state.(*MatchState)
	payload := message.GetData()
	var event BattleEndEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		logger.Error("unable to unmarshal roster: %v", err)
		return
	}
	if _, ok := mState.factions[event.Faction]; ok {
		mState.battleResult[event.Faction] = event.IsVictory
	}
	if len(mState.battleResult) == mState.totalFaction {
		mState.battleEnd = true
	}
}

func startBattle(logger runtime.Logger, dispatcher runtime.MatchDispatcher, state interface{}) {
	mState, _ := state.(*MatchState)
	mState.totalFaction = len(mState.factions)
	event := StartMatchEvent{
		PresenceFactions: map[string]int{},
	}
	for faction, p := range mState.factions {
		event.PresenceFactions[p.GetSessionId()] = faction
		if faction == FactionRedSide {
			event.RedSideRoster = p.Roster
		} else if faction == FactionBlueSide {
			event.BlueSideRoster = p.Roster
		}
	}
	event.StrategicProperties = mState.strategicProperties
	if err := dispatcher.BroadcastMessage(OpCodeMatchStart, DumpEvent(event), nil, nil, true); err != nil {
		logger.Error("err in broadcasting message: %v", err)
	}
}

func init() {
	handlers[OpCodeOpponentReady] = onOpponentReady
	handlers[OpCodeBattleEnd] = onBattleEnd
}
