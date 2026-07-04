package matches

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/samber/lo"
	"online-championships/pkg/protocol"
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
	if mState.roomRuntime != nil {
		logger.Info("legacy opponent ready ignored for v1 room")
		return
	}
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
	if mState.roomRuntime != nil {
		logger.Info("legacy battle end ignored for v1 room")
		return
	}
	senderFaction := factionForSession(mState, message.GetSessionId())
	if senderFaction == FactionNone {
		logger.Info("battle end ignored for unseated session %s", message.GetSessionId())
		return
	}
	payload := message.GetData()
	var event BattleEndEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		logger.Error("unable to unmarshal roster: %v", err)
		return
	}
	if event.Faction != senderFaction {
		logger.Info("battle end ignored for mismatched faction: sender=%d event=%d", senderFaction, event.Faction)
		return
	}
	if event.Faction == FactionRedSide || event.Faction == FactionBlueSide {
		mState.battleResult[event.Faction] = event.IsVictory
	}
	if mState.totalFaction == 0 && mState.roomRuntime != nil {
		mState.totalFaction = 2
	}
	if len(mState.battleResult) == mState.totalFaction {
		mState.battleEnd = true
	}
}

func onOpponentForceQuit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, state interface{}, message runtime.MatchData) {
	mState, _ := state.(*MatchState)
	if mState.roomRuntime != nil {
		logger.Info("legacy force quit ignored for v1 room")
		return
	}
	loserFaction := factionForSession(mState, message.GetSessionId())
	winnerFaction := oppositeFaction(loserFaction)
	if loserFaction == FactionNone || winnerFaction == FactionNone {
		logger.Info("force quit ignored for unseated session %s", message.GetSessionId())
		return
	}

	mState.battleStarted = true
	mState.battleEnd = true
	mState.battleEndEventBroadcast = true
	mState.totalFaction = 2
	mState.battleResult[loserFaction] = false
	mState.battleResult[winnerFaction] = true
	if mState.roomRuntime != nil {
		mState.roomRuntime.Status = protocol.RoomStatusClosed
	}

	if err := dispatcher.BroadcastMessage(OpCodeBattleEnd, DumpEvent(BattleEndEvent{
		Faction:   winnerFaction,
		IsVictory: true,
	}), nil, message, true); err != nil {
		logger.Error("err in broadcasting force quit battle result: %v", err)
	}
}

func startBattle(logger runtime.Logger, dispatcher runtime.MatchDispatcher, state interface{}) {
	mState, _ := state.(*MatchState)
	if mState.roomRuntime != nil {
		logger.Info("legacy start battle ignored for v1 room")
		return
	}
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

func factionForSession(mState *MatchState, sessionID string) int {
	if mState.roomRuntime != nil {
		if seat, ok := mState.roomRuntime.SeatForSession(sessionID); ok {
			return factionForSeat(seat)
		}
	}
	if p := mState.presences[sessionID]; p != nil {
		return p.Faction
	}
	return FactionNone
}

func factionForSeat(seat protocol.Seat) int {
	switch seat {
	case protocol.SeatRed:
		return FactionRedSide
	case protocol.SeatBlue:
		return FactionBlueSide
	default:
		return FactionNone
	}
}

func oppositeFaction(faction int) int {
	switch faction {
	case FactionRedSide:
		return FactionBlueSide
	case FactionBlueSide:
		return FactionRedSide
	default:
		return FactionNone
	}
}

func init() {
	handlers[OpCodeOpponentReady] = onOpponentReady
	handlers[OpCodeBattleEnd] = onBattleEnd
	handlers[OpCodeOpponentForceQuit] = onOpponentForceQuit
}
