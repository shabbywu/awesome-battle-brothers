package matches

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/heroiclabs/nakama-common/runtime"
	"online-championships/pkg/matches/strategicProperties"
	"strconv"
)

type MatchState struct {
	factions                map[int]*Presence
	presences               map[string]*Presence
	totalFaction            int
	battleResult            map[int]bool
	orderedPresences        []string
	battleStarted           bool
	battleEnd               bool
	battleEndEventBroadcast bool
	strategicProperties     strategicProperties.StrategicProperties
	features                map[int]bool
}

type Match struct{}

func NewMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (m runtime.Match, err error) {
	return &Match{}, nil
}

func (m *Match) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	state := &MatchState{
		factions:     map[int]*Presence{},
		presences:    map[string]*Presence{},
		battleResult: map[int]bool{},
		features:     map[int]bool{},
	}

	createdParams := MatchCreateParams{}
	_ = json.Unmarshal([]byte(params["createdParams"].(string)), &createdParams)
	state.strategicProperties = strategicProperties.BuildStrategicProperties(createdParams.Scenario, createdParams.MapSeed, createdParams.CombatSeed)

	tickRate := 10
	name := "EmptyRoom"
	if createdParams.Name != "" {
		name = createdParams.Name
		state.features[FeatureCustomMatchName] = true
	}
	description := createdParams.Description
	label := fmt.Sprintf(`{"name": "%s", "%s": ""}`, name, description)
	return state, tickRate, label
}

func (m *Match) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	acceptUser := true
	return state, acceptUser, ""
}

func (m *Match) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	mState, _ := state.(*MatchState)

	for _, p := range presences {
		mState.presences[p.GetSessionId()] = spawnPlayer(p)
		mState.orderedPresences = append(mState.orderedPresences, p.GetSessionId())
	}
	return mState
}

func (m *Match) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	mState, _ := state.(*MatchState)

	combatResult := CombatResultNone
	for _, p := range presences {
		o := mState.presences[p.GetSessionId()]
		if o.Faction != FactionNone {
			if mState.battleStarted {
				combatResult = convertMatchResult(combatResult, o.Faction)
			}
		}
		delete(mState.presences, p.GetSessionId())
	}

	// rebuild orderedPresences
	var orderedPresences []string
	for _, o := range mState.orderedPresences {
		if _, ok := mState.presences[o]; ok {
			orderedPresences = append(orderedPresences, o)
		}
	}
	mState.orderedPresences = orderedPresences

	if mState.battleStarted && !mState.battleEnd {
		switch combatResult {
		case CombatResultBlueSideWin:
			mState.battleResult[FactionBlueSide] = true
			mState.battleResult[FactionRedSide] = false
			mState.battleEnd = true
		case CombatResultRedSideWin:
			mState.battleResult[FactionBlueSide] = false
			mState.battleResult[FactionRedSide] = true
			mState.battleEnd = true
		case CombatResultDraw:
			mState.battleResult[FactionBlueSide] = false
			mState.battleResult[FactionRedSide] = false
			mState.battleEnd = true
		default:
			break
		}
	}
	if len(mState.presences) == 0 {
		return nil
	}
	return mState
}

func (m *Match) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	mState, _ := state.(*MatchState)

	for _, message := range messages {
		if handle, ok := handlers[message.GetOpCode()]; ok {
			handle(ctx, logger, db, nk, dispatcher, state, message)
		} else {
			broadcastMessageToAllOtherPresences(ctx, logger, db, nk, dispatcher, state, message)
		}
	}

	if !mState.battleStarted {
		m.balanceFaction(logger, state, dispatcher)
	} else if len(mState.battleResult) == mState.totalFaction && !mState.battleEndEventBroadcast {
		m.broadcastBattleResult(logger, state, dispatcher)
	}

	return mState
}

func (m *Match) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	message := "Server shutting down in " + strconv.Itoa(graceSeconds) + " seconds."
	// _ = dispatcher.BroadcastMessage(OpCodeMatchTerminate, []byte(message), nil, nil, true)
	logger.Info(message)
	return state
}

func (m *Match) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	logger.Info("signal received: " + data)
	return state, "signal received: " + data
}

func (m *Match) balanceFaction(logger runtime.Logger, state interface{}, dispatcher runtime.MatchDispatcher) {
	mState, _ := state.(*MatchState)
	if len(mState.factions) == 2 {
		return
	}

	var players []string
	for _, faction := range []int{FactionRedSide, FactionBlueSide} {
		if p, used := mState.factions[faction]; used {
			players = append(players, p.GetUsername())
			continue
		}
		for _, sessionId := range mState.orderedPresences {
			p := mState.presences[sessionId]
			if p.Faction == FactionNone {
				p.Faction = faction
				mState.factions[faction] = p
				players = append(players, p.GetUsername())

				// broadcast balance event
				err := dispatcher.BroadcastMessageDeferred(OpCodeMatchFactionDispatch, DumpEvent(FactionDispatchEvent{
					Presence: sessionId,
					Faction:  faction,
				}), nil, nil, true)
				if err != nil {
					logger.Error("broadcast error: %v", err)
				}
			}
		}
	}
	if !mState.features[FeatureCustomMatchName] {
		var name string
		switch len(players) {
		case 1:
			name = fmt.Sprintf("vs %s", players[0])
			break
		case 2:
			name = fmt.Sprintf("%s vs %s", players[0], players[1])
			break
		default:
			name = "EmptyRoom"
		}
		label := fmt.Sprintf(`{"name": "%s", "description": ""}`, name)
		if err := dispatcher.MatchLabelUpdate(label); err != nil {
			logger.Error("MatchLabelUpdate error: %v", err)
		}
	}
}

func (m *Match) broadcastBattleResult(logger runtime.Logger, state interface{}, dispatcher runtime.MatchDispatcher) {
	mState, _ := state.(*MatchState)
	mState.battleEndEventBroadcast = true
	var event BattleEndEvent
	for f, IsVictory := range mState.battleResult {
		if IsVictory {
			event.IsVictory = IsVictory
			event.Faction = f
		}
	}

	if err := dispatcher.BroadcastMessage(OpCodeBattleEnd, DumpEvent(event), nil, nil, true); err != nil {
		logger.Error("broadcast error: %v", err)
	}
}
