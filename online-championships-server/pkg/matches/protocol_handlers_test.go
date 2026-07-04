package matches

import (
	"context"
	"encoding/json"
	"online-championships/pkg/matches/strategicProperties"
	"online-championships/pkg/protocol"
	"strings"
	"testing"

	"github.com/heroiclabs/nakama-common/runtime"
)

func TestMatchLifecycleAssignsV1Seats(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state := newProtocolTestState()
	red := testPresence{userID: "u-red", sessionID: "s-red", username: "red"}
	blue := testPresence{userID: "u-blue", sessionID: "s-blue", username: "blue"}

	var accepted bool
	stateAny, accepted, reason := match.MatchJoinAttempt(context.Background(), logger, nil, nil, dispatcher, 0, state, red, map[string]string{"seat": "red", "matchId": "match-1"})
	if !accepted {
		t.Fatalf("red join rejected: %s", reason)
	}
	state = stateAny.(*MatchState)
	state = match.MatchJoin(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.Presence{red}).(*MatchState)

	stateAny, accepted, reason = match.MatchJoinAttempt(context.Background(), logger, nil, nil, dispatcher, 0, state, blue, map[string]string{"seat": "blue", "matchId": "match-1"})
	if !accepted {
		t.Fatalf("blue join rejected: %s", reason)
	}
	state = stateAny.(*MatchState)
	state = match.MatchJoin(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.Presence{blue}).(*MatchState)

	if seat, ok := state.roomRuntime.SeatForSession("s-red"); !ok || seat != protocol.SeatRed {
		t.Fatalf("red assigned seat = %q, ok=%v", seat, ok)
	}
	if seat, ok := state.roomRuntime.SeatForSession("s-blue"); !ok || seat != protocol.SeatBlue {
		t.Fatalf("blue assigned seat = %q, ok=%v", seat, ok)
	}
	if dispatcher.countDeferred(protocol.OpCodeSeatAssigned) != 2 {
		t.Fatalf("seat assignment broadcasts = %d, want 2", dispatcher.countDeferred(protocol.OpCodeSeatAssigned))
	}
	if dispatcher.countBroadcast(protocol.OpCodeRoomState) != 2 {
		t.Fatalf("room state broadcasts = %d, want 2", dispatcher.countBroadcast(protocol.OpCodeRoomState))
	}
	roomState := dispatcher.lastBroadcast(protocol.OpCodeRoomState)
	if roomState == nil {
		t.Fatalf("room state was not broadcast")
	}
	var stateMessage protocol.RoomState
	if err := json.Unmarshal(roomState.data, &stateMessage); err != nil {
		t.Fatalf("decode room state: %v", err)
	}
	if stateMessage.RoomID != "match-1" {
		t.Fatalf("room state room id = %q, want match-1", stateMessage.RoomID)
	}
}

func TestMatchJoinAttemptDoesNotMutateRoomRuntimeBeforeJoin(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state := newProtocolTestState()
	red := testPresence{userID: "u-red", sessionID: "s-red", username: "red"}

	stateAny, accepted, reason := match.MatchJoinAttempt(context.Background(), logger, nil, nil, dispatcher, 1, state, red, map[string]string{"seat": "red", "matchId": "match-1"})
	if !accepted {
		t.Fatalf("red join rejected: %s", reason)
	}
	state = stateAny.(*MatchState)
	if _, ok := state.roomRuntime.Seats[protocol.SeatRed]; ok {
		t.Fatalf("join attempt should not mutate room runtime seats")
	}
	if _, ok := state.joinReservations["s-red"]; !ok {
		t.Fatalf("join attempt should create reservation")
	}

	state = match.MatchJoin(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.Presence{red}).(*MatchState)
	if seat, ok := state.roomRuntime.SeatForSession("s-red"); !ok || seat != protocol.SeatRed {
		t.Fatalf("final join assigned seat = %q ok=%v, want red true", seat, ok)
	}
	if _, ok := state.joinReservations["s-red"]; ok {
		t.Fatalf("final join should clear reservation")
	}
}

func TestMatchJoinAttemptReservationBlocksDuplicateSeatUntilExpired(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state := newProtocolTestState()
	red := testPresence{userID: "u-red", sessionID: "s-red", username: "red"}
	other := testPresence{userID: "u-other", sessionID: "s-other", username: "other"}

	stateAny, accepted, reason := match.MatchJoinAttempt(context.Background(), logger, nil, nil, dispatcher, 1, state, red, map[string]string{"seat": "red", "matchId": "match-1"})
	if !accepted {
		t.Fatalf("red join rejected: %s", reason)
	}
	state = stateAny.(*MatchState)
	stateAny, accepted, reason = match.MatchJoinAttempt(context.Background(), logger, nil, nil, dispatcher, 2, state, other, map[string]string{"seat": "red", "matchId": "match-1"})
	if accepted {
		t.Fatalf("duplicate red join should be rejected before reservation expiry")
	}
	if reason == "" {
		t.Fatalf("duplicate red join should include rejection reason")
	}

	state = stateAny.(*MatchState)
	stateAny, accepted, reason = match.MatchJoinAttempt(context.Background(), logger, nil, nil, dispatcher, 1+DefaultJoinReservationTimeoutTicks, state, other, map[string]string{"seat": "red", "matchId": "match-1"})
	if !accepted {
		t.Fatalf("red join after reservation expiry rejected: %s", reason)
	}
	state = stateAny.(*MatchState)
	if _, ok := state.joinReservations["s-red"]; ok {
		t.Fatalf("expired reservation should be cleared")
	}
	if _, ok := state.roomRuntime.Seats[protocol.SeatRed]; ok {
		t.Fatalf("expired reservation should not have mutated runtime seat")
	}
}

func TestMatchJoinAttemptSpectatorReservationCountsTowardLimit(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state := newProtocolTestState()
	state.roomRuntime.Status = protocol.RoomStatusInBattle
	metadata := state.roomRuntime.Metadata
	metadata.Spectators.Enabled = true
	metadata.Spectators.Limit = 1
	state.roomRuntime.Metadata = metadata
	spec1 := testPresence{userID: "u-spec-1", sessionID: "s-spec-1", username: "spec1"}
	spec2 := testPresence{userID: "u-spec-2", sessionID: "s-spec-2", username: "spec2"}

	stateAny, accepted, reason := match.MatchJoinAttempt(context.Background(), logger, nil, nil, dispatcher, 1, state, spec1, map[string]string{"seat": "spectator", "matchId": "match-1"})
	if !accepted {
		t.Fatalf("first spectator join rejected: %s", reason)
	}
	state = stateAny.(*MatchState)
	if len(state.roomRuntime.Spectators) != 0 {
		t.Fatalf("spectator join attempt should not mutate runtime spectators")
	}
	stateAny, accepted, reason = match.MatchJoinAttempt(context.Background(), logger, nil, nil, dispatcher, 2, state, spec2, map[string]string{"seat": "spectator", "matchId": "match-1"})
	if accepted {
		t.Fatalf("second spectator should be rejected while first reservation holds limit")
	}
	if reason == "" {
		t.Fatalf("second spectator rejection should include reason")
	}

	state = stateAny.(*MatchState)
	stateAny, accepted, reason = match.MatchJoinAttempt(context.Background(), logger, nil, nil, dispatcher, 1+DefaultJoinReservationTimeoutTicks, state, spec2, map[string]string{"seat": "spectator", "matchId": "match-1"})
	if !accepted {
		t.Fatalf("second spectator after reservation expiry rejected: %s", reason)
	}
	state = stateAny.(*MatchState)
	if _, ok := state.joinReservations["s-spec-1"]; ok {
		t.Fatalf("expired spectator reservation should be cleared")
	}
}

func TestMatchJoinAttemptReconnectDoesNotClearAbandonUntilFinalJoin(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)
	state = match.MatchLeave(context.Background(), logger, nil, nil, dispatcher, 10, state, []runtime.Presence{red}).(*MatchState)
	if state.roomRuntime.Seats[protocol.SeatRed].SessionID != "" {
		t.Fatalf("red session should be cleared after leave")
	}
	if _, ok := state.roomRuntime.AbandonDeadlines[protocol.SeatRed]; !ok {
		t.Fatalf("red abandon deadline should exist after in-battle leave")
	}

	redReconnect := testPresence{userID: "u-red", sessionID: "s-red-2", username: "red"}
	stateAny, accepted, reason := match.MatchJoinAttempt(context.Background(), logger, nil, nil, dispatcher, 11, state, redReconnect, map[string]string{"seat": "auto", "matchId": "match-1"})
	if !accepted {
		t.Fatalf("red reconnect join attempt rejected: %s", reason)
	}
	state = stateAny.(*MatchState)
	if state.roomRuntime.Seats[protocol.SeatRed].SessionID != "" {
		t.Fatalf("join attempt must not restore red session before final join")
	}
	if _, ok := state.roomRuntime.AbandonDeadlines[protocol.SeatRed]; !ok {
		t.Fatalf("join attempt must not clear abandon deadline before final join")
	}

	state = match.MatchJoin(context.Background(), logger, nil, nil, dispatcher, 11, state, []runtime.Presence{redReconnect}).(*MatchState)
	if state.roomRuntime.Seats[protocol.SeatRed].SessionID != "s-red-2" {
		t.Fatalf("final join did not restore red session: %+v", state.roomRuntime.Seats[protocol.SeatRed])
	}
	if _, ok := state.roomRuntime.AbandonDeadlines[protocol.SeatRed]; ok {
		t.Fatalf("final join should clear abandon deadline")
	}
}

func TestExpiredJoinReservationDoesNotFallbackToAutoSeat(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state := newProtocolTestState()
	red := testPresence{userID: "u-red", sessionID: "s-red", username: "red"}
	other := testPresence{userID: "u-other", sessionID: "s-other", username: "other"}

	stateAny, accepted, reason := match.MatchJoinAttempt(context.Background(), logger, nil, nil, dispatcher, 1, state, red, map[string]string{"seat": "red", "matchId": "match-1"})
	if !accepted {
		t.Fatalf("red join rejected: %s", reason)
	}
	state = stateAny.(*MatchState)
	stateAny, accepted, reason = match.MatchJoinAttempt(context.Background(), logger, nil, nil, dispatcher, 1+DefaultJoinReservationTimeoutTicks, state, other, map[string]string{"seat": "red", "matchId": "match-1"})
	if !accepted {
		t.Fatalf("other red join after reservation expiry rejected: %s", reason)
	}
	state = stateAny.(*MatchState)
	state = match.MatchJoin(context.Background(), logger, nil, nil, dispatcher, 1+DefaultJoinReservationTimeoutTicks, state, []runtime.Presence{other}).(*MatchState)
	if seat, ok := state.roomRuntime.SeatForSession("s-other"); !ok || seat != protocol.SeatRed {
		t.Fatalf("other final join assigned seat = %q ok=%v, want red true", seat, ok)
	}
	seatAssignments := dispatcher.countDeferred(protocol.OpCodeSeatAssigned)

	state = match.MatchJoin(context.Background(), logger, nil, nil, dispatcher, 1+DefaultJoinReservationTimeoutTicks, state, []runtime.Presence{red}).(*MatchState)
	if _, ok := state.roomRuntime.SeatForSession("s-red"); ok {
		t.Fatalf("expired red reservation must not be auto-assigned to another seat")
	}
	if _, ok := state.roomRuntime.Seats[protocol.SeatBlue]; ok {
		t.Fatalf("expired red reservation must not fall back to blue")
	}
	if _, ok := state.presences["s-red"]; ok {
		t.Fatalf("expired red reservation must not create presence")
	}
	if got := dispatcher.countDeferred(protocol.OpCodeSeatAssigned); got != seatAssignments {
		t.Fatalf("expired red reservation broadcast seat assignments = %d, want %d", got, seatAssignments)
	}
	if got := dispatcher.countKicked("s-red"); got != 1 {
		t.Fatalf("expired red reservation kicks = %d, want 1", got)
	}
}

func TestExpiredSpectatorReservationDoesNotFallbackToPlayerSeat(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state := newProtocolTestState()
	metadata := state.roomRuntime.Metadata
	metadata.Spectators.Enabled = true
	state.roomRuntime.Metadata = metadata
	spectator := testPresence{userID: "u-spec", sessionID: "s-spec", username: "spec"}

	stateAny, accepted, reason := match.MatchJoinAttempt(context.Background(), logger, nil, nil, dispatcher, 1, state, spectator, map[string]string{"seat": "spectator", "matchId": "match-1"})
	if !accepted {
		t.Fatalf("spectator join rejected: %s", reason)
	}
	state = stateAny.(*MatchState)
	state = match.MatchJoin(context.Background(), logger, nil, nil, dispatcher, 1+DefaultJoinReservationTimeoutTicks, state, []runtime.Presence{spectator}).(*MatchState)
	if _, ok := state.roomRuntime.SeatForSession("s-spec"); ok {
		t.Fatalf("expired spectator reservation must not be auto-assigned to a player seat")
	}
	if _, ok := state.roomRuntime.Spectators["s-spec"]; ok {
		t.Fatalf("expired spectator reservation must not create spectator presence")
	}
	if _, ok := state.presences["s-spec"]; ok {
		t.Fatalf("expired spectator reservation must not create match presence")
	}
	if got := dispatcher.countDeferred(protocol.OpCodeSeatAssigned); got != 0 {
		t.Fatalf("expired spectator reservation seat assignments = %d, want 0", got)
	}
	if got := dispatcher.countKicked("s-spec"); got != 1 {
		t.Fatalf("expired spectator reservation kicks = %d, want 1", got)
	}
}

func TestMatchJoinWithoutReservationDoesNotCreatePresence(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state := newProtocolTestState()
	red := testPresence{userID: "u-red", sessionID: "s-red", username: "red"}

	state = match.MatchJoin(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.Presence{red}).(*MatchState)
	if _, ok := state.roomRuntime.SeatForSession("s-red"); ok {
		t.Fatalf("join without reservation must not assign a seat")
	}
	if _, ok := state.presences["s-red"]; ok {
		t.Fatalf("join without reservation must not create presence")
	}
	if got := dispatcher.countDeferred(protocol.OpCodeSeatAssigned); got != 0 {
		t.Fatalf("join without reservation seat assignments = %d, want 0", got)
	}
	if got := dispatcher.countKicked("s-red"); got != 1 {
		t.Fatalf("join without reservation kicks = %d, want 1", got)
	}
}

func TestMatchInitPopulatesRoomMetadataVersionAndCreatedAt(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	stateAny, tickRate, labelText := match.MatchInit(context.Background(), logger, nil, nil, map[string]interface{}{
		"createdParams": string(DumpEvent(MatchCreateParams{
			Scenario:    strategicProperties.ScenarioArena,
			MapSeed:     101,
			CombatSeed:  202,
			Name:        "Arena Room",
			Description: "MVP room",
		})),
	})
	if tickRate != 10 {
		t.Fatalf("tick rate = %d, want 10", tickRate)
	}
	state := stateAny.(*MatchState)
	var label MatchLabel
	if err := json.Unmarshal([]byte(labelText), &label); err != nil {
		t.Fatalf("decode label: %v", err)
	}
	metadata := label.Metadata
	if metadata.SchemaVersion != protocol.SchemaVersion ||
		metadata.ProtocolVersion != protocol.ProtocolVersion ||
		metadata.ServerVersion != protocol.ServerVersion ||
		metadata.RuntimeVersion != protocol.RuntimeVersion ||
		metadata.ResultContractVersion != protocol.ResultContractVersion ||
		metadata.RankedPolicyVersion != protocol.RankedPolicyVersion ||
		metadata.RankedMode != protocol.RankedModeDisabled {
		t.Fatalf("metadata version fields are invalid: %+v", metadata)
	}
	if metadata.Name != "Arena Room" || metadata.Description != "MVP room" {
		t.Fatalf("metadata name/description = %q/%q", metadata.Name, metadata.Description)
	}
	if metadata.Scenario != strategicProperties.ScenarioArena || metadata.MapSeed != 101 || metadata.CombatSeed != 202 {
		t.Fatalf("metadata scenario/seeds are invalid: %+v", metadata)
	}
	if metadata.CreatedAt == "" || metadata.CreatedAt == "unknown" {
		t.Fatalf("metadata createdAt was not populated: %+v", metadata)
	}
	if !metadata.Spectators.Enabled || metadata.Spectators.Limit != 0 {
		t.Fatalf("default spectator metadata is invalid: %+v", metadata.Spectators)
	}
	assertTimeoutPolicy(t, metadata.TimeoutPolicy, DefaultCommandResultTimeoutTicks, DefaultAbandonGraceTicks, DefaultBattleEndReportTimeoutTicks)
	if state.roomRuntime.Metadata != metadata {
		t.Fatalf("runtime metadata does not match label metadata: runtime=%+v label=%+v", state.roomRuntime.Metadata, metadata)
	}
}

func TestMatchInitPopulatesRoomMetadataSpectatorPolicy(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	stateAny, _, labelText := match.MatchInit(context.Background(), logger, nil, nil, map[string]interface{}{
		"createdParams": string(DumpEvent(MatchCreateParams{
			Scenario:    strategicProperties.ScenarioArena,
			MapSeed:     101,
			CombatSeed:  202,
			Name:        "Closed Room",
			Description: "No spectators",
			Spectators: &protocol.RoomSpectatorConfig{
				Enabled: false,
				Limit:   2,
			},
		})),
	})
	state := stateAny.(*MatchState)
	var label MatchLabel
	if err := json.Unmarshal([]byte(labelText), &label); err != nil {
		t.Fatalf("decode label: %v", err)
	}
	if label.Metadata.Spectators.Enabled || label.Metadata.Spectators.Limit != 2 {
		t.Fatalf("metadata spectator policy = %+v, want disabled limit 2", label.Metadata.Spectators)
	}
	if state.roomRuntime.Metadata.Spectators != label.Metadata.Spectators {
		t.Fatalf("runtime spectator policy = %+v, want %+v", state.roomRuntime.Metadata.Spectators, label.Metadata.Spectators)
	}
}

func TestRosterLockStartsV1Match(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)

	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)

	if state.roomRuntime.Status != protocol.RoomStatusInBattle {
		t.Fatalf("room status = %q, want InBattle", state.roomRuntime.Status)
	}
	if state.roomRuntime.LastAcceptedHash == "" {
		t.Fatalf("start hash was not initialized")
	}
	start := dispatcher.lastBroadcast(protocol.OpCodeMatchStart)
	if start == nil {
		t.Fatalf("MATCH_START was not broadcast")
	}
	var snapshot protocol.MatchStartSnapshot
	if err := json.Unmarshal(start.data, &snapshot); err != nil {
		t.Fatalf("decode start snapshot: %v", err)
	}
	if snapshot.StartHash != state.roomRuntime.LastAcceptedHash {
		t.Fatalf("snapshot start hash = %q, want %q", snapshot.StartHash, state.roomRuntime.LastAcceptedHash)
	}
	if err := protocol.ValidateSHA256Digest("startHash", snapshot.StartHash); err != nil {
		t.Fatalf("start hash format is invalid: %v", err)
	}
	if snapshot.RoomID != "match-1" {
		t.Fatalf("snapshot room id = %q, want match-1", snapshot.RoomID)
	}
	if snapshot.MapSeed != 101 || snapshot.CombatSeed != 202 {
		t.Fatalf("snapshot seeds = %d/%d, want 101/202", snapshot.MapSeed, snapshot.CombatSeed)
	}
	if snapshot.RulesetHash != "rules" {
		t.Fatalf("snapshot ruleset hash = %q, want rules", snapshot.RulesetHash)
	}
	if snapshot.ExpectedActiveSeat != protocol.SeatRed || snapshot.ExpectedActiveNetID != "red:0" {
		t.Fatalf("unexpected active actor: %s %s", snapshot.ExpectedActiveSeat, snapshot.ExpectedActiveNetID)
	}
	roomState := dispatcher.lastBroadcast(protocol.OpCodeRoomState)
	if roomState == nil {
		t.Fatalf("ROOM_STATE was not broadcast after match start")
	}
	var stateMessage protocol.RoomState
	if err := json.Unmarshal(roomState.data, &stateMessage); err != nil {
		t.Fatalf("decode room state: %v", err)
	}
	if stateMessage.Status != protocol.RoomStatusInBattle ||
		stateMessage.LastAcceptedHash != snapshot.StartHash ||
		stateMessage.ExpectedActiveSeat != protocol.SeatRed ||
		stateMessage.ExpectedActiveNetID != "red:0" {
		t.Fatalf("unexpected room state after start: %+v", stateMessage)
	}
}

func TestMatchLoopEchoesAcceptedCommand(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)
	dispatcher.clear()

	cmd := protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandWaitTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         state.roomRuntime.LastAcceptedHash,
	}
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeCmdWaitTurn, data: DumpEvent(cmd)},
	}).(*MatchState)

	echo := dispatcher.lastBroadcast(protocol.OpCodeCmdWaitTurn)
	if echo == nil {
		t.Fatalf("accepted command was not echoed")
	}
	var accepted protocol.CommandEnvelope
	if err := json.Unmarshal(echo.data, &accepted); err != nil {
		t.Fatalf("decode accepted command: %v", err)
	}
	if accepted.ServerSeq != 1 {
		t.Fatalf("serverSeq = %d, want 1", accepted.ServerSeq)
	}
	if len(state.roomRuntime.PendingCheckpoints) != 1 {
		t.Fatalf("pending checkpoints = %d, want 1", len(state.roomRuntime.PendingCheckpoints))
	}
	roomState := dispatcher.lastBroadcast(protocol.OpCodeRoomState)
	if roomState == nil {
		t.Fatalf("ROOM_STATE was not broadcast after command accept")
	}
	var stateMessage protocol.RoomState
	if err := json.Unmarshal(roomState.data, &stateMessage); err != nil {
		t.Fatalf("decode room state: %v", err)
	}
	if stateMessage.PendingServerSeq != 1 || stateMessage.LastCompletedSeq != 0 || stateMessage.CommandLogLength != 1 {
		t.Fatalf("unexpected room state after command accept: %+v", stateMessage)
	}
}

func TestMatchLoopRejectsCommandForWrongMatchID(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)
	dispatcher.clear()

	cmd := protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "other-match",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandWaitTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         state.roomRuntime.LastAcceptedHash,
	}
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeCmdWaitTurn, data: DumpEvent(cmd)},
	}).(*MatchState)

	if dispatcher.countBroadcast(protocol.OpCodeCmdWaitTurn) != 0 {
		t.Fatalf("wrong match command should not be echoed")
	}
	errorMessage := dispatcher.lastBroadcast(protocol.OpCodeProtocolError)
	if errorMessage == nil {
		t.Fatalf("protocol error was not returned")
	}
	var event protocol.ProtocolErrorEvent
	if err := json.Unmarshal(errorMessage.data, &event); err != nil {
		t.Fatalf("decode protocol error: %v", err)
	}
	if event.Code != protocol.ErrorInvalidCommand || event.ClientCommandID != "client-1" {
		t.Fatalf("unexpected protocol error: %+v", event)
	}
	if len(state.roomRuntime.CommandLog) != 0 {
		t.Fatalf("wrong match command should not enter command log")
	}
}

func TestV1SingleRosterLockDoesNotEndRoom(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, _ := joinedProtocolRoom(t, match, logger, dispatcher)

	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
	}).(*MatchState)

	if state.roomRuntime.Status != protocol.RoomStatusOpen {
		t.Fatalf("single roster lock should keep room open, got %q", state.roomRuntime.Status)
	}
	if state.battleEnd || state.battleEndEventBroadcast {
		t.Fatalf("single roster lock should not trigger legacy battle end")
	}
	if result := dispatcher.lastBroadcast(OpCodeBattleEnd); result != nil {
		t.Fatalf("single roster lock should not broadcast legacy battle end")
	}
}

func TestMatchLoopSendsProtocolErrorForInactiveSeatCommand(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)
	dispatcher.clear()

	cmd := protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-blue-1",
		Seat:            protocol.SeatBlue,
		ActiveNetID:     "blue:0",
		Op:              protocol.CommandWaitTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         state.roomRuntime.LastAcceptedHash,
	}
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.MatchData{
		testMatchData{testPresence: blue, opCode: protocol.OpCodeCmdWaitTurn, data: DumpEvent(cmd)},
	}).(*MatchState)

	if dispatcher.countBroadcast(protocol.OpCodeCmdWaitTurn) != 0 {
		t.Fatalf("inactive seat command should not be echoed")
	}
	errorMessage := dispatcher.lastBroadcast(protocol.OpCodeProtocolError)
	if errorMessage == nil {
		t.Fatalf("protocol error was not returned")
	}
	var event protocol.ProtocolErrorEvent
	if err := json.Unmarshal(errorMessage.data, &event); err != nil {
		t.Fatalf("decode protocol error: %v", err)
	}
	if event.Code != protocol.ErrorInvalidSeat {
		t.Fatalf("protocol error code = %q, want invalid_seat", event.Code)
	}
	if event.OpCode != protocol.OpCodeCmdWaitTurn {
		t.Fatalf("protocol error opCode = %d, want %d", event.OpCode, protocol.OpCodeCmdWaitTurn)
	}
	if event.ClientCommandID != "client-blue-1" {
		t.Fatalf("protocol error clientCommandId = %q, want client-blue-1", event.ClientCommandID)
	}
	if len(errorMessage.targetSessionIDs) != 1 || errorMessage.targetSessionIDs[0] != "s-blue" {
		t.Fatalf("protocol error targets = %v, want [s-blue]", errorMessage.targetSessionIDs)
	}
}

func TestMatchLoopSendsProtocolErrorForUnknownOpcode(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, _ := joinedProtocolRoom(t, match, logger, dispatcher)

	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: 9999, data: []byte(`{}`)},
	}).(*MatchState)

	errorMessage := dispatcher.lastBroadcast(protocol.OpCodeProtocolError)
	if errorMessage == nil {
		t.Fatalf("protocol error was not returned")
	}
	var event protocol.ProtocolErrorEvent
	if err := json.Unmarshal(errorMessage.data, &event); err != nil {
		t.Fatalf("decode protocol error: %v", err)
	}
	if event.Code != protocol.ErrorUnknownOpcode {
		t.Fatalf("protocol error code = %q, want unknown_opcode", event.Code)
	}
	if event.OpCode != 9999 {
		t.Fatalf("protocol error opCode = %d, want 9999", event.OpCode)
	}
	if len(errorMessage.targetSessionIDs) != 1 || errorMessage.targetSessionIDs[0] != "s-red" {
		t.Fatalf("protocol error targets = %v, want [s-red]", errorMessage.targetSessionIDs)
	}
}

func TestMatchLoopBroadcastsCompletedCommandResult(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)
	dispatcher.clear()

	cmd := protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandEndTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         state.roomRuntime.LastAcceptedHash,
	}
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeCmdEndTurn, data: DumpEvent(cmd)},
	}).(*MatchState)
	dispatcher.clear()

	postHash := testDigest("b", 20)
	redResult := protocol.CommandResult{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ServerSeq:       1,
		Seat:            protocol.SeatRed,
		PostHash:        postHash,
		PostSummary:     testPostSummary(postHash, 1),
		NextActiveSeat:  protocol.SeatBlue,
		NextActiveNetID: "blue:0",
		Round:           1,
		Turn:            2,
	}
	blueResult := redResult
	blueResult.Seat = protocol.SeatBlue

	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 2, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeCmdResult, data: DumpEvent(redResult)},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeCmdResult, data: DumpEvent(blueResult)},
	}).(*MatchState)

	result := dispatcher.lastBroadcast(protocol.OpCodeCmdResult)
	if result == nil {
		t.Fatalf("completed command result was not broadcast")
	}
	var broadcast protocol.CommandResult
	if err := json.Unmarshal(result.data, &broadcast); err != nil {
		t.Fatalf("decode command result broadcast: %v", err)
	}
	if broadcast.PostHash != postHash || broadcast.NextActiveSeat != protocol.SeatBlue || broadcast.NextActiveNetID != "blue:0" {
		t.Fatalf("unexpected command result broadcast: %+v", broadcast)
	}
	if state.roomRuntime.LastAcceptedHash != postHash {
		t.Fatalf("last accepted hash = %q, want %q", state.roomRuntime.LastAcceptedHash, postHash)
	}
	if len(state.roomRuntime.PendingCheckpoints) != 0 {
		t.Fatalf("pending checkpoints = %d, want 0", len(state.roomRuntime.PendingCheckpoints))
	}
	roomState := dispatcher.lastBroadcast(protocol.OpCodeRoomState)
	if roomState == nil {
		t.Fatalf("ROOM_STATE was not broadcast after checkpoint completion")
	}
	var stateMessage protocol.RoomState
	if err := json.Unmarshal(roomState.data, &stateMessage); err != nil {
		t.Fatalf("decode room state: %v", err)
	}
	if stateMessage.LastCompletedSeq != 1 ||
		stateMessage.PendingServerSeq != 0 ||
		stateMessage.LastAcceptedHash != postHash ||
		stateMessage.ExpectedActiveSeat != protocol.SeatBlue ||
		stateMessage.ExpectedActiveNetID != "blue:0" {
		t.Fatalf("unexpected room state after checkpoint completion: %+v", stateMessage)
	}
}

func TestMatchLoopRejectsCommandResultForWrongMatchID(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)
	dispatcher.clear()

	cmd := protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandEndTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         state.roomRuntime.LastAcceptedHash,
	}
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeCmdEndTurn, data: DumpEvent(cmd)},
	}).(*MatchState)
	dispatcher.clear()

	postHash := testDigest("b", 20)
	result := protocol.CommandResult{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "other-match",
		ServerSeq:       1,
		Seat:            protocol.SeatRed,
		PostHash:        postHash,
		PostSummary:     testPostSummaryForMatch("other-match", postHash, 1),
		NextActiveSeat:  protocol.SeatBlue,
		NextActiveNetID: "blue:0",
		Round:           1,
		Turn:            2,
	}
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 2, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeCmdResult, data: DumpEvent(result)},
	}).(*MatchState)

	if dispatcher.countBroadcast(protocol.OpCodeCmdResult) != 0 {
		t.Fatalf("wrong match result should not be broadcast")
	}
	errorMessage := dispatcher.lastBroadcast(protocol.OpCodeProtocolError)
	if errorMessage == nil {
		t.Fatalf("protocol error was not returned")
	}
	var event protocol.ProtocolErrorEvent
	if err := json.Unmarshal(errorMessage.data, &event); err != nil {
		t.Fatalf("decode protocol error: %v", err)
	}
	if event.Code != protocol.ErrorInvalidCommand || event.ServerSeq != 1 {
		t.Fatalf("unexpected protocol error: %+v", event)
	}
	if len(state.roomRuntime.PendingCheckpoints[1].Results) != 0 {
		t.Fatalf("wrong match result should not enter checkpoint: %+v", state.roomRuntime.PendingCheckpoints[1].Results)
	}
}

func TestMatchLoopBroadcastsDesyncWithCheckpointEvidence(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)
	dispatcher.clear()

	cmd := protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandEndTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         state.roomRuntime.LastAcceptedHash,
	}
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeCmdEndTurn, data: DumpEvent(cmd)},
	}).(*MatchState)
	dispatcher.clear()

	redHash := testDigest("c", 21)
	blueHash := testDigest("d", 22)
	redResult := protocol.CommandResult{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ServerSeq:       1,
		Seat:            protocol.SeatRed,
		PostHash:        redHash,
		PostSummary:     testPostSummary(redHash, 1),
		NextActiveSeat:  protocol.SeatBlue,
		NextActiveNetID: "blue:0",
		Round:           1,
		Turn:            2,
	}
	blueResult := redResult
	blueResult.Seat = protocol.SeatBlue
	blueResult.PostHash = blueHash
	blueResult.PostSummary = testPostSummary(blueHash, 1)

	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 2, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeCmdResult, data: DumpEvent(redResult)},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeCmdResult, data: DumpEvent(blueResult)},
	}).(*MatchState)

	desync := dispatcher.lastBroadcast(protocol.OpCodeMatchDesync)
	if desync == nil {
		t.Fatalf("desync was not broadcast")
	}
	var event protocol.MatchDesyncEvent
	if err := json.Unmarshal(desync.data, &event); err != nil {
		t.Fatalf("decode desync event: %v", err)
	}
	if event.Checkpoint == nil {
		t.Fatalf("desync checkpoint evidence is missing")
	}
	if event.Checkpoint.RedResult.PostHash != redHash || event.Checkpoint.BlueResult.PostHash != blueHash {
		t.Fatalf("unexpected desync evidence: %+v", event.Checkpoint)
	}
	roomState := dispatcher.lastBroadcast(protocol.OpCodeRoomState)
	if roomState == nil {
		t.Fatalf("ROOM_STATE was not broadcast after desync")
	}
	var stateMessage protocol.RoomState
	if err := json.Unmarshal(roomState.data, &stateMessage); err != nil {
		t.Fatalf("decode room state: %v", err)
	}
	if stateMessage.Status != protocol.RoomStatusDesynced || stateMessage.PendingServerSeq != 1 {
		t.Fatalf("unexpected room state after desync: %+v", stateMessage)
	}
}

func TestMatchLoopBroadcastsDesyncAndRoomStateOnResultTimeout(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)
	state.roomRuntime.CommandResultTimeoutTicks = 3
	dispatcher.clear()

	cmd := protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandEndTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         state.roomRuntime.LastAcceptedHash,
	}
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 10, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeCmdEndTurn, data: DumpEvent(cmd)},
	}).(*MatchState)
	if state.roomRuntime.PendingCheckpoints[1].DeadlineTick != 13 {
		t.Fatalf("checkpoint deadline = %d, want 13", state.roomRuntime.PendingCheckpoints[1].DeadlineTick)
	}
	dispatcher.clear()

	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 12, state, nil).(*MatchState)
	if state.roomRuntime.Status != protocol.RoomStatusInBattle {
		t.Fatalf("room should not desync before deadline, got %q", state.roomRuntime.Status)
	}
	if dispatcher.countBroadcast(protocol.OpCodeMatchDesync) != 0 {
		t.Fatalf("timeout should not broadcast before deadline")
	}

	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 13, state, nil).(*MatchState)
	if state.roomRuntime.Status != protocol.RoomStatusDesynced {
		t.Fatalf("room status = %q, want Desynced", state.roomRuntime.Status)
	}
	desync := dispatcher.lastBroadcast(protocol.OpCodeMatchDesync)
	if desync == nil {
		t.Fatalf("MATCH_DESYNC was not broadcast on timeout")
	}
	var event protocol.MatchDesyncEvent
	if err := json.Unmarshal(desync.data, &event); err != nil {
		t.Fatalf("decode desync event: %v", err)
	}
	if event.Code != protocol.ErrorDesync || event.ServerSeq != 1 || event.Checkpoint != nil {
		t.Fatalf("unexpected timeout desync event: %+v", event)
	}
	if !strings.Contains(event.Message, "checkpoint result timeout") {
		t.Fatalf("desync message should mention timeout: %+v", event)
	}
	roomState := dispatcher.lastBroadcast(protocol.OpCodeRoomState)
	if roomState == nil {
		t.Fatalf("ROOM_STATE was not broadcast on timeout")
	}
	var stateMessage protocol.RoomState
	if err := json.Unmarshal(roomState.data, &stateMessage); err != nil {
		t.Fatalf("decode room state: %v", err)
	}
	if stateMessage.Status != protocol.RoomStatusDesynced || stateMessage.PendingServerSeq != 1 || stateMessage.LastCompletedSeq != 0 {
		t.Fatalf("unexpected room state after timeout: %+v", stateMessage)
	}
	if len(state.roomRuntime.CompletedResults) != 0 || len(state.roomRuntime.CompletedCheckpoints) != 0 {
		t.Fatalf("timed out checkpoint must not be completed")
	}
	if result := dispatcher.lastBroadcast(protocol.OpCodeMatchResult); result != nil {
		t.Fatalf("timeout must not broadcast ranked match result")
	}
}

func TestMatchLoopForfeitBroadcastsUnrankedMatchResultAndRoomState(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)
	dispatcher.clear()

	request := protocol.MatchForfeitRequest{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		Seat:            protocol.SeatRed,
		Reason:          protocol.MatchResultReasonForfeit,
		ClientRequestID: "forfeit-red-1",
	}
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeMatchForfeitRequest, data: DumpEvent(request)},
	}).(*MatchState)

	if state.roomRuntime.Status != protocol.RoomStatusEnded {
		t.Fatalf("forfeit should end room, got %q", state.roomRuntime.Status)
	}
	resultMessage := dispatcher.lastBroadcast(protocol.OpCodeMatchResult)
	if resultMessage == nil {
		t.Fatalf("MATCH_RESULT was not broadcast")
	}
	var result protocol.MatchResultEvent
	if err := json.Unmarshal(resultMessage.data, &result); err != nil {
		t.Fatalf("decode match result: %v", err)
	}
	if result.WinnerSeat != protocol.SeatBlue ||
		result.LoserSeat != protocol.SeatRed ||
		result.Reason != protocol.MatchResultReasonForfeit ||
		result.RankedEligible {
		t.Fatalf("unexpected forfeit result: %+v", result)
	}
	if result.RankedReason == "" || result.SourceUserID != "u-red" {
		t.Fatalf("forfeit result missing source/ranked reason: %+v", result)
	}
	if result.ResolvedAtTick != 1 || result.ClientRequestID != "forfeit-red-1" {
		t.Fatalf("forfeit result missing resolution evidence: %+v", result)
	}
	assertCheckpointBinding(t, result, state.roomRuntime.StartHash, 0, state.roomRuntime.StartHash, 0, false)
	assertParticipantSnapshot(t, result.Participants, "match-1")
	assertTimeoutPolicy(t, result.TimeoutPolicy, DefaultCommandResultTimeoutTicks, DefaultAbandonGraceTicks, DefaultBattleEndReportTimeoutTicks)
	roomState := dispatcher.lastBroadcast(protocol.OpCodeRoomState)
	if roomState == nil {
		t.Fatalf("ROOM_STATE was not broadcast after forfeit")
	}
	var stateMessage protocol.RoomState
	if err := json.Unmarshal(roomState.data, &stateMessage); err != nil {
		t.Fatalf("decode room state: %v", err)
	}
	if stateMessage.Status != protocol.RoomStatusEnded ||
		stateMessage.Result == nil ||
		stateMessage.Result.LoserSeat != protocol.SeatRed {
		t.Fatalf("unexpected room state after forfeit: %+v", stateMessage)
	}
}

func TestMatchLoopRejectsSpectatorForfeit(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)
	spectator := testPresence{userID: "u-spec", sessionID: "s-spec", username: "spec"}
	stateAny, accepted, reason := match.MatchJoinAttempt(context.Background(), logger, nil, nil, dispatcher, 1, state, spectator, map[string]string{"seat": "spectator", "matchId": "match-1"})
	if !accepted {
		t.Fatalf("spectator join rejected: %s", reason)
	}
	state = stateAny.(*MatchState)
	state = match.MatchJoin(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.Presence{spectator}).(*MatchState)
	dispatcher.clear()

	request := protocol.MatchForfeitRequest{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		Seat:            protocol.SeatSpectator,
		Reason:          protocol.MatchResultReasonForfeit,
		ClientRequestID: "forfeit-spec-1",
	}
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 2, state, []runtime.MatchData{
		testMatchData{testPresence: spectator, opCode: protocol.OpCodeMatchForfeitRequest, data: DumpEvent(request)},
	}).(*MatchState)

	if state.roomRuntime.Status != protocol.RoomStatusInBattle {
		t.Fatalf("spectator forfeit should not end room, got %q", state.roomRuntime.Status)
	}
	if result := dispatcher.lastBroadcast(protocol.OpCodeMatchResult); result != nil {
		t.Fatalf("spectator forfeit must not broadcast match result")
	}
	errorMessage := dispatcher.lastBroadcast(protocol.OpCodeProtocolError)
	if errorMessage == nil {
		t.Fatalf("spectator forfeit should return protocol error")
	}
	var event protocol.ProtocolErrorEvent
	if err := json.Unmarshal(errorMessage.data, &event); err != nil {
		t.Fatalf("decode protocol error: %v", err)
	}
	if event.Code != protocol.ErrorInvalidSeat || event.ClientCommandID != "forfeit-spec-1" {
		t.Fatalf("unexpected protocol error: %+v", event)
	}
}

func TestMatchLoopRejectsForfeitBeforeBattleStart(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, _ := joinedProtocolRoom(t, match, logger, dispatcher)

	request := protocol.MatchForfeitRequest{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		Seat:            protocol.SeatRed,
		Reason:          protocol.MatchResultReasonForfeit,
		ClientRequestID: "forfeit-red-pre-start",
	}
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeMatchForfeitRequest, data: DumpEvent(request)},
	}).(*MatchState)

	if state.roomRuntime.Status != protocol.RoomStatusOpen {
		t.Fatalf("pre-start forfeit should not end room, got %q", state.roomRuntime.Status)
	}
	if result := dispatcher.lastBroadcast(protocol.OpCodeMatchResult); result != nil {
		t.Fatalf("pre-start forfeit must not broadcast match result")
	}
	errorMessage := dispatcher.lastBroadcast(protocol.OpCodeProtocolError)
	if errorMessage == nil {
		t.Fatalf("pre-start forfeit should return protocol error")
	}
	var event protocol.ProtocolErrorEvent
	if err := json.Unmarshal(errorMessage.data, &event); err != nil {
		t.Fatalf("decode protocol error: %v", err)
	}
	if event.Code != protocol.ErrorInvalidState || event.ClientCommandID != "forfeit-red-pre-start" {
		t.Fatalf("unexpected protocol error: %+v", event)
	}
}

func TestMatchLoopAbandonAfterGraceBroadcastsUnrankedMatchResult(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)
	state.roomRuntime.AbandonGraceTicks = 3
	dispatcher.clear()

	state = match.MatchLeave(context.Background(), logger, nil, nil, dispatcher, 10, state, []runtime.Presence{red}).(*MatchState)
	if state.roomRuntime.Status != protocol.RoomStatusInBattle {
		t.Fatalf("player leave should keep room in battle during grace, got %q", state.roomRuntime.Status)
	}
	if result := dispatcher.lastBroadcast(protocol.OpCodeMatchResult); result != nil {
		t.Fatalf("player leave must not immediately broadcast match result")
	}
	deadline := state.roomRuntime.AbandonDeadlines[protocol.SeatRed]
	if deadline.DeadlineTick != 13 {
		t.Fatalf("abandon deadline = %d, want 13", deadline.DeadlineTick)
	}
	dispatcher.clear()

	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 12, state, nil).(*MatchState)
	if state.roomRuntime.Status != protocol.RoomStatusInBattle {
		t.Fatalf("room should remain in battle before abandon deadline, got %q", state.roomRuntime.Status)
	}
	if result := dispatcher.lastBroadcast(protocol.OpCodeMatchResult); result != nil {
		t.Fatalf("abandon should not broadcast before deadline")
	}

	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 13, state, nil).(*MatchState)
	if state.roomRuntime.Status != protocol.RoomStatusEnded {
		t.Fatalf("abandon should end room, got %q", state.roomRuntime.Status)
	}
	resultMessage := dispatcher.lastBroadcast(protocol.OpCodeMatchResult)
	if resultMessage == nil {
		t.Fatalf("MATCH_RESULT was not broadcast on abandon")
	}
	var result protocol.MatchResultEvent
	if err := json.Unmarshal(resultMessage.data, &result); err != nil {
		t.Fatalf("decode abandon result: %v", err)
	}
	if result.Reason != protocol.MatchResultReasonAbandon ||
		result.WinnerSeat != protocol.SeatBlue ||
		result.LoserSeat != protocol.SeatRed ||
		result.SourceUserID != "u-red" ||
		result.ResolvedAtTick != 13 ||
		result.RankedEligible {
		t.Fatalf("unexpected abandon result: %+v", result)
	}
	assertCheckpointBinding(t, result, state.roomRuntime.StartHash, 0, state.roomRuntime.StartHash, 0, false)
	assertParticipantSnapshot(t, result.Participants, "match-1")
	assertTimeoutPolicy(t, result.TimeoutPolicy, DefaultCommandResultTimeoutTicks, 3, DefaultBattleEndReportTimeoutTicks)
	roomState := dispatcher.lastBroadcast(protocol.OpCodeRoomState)
	if roomState == nil {
		t.Fatalf("ROOM_STATE was not broadcast after abandon")
	}
	var stateMessage protocol.RoomState
	if err := json.Unmarshal(roomState.data, &stateMessage); err != nil {
		t.Fatalf("decode room state: %v", err)
	}
	if stateMessage.Status != protocol.RoomStatusEnded ||
		stateMessage.Result == nil ||
		stateMessage.Result.Reason != protocol.MatchResultReasonAbandon {
		t.Fatalf("unexpected room state after abandon: %+v", stateMessage)
	}
}

func TestSpectatorJoinReplaysTerminalMatchResult(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)
	request := protocol.MatchForfeitRequest{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		Seat:            protocol.SeatRed,
		Reason:          protocol.MatchResultReasonForfeit,
		ClientRequestID: "forfeit-red-1",
	}
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeMatchForfeitRequest, data: DumpEvent(request)},
	}).(*MatchState)
	dispatcher.clear()

	spectator := testPresence{userID: "u-spec", sessionID: "s-spec", username: "spec"}
	stateAny, accepted, reason := match.MatchJoinAttempt(context.Background(), logger, nil, nil, dispatcher, 2, state, spectator, map[string]string{"seat": "spectator", "matchId": "match-1"})
	if !accepted {
		t.Fatalf("spectator join rejected: %s", reason)
	}
	state = stateAny.(*MatchState)
	state = match.MatchJoin(context.Background(), logger, nil, nil, dispatcher, 2, state, []runtime.Presence{spectator}).(*MatchState)

	want := []int64{
		protocol.OpCodeSeatAssigned,
		protocol.OpCodeRoomState,
		protocol.OpCodeMatchStart,
		protocol.OpCodeMatchResult,
	}
	got := dispatcher.deferredOpCodes()
	if len(got) != len(want) {
		t.Fatalf("terminal replay op count = %d (%v), want %d (%v)", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("terminal replay op[%d] = %d, want %d; all=%v", i, got[i], want[i], got)
		}
	}
	var replayResult protocol.MatchResultEvent
	if err := json.Unmarshal(dispatcher.broadcasts[3].data, &replayResult); err != nil {
		t.Fatalf("decode replay match result: %v", err)
	}
	if replayResult.Reason != protocol.MatchResultReasonForfeit || replayResult.LoserSeat != protocol.SeatRed {
		t.Fatalf("unexpected replay match result: %+v", replayResult)
	}
	if !replayResult.Replay {
		t.Fatalf("replayed match result was not marked replay: %+v", replayResult)
	}
}

func TestMatchLoopBattleEndReportsBroadcastUnrankedMatchResultAfterBothSeats(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)
	dispatcher.clear()

	redReport := battleEndReportFromState(state, protocol.SeatRed, protocol.SeatRed, protocol.SeatBlue, "battle-end-red-1")
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeMatchBattleEndReport, data: DumpEvent(redReport)},
	}).(*MatchState)
	if state.roomRuntime.Status != protocol.RoomStatusInBattle {
		t.Fatalf("single battle end report should keep room in battle, got %q", state.roomRuntime.Status)
	}
	if result := dispatcher.lastBroadcast(protocol.OpCodeMatchResult); result != nil {
		t.Fatalf("single battle end report must not broadcast match result")
	}
	dispatcher.clear()

	blueReport := battleEndReportFromState(state, protocol.SeatBlue, protocol.SeatRed, protocol.SeatBlue, "battle-end-blue-1")
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 2, state, []runtime.MatchData{
		testMatchData{testPresence: blue, opCode: protocol.OpCodeMatchBattleEndReport, data: DumpEvent(blueReport)},
	}).(*MatchState)
	if state.roomRuntime.Status != protocol.RoomStatusEnded {
		t.Fatalf("matching battle end reports should end room, got %q", state.roomRuntime.Status)
	}
	resultMessage := dispatcher.lastBroadcast(protocol.OpCodeMatchResult)
	if resultMessage == nil {
		t.Fatalf("MATCH_RESULT was not broadcast")
	}
	var result protocol.MatchResultEvent
	if err := json.Unmarshal(resultMessage.data, &result); err != nil {
		t.Fatalf("decode match result: %v", err)
	}
	if result.Reason != protocol.MatchResultReasonBattleEnd ||
		result.WinnerSeat != protocol.SeatRed ||
		result.LoserSeat != protocol.SeatBlue ||
		result.RankedEligible {
		t.Fatalf("unexpected battle end result: %+v", result)
	}
	if result.ReportClientRequestIDs[protocol.SeatRed] != "battle-end-red-1" ||
		result.ReportClientRequestIDs[protocol.SeatBlue] != "battle-end-blue-1" {
		t.Fatalf("report request ids missing: %+v", result.ReportClientRequestIDs)
	}
	assertCheckpointBinding(t, result, state.roomRuntime.StartHash, 0, state.roomRuntime.StartHash, 0, false)
	assertParticipantSnapshot(t, result.Participants, "match-1")
	assertTimeoutPolicy(t, result.TimeoutPolicy, DefaultCommandResultTimeoutTicks, DefaultAbandonGraceTicks, DefaultBattleEndReportTimeoutTicks)
	roomState := dispatcher.lastBroadcast(protocol.OpCodeRoomState)
	if roomState == nil {
		t.Fatalf("ROOM_STATE was not broadcast after battle end")
	}
	var stateMessage protocol.RoomState
	if err := json.Unmarshal(roomState.data, &stateMessage); err != nil {
		t.Fatalf("decode room state: %v", err)
	}
	if stateMessage.Status != protocol.RoomStatusEnded ||
		stateMessage.Result == nil ||
		stateMessage.Result.Reason != protocol.MatchResultReasonBattleEnd {
		t.Fatalf("unexpected room state after battle end: %+v", stateMessage)
	}
}

func TestMatchLoopBattleEndReportsWaitForTerminalCommandResult(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)
	dispatcher.clear()

	cmd := protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "terminal-command-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandEndTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         state.roomRuntime.LastAcceptedHash,
	}
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeCmdEndTurn, data: DumpEvent(cmd)},
	}).(*MatchState)
	dispatcher.clear()

	redReport := protocol.MatchBattleEndReport{
		SchemaVersion:    protocol.SchemaVersion,
		MatchID:          "match-1",
		Seat:             protocol.SeatRed,
		WinnerSeat:       protocol.SeatRed,
		LoserSeat:        protocol.SeatBlue,
		ClientRequestID:  "battle-end-red-terminal",
		LastCompletedSeq: 0,
		LastHash:         state.roomRuntime.LastAcceptedHash,
	}
	blueReport := redReport
	blueReport.Seat = protocol.SeatBlue
	blueReport.ClientRequestID = "battle-end-blue-terminal"
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 2, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeMatchBattleEndReport, data: DumpEvent(redReport)},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeMatchBattleEndReport, data: DumpEvent(blueReport)},
	}).(*MatchState)
	if state.roomRuntime.Status != protocol.RoomStatusInBattle {
		t.Fatalf("battle end reports should wait for pending checkpoint, got %q", state.roomRuntime.Status)
	}
	if result := dispatcher.lastBroadcast(protocol.OpCodeMatchResult); result != nil {
		t.Fatalf("battle end reports must not resolve before pending checkpoint completes")
	}
	dispatcher.clear()

	postHash := testDigest("8", 20)
	redResult := protocol.CommandResult{
		SchemaVersion:  protocol.SchemaVersion,
		MatchID:        "match-1",
		ServerSeq:      1,
		Seat:           protocol.SeatRed,
		PostHash:       postHash,
		PostSummary:    testPostSummary(postHash, 1),
		Terminal:       true,
		NextActiveSeat: "",
		Round:          3,
		Turn:           9,
	}
	blueResult := redResult
	blueResult.Seat = protocol.SeatBlue
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 3, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeCmdResult, data: DumpEvent(redResult)},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeCmdResult, data: DumpEvent(blueResult)},
	}).(*MatchState)

	var seenCmdResult bool
	var seenMatchResult bool
	for _, broadcast := range dispatcher.broadcasts {
		if broadcast.deferred {
			continue
		}
		if broadcast.opCode == protocol.OpCodeMatchResult && !seenCmdResult {
			t.Fatalf("MATCH_RESULT was broadcast before terminal CMD_RESULT: opcodes=%v", dispatcher.broadcastOpCodes())
		}
		if broadcast.opCode == protocol.OpCodeCmdResult {
			seenCmdResult = true
		}
		if broadcast.opCode == protocol.OpCodeMatchResult {
			seenMatchResult = true
		}
	}
	if !seenCmdResult || !seenMatchResult {
		t.Fatalf("terminal CMD_RESULT and MATCH_RESULT should both broadcast, got %v", dispatcher.broadcastOpCodes())
	}

	cmdResultMessage := dispatcher.lastBroadcast(protocol.OpCodeCmdResult)
	if cmdResultMessage == nil {
		t.Fatalf("terminal CMD_RESULT was not broadcast")
	}
	var completed protocol.CommandResult
	if err := json.Unmarshal(cmdResultMessage.data, &completed); err != nil {
		t.Fatalf("decode terminal command result: %v", err)
	}
	if !completed.Terminal || completed.NextActiveSeat != "" || completed.NextActiveNetID != "" {
		t.Fatalf("unexpected terminal command result: %+v", completed)
	}

	resultMessage := dispatcher.lastBroadcast(protocol.OpCodeMatchResult)
	if resultMessage == nil {
		t.Fatalf("MATCH_RESULT was not broadcast after terminal checkpoint")
	}
	var result protocol.MatchResultEvent
	if err := json.Unmarshal(resultMessage.data, &result); err != nil {
		t.Fatalf("decode match result: %v", err)
	}
	if result.Reason != protocol.MatchResultReasonBattleEnd ||
		result.WinnerSeat != protocol.SeatRed ||
		result.LoserSeat != protocol.SeatBlue {
		t.Fatalf("unexpected battle end result: %+v", result)
	}
	assertCheckpointBinding(t, result, cmd.PreHash, 1, postHash, 0, true)
	if state.roomRuntime.Status != protocol.RoomStatusEnded ||
		state.roomRuntime.ExpectedActiveSeat != "" ||
		state.roomRuntime.ExpectedActiveNetID != "" {
		t.Fatalf("terminal checkpoint should end room and clear expected active: status=%q seat=%q netId=%q", state.roomRuntime.Status, state.roomRuntime.ExpectedActiveSeat, state.roomRuntime.ExpectedActiveNetID)
	}
	roomState := dispatcher.lastBroadcast(protocol.OpCodeRoomState)
	if roomState == nil {
		t.Fatalf("ROOM_STATE was not broadcast after terminal battle result")
	}
	var stateMessage protocol.RoomState
	if err := json.Unmarshal(roomState.data, &stateMessage); err != nil {
		t.Fatalf("decode room state: %v", err)
	}
	if stateMessage.Status != protocol.RoomStatusEnded ||
		stateMessage.ExpectedActiveSeat != "" ||
		stateMessage.ExpectedActiveNetID != "" ||
		stateMessage.Result == nil ||
		stateMessage.Result.Reason != protocol.MatchResultReasonBattleEnd {
		t.Fatalf("unexpected terminal room state: %+v", stateMessage)
	}
}

func TestSpectatorJoinReplaysTerminalCommandResultAndMatchResult(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)
	dispatcher.clear()

	cmd := protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "terminal-replay-command-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandEndTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         state.roomRuntime.LastAcceptedHash,
	}
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeCmdEndTurn, data: DumpEvent(cmd)},
	}).(*MatchState)
	redReport := battleEndReportFromStateAt(state, protocol.SeatRed, protocol.SeatRed, protocol.SeatBlue, "battle-end-red-terminal-replay", state.roomRuntime.PendingServerSeq())
	blueReport := redReport
	blueReport.Seat = protocol.SeatBlue
	blueReport.ClientRequestID = "battle-end-blue-terminal-replay"
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 2, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeMatchBattleEndReport, data: DumpEvent(redReport)},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeMatchBattleEndReport, data: DumpEvent(blueReport)},
	}).(*MatchState)
	postHash := testDigest("9", 20)
	redResult := protocol.CommandResult{
		SchemaVersion: protocol.SchemaVersion,
		MatchID:       "match-1",
		ServerSeq:     1,
		Seat:          protocol.SeatRed,
		PostHash:      postHash,
		PostSummary:   testPostSummary(postHash, 1),
		Terminal:      true,
		Round:         3,
		Turn:          9,
	}
	blueResult := redResult
	blueResult.Seat = protocol.SeatBlue
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 3, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeCmdResult, data: DumpEvent(redResult)},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeCmdResult, data: DumpEvent(blueResult)},
	}).(*MatchState)
	if state.roomRuntime.Status != protocol.RoomStatusEnded {
		t.Fatalf("terminal battle end should end room, got %q", state.roomRuntime.Status)
	}
	dispatcher.clear()

	spectator := testPresence{userID: "u-spec-terminal", sessionID: "s-spec-terminal", username: "spec-terminal"}
	stateAny, accepted, reason := match.MatchJoinAttempt(context.Background(), logger, nil, nil, dispatcher, 4, state, spectator, map[string]string{"seat": "spectator", "matchId": "match-1"})
	if !accepted {
		t.Fatalf("spectator join rejected: %s", reason)
	}
	state = stateAny.(*MatchState)
	state = match.MatchJoin(context.Background(), logger, nil, nil, dispatcher, 4, state, []runtime.Presence{spectator}).(*MatchState)

	want := []int64{
		protocol.OpCodeSeatAssigned,
		protocol.OpCodeRoomState,
		protocol.OpCodeMatchStart,
		protocol.OpCodeCmdEndTurn,
		protocol.OpCodeCmdResult,
		protocol.OpCodeMatchResult,
	}
	got := dispatcher.deferredOpCodes()
	if len(got) != len(want) {
		t.Fatalf("terminal replay op count = %d (%v), want %d (%v)", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("terminal replay op[%d] = %d, want %d; all=%v", i, got[i], want[i], got)
		}
	}
	var replayResult protocol.CommandResult
	if err := json.Unmarshal(dispatcher.broadcasts[4].data, &replayResult); err != nil {
		t.Fatalf("decode replay terminal command result: %v", err)
	}
	if !replayResult.Replay || !replayResult.Terminal || replayResult.NextActiveSeat != "" || replayResult.NextActiveNetID != "" {
		t.Fatalf("unexpected replay terminal command result: %+v", replayResult)
	}
	var matchResult protocol.MatchResultEvent
	if err := json.Unmarshal(dispatcher.broadcasts[5].data, &matchResult); err != nil {
		t.Fatalf("decode replay match result: %v", err)
	}
	if matchResult.Reason != protocol.MatchResultReasonBattleEnd ||
		matchResult.WinnerSeat != protocol.SeatRed ||
		matchResult.LoserSeat != protocol.SeatBlue {
		t.Fatalf("unexpected replay match result: %+v", matchResult)
	}
	if !matchResult.Replay {
		t.Fatalf("replayed terminal match result was not marked replay: %+v", matchResult)
	}
}

func TestMatchLoopBattleEndReportConflictBroadcastsDesync(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)
	dispatcher.clear()

	redReport := battleEndReportFromState(state, protocol.SeatRed, protocol.SeatRed, protocol.SeatBlue, "")
	blueReport := battleEndReportFromState(state, protocol.SeatBlue, protocol.SeatBlue, protocol.SeatRed, "")
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeMatchBattleEndReport, data: DumpEvent(redReport)},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeMatchBattleEndReport, data: DumpEvent(blueReport)},
	}).(*MatchState)

	if state.roomRuntime.Status != protocol.RoomStatusDesynced {
		t.Fatalf("conflicting battle end reports should desync room, got %q", state.roomRuntime.Status)
	}
	if result := dispatcher.lastBroadcast(protocol.OpCodeMatchResult); result != nil {
		t.Fatalf("conflicting battle end reports must not broadcast match result")
	}
	desync := dispatcher.lastBroadcast(protocol.OpCodeMatchDesync)
	if desync == nil {
		t.Fatalf("MATCH_DESYNC was not broadcast")
	}
	var event protocol.MatchDesyncEvent
	if err := json.Unmarshal(desync.data, &event); err != nil {
		t.Fatalf("decode desync: %v", err)
	}
	if event.Code != protocol.ErrorDesync {
		t.Fatalf("unexpected desync event: %+v", event)
	}
}

func TestMatchLoopBattleEndReportTimeoutDesyncsWithoutAwardingWinner(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)
	state.roomRuntime.BattleEndReportTimeoutTicks = 3
	dispatcher.clear()

	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 10, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeMatchBattleEndReport, data: DumpEvent(battleEndReportFromState(state, protocol.SeatRed, protocol.SeatRed, protocol.SeatBlue, "battle-end-red-1"))},
	}).(*MatchState)
	if state.roomRuntime.Status != protocol.RoomStatusInBattle {
		t.Fatalf("single battle end report should keep room in battle, got %q", state.roomRuntime.Status)
	}
	if result := dispatcher.lastBroadcast(protocol.OpCodeMatchResult); result != nil {
		t.Fatalf("single battle end report must not broadcast result")
	}
	dispatcher.clear()

	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 13, state, nil).(*MatchState)
	if state.roomRuntime.Status != protocol.RoomStatusDesynced {
		t.Fatalf("battle end report timeout should desync room, got %q", state.roomRuntime.Status)
	}
	if result := dispatcher.lastBroadcast(protocol.OpCodeMatchResult); result != nil {
		t.Fatalf("battle end report timeout must not award a winner")
	}
	desync := dispatcher.lastBroadcast(protocol.OpCodeMatchDesync)
	if desync == nil {
		t.Fatalf("MATCH_DESYNC was not broadcast on battle end report timeout")
	}
	var event protocol.MatchDesyncEvent
	if err := json.Unmarshal(desync.data, &event); err != nil {
		t.Fatalf("decode desync: %v", err)
	}
	if event.Code != protocol.ErrorDesync || !strings.Contains(event.Message, "battle end report timeout") {
		t.Fatalf("unexpected timeout desync event: %+v", event)
	}
	roomState := dispatcher.lastBroadcast(protocol.OpCodeRoomState)
	if roomState == nil {
		t.Fatalf("ROOM_STATE was not broadcast on battle end report timeout")
	}
	var stateMessage protocol.RoomState
	if err := json.Unmarshal(roomState.data, &stateMessage); err != nil {
		t.Fatalf("decode room state: %v", err)
	}
	if stateMessage.Status != protocol.RoomStatusDesynced || stateMessage.Result != nil {
		t.Fatalf("unexpected room state after battle end report timeout: %+v", stateMessage)
	}
}

func TestSpectatorJoinReceivesReplay(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)

	cmd := protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandEndTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         state.roomRuntime.LastAcceptedHash,
	}
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeCmdEndTurn, data: DumpEvent(cmd)},
	}).(*MatchState)

	postHash := testDigest("e", 20)
	redResult := protocol.CommandResult{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ServerSeq:       1,
		Seat:            protocol.SeatRed,
		PostHash:        postHash,
		PostSummary:     testPostSummary(postHash, 1),
		NextActiveSeat:  protocol.SeatBlue,
		NextActiveNetID: "blue:0",
		Round:           1,
		Turn:            2,
	}
	blueResult := redResult
	blueResult.Seat = protocol.SeatBlue
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 2, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeCmdResult, data: DumpEvent(redResult)},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeCmdResult, data: DumpEvent(blueResult)},
	}).(*MatchState)
	dispatcher.clear()

	spectator := testPresence{userID: "u-spec", sessionID: "s-spec", username: "spec"}
	stateAny, accepted, reason := match.MatchJoinAttempt(context.Background(), logger, nil, nil, dispatcher, 3, state, spectator, map[string]string{"seat": "spectator", "matchId": "match-1"})
	if !accepted {
		t.Fatalf("spectator join rejected: %s", reason)
	}
	state = stateAny.(*MatchState)
	state = match.MatchJoin(context.Background(), logger, nil, nil, dispatcher, 3, state, []runtime.Presence{spectator}).(*MatchState)

	want := []int64{
		protocol.OpCodeSeatAssigned,
		protocol.OpCodeRoomState,
		protocol.OpCodeMatchStart,
		protocol.OpCodeCmdEndTurn,
		protocol.OpCodeCmdResult,
	}
	got := dispatcher.deferredOpCodes()
	if len(got) != len(want) {
		t.Fatalf("replay op count = %d (%v), want %d (%v)", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("replay op[%d] = %d, want %d; all=%v", i, got[i], want[i], got)
		}
	}
	var replayRoomState protocol.RoomState
	if err := json.Unmarshal(dispatcher.broadcasts[1].data, &replayRoomState); err != nil {
		t.Fatalf("decode replay room state: %v", err)
	}
	if replayRoomState.RoomID != "match-1" {
		t.Fatalf("replay room state room id = %q, want match-1", replayRoomState.RoomID)
	}
	var replayCmd protocol.CommandEnvelope
	if err := json.Unmarshal(dispatcher.broadcasts[3].data, &replayCmd); err != nil {
		t.Fatalf("decode replay command: %v", err)
	}
	if !replayCmd.Replay {
		t.Fatalf("replayed command was not marked replay")
	}
	var replayResult protocol.CommandResult
	if err := json.Unmarshal(dispatcher.broadcasts[4].data, &replayResult); err != nil {
		t.Fatalf("decode replay result: %v", err)
	}
	if !replayResult.Replay {
		t.Fatalf("replayed result was not marked replay")
	}
	roomState := dispatcher.lastBroadcast(protocol.OpCodeRoomState)
	if roomState == nil {
		t.Fatalf("global room state was not broadcast after spectator join")
	}
	var stateMessage protocol.RoomState
	if err := json.Unmarshal(roomState.data, &stateMessage); err != nil {
		t.Fatalf("decode global room state: %v", err)
	}
	if len(stateMessage.Spectators) != 1 {
		t.Fatalf("global room state spectator count = %d, want 1", len(stateMessage.Spectators))
	}
}

func TestMatchSignalCommandLogQueryReturnsCompletedPrefix(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)

	cmd := protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandEndTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         state.roomRuntime.LastAcceptedHash,
	}
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeCmdEndTurn, data: DumpEvent(cmd)},
	}).(*MatchState)

	postHash := testDigest("e", 20)
	redResult := protocol.CommandResult{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ServerSeq:       1,
		Seat:            protocol.SeatRed,
		PostHash:        postHash,
		PostSummary:     testPostSummary(postHash, 1),
		NextActiveSeat:  protocol.SeatBlue,
		NextActiveNetID: "blue:0",
		Round:           1,
		Turn:            2,
	}
	blueResult := redResult
	blueResult.Seat = protocol.SeatBlue
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 2, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeCmdResult, data: DumpEvent(redResult)},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeCmdResult, data: DumpEvent(blueResult)},
	}).(*MatchState)
	originalLogLength := len(state.roomRuntime.CommandLog)

	query := MatchSignalRequest{
		SchemaVersion: protocol.SchemaVersion,
		Kind:          MatchSignalKindCommandLogQuery,
		CommandLog: &CommandLogQueryRequest{
			MatchID:                 "match-1",
			FromServerSeq:           1,
			Limit:                   10,
			IncludeRoomState:        true,
			IncludeStartSnapshot:    true,
			IncludeCompletedResults: true,
			IncludeTerminalResult:   true,
		},
	}
	nextState, raw := match.MatchSignal(context.Background(), logger, nil, nil, dispatcher, 3, state, string(DumpEvent(query)))
	if nextState != state {
		t.Fatalf("command log query should not replace match state")
	}
	var response CommandLogQueryResponse
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		t.Fatalf("decode command log response: %v", err)
	}
	if response.Status != "ok" || response.Kind != MatchSignalKindCommandLogQuery {
		t.Fatalf("unexpected command log response status: %+v", response)
	}
	if response.Persistence != "in-memory" || response.CanPromoteMVP || response.CanPromotePersistence {
		t.Fatalf("command log query must remain non-promoting in-memory evidence: %+v", response)
	}
	if response.RoomState == nil || response.RoomState.CommandLogLength != 1 || response.RoomState.CompletedResultCount != 1 {
		t.Fatalf("missing room state watermarks: %+v", response.RoomState)
	}
	if response.StartSnapshot == nil || response.StartSnapshot.RoomID != "match-1" {
		t.Fatalf("missing start snapshot: %+v", response.StartSnapshot)
	}
	if len(response.Commands) != 1 || response.Commands[0].ServerSeq != 1 || response.Commands[0].Replay {
		t.Fatalf("unexpected command log rows: %+v", response.Commands)
	}
	if len(response.CompletedResults) != 1 || response.CompletedResults[0].ServerSeq != 1 || response.CompletedResults[0].Replay {
		t.Fatalf("unexpected completed results: %+v", response.CompletedResults)
	}
	if response.HasMore || response.NextFromServerSeq != 2 {
		t.Fatalf("unexpected pagination: hasMore=%v next=%d", response.HasMore, response.NextFromServerSeq)
	}
	if len(state.roomRuntime.CommandLog) != originalLogLength {
		t.Fatalf("command log query mutated stored log")
	}
}

func TestMatchSignalCommandLogQueryReturnsPendingCheckpoint(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)

	cmd := protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandEndTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         state.roomRuntime.LastAcceptedHash,
	}
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeCmdEndTurn, data: DumpEvent(cmd)},
	}).(*MatchState)

	query := MatchSignalRequest{
		SchemaVersion: protocol.SchemaVersion,
		Kind:          MatchSignalKindCommandLogQuery,
		CommandLog: &CommandLogQueryRequest{
			MatchID:                 "match-1",
			IncludeRoomState:        true,
			IncludeCompletedResults: true,
			IncludePending:          true,
		},
	}
	_, raw := match.MatchSignal(context.Background(), logger, nil, nil, dispatcher, 2, state, string(DumpEvent(query)))
	var response CommandLogQueryResponse
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		t.Fatalf("decode pending command log response: %v", err)
	}
	if response.Status != "ok" {
		t.Fatalf("unexpected pending response status: %+v", response)
	}
	if len(response.Commands) != 1 || len(response.CompletedResults) != 0 {
		t.Fatalf("pending query should include command without completed result: %+v", response)
	}
	if response.PendingCommand == nil || response.PendingCommand.ServerSeq != 1 || response.PendingCommand.Replay {
		t.Fatalf("pending command missing or mutated: %+v", response.PendingCommand)
	}
	if response.RoomState == nil || response.RoomState.PendingServerSeq != 1 || response.PendingServerSeq != 1 {
		t.Fatalf("pending watermarks missing: response=%+v room=%+v", response, response.RoomState)
	}
}

func TestMatchSignalCommandLogQueryRejectsWrongMatchID(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, _, _ := joinedProtocolRoom(t, match, logger, dispatcher)
	query := MatchSignalRequest{
		SchemaVersion: protocol.SchemaVersion,
		Kind:          MatchSignalKindCommandLogQuery,
		CommandLog: &CommandLogQueryRequest{
			MatchID: "other-match",
		},
	}

	_, raw := match.MatchSignal(context.Background(), logger, nil, nil, dispatcher, 1, state, string(DumpEvent(query)))
	var response MatchSignalErrorResponse
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		t.Fatalf("decode command log error response: %v", err)
	}
	if response.Status != "error" || response.Code != protocol.ErrorInvalidCommand {
		t.Fatalf("unexpected command log error response: %+v", response)
	}
}

func TestMatchSignalCommandLogQueryPaginatesWithoutMutatingReplayFlags(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state := newProtocolTestState()
	state.roomID = "match-1"
	state.roomRuntime.CommandLog = []protocol.CommandEnvelope{
		{
			SchemaVersion:   protocol.SchemaVersion,
			MatchID:         "match-1",
			ServerSeq:       1,
			ClientCommandID: "client-1",
			Seat:            protocol.SeatRed,
			ActiveNetID:     "red:0",
			Op:              protocol.CommandWaitTurn,
			Payload:         json.RawMessage(`{}`),
			PreHash:         testDigest("1", 10),
		},
		{
			SchemaVersion:   protocol.SchemaVersion,
			MatchID:         "match-1",
			ServerSeq:       2,
			ClientCommandID: "client-2",
			Seat:            protocol.SeatBlue,
			ActiveNetID:     "blue:0",
			Op:              protocol.CommandEndTurn,
			Payload:         json.RawMessage(`{}`),
			PreHash:         testDigest("2", 10),
		},
		{
			SchemaVersion:   protocol.SchemaVersion,
			MatchID:         "match-1",
			ServerSeq:       3,
			ClientCommandID: "client-3",
			Seat:            protocol.SeatRed,
			ActiveNetID:     "red:0",
			Op:              protocol.CommandWaitTurn,
			Payload:         json.RawMessage(`{}`),
			PreHash:         testDigest("3", 10),
		},
	}
	makeResult := func(seat protocol.Seat, seq int64, postHash string, nextSeat protocol.Seat, nextNetID string) protocol.CommandResult {
		return protocol.CommandResult{
			SchemaVersion:   protocol.SchemaVersion,
			MatchID:         "match-1",
			ServerSeq:       seq,
			Seat:            seat,
			PostHash:        postHash,
			PostSummary:     testPostSummary(postHash, seq),
			NextActiveSeat:  nextSeat,
			NextActiveNetID: nextNetID,
			Round:           1,
			Turn:            int(seq + 1),
		}
	}
	state.roomRuntime.CompletedResults[1] = makeResult(protocol.SeatRed, 1, testDigest("a", 20), protocol.SeatBlue, "blue:0")
	state.roomRuntime.CompletedResults[2] = makeResult(protocol.SeatBlue, 2, testDigest("b", 20), protocol.SeatRed, "red:0")
	state.roomRuntime.CompletedResults[3] = makeResult(protocol.SeatRed, 3, testDigest("c", 20), protocol.SeatBlue, "blue:0")
	state.roomRuntime.NextServerSeq = 4

	query := MatchSignalRequest{
		SchemaVersion: protocol.SchemaVersion,
		Kind:          MatchSignalKindCommandLogQuery,
		CommandLog: &CommandLogQueryRequest{
			MatchID:                 "match-1",
			FromServerSeq:           2,
			Limit:                   1,
			IncludeCompletedResults: true,
		},
	}
	_, raw := match.MatchSignal(context.Background(), logger, nil, nil, dispatcher, 4, state, string(DumpEvent(query)))
	var response CommandLogQueryResponse
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		t.Fatalf("decode paged command log response: %v", err)
	}
	if len(response.Commands) != 1 || response.Commands[0].ServerSeq != 2 {
		t.Fatalf("paged command rows = %+v, want only seq 2", response.Commands)
	}
	if len(response.CompletedResults) != 1 || response.CompletedResults[0].ServerSeq != 2 {
		t.Fatalf("paged completed results = %+v, want only seq 2", response.CompletedResults)
	}
	if !response.HasMore || response.NextFromServerSeq != 3 {
		t.Fatalf("unexpected pagination: hasMore=%v next=%d", response.HasMore, response.NextFromServerSeq)
	}
	if response.Commands[0].Replay || response.CompletedResults[0].Replay {
		t.Fatalf("query must not mark stored rows as replay: command=%+v result=%+v", response.Commands[0], response.CompletedResults[0])
	}
	if state.roomRuntime.CommandLog[1].Replay || state.roomRuntime.CompletedResults[2].Replay {
		t.Fatalf("query mutated stored replay flags")
	}
}

func TestMatchSignalCommandLogQueryReturnsTerminalResult(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeMatchForfeitRequest, data: DumpEvent(protocol.MatchForfeitRequest{
			SchemaVersion:   protocol.SchemaVersion,
			MatchID:         "match-1",
			Seat:            protocol.SeatRed,
			Reason:          protocol.MatchResultReasonForfeit,
			ClientRequestID: "forfeit-red-1",
		})},
	}).(*MatchState)

	query := MatchSignalRequest{
		SchemaVersion: protocol.SchemaVersion,
		Kind:          MatchSignalKindCommandLogQuery,
		CommandLog: &CommandLogQueryRequest{
			MatchID:               "match-1",
			IncludeRoomState:      true,
			IncludeTerminalResult: true,
		},
	}
	_, raw := match.MatchSignal(context.Background(), logger, nil, nil, dispatcher, 2, state, string(DumpEvent(query)))
	var response CommandLogQueryResponse
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		t.Fatalf("decode terminal command log response: %v", err)
	}
	if response.TerminalResult == nil {
		t.Fatalf("terminal result missing from command log query")
	}
	if response.TerminalResult.Reason != protocol.MatchResultReasonForfeit || response.TerminalResult.Replay {
		t.Fatalf("unexpected terminal result: %+v", response.TerminalResult)
	}
	if response.RoomState == nil || response.RoomState.Status != protocol.RoomStatusEnded || response.RoomState.Result == nil {
		t.Fatalf("terminal room state missing result: %+v", response.RoomState)
	}
}

func TestMatchSignalRejectsMalformedOrUnknownSignal(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state := newProtocolTestState()

	_, raw := match.MatchSignal(context.Background(), logger, nil, nil, dispatcher, 1, state, "{not-json")
	var malformed MatchSignalErrorResponse
	if err := json.Unmarshal([]byte(raw), &malformed); err != nil {
		t.Fatalf("decode malformed signal response: %v", err)
	}
	if malformed.Status != "error" || malformed.Code != protocol.ErrorInvalidCommand {
		t.Fatalf("unexpected malformed signal response: %+v", malformed)
	}

	unknown := MatchSignalRequest{SchemaVersion: protocol.SchemaVersion, Kind: "unknown-signal"}
	_, raw = match.MatchSignal(context.Background(), logger, nil, nil, dispatcher, 1, state, string(DumpEvent(unknown)))
	var response MatchSignalErrorResponse
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		t.Fatalf("decode unknown signal response: %v", err)
	}
	if response.Status != "error" || response.Kind != "unknown-signal" || response.Code != protocol.ErrorInvalidCommand {
		t.Fatalf("unexpected unknown signal response: %+v", response)
	}
}

func TestSpectatorJoinReplaysCompletedPrefixAndPendingCommandWithoutResultRequirement(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)

	cmd := protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandEndTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         state.roomRuntime.LastAcceptedHash,
	}
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeCmdEndTurn, data: DumpEvent(cmd)},
	}).(*MatchState)
	dispatcher.clear()

	spectator := testPresence{userID: "u-spec", sessionID: "s-spec", username: "spec"}
	stateAny, accepted, reason := match.MatchJoinAttempt(context.Background(), logger, nil, nil, dispatcher, 2, state, spectator, map[string]string{"seat": "spectator", "matchId": "match-1"})
	if !accepted {
		t.Fatalf("spectator join rejected: %s", reason)
	}
	state = stateAny.(*MatchState)
	state = match.MatchJoin(context.Background(), logger, nil, nil, dispatcher, 2, state, []runtime.Presence{spectator}).(*MatchState)

	want := []int64{
		protocol.OpCodeSeatAssigned,
		protocol.OpCodeRoomState,
		protocol.OpCodeMatchStart,
		protocol.OpCodeCmdEndTurn,
	}
	got := dispatcher.deferredOpCodes()
	if len(got) != len(want) {
		t.Fatalf("replay op count = %d (%v), want %d (%v)", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("replay op[%d] = %d, want %d; all=%v", i, got[i], want[i], got)
		}
	}
	var replayRoomState protocol.RoomState
	if err := json.Unmarshal(dispatcher.broadcasts[1].data, &replayRoomState); err != nil {
		t.Fatalf("decode replay room state: %v", err)
	}
	if replayRoomState.PendingServerSeq != 1 || replayRoomState.LastCompletedSeq != 0 {
		t.Fatalf("unexpected replay watermarks: %+v", replayRoomState)
	}
	if replayRoomState.CommandLogLength != 1 || replayRoomState.CompletedResultCount != 0 {
		t.Fatalf("unexpected replay log counts: %+v", replayRoomState)
	}
	var replayCmd protocol.CommandEnvelope
	if err := json.Unmarshal(dispatcher.broadcasts[3].data, &replayCmd); err != nil {
		t.Fatalf("decode replay pending command: %v", err)
	}
	if !replayCmd.Replay {
		t.Fatalf("pending command replay was not marked replay")
	}
	if replayCmd.ReplayRequiresResult {
		t.Fatalf("spectator pending replay should not require result: %+v", replayCmd)
	}
	if dispatcher.countDeferred(protocol.OpCodeCmdResult) != 0 {
		t.Fatalf("pending replay should not include command result")
	}
}

func TestPlayerReconnectPendingCommandRequiresMissingResult(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)

	cmd := protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandEndTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         state.roomRuntime.LastAcceptedHash,
	}
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeCmdEndTurn, data: DumpEvent(cmd)},
	}).(*MatchState)
	dispatcher.clear()

	state = match.MatchLeave(context.Background(), logger, nil, nil, dispatcher, 2, state, []runtime.Presence{red}).(*MatchState)
	dispatcher.clear()

	redReconnect := testPresence{userID: "u-red", sessionID: "s-red-2", username: "red"}
	stateAny, accepted, reason := match.MatchJoinAttempt(context.Background(), logger, nil, nil, dispatcher, 3, state, redReconnect, map[string]string{"seat": "auto", "matchId": "match-1"})
	if !accepted {
		t.Fatalf("red reconnect rejected: %s", reason)
	}
	state = stateAny.(*MatchState)
	state = match.MatchJoin(context.Background(), logger, nil, nil, dispatcher, 3, state, []runtime.Presence{redReconnect}).(*MatchState)

	want := []int64{
		protocol.OpCodeSeatAssigned,
		protocol.OpCodeRoomState,
		protocol.OpCodeMatchStart,
		protocol.OpCodeCmdEndTurn,
	}
	got := dispatcher.deferredOpCodes()
	if len(got) != len(want) {
		t.Fatalf("reconnect replay op count = %d (%v), want %d (%v)", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("reconnect replay op[%d] = %d, want %d; all=%v", i, got[i], want[i], got)
		}
	}
	var replayCmd protocol.CommandEnvelope
	if err := json.Unmarshal(dispatcher.broadcasts[3].data, &replayCmd); err != nil {
		t.Fatalf("decode replay pending command: %v", err)
	}
	if !replayCmd.Replay || !replayCmd.ReplayRequiresResult {
		t.Fatalf("reconnected player with missing checkpoint should replay and report result: %+v", replayCmd)
	}
	if replayCmd.ServerSeq != 1 || replayCmd.ActiveNetID != "red:0" {
		t.Fatalf("unexpected pending replay command: %+v", replayCmd)
	}
}

func TestPlayerReconnectReceivesReplay(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)

	cmd := protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandEndTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         state.roomRuntime.LastAcceptedHash,
	}
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeCmdEndTurn, data: DumpEvent(cmd)},
	}).(*MatchState)

	postHash := testDigest("f", 20)
	redResult := protocol.CommandResult{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ServerSeq:       1,
		Seat:            protocol.SeatRed,
		PostHash:        postHash,
		PostSummary:     testPostSummary(postHash, 1),
		NextActiveSeat:  protocol.SeatBlue,
		NextActiveNetID: "blue:0",
		Round:           1,
		Turn:            2,
	}
	blueResult := redResult
	blueResult.Seat = protocol.SeatBlue
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 2, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeCmdResult, data: DumpEvent(redResult)},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeCmdResult, data: DumpEvent(blueResult)},
	}).(*MatchState)
	dispatcher.clear()

	state = match.MatchLeave(context.Background(), logger, nil, nil, dispatcher, 3, state, []runtime.Presence{red}).(*MatchState)
	if state.roomRuntime.Seats[protocol.SeatRed].SessionID != "" {
		t.Fatalf("red session should be cleared after leave")
	}
	dispatcher.clear()

	redReconnect := testPresence{userID: "u-red", sessionID: "s-red-2", username: "red"}
	stateAny, accepted, reason := match.MatchJoinAttempt(context.Background(), logger, nil, nil, dispatcher, 4, state, redReconnect, map[string]string{"seat": "auto", "matchId": "match-1"})
	if !accepted {
		t.Fatalf("red reconnect rejected: %s", reason)
	}
	state = stateAny.(*MatchState)
	state = match.MatchJoin(context.Background(), logger, nil, nil, dispatcher, 4, state, []runtime.Presence{redReconnect}).(*MatchState)

	if seat, ok := state.roomRuntime.SeatForSession("s-red-2"); !ok || seat != protocol.SeatRed {
		t.Fatalf("reconnected seat = %q ok=%v, want red", seat, ok)
	}
	want := []int64{
		protocol.OpCodeSeatAssigned,
		protocol.OpCodeRoomState,
		protocol.OpCodeMatchStart,
		protocol.OpCodeCmdEndTurn,
		protocol.OpCodeCmdResult,
	}
	got := dispatcher.deferredOpCodes()
	if len(got) != len(want) {
		t.Fatalf("reconnect replay op count = %d (%v), want %d (%v)", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("reconnect replay op[%d] = %d, want %d; all=%v", i, got[i], want[i], got)
		}
	}
	var assigned protocol.SeatAssignedEvent
	if err := json.Unmarshal(dispatcher.broadcasts[0].data, &assigned); err != nil {
		t.Fatalf("decode reconnect seat assignment: %v", err)
	}
	if assigned.Seat != protocol.SeatRed || assigned.SessionID != "s-red-2" {
		t.Fatalf("unexpected reconnect seat assignment: %+v", assigned)
	}
	var replayRoomState protocol.RoomState
	if err := json.Unmarshal(dispatcher.broadcasts[1].data, &replayRoomState); err != nil {
		t.Fatalf("decode reconnect room state: %v", err)
	}
	if !replayRoomState.RedConnected || !replayRoomState.BlueConnected {
		t.Fatalf("reconnect replay room state should show both players connected: %+v", replayRoomState)
	}
	var replayCmd protocol.CommandEnvelope
	if err := json.Unmarshal(dispatcher.broadcasts[3].data, &replayCmd); err != nil {
		t.Fatalf("decode replay command: %v", err)
	}
	if !replayCmd.Replay {
		t.Fatalf("replayed command was not marked replay")
	}
	var replayResult protocol.CommandResult
	if err := json.Unmarshal(dispatcher.broadcasts[4].data, &replayResult); err != nil {
		t.Fatalf("decode replay result: %v", err)
	}
	if !replayResult.Replay || replayResult.PostHash != postHash {
		t.Fatalf("unexpected replay result: %+v", replayResult)
	}
}

func TestMatchLeaveUpdatesRoomStateForSpectator(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, _, _ := joinedProtocolRoom(t, match, logger, dispatcher)

	spectator := testPresence{userID: "u-spec", sessionID: "s-spec", username: "spec"}
	stateAny, accepted, reason := match.MatchJoinAttempt(context.Background(), logger, nil, nil, dispatcher, 1, state, spectator, map[string]string{"seat": "spectator"})
	if !accepted {
		t.Fatalf("spectator join rejected: %s", reason)
	}
	state = stateAny.(*MatchState)
	state = match.MatchJoin(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.Presence{spectator}).(*MatchState)
	dispatcher.clear()

	state = match.MatchLeave(context.Background(), logger, nil, nil, dispatcher, 2, state, []runtime.Presence{spectator}).(*MatchState)
	if _, ok := state.roomRuntime.Spectators["s-spec"]; ok {
		t.Fatalf("spectator was not removed from runtime")
	}
	roomState := dispatcher.lastBroadcast(protocol.OpCodeRoomState)
	if roomState == nil {
		t.Fatalf("room state was not broadcast after spectator leave")
	}
	var stateMessage protocol.RoomState
	if err := json.Unmarshal(roomState.data, &stateMessage); err != nil {
		t.Fatalf("decode room state: %v", err)
	}
	if len(stateMessage.Spectators) != 0 {
		t.Fatalf("spectator still present in room state: %+v", stateMessage)
	}
}

func TestMatchLeaveKeepsInBattleRuntimeForReconnect(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)
	dispatcher.clear()

	stateAny := match.MatchLeave(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.Presence{red, blue})
	state, ok := stateAny.(*MatchState)
	if !ok || state == nil {
		t.Fatalf("in-battle match state should be retained for reconnect")
	}
	if state.roomRuntime.Seats[protocol.SeatRed].SessionID != "" || state.roomRuntime.Seats[protocol.SeatBlue].SessionID != "" {
		t.Fatalf("player sessions should be cleared after leave: %+v", state.roomRuntime.Seats)
	}
}

func TestV1RoomRejectsLegacyOpponentForceQuitAfterStart(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)
	dispatcher.clear()

	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: OpCodeOpponentForceQuit, data: []byte(`{}`)},
	}).(*MatchState)

	if state.roomRuntime.Status != protocol.RoomStatusInBattle {
		t.Fatalf("legacy force quit should not close v1 room, got %q", state.roomRuntime.Status)
	}
	if state.battleEnd || state.battleEndEventBroadcast {
		t.Fatalf("legacy force quit should not mark battle ended in v1 room")
	}
	if result := dispatcher.lastBroadcast(OpCodeBattleEnd); result != nil {
		t.Fatalf("legacy force quit should not broadcast battle end in v1 room")
	}
	errorMessage := dispatcher.lastBroadcast(protocol.OpCodeProtocolError)
	if errorMessage == nil {
		t.Fatalf("legacy force quit should return protocol error")
	}
	var event protocol.ProtocolErrorEvent
	if err := json.Unmarshal(errorMessage.data, &event); err != nil {
		t.Fatalf("decode protocol error: %v", err)
	}
	if event.Code != protocol.ErrorInvalidCommand || event.OpCode != OpCodeOpponentForceQuit {
		t.Fatalf("unexpected protocol error: %+v", event)
	}
}

func TestV1RoomRejectsLegacyBattleEndAfterStart(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)
	dispatcher.clear()

	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: OpCodeBattleEnd, data: DumpEvent(BattleEndEvent{Faction: FactionRedSide, IsVictory: true})},
		testMatchData{testPresence: blue, opCode: OpCodeBattleEnd, data: DumpEvent(BattleEndEvent{Faction: FactionBlueSide, IsVictory: false})},
	}).(*MatchState)

	if state.roomRuntime.Status != protocol.RoomStatusInBattle {
		t.Fatalf("legacy battle end should not end v1 room, got %q", state.roomRuntime.Status)
	}
	if result := dispatcher.lastBroadcast(OpCodeBattleEnd); result == nil {
		// Expected: no legacy result broadcast.
	} else {
		t.Fatalf("legacy battle end should not broadcast result in v1 room")
	}
	if state.battleEnd || len(state.battleResult) != 0 {
		t.Fatalf("legacy battle end should not mutate v1 battle result: end=%v results=%+v", state.battleEnd, state.battleResult)
	}
	if errors := dispatcher.countBroadcast(protocol.OpCodeProtocolError); errors != 2 {
		t.Fatalf("legacy battle end submissions should return protocol errors, got %d", errors)
	}
}

func TestV1RoomIgnoresLegacyBattleResultPollution(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatRed))},
		testMatchData{testPresence: blue, opCode: protocol.OpCodeRosterLocked, data: DumpEvent(roster(protocol.SeatBlue))},
	}).(*MatchState)
	dispatcher.clear()

	state.totalFaction = 2
	state.battleEnd = true
	state.battleResult[FactionRedSide] = true
	state.battleResult[FactionBlueSide] = false
	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 1, state, nil).(*MatchState)

	if result := dispatcher.lastBroadcast(OpCodeBattleEnd); result != nil {
		t.Fatalf("v1 room must not broadcast legacy battle result when legacy fields are polluted")
	}
	if state.battleEndEventBroadcast {
		t.Fatalf("v1 room must not mark legacy battle result as broadcast")
	}
	if state.roomRuntime.Status != protocol.RoomStatusInBattle {
		t.Fatalf("v1 room runtime status = %q, want InBattle", state.roomRuntime.Status)
	}
	if state.roomRuntime.TerminalResult != nil {
		t.Fatalf("v1 legacy pollution must not create terminal result: %+v", state.roomRuntime.TerminalResult)
	}
}

func TestV1RoomRejectsLegacyOpponentReadyEvenIfLegacyFactionsPopulated(t *testing.T) {
	match := &Match{}
	logger := testLogger{}
	dispatcher := &testDispatcher{}
	state, red, blue := joinedProtocolRoom(t, match, logger, dispatcher)
	state.factions[FactionRedSide] = state.presences[red.sessionID]
	state.factions[FactionBlueSide] = state.presences[blue.sessionID]
	dispatcher.clear()

	state = match.MatchLoop(context.Background(), logger, nil, nil, dispatcher, 2, state, []runtime.MatchData{
		testMatchData{testPresence: red, opCode: OpCodeOpponentReady, data: []byte(`{"legacy":"roster-red"}`)},
		testMatchData{testPresence: blue, opCode: OpCodeOpponentReady, data: []byte(`{"legacy":"roster-blue"}`)},
	}).(*MatchState)

	if state.battleStarted {
		t.Fatalf("legacy ready should not start v1 room")
	}
	if result := dispatcher.lastBroadcast(OpCodeMatchStart); result != nil {
		t.Fatalf("legacy ready should not broadcast old match start")
	}
	if errors := dispatcher.countBroadcast(protocol.OpCodeProtocolError); errors != 2 {
		t.Fatalf("legacy ready submissions should return protocol errors, got %d", errors)
	}
}

func joinedProtocolRoom(t *testing.T, match *Match, logger runtime.Logger, dispatcher *testDispatcher) (*MatchState, testPresence, testPresence) {
	t.Helper()

	state := newProtocolTestState()
	red := testPresence{userID: "u-red", sessionID: "s-red", username: "red"}
	blue := testPresence{userID: "u-blue", sessionID: "s-blue", username: "blue"}
	stateAny, accepted, reason := match.MatchJoinAttempt(context.Background(), logger, nil, nil, dispatcher, 0, state, red, map[string]string{"seat": "red", "matchId": "match-1"})
	if !accepted {
		t.Fatalf("red join rejected: %s", reason)
	}
	state = stateAny.(*MatchState)
	state = match.MatchJoin(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.Presence{red}).(*MatchState)
	stateAny, accepted, reason = match.MatchJoinAttempt(context.Background(), logger, nil, nil, dispatcher, 0, state, blue, map[string]string{"seat": "blue", "matchId": "match-1"})
	if !accepted {
		t.Fatalf("blue join rejected: %s", reason)
	}
	state = stateAny.(*MatchState)
	state = match.MatchJoin(context.Background(), logger, nil, nil, dispatcher, 0, state, []runtime.Presence{blue}).(*MatchState)
	dispatcher.clear()
	return state, red, blue
}

func battleEndReportFromState(state *MatchState, seat protocol.Seat, winner protocol.Seat, loser protocol.Seat, requestID string) protocol.MatchBattleEndReport {
	return battleEndReportFromStateAt(state, seat, winner, loser, requestID, state.roomRuntime.LastCompletedSeq())
}

func battleEndReportFromStateAt(state *MatchState, seat protocol.Seat, winner protocol.Seat, loser protocol.Seat, requestID string, lastCompletedSeq int64) protocol.MatchBattleEndReport {
	return protocol.MatchBattleEndReport{
		SchemaVersion:    protocol.SchemaVersion,
		MatchID:          "match-1",
		Seat:             seat,
		WinnerSeat:       winner,
		LoserSeat:        loser,
		ClientRequestID:  requestID,
		LastCompletedSeq: lastCompletedSeq,
		LastHash:         state.roomRuntime.LastAcceptedHash,
	}
}

func newProtocolTestState() *MatchState {
	return &MatchState{
		factions:            map[int]*Presence{},
		presences:           map[string]*Presence{},
		battleResult:        map[int]bool{},
		features:            map[int]bool{},
		roomRuntime:         NewRoomRuntime(""),
		joinSeats:           map[string]protocol.Seat{},
		joinReservations:    map[string]JoinReservation{},
		strategicProperties: strategicProperties.BuildStrategicProperties(strategicProperties.ScenarioArena, 101, 202),
	}
}

type testPresence struct {
	userID    string
	sessionID string
	username  string
}

func (p testPresence) GetUserId() string                 { return p.userID }
func (p testPresence) GetSessionId() string              { return p.sessionID }
func (p testPresence) GetNodeId() string                 { return "node" }
func (p testPresence) GetHidden() bool                   { return false }
func (p testPresence) GetPersistence() bool              { return false }
func (p testPresence) GetUsername() string               { return p.username }
func (p testPresence) GetStatus() string                 { return "" }
func (p testPresence) GetReason() runtime.PresenceReason { return 0 }

type testMatchData struct {
	testPresence
	opCode int64
	data   []byte
}

func (m testMatchData) GetOpCode() int64      { return m.opCode }
func (m testMatchData) GetData() []byte       { return m.data }
func (m testMatchData) GetReliable() bool     { return true }
func (m testMatchData) GetReceiveTime() int64 { return 0 }

type testBroadcast struct {
	opCode           int64
	data             []byte
	deferred         bool
	targetSessionIDs []string
}

type testDispatcher struct {
	broadcasts       []testBroadcast
	kickedSessionIDs []string
}

func (d *testDispatcher) BroadcastMessage(opCode int64, data []byte, presences []runtime.Presence, sender runtime.Presence, reliable bool) error {
	d.broadcasts = append(d.broadcasts, testBroadcast{opCode: opCode, data: data, targetSessionIDs: targetSessionIDs(presences)})
	return nil
}

func (d *testDispatcher) BroadcastMessageDeferred(opCode int64, data []byte, presences []runtime.Presence, sender runtime.Presence, reliable bool) error {
	d.broadcasts = append(d.broadcasts, testBroadcast{opCode: opCode, data: data, deferred: true, targetSessionIDs: targetSessionIDs(presences)})
	return nil
}

func (d *testDispatcher) MatchKick(presences []runtime.Presence) error {
	d.kickedSessionIDs = append(d.kickedSessionIDs, targetSessionIDs(presences)...)
	return nil
}
func (d *testDispatcher) MatchLabelUpdate(label string) error { return nil }

func (d *testDispatcher) clear() {
	d.broadcasts = nil
	d.kickedSessionIDs = nil
}

func (d *testDispatcher) countDeferred(opCode int64) int {
	count := 0
	for _, b := range d.broadcasts {
		if b.opCode == opCode && b.deferred {
			count++
		}
	}
	return count
}

func (d *testDispatcher) countBroadcast(opCode int64) int {
	count := 0
	for _, b := range d.broadcasts {
		if b.opCode == opCode && !b.deferred {
			count++
		}
	}
	return count
}

func (d *testDispatcher) countKicked(sessionID string) int {
	count := 0
	for _, kickedSessionID := range d.kickedSessionIDs {
		if kickedSessionID == sessionID {
			count++
		}
	}
	return count
}

func (d *testDispatcher) lastBroadcast(opCode int64) *testBroadcast {
	for i := len(d.broadcasts) - 1; i >= 0; i-- {
		if d.broadcasts[i].opCode == opCode && !d.broadcasts[i].deferred {
			return &d.broadcasts[i]
		}
	}
	return nil
}

func (d *testDispatcher) deferredOpCodes() []int64 {
	var opCodes []int64
	for _, b := range d.broadcasts {
		if b.deferred {
			opCodes = append(opCodes, b.opCode)
		}
	}
	return opCodes
}

func (d *testDispatcher) broadcastOpCodes() []int64 {
	opCodes := make([]int64, 0, len(d.broadcasts))
	for _, b := range d.broadcasts {
		opCodes = append(opCodes, b.opCode)
	}
	return opCodes
}

func targetSessionIDs(presences []runtime.Presence) []string {
	if len(presences) == 0 {
		return nil
	}
	ids := make([]string, 0, len(presences))
	for _, presence := range presences {
		ids = append(ids, presence.GetSessionId())
	}
	return ids
}

type testLogger struct{}

func (testLogger) Debug(format string, v ...interface{})              {}
func (testLogger) Info(format string, v ...interface{})               {}
func (testLogger) Warn(format string, v ...interface{})               {}
func (testLogger) Error(format string, v ...interface{})              {}
func (testLogger) WithField(key string, v interface{}) runtime.Logger { return testLogger{} }
func (testLogger) WithFields(fields map[string]interface{}) runtime.Logger {
	return testLogger{}
}
func (testLogger) Fields() map[string]interface{} { return map[string]interface{}{} }

var _ runtime.MatchDispatcher = (*testDispatcher)(nil)
var _ runtime.Presence = testPresence{}
var _ runtime.MatchData = testMatchData{}
var _ runtime.Logger = testLogger{}
