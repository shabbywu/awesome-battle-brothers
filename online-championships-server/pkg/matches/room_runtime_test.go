package matches

import (
	"encoding/json"
	"fmt"
	"online-championships/pkg/protocol"
	"strings"
	"testing"
)

func TestRoomRuntimeAcceptsCommandAndAdvancesCheckpoint(t *testing.T) {
	room := readyRoom(t)

	cmd := protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandWaitTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         "start-hash",
	}

	accepted, err := room.AcceptCommand(cmd)
	if err != nil {
		t.Fatalf("command rejected: %v", err)
	}
	if accepted.ServerSeq != 1 {
		t.Fatalf("serverSeq = %d, want 1", accepted.ServerSeq)
	}
	if len(room.CommandLog) != 1 {
		t.Fatalf("command log len = %d, want 1", len(room.CommandLog))
	}

	postHash := testDigest("b", 20)
	recordResult(t, room, protocol.SeatRed, 1, postHash, protocol.SeatBlue, "blue:0")
	recordResult(t, room, protocol.SeatBlue, 1, postHash, protocol.SeatBlue, "blue:0")

	if room.LastAcceptedHash != postHash {
		t.Fatalf("last hash = %q, want %q", room.LastAcceptedHash, postHash)
	}
	if room.ExpectedActiveSeat != protocol.SeatBlue {
		t.Fatalf("next active seat = %q, want blue", room.ExpectedActiveSeat)
	}
	if len(room.PendingCheckpoints) != 0 {
		t.Fatalf("pending checkpoints were not cleared")
	}
}

func TestRoomRuntimeRejectsNonActiveSeat(t *testing.T) {
	room := readyRoom(t)

	_, err := room.AcceptCommand(protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatBlue,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandWaitTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         "start-hash",
	})
	if protocol.ErrorCodeOf(err) != protocol.ErrorInvalidSeat {
		t.Fatalf("non-active seat should be rejected, got %v", err)
	}
}

func TestRoomRuntimePreHashMismatchDesyncsRoom(t *testing.T) {
	room := readyRoom(t)

	_, err := room.AcceptCommand(protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandWaitTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         "stale-hash",
	})
	if protocol.ErrorCodeOf(err) != protocol.ErrorDesync {
		t.Fatalf("preHash mismatch should desync, got %v", err)
	}
	if room.Status != protocol.RoomStatusDesynced {
		t.Fatalf("room status = %q, want Desynced", room.Status)
	}

	_, err = room.AcceptCommand(protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-2",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandWaitTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         "start-hash",
	})
	if protocol.ErrorCodeOf(err) != protocol.ErrorInvalidState {
		t.Fatalf("desynced room should reject later commands, got %v", err)
	}
}

func TestRoomRuntimeReturnsSameAcceptedCommandForDuplicateClientCommandID(t *testing.T) {
	room := readyRoom(t)
	cmd := protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandWaitTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         "start-hash",
	}
	first, err := room.AcceptCommand(cmd)
	if err != nil {
		t.Fatalf("first command rejected: %v", err)
	}
	duplicate, err := room.AcceptCommand(cmd)
	if err != nil {
		t.Fatalf("duplicate command rejected: %v", err)
	}
	if duplicate.ServerSeq != first.ServerSeq {
		t.Fatalf("duplicate serverSeq = %d, want %d", duplicate.ServerSeq, first.ServerSeq)
	}
	if len(room.CommandLog) != 1 {
		t.Fatalf("command log len = %d, want 1", len(room.CommandLog))
	}
}

func TestRoomRuntimeRejectsRosterLockAfterStart(t *testing.T) {
	room := readyRoom(t)
	err := room.LockRoster(protocol.SeatRed, roster(protocol.SeatRed))
	if protocol.ErrorCodeOf(err) != protocol.ErrorInvalidState {
		t.Fatalf("roster lock after start should be rejected, got %v", err)
	}
}

func TestRoomRuntimeRejectsSpoofedRosterOwnerUserID(t *testing.T) {
	room := NewRoomRuntime("")
	if _, err := room.AssignSeat(SeatPresence{UserID: "u-red", SessionID: "s-red", Username: "red"}, protocol.SeatRed); err != nil {
		t.Fatalf("assign red: %v", err)
	}
	spoofed := roster(protocol.SeatRed)
	spoofed.OwnerUserID = "u-blue"
	if err := room.LockRoster(protocol.SeatRed, spoofed); protocol.ErrorCodeOf(err) != protocol.ErrorInvalidCommand {
		t.Fatalf("spoofed roster owner should be rejected, got %v", err)
	}

	blank := roster(protocol.SeatRed)
	blank.OwnerUserID = ""
	if err := room.LockRoster(protocol.SeatRed, blank); err != nil {
		t.Fatalf("blank roster owner should be server-filled: %v", err)
	}
	if got := room.Rosters[protocol.SeatRed].OwnerUserID; got != "u-red" {
		t.Fatalf("roster ownerUserId = %q, want u-red", got)
	}
}

func TestRoomRuntimeRejectsRosterRulesetMismatch(t *testing.T) {
	room := NewRoomRuntime("")
	if _, err := room.AssignSeat(SeatPresence{UserID: "u-red", SessionID: "s-red", Username: "red"}, protocol.SeatRed); err != nil {
		t.Fatalf("assign red: %v", err)
	}
	if _, err := room.AssignSeat(SeatPresence{UserID: "u-blue", SessionID: "s-blue", Username: "blue"}, protocol.SeatBlue); err != nil {
		t.Fatalf("assign blue: %v", err)
	}
	if err := room.LockRoster(protocol.SeatRed, roster(protocol.SeatRed)); err != nil {
		t.Fatalf("lock red roster: %v", err)
	}
	blue := roster(protocol.SeatBlue)
	blue.RulesetHash = "other-rules"
	if err := room.LockRoster(protocol.SeatBlue, blue); protocol.ErrorCodeOf(err) != protocol.ErrorInvalidCommand {
		t.Fatalf("ruleset mismatch should be rejected, got %v", err)
	}
}

func TestRoomRuntimeRejectsRosterGameSHA256Mismatch(t *testing.T) {
	room := NewRoomRuntime("")
	if _, err := room.AssignSeat(SeatPresence{UserID: "u-red", SessionID: "s-red", Username: "red"}, protocol.SeatRed); err != nil {
		t.Fatalf("assign red: %v", err)
	}
	if _, err := room.AssignSeat(SeatPresence{UserID: "u-blue", SessionID: "s-blue", Username: "blue"}, protocol.SeatBlue); err != nil {
		t.Fatalf("assign blue: %v", err)
	}
	if err := room.LockRoster(protocol.SeatRed, roster(protocol.SeatRed)); err != nil {
		t.Fatalf("lock red roster: %v", err)
	}
	blue := roster(protocol.SeatBlue)
	blue.GameSHA256 = strings.Repeat("b", 64)
	if err := room.LockRoster(protocol.SeatBlue, blue); protocol.ErrorCodeOf(err) != protocol.ErrorInvalidCommand {
		t.Fatalf("gameSha256 mismatch should be rejected, got %v", err)
	}
}

func TestRoomRuntimeAllowsSpectatorJoinAfterStart(t *testing.T) {
	room := readyRoom(t)
	seat, err := room.AssignSeat(SeatPresence{UserID: "u-spec", SessionID: "s-spec", Username: "spec"}, protocol.SeatSpectator)
	if err != nil {
		t.Fatalf("spectator join after start rejected: %v", err)
	}
	if seat != protocol.SeatSpectator {
		t.Fatalf("seat = %q, want spectator", seat)
	}

	_, err = room.AssignSeat(SeatPresence{UserID: "u-new", SessionID: "s-new", Username: "new"}, protocol.SeatAuto)
	if protocol.ErrorCodeOf(err) != protocol.ErrorInvalidState {
		t.Fatalf("new player join after start should be rejected, got %v", err)
	}
}

func TestRoomRuntimeRejectsSpectatorWhenDisabled(t *testing.T) {
	room := readyRoom(t)
	metadata := room.Metadata
	metadata.Spectators.Enabled = false
	room.Metadata = metadata

	_, err := room.AssignSeat(SeatPresence{UserID: "u-spec", SessionID: "s-spec", Username: "spec"}, protocol.SeatSpectator)
	if protocol.ErrorCodeOf(err) != protocol.ErrorInvalidState {
		t.Fatalf("disabled spectator join should be rejected, got %v", err)
	}
	if len(room.Spectators) != 0 {
		t.Fatalf("disabled spectator join should not mutate spectators: %+v", room.Spectators)
	}
}

func TestRoomRuntimeRejectsSpectatorAboveLimit(t *testing.T) {
	room := readyRoom(t)
	metadata := room.Metadata
	metadata.Spectators.Enabled = true
	metadata.Spectators.Limit = 1
	room.Metadata = metadata

	if _, err := room.AssignSeat(SeatPresence{UserID: "u-spec-1", SessionID: "s-spec-1", Username: "spec1"}, protocol.SeatSpectator); err != nil {
		t.Fatalf("first spectator rejected: %v", err)
	}
	_, err := room.AssignSeat(SeatPresence{UserID: "u-spec-2", SessionID: "s-spec-2", Username: "spec2"}, protocol.SeatSpectator)
	if protocol.ErrorCodeOf(err) != protocol.ErrorInvalidState {
		t.Fatalf("spectator above limit should be rejected, got %v", err)
	}
	if len(room.Spectators) != 1 {
		t.Fatalf("spectator limit rejection should preserve existing spectators, got %d", len(room.Spectators))
	}
}

func TestRoomRuntimeSpectatorLeaveFreesLimitSlot(t *testing.T) {
	room := readyRoom(t)
	metadata := room.Metadata
	metadata.Spectators.Enabled = true
	metadata.Spectators.Limit = 1
	room.Metadata = metadata

	if _, err := room.AssignSeat(SeatPresence{UserID: "u-spec-1", SessionID: "s-spec-1", Username: "spec1"}, protocol.SeatSpectator); err != nil {
		t.Fatalf("first spectator rejected: %v", err)
	}
	if seat, ok := room.RemovePresence("s-spec-1"); !ok || seat != protocol.SeatSpectator {
		t.Fatalf("remove spectator = %q ok=%v, want spectator true", seat, ok)
	}
	if _, err := room.AssignSeat(SeatPresence{UserID: "u-spec-2", SessionID: "s-spec-2", Username: "spec2"}, protocol.SeatSpectator); err != nil {
		t.Fatalf("spectator slot was not freed: %v", err)
	}
}

func TestRoomRuntimeSpectatorReconnectDoesNotConsumeLimitSlot(t *testing.T) {
	room := readyRoom(t)
	metadata := room.Metadata
	metadata.Spectators.Enabled = true
	metadata.Spectators.Limit = 1
	room.Metadata = metadata

	if _, err := room.AssignSeat(SeatPresence{UserID: "u-spec", SessionID: "s-spec-1", Username: "spec"}, protocol.SeatSpectator); err != nil {
		t.Fatalf("first spectator rejected: %v", err)
	}
	if _, err := room.AssignSeat(SeatPresence{UserID: "u-spec", SessionID: "s-spec-2", Username: "spec"}, protocol.SeatSpectator); err != nil {
		t.Fatalf("same spectator reconnect rejected: %v", err)
	}
	if _, ok := room.Spectators["s-spec-1"]; ok {
		t.Fatalf("old spectator session should be replaced")
	}
	if _, ok := room.Spectators["s-spec-2"]; !ok || len(room.Spectators) != 1 {
		t.Fatalf("reconnected spectator was not retained correctly: %+v", room.Spectators)
	}
}

func TestRoomRuntimeRemovePresenceBeforeStartFreesSeatAndRoster(t *testing.T) {
	room := NewRoomRuntime("")
	if _, err := room.AssignSeat(SeatPresence{UserID: "u-red", SessionID: "s-red", Username: "red"}, protocol.SeatRed); err != nil {
		t.Fatalf("assign red: %v", err)
	}
	if err := room.LockRoster(protocol.SeatRed, roster(protocol.SeatRed)); err != nil {
		t.Fatalf("lock red roster: %v", err)
	}

	seat, ok := room.RemovePresence("s-red")
	if !ok || seat != protocol.SeatRed {
		t.Fatalf("removed seat = %q, ok=%v", seat, ok)
	}
	if _, ok := room.Seats[protocol.SeatRed]; ok {
		t.Fatalf("red seat was not freed")
	}
	if _, ok := room.Rosters[protocol.SeatRed]; ok {
		t.Fatalf("red roster was not removed")
	}
}

func TestRoomRuntimeRemovePresenceAfterStartKeepsSeatForReconnect(t *testing.T) {
	room := readyRoom(t)
	seat, ok := room.RemovePresence("s-red")
	if !ok || seat != protocol.SeatRed {
		t.Fatalf("removed seat = %q, ok=%v", seat, ok)
	}
	if _, ok := room.Seats[protocol.SeatRed]; !ok {
		t.Fatalf("red seat should be retained after start")
	}
	if room.Seats[protocol.SeatRed].SessionID != "" {
		t.Fatalf("red session id should be cleared after leave")
	}
	state := room.RoomState("room-1")
	if state.RedUserID != "u-red" || state.RedConnected {
		t.Fatalf("unexpected red room state after leave: %+v", state)
	}

	seat, err := room.AssignSeat(SeatPresence{UserID: "u-red", SessionID: "s-red-2", Username: "red"}, protocol.SeatAuto)
	if err != nil {
		t.Fatalf("reconnect rejected: %v", err)
	}
	if seat != protocol.SeatRed {
		t.Fatalf("reconnected seat = %q, want red", seat)
	}
	if room.Seats[protocol.SeatRed].SessionID != "s-red-2" {
		t.Fatalf("red session id was not refreshed")
	}
}

func TestRoomRuntimeRemoveSpectator(t *testing.T) {
	room := readyRoom(t)
	if _, err := room.AssignSeat(SeatPresence{UserID: "u-spec", SessionID: "s-spec", Username: "spec"}, protocol.SeatSpectator); err != nil {
		t.Fatalf("assign spectator: %v", err)
	}
	if len(room.RoomState("room-1").Spectators) != 1 {
		t.Fatalf("spectator was not added")
	}
	seat, ok := room.RemovePresence("s-spec")
	if !ok || seat != protocol.SeatSpectator {
		t.Fatalf("removed seat = %q, ok=%v", seat, ok)
	}
	if len(room.RoomState("room-1").Spectators) != 0 {
		t.Fatalf("spectator was not removed")
	}
}

func TestRoomRuntimeRejectsJoinAfterTerminalStatus(t *testing.T) {
	room := readyRoom(t)
	room.Status = protocol.RoomStatusClosed

	if _, err := room.AssignSeat(SeatPresence{UserID: "u-red", SessionID: "s-red-2", Username: "red"}, protocol.SeatAuto); protocol.ErrorCodeOf(err) != protocol.ErrorInvalidState {
		t.Fatalf("closed room should reject same user reconnect, got %v", err)
	}
	if _, err := room.AssignSeat(SeatPresence{UserID: "u-spec", SessionID: "s-spec", Username: "spec"}, protocol.SeatSpectator); protocol.ErrorCodeOf(err) != protocol.ErrorInvalidState {
		t.Fatalf("closed room should reject spectator join, got %v", err)
	}

	room.Status = protocol.RoomStatusEnded
	if seat, err := room.AssignSeat(SeatPresence{UserID: "u-blue", SessionID: "s-blue-2", Username: "blue"}, protocol.SeatAuto); err != nil || seat != protocol.SeatBlue {
		t.Fatalf("ended room should allow same user reconnect for replay, seat=%q err=%v", seat, err)
	}
	if seat, err := room.AssignSeat(SeatPresence{UserID: "u-spec", SessionID: "s-spec", Username: "spec"}, protocol.SeatSpectator); err != nil || seat != protocol.SeatSpectator {
		t.Fatalf("ended room should allow spectator replay, seat=%q err=%v", seat, err)
	}
	if _, err := room.AssignSeat(SeatPresence{UserID: "u-new", SessionID: "s-new", Username: "new"}, protocol.SeatAuto); protocol.ErrorCodeOf(err) != protocol.ErrorInvalidState {
		t.Fatalf("ended room should reject new player join, got %v", err)
	}
}

func TestRoomRuntimeRoomStateIncludesReplayWatermarks(t *testing.T) {
	room := readyRoom(t)

	accepted, err := room.AcceptCommand(protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandWaitTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         "start-hash",
	})
	if err != nil {
		t.Fatalf("command rejected: %v", err)
	}

	state := room.RoomState("room-1")
	if state.LastAcceptedHash != "start-hash" {
		t.Fatalf("last accepted hash = %q, want start-hash", state.LastAcceptedHash)
	}
	if state.LastCompletedSeq != 0 || state.PendingServerSeq != accepted.ServerSeq {
		t.Fatalf("unexpected pending watermarks: completed=%d pending=%d", state.LastCompletedSeq, state.PendingServerSeq)
	}
	if state.NextServerSeq != 2 || state.CommandLogLength != 1 || state.CompletedResultCount != 0 {
		t.Fatalf("unexpected command watermarks: next=%d log=%d completed=%d", state.NextServerSeq, state.CommandLogLength, state.CompletedResultCount)
	}
	if state.ExpectedActiveSeat != protocol.SeatRed || state.ExpectedActiveNetID != "red:0" {
		t.Fatalf("unexpected expected active actor: %s %s", state.ExpectedActiveSeat, state.ExpectedActiveNetID)
	}

	postHash := testDigest("a", 20)
	recordResult(t, room, protocol.SeatRed, accepted.ServerSeq, postHash, protocol.SeatBlue, "blue:0")
	recordResult(t, room, protocol.SeatBlue, accepted.ServerSeq, postHash, protocol.SeatBlue, "blue:0")

	state = room.RoomState("room-1")
	if state.LastAcceptedHash != postHash {
		t.Fatalf("last accepted hash after checkpoint = %q, want %q", state.LastAcceptedHash, postHash)
	}
	if state.LastCompletedSeq != accepted.ServerSeq || state.PendingServerSeq != 0 {
		t.Fatalf("unexpected completed watermarks: completed=%d pending=%d", state.LastCompletedSeq, state.PendingServerSeq)
	}
	if state.CompletedResultCount != 1 || state.ExpectedActiveSeat != protocol.SeatBlue || state.ExpectedActiveNetID != "blue:0" {
		t.Fatalf("unexpected post-checkpoint state: %+v", state)
	}
}

func TestRoomStateIncludesRoomMetadataVersionAndCreatedAt(t *testing.T) {
	room := readyRoom(t)
	metadata := defaultRoomMetadata()
	metadata.Name = "Arena Room"
	metadata.Description = "MVP room"
	metadata.Scenario = 1
	metadata.MapSeed = 101
	metadata.CombatSeed = 202
	metadata.CreatedAt = "2026-07-03T00:00:00Z"
	room.Metadata = metadata

	state := room.RoomState("room-1")
	assertTimeoutPolicy(t, state.Metadata.TimeoutPolicy, DefaultCommandResultTimeoutTicks, DefaultAbandonGraceTicks, DefaultBattleEndReportTimeoutTicks)
	if state.Metadata != metadata {
		metadata.TimeoutPolicy = state.Metadata.TimeoutPolicy
		if state.Metadata != metadata {
			t.Fatalf("room state metadata = %+v, want %+v", state.Metadata, metadata)
		}
	}
	if state.Metadata.ProtocolVersion != protocol.ProtocolVersion ||
		state.Metadata.ServerVersion != protocol.ServerVersion ||
		state.Metadata.RuntimeVersion != protocol.RuntimeVersion ||
		state.Metadata.ResultContractVersion != protocol.ResultContractVersion ||
		state.Metadata.RankedPolicyVersion != protocol.RankedPolicyVersion ||
		state.Metadata.RankedMode != protocol.RankedModeDisabled ||
		state.Metadata.CreatedAt == "" {
		t.Fatalf("room state metadata version/createdAt invalid: %+v", state.Metadata)
	}
}

func TestRoomStateIncludesRuntimeTimeoutPolicySnapshot(t *testing.T) {
	room := readyRoom(t)
	room.CommandResultTimeoutTicks = 7
	room.AbandonGraceTicks = 8
	room.BattleEndReportTimeoutTicks = 9

	state := room.RoomState("room-1")
	assertTimeoutPolicy(t, state.Metadata.TimeoutPolicy, 7, 8, 9)
}

func TestRoomRuntimeRetainsSuccessfulCheckpointEvidenceForBothSeats(t *testing.T) {
	room := readyRoom(t)
	accepted, err := room.AcceptCommand(protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandWaitTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         "start-hash",
	})
	if err != nil {
		t.Fatalf("command rejected: %v", err)
	}

	postHash := testDigest("a", 20)
	redSummary := testPostSummaryCanonical(postHash, accepted.ServerSeq, "state-hash-v1|seat=red")
	blueSummary := testPostSummaryCanonical(postHash, accepted.ServerSeq, "state-hash-v1|seat=blue")
	if err := room.RecordCommandResult(protocol.CommandResult{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ServerSeq:       accepted.ServerSeq,
		Seat:            protocol.SeatRed,
		PostHash:        postHash,
		PostSummary:     redSummary,
		NextActiveSeat:  protocol.SeatBlue,
		NextActiveNetID: "blue:0",
		Round:           1,
		Turn:            1,
	}); err != nil {
		t.Fatalf("red result rejected: %v", err)
	}
	if err := room.RecordCommandResult(protocol.CommandResult{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ServerSeq:       accepted.ServerSeq,
		Seat:            protocol.SeatBlue,
		PostHash:        postHash,
		PostSummary:     blueSummary,
		NextActiveSeat:  protocol.SeatBlue,
		NextActiveNetID: "blue:0",
		Round:           1,
		Turn:            1,
	}); err != nil {
		t.Fatalf("blue result rejected: %v", err)
	}

	completed, ok := room.CompletedCheckpoint(accepted.ServerSeq)
	if !ok {
		t.Fatalf("completed checkpoint was not retained")
	}
	if completed.Command.ServerSeq != accepted.ServerSeq {
		t.Fatalf("completed command seq = %d, want %d", completed.Command.ServerSeq, accepted.ServerSeq)
	}
	if string(completed.Results[protocol.SeatRed].PostSummary) != string(redSummary) {
		t.Fatalf("red post summary was not retained")
	}
	if string(completed.Results[protocol.SeatBlue].PostSummary) != string(blueSummary) {
		t.Fatalf("blue post summary was not retained")
	}
	if string(completed.Results[protocol.SeatRed].PostSummary) == string(completed.Results[protocol.SeatBlue].PostSummary) {
		t.Fatalf("test expected distinct red/blue post summaries to remain distinguishable")
	}
	replayResult, ok := room.CompletedResult(accepted.ServerSeq)
	if !ok {
		t.Fatalf("completed replay result was not retained")
	}
	if replayResult.Seat != protocol.SeatRed || replayResult.PostHash != postHash {
		t.Fatalf("unexpected replay result: %+v", replayResult)
	}
}

func TestRoomRuntimeCommandResultTimeoutDesyncsPendingCheckpoint(t *testing.T) {
	room := readyRoom(t)
	room.CommandResultTimeoutTicks = 3
	accepted, err := room.AcceptCommandAt(protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandWaitTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         "start-hash",
	}, 10)
	if err != nil {
		t.Fatalf("command rejected: %v", err)
	}
	pending := room.PendingCheckpoints[accepted.ServerSeq]
	if pending.AcceptedTick != 10 || pending.DeadlineTick != 13 {
		t.Fatalf("unexpected checkpoint deadline: %+v", pending)
	}
	postHash := testDigest("a", 20)
	recordResult(t, room, protocol.SeatRed, accepted.ServerSeq, postHash, protocol.SeatBlue, "blue:0")

	if seq, err := room.ExpirePendingCheckpoints(12); err != nil || seq != 0 {
		t.Fatalf("checkpoint expired before deadline: seq=%d err=%v", seq, err)
	}
	seq, err := room.ExpirePendingCheckpoints(13)
	if protocol.ErrorCodeOf(err) != protocol.ErrorDesync {
		t.Fatalf("checkpoint timeout should desync, seq=%d err=%v", seq, err)
	}
	if seq != accepted.ServerSeq {
		t.Fatalf("expired seq = %d, want %d", seq, accepted.ServerSeq)
	}
	if room.Status != protocol.RoomStatusDesynced {
		t.Fatalf("room status = %q, want Desynced", room.Status)
	}
	if _, ok := room.CompletedResult(accepted.ServerSeq); ok {
		t.Fatalf("timed out checkpoint must not be completed as success")
	}
	if _, ok := room.CompletedCheckpoint(accepted.ServerSeq); ok {
		t.Fatalf("timed out checkpoint must not create successful evidence")
	}
	if len(room.PendingCheckpoints) != 1 {
		t.Fatalf("timed out checkpoint evidence should remain pending, got %d", len(room.PendingCheckpoints))
	}
}

func TestRoomRuntimeForfeitEndsRoomWithoutRankedResolution(t *testing.T) {
	room := readyRoom(t)
	result, err := room.Forfeit("match-1", protocol.SeatRed, "u-red", protocol.MatchResultReasonForfeit)
	if err != nil {
		t.Fatalf("forfeit rejected: %v", err)
	}
	if room.Status != protocol.RoomStatusEnded {
		t.Fatalf("room status = %q, want Ended", room.Status)
	}
	if result.WinnerSeat != protocol.SeatBlue || result.LoserSeat != protocol.SeatRed {
		t.Fatalf("unexpected forfeit winner/loser: %+v", result)
	}
	assertMVPUnrankedResult(t, result)
	assertCheckpointBinding(t, result, "start-hash", 0, "start-hash", 0, false)
	if room.TerminalResult == nil || room.TerminalResult.WinnerSeat != protocol.SeatBlue {
		t.Fatalf("terminal result was not retained: %+v", room.TerminalResult)
	}
	state := room.RoomState("match-1")
	if state.Result == nil || state.Result.LoserSeat != protocol.SeatRed {
		t.Fatalf("room state did not include terminal result: %+v", state)
	}
}

func TestRoomRuntimeForfeitClearsPendingCheckpoint(t *testing.T) {
	room := readyRoom(t)
	accepted, err := room.AcceptCommand(protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandWaitTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         "start-hash",
	})
	if err != nil {
		t.Fatalf("command rejected: %v", err)
	}
	if accepted.ServerSeq != 1 || len(room.PendingCheckpoints) != 1 {
		t.Fatalf("test expected one pending checkpoint")
	}
	result, err := room.Forfeit("match-1", protocol.SeatRed, "u-red", protocol.MatchResultReasonForfeit)
	if err != nil {
		t.Fatalf("forfeit rejected: %v", err)
	}
	assertCheckpointBinding(t, result, "start-hash", 0, "start-hash", accepted.ServerSeq, false)
	if len(room.PendingCheckpoints) != 0 {
		t.Fatalf("forfeit should clear pending checkpoints, got %d", len(room.PendingCheckpoints))
	}
}

func TestRoomRuntimeRecordsAndExpiresAbandonDeadline(t *testing.T) {
	room := readyRoom(t)
	room.AbandonGraceTicks = 3
	seat, ok := room.RemovePresenceAt("s-red", 10)
	if !ok || seat != protocol.SeatRed {
		t.Fatalf("remove red = %q ok=%v, want red true", seat, ok)
	}
	deadline := room.AbandonDeadlines[protocol.SeatRed]
	if deadline.AcceptedTick != 10 || deadline.DeadlineTick != 13 || deadline.UserID != "u-red" {
		t.Fatalf("unexpected abandon deadline: %+v", deadline)
	}
	if result, expired, err := room.ExpireAbandonedSeats("match-1", 12); err != nil || expired || result.MatchID != "" {
		t.Fatalf("abandon expired before deadline: result=%+v expired=%v err=%v", result, expired, err)
	}
	result, expired, err := room.ExpireAbandonedSeats("match-1", 13)
	if err != nil || !expired {
		t.Fatalf("abandon should expire at deadline: result=%+v expired=%v err=%v", result, expired, err)
	}
	if result.Reason != protocol.MatchResultReasonAbandon ||
		result.WinnerSeat != protocol.SeatBlue ||
		result.LoserSeat != protocol.SeatRed ||
		result.ResolvedAtTick != 13 {
		t.Fatalf("unexpected abandon result: %+v", result)
	}
	assertMVPUnrankedResult(t, result)
	assertCheckpointBinding(t, result, "start-hash", 0, "start-hash", 0, false)
	if room.Status != protocol.RoomStatusEnded {
		t.Fatalf("room status = %q, want Ended", room.Status)
	}
}

func TestRoomRuntimeReconnectClearsAbandonDeadline(t *testing.T) {
	room := readyRoom(t)
	room.AbandonGraceTicks = 3
	if _, ok := room.RemovePresenceAt("s-red", 10); !ok {
		t.Fatalf("remove red failed")
	}
	if _, err := room.AssignSeat(SeatPresence{UserID: "u-red", SessionID: "s-red-2", Username: "red"}, protocol.SeatAuto); err != nil {
		t.Fatalf("red reconnect rejected: %v", err)
	}
	if _, ok := room.AbandonDeadlines[protocol.SeatRed]; ok {
		t.Fatalf("red reconnect should clear abandon deadline")
	}
	if result, expired, err := room.ExpireAbandonedSeats("match-1", 13); err != nil || expired || result.MatchID != "" {
		t.Fatalf("reconnected player should not abandon: result=%+v expired=%v err=%v", result, expired, err)
	}
}

func TestRoomRuntimeDesyncDoesNotExpireAbandonAsResult(t *testing.T) {
	room := readyRoom(t)
	room.AbandonGraceTicks = 3
	if _, ok := room.RemovePresenceAt("s-red", 10); !ok {
		t.Fatalf("remove red failed")
	}
	room.Status = protocol.RoomStatusDesynced
	if result, expired, err := room.ExpireAbandonedSeats("match-1", 13); err != nil || expired || result.MatchID != "" {
		t.Fatalf("desynced room must not convert abandon to result: result=%+v expired=%v err=%v", result, expired, err)
	}
}

func TestRoomRuntimeBattleEndRequiresBothReports(t *testing.T) {
	room := readyRoom(t)
	room.BattleEndReportTimeoutTicks = 3
	result, resolved, err := room.RecordBattleEndReport(battleEndReport(room, protocol.SeatRed, protocol.SeatRed, protocol.SeatBlue, "battle-end-red-1"), 10)
	if err != nil || resolved || result.MatchID != "" {
		t.Fatalf("single battle end report should wait: result=%+v resolved=%v err=%v", result, resolved, err)
	}
	if room.Status != protocol.RoomStatusInBattle {
		t.Fatalf("single report should keep room in battle, got %q", room.Status)
	}
	if err := room.ExpireBattleEndReports(12); err != nil {
		t.Fatalf("single report expired before deadline: %v", err)
	}
	if err := room.ExpireBattleEndReports(13); protocol.ErrorCodeOf(err) != protocol.ErrorDesync {
		t.Fatalf("single report should desync at deadline, got %v", err)
	}
	if room.Status != protocol.RoomStatusDesynced {
		t.Fatalf("expired battle end report status = %q, want Desynced", room.Status)
	}
}

func TestRoomRuntimeBattleEndResolvesMatchingReports(t *testing.T) {
	room := readyRoom(t)
	if _, resolved, err := room.RecordBattleEndReport(battleEndReport(room, protocol.SeatRed, protocol.SeatRed, protocol.SeatBlue, "battle-end-red-1"), 10); err != nil || resolved {
		t.Fatalf("first report should wait: resolved=%v err=%v", resolved, err)
	}
	result, resolved, err := room.RecordBattleEndReport(battleEndReport(room, protocol.SeatBlue, protocol.SeatRed, protocol.SeatBlue, "battle-end-blue-1"), 11)
	if err != nil || !resolved {
		t.Fatalf("matching reports should resolve: result=%+v resolved=%v err=%v", result, resolved, err)
	}
	if result.Reason != protocol.MatchResultReasonBattleEnd ||
		result.WinnerSeat != protocol.SeatRed ||
		result.LoserSeat != protocol.SeatBlue ||
		result.SourceSeat != protocol.SeatRed {
		t.Fatalf("unexpected battle end result: %+v", result)
	}
	assertMVPUnrankedResult(t, result)
	assertCheckpointBinding(t, result, "start-hash", 0, "start-hash", 0, false)
	if result.ReportClientRequestIDs[protocol.SeatRed] != "battle-end-red-1" ||
		result.ReportClientRequestIDs[protocol.SeatBlue] != "battle-end-blue-1" {
		t.Fatalf("report request ids were not retained: %+v", result.ReportClientRequestIDs)
	}
	if room.Status != protocol.RoomStatusEnded || room.TerminalResult == nil {
		t.Fatalf("battle end should make terminal room: status=%q result=%+v", room.Status, room.TerminalResult)
	}
}

func TestRoomRuntimeBattleEndConflictDesyncs(t *testing.T) {
	room := readyRoom(t)
	if _, resolved, err := room.RecordBattleEndReport(battleEndReport(room, protocol.SeatRed, protocol.SeatRed, protocol.SeatBlue, ""), 10); err != nil || resolved {
		t.Fatalf("first report should wait: resolved=%v err=%v", resolved, err)
	}
	if result, resolved, err := room.RecordBattleEndReport(battleEndReport(room, protocol.SeatBlue, protocol.SeatBlue, protocol.SeatRed, ""), 11); protocol.ErrorCodeOf(err) != protocol.ErrorDesync || resolved || result.MatchID != "" {
		t.Fatalf("conflicting reports should desync: result=%+v resolved=%v err=%v", result, resolved, err)
	}
	if room.Status != protocol.RoomStatusDesynced {
		t.Fatalf("conflict status = %q, want Desynced", room.Status)
	}
}

func TestRoomRuntimeBattleEndWatermarkMismatchDesyncs(t *testing.T) {
	room := readyRoom(t)
	report := battleEndReport(room, protocol.SeatRed, protocol.SeatRed, protocol.SeatBlue, "battle-end-red-1")
	report.LastHash = "stale-hash"
	if result, resolved, err := room.RecordBattleEndReport(report, 10); protocol.ErrorCodeOf(err) != protocol.ErrorDesync || resolved || result.MatchID != "" {
		t.Fatalf("hash mismatch should desync: result=%+v resolved=%v err=%v", result, resolved, err)
	}
	if room.Status != protocol.RoomStatusDesynced {
		t.Fatalf("hash mismatch status = %q, want Desynced", room.Status)
	}
}

func TestRoomRuntimeBattleEndSeqMismatchDesyncs(t *testing.T) {
	room := readyRoom(t)
	report := battleEndReport(room, protocol.SeatRed, protocol.SeatRed, protocol.SeatBlue, "battle-end-red-1")
	report.LastCompletedSeq = 99
	if result, resolved, err := room.RecordBattleEndReport(report, 10); protocol.ErrorCodeOf(err) != protocol.ErrorDesync || resolved || result.MatchID != "" {
		t.Fatalf("seq mismatch should desync: result=%+v resolved=%v err=%v", result, resolved, err)
	}
	if room.Status != protocol.RoomStatusDesynced {
		t.Fatalf("seq mismatch status = %q, want Desynced", room.Status)
	}
}

func TestRoomRuntimeBattleEndWaitsForPendingCheckpoint(t *testing.T) {
	room := readyRoom(t)
	room.BattleEndReportTimeoutTicks = 3
	accepted, err := room.AcceptCommand(protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandWaitTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         "start-hash",
	})
	if err != nil {
		t.Fatalf("command rejected: %v", err)
	}
	redReport := battleEndReportAt(room, protocol.SeatRed, protocol.SeatRed, protocol.SeatBlue, "battle-end-red-1", accepted.ServerSeq)
	blueReport := battleEndReportAt(room, protocol.SeatBlue, protocol.SeatRed, protocol.SeatBlue, "battle-end-blue-1", accepted.ServerSeq)
	if _, resolved, err := room.RecordBattleEndReport(redReport, 10); err != nil || resolved {
		t.Fatalf("first report should wait: resolved=%v err=%v", resolved, err)
	}
	if _, resolved, err := room.RecordBattleEndReport(blueReport, 11); err != nil || resolved {
		t.Fatalf("matching reports should wait for pending checkpoint: resolved=%v err=%v", resolved, err)
	}
	if err := room.ExpireBattleEndReports(14); err != nil {
		t.Fatalf("matching reports should not timeout while waiting for pending checkpoint: %v", err)
	}
	if room.Status != protocol.RoomStatusInBattle {
		t.Fatalf("matching reports should keep waiting in battle, got %q", room.Status)
	}

	postHash := testDigest("b", 20)
	redResult := protocol.CommandResult{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ServerSeq:       accepted.ServerSeq,
		Seat:            protocol.SeatRed,
		PostHash:        postHash,
		PostSummary:     testPostSummary(postHash, accepted.ServerSeq),
		NextActiveSeat:  protocol.SeatBlue,
		NextActiveNetID: "blue:0",
		Round:           1,
		Turn:            2,
	}
	blueResult := redResult
	blueResult.Seat = protocol.SeatBlue
	if err := room.RecordCommandResult(redResult); err != nil {
		t.Fatalf("red result rejected: %v", err)
	}
	if err := room.RecordCommandResult(blueResult); err != nil {
		t.Fatalf("blue result rejected: %v", err)
	}
	result, resolved, err := room.RecordBattleEndReport(redReport, 12)
	if err != nil || !resolved {
		t.Fatalf("duplicate report should resolve after pending checkpoint: result=%+v resolved=%v err=%v", result, resolved, err)
	}
	if result.PendingServerSeq != 0 || result.LastCompletedSeq != accepted.ServerSeq {
		t.Fatalf("unexpected battle end watermarks: %+v", result)
	}
	assertCheckpointBinding(t, result, "start-hash", accepted.ServerSeq, postHash, 0, false)
}

func TestRoomRuntimeTerminalCheckpointCompletesWithoutNextActive(t *testing.T) {
	room := readyRoom(t)
	accepted, err := room.AcceptCommand(protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandWaitTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         "start-hash",
	})
	if err != nil {
		t.Fatalf("command rejected: %v", err)
	}
	postHash := testDigest("c", 20)
	redResult := protocol.CommandResult{
		SchemaVersion: protocol.SchemaVersion,
		MatchID:       "match-1",
		ServerSeq:     accepted.ServerSeq,
		Seat:          protocol.SeatRed,
		PostHash:      postHash,
		PostSummary:   testPostSummary(postHash, accepted.ServerSeq),
		Round:         1,
		Turn:          2,
		Terminal:      true,
	}
	blueResult := redResult
	blueResult.Seat = protocol.SeatBlue
	if err := room.RecordCommandResult(redResult); err != nil {
		t.Fatalf("red terminal result rejected: %v", err)
	}
	if err := room.RecordCommandResult(blueResult); err != nil {
		t.Fatalf("blue terminal result rejected: %v", err)
	}
	if len(room.PendingCheckpoints) != 0 {
		t.Fatalf("terminal checkpoint should complete, pending=%d", len(room.PendingCheckpoints))
	}
	if room.ExpectedActiveSeat != "" || room.ExpectedActiveNetID != "" {
		t.Fatalf("terminal checkpoint should clear expected active, seat=%q netId=%q", room.ExpectedActiveSeat, room.ExpectedActiveNetID)
	}
	completed, ok := room.CompletedResult(accepted.ServerSeq)
	if !ok || !completed.Terminal {
		t.Fatalf("terminal completed result missing: %+v ok=%v", completed, ok)
	}
}

func TestRoomRuntimeTerminalCheckpointMismatchDesyncs(t *testing.T) {
	room := readyRoom(t)
	accepted, err := room.AcceptCommand(protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandWaitTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         "start-hash",
	})
	if err != nil {
		t.Fatalf("command rejected: %v", err)
	}
	postHash := testDigest("d", 20)
	redResult := protocol.CommandResult{
		SchemaVersion: protocol.SchemaVersion,
		MatchID:       "match-1",
		ServerSeq:     accepted.ServerSeq,
		Seat:          protocol.SeatRed,
		PostHash:      postHash,
		PostSummary:   testPostSummary(postHash, accepted.ServerSeq),
		Round:         1,
		Turn:          2,
		Terminal:      true,
	}
	blueResult := redResult
	blueResult.Seat = protocol.SeatBlue
	blueResult.Terminal = false
	blueResult.NextActiveSeat = protocol.SeatBlue
	blueResult.NextActiveNetID = "blue:0"
	if err := room.RecordCommandResult(redResult); err != nil {
		t.Fatalf("red terminal result rejected: %v", err)
	}
	if err := room.RecordCommandResult(blueResult); protocol.ErrorCodeOf(err) != protocol.ErrorDesync {
		t.Fatalf("terminal mismatch should desync, got %v", err)
	}
	if room.Status != protocol.RoomStatusDesynced || room.LastMismatch == nil {
		t.Fatalf("terminal mismatch should retain desync evidence: status=%q mismatch=%+v", room.Status, room.LastMismatch)
	}
}

func TestRoomRuntimeDetectsCheckpointMismatch(t *testing.T) {
	room := readyRoom(t)

	accepted, err := room.AcceptCommand(protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandEndTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         "start-hash",
	})
	if err != nil {
		t.Fatalf("command rejected: %v", err)
	}

	redHash := testDigest("c", 21)
	blueHash := testDigest("d", 22)
	recordResult(t, room, protocol.SeatRed, accepted.ServerSeq, redHash, protocol.SeatBlue, "blue:0")
	err = room.RecordCommandResult(protocol.CommandResult{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ServerSeq:       accepted.ServerSeq,
		Seat:            protocol.SeatBlue,
		PostHash:        blueHash,
		PostSummary:     testPostSummary(blueHash, accepted.ServerSeq),
		NextActiveSeat:  protocol.SeatBlue,
		NextActiveNetID: "blue:0",
		Round:           1,
		Turn:            1,
	})
	if protocol.ErrorCodeOf(err) != protocol.ErrorDesync {
		t.Fatalf("checkpoint mismatch should desync, got %v", err)
	}
	if room.Status != protocol.RoomStatusDesynced {
		t.Fatalf("room status = %q, want Desynced", room.Status)
	}
	if room.LastMismatch == nil {
		t.Fatalf("checkpoint mismatch evidence was not recorded")
	}
	if string(room.LastMismatch.BlueResult.PostSummary) == "" {
		t.Fatalf("blue post summary was not retained in mismatch evidence")
	}
}

func TestRoomRuntimeRejectsDuplicateCommandResultFromSameSeat(t *testing.T) {
	room := readyRoom(t)
	accepted, err := room.AcceptCommand(protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandWaitTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         "start-hash",
	})
	if err != nil {
		t.Fatalf("command rejected: %v", err)
	}

	result := protocol.CommandResult{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ServerSeq:       accepted.ServerSeq,
		Seat:            protocol.SeatRed,
		PostHash:        testDigest("e", 20),
		PostSummary:     testPostSummary(testDigest("e", 20), accepted.ServerSeq),
		NextActiveSeat:  protocol.SeatBlue,
		NextActiveNetID: "blue:0",
		Round:           1,
		Turn:            1,
	}
	if err := room.RecordCommandResult(result); err != nil {
		t.Fatalf("first result rejected: %v", err)
	}
	if err := room.RecordCommandResult(result); protocol.ErrorCodeOf(err) != protocol.ErrorInvalidCommand {
		t.Fatalf("duplicate result should be rejected, got %v", err)
	}
}

func TestRoomRuntimeRejectsCommandResultForWrongMatch(t *testing.T) {
	room := readyRoom(t)
	accepted, err := room.AcceptCommand(protocol.CommandEnvelope{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            protocol.SeatRed,
		ActiveNetID:     "red:0",
		Op:              protocol.CommandWaitTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         "start-hash",
	})
	if err != nil {
		t.Fatalf("command rejected: %v", err)
	}

	err = room.RecordCommandResult(protocol.CommandResult{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "other-match",
		ServerSeq:       accepted.ServerSeq,
		Seat:            protocol.SeatRed,
		PostHash:        testDigest("f", 20),
		NextActiveSeat:  protocol.SeatBlue,
		NextActiveNetID: "blue:0",
		Round:           1,
		Turn:            1,
	})
	if protocol.ErrorCodeOf(err) != protocol.ErrorInvalidCommand {
		t.Fatalf("wrong match result should be rejected, got %v", err)
	}
}

func readyRoom(t *testing.T) *RoomRuntime {
	t.Helper()

	room := NewRoomRuntime("start-hash")
	if _, err := room.AssignSeat(SeatPresence{UserID: "u-red", SessionID: "s-red", Username: "red"}, protocol.SeatRed); err != nil {
		t.Fatalf("assign red: %v", err)
	}
	if _, err := room.AssignSeat(SeatPresence{UserID: "u-blue", SessionID: "s-blue", Username: "blue"}, protocol.SeatBlue); err != nil {
		t.Fatalf("assign blue: %v", err)
	}
	if err := room.LockRoster(protocol.SeatRed, roster(protocol.SeatRed)); err != nil {
		t.Fatalf("lock red roster: %v", err)
	}
	if err := room.LockRoster(protocol.SeatBlue, roster(protocol.SeatBlue)); err != nil {
		t.Fatalf("lock blue roster: %v", err)
	}
	if err := room.Start("start-hash", protocol.SeatRed, "red:0"); err != nil {
		t.Fatalf("start room: %v", err)
	}
	return room
}

func battleEndReport(room *RoomRuntime, seat protocol.Seat, winner protocol.Seat, loser protocol.Seat, requestID string) protocol.MatchBattleEndReport {
	return battleEndReportAt(room, seat, winner, loser, requestID, room.LastCompletedSeq())
}

func battleEndReportAt(room *RoomRuntime, seat protocol.Seat, winner protocol.Seat, loser protocol.Seat, requestID string, lastCompletedSeq int64) protocol.MatchBattleEndReport {
	return protocol.MatchBattleEndReport{
		SchemaVersion:    protocol.SchemaVersion,
		MatchID:          "match-1",
		Seat:             seat,
		WinnerSeat:       winner,
		LoserSeat:        loser,
		ClientRequestID:  requestID,
		LastCompletedSeq: lastCompletedSeq,
		LastHash:         room.LastAcceptedHash,
	}
}

func roster(seat protocol.Seat) protocol.RosterSnapshot {
	return protocol.RosterSnapshot{
		SchemaVersion:       protocol.SchemaVersion,
		CodecVersion:        1,
		GameSHA256:          strings.Repeat("a", 64),
		GameVersion:         "1.5.1.8",
		ModVersion:          "mvp",
		RulesetHash:         "rules",
		OwnerUserID:         fmt.Sprintf("u-%s", seat),
		Seat:                seat,
		Entries:             rosterEntries(seat),
		HashAlgorithm:       protocol.RosterHashAlgorithmSHA256,
		RosterHash:          testDigest("a", len(seat)+7),
		VisualStatusSummary: "unknown",
	}
}

func rosterEntries(seat protocol.Seat) json.RawMessage {
	return json.RawMessage(fmt.Sprintf(`[
		{
			"netId": "%[1]s:0",
			"ownerSeat": "%[1]s",
			"entryIndex": 0,
			"scriptName": "scripts/entity/tactical/player",
			"placeInFormation": 0,
			"faction": 1,
			"logicState": {
				"source": "squirrel-onSerialize-bufferio",
				"schemaVersion": 1,
				"bufferEncoding": "base64",
				"bufferSize": 4,
				"buffer": "AAAA",
				"bufferChecksum": "sha256:0000000000000000000000000000000000000000000000000000000000000000:len:4",
				"logicHash": "sha256:1111111111111111111111111111111111111111111111111111111111111111:len:10",
				"decodeStatus": "not-validated",
				"decodePolicy": "reject-match-start-on-failure"
			},
			"visualState": {
				"source": "derived-v1",
				"status": "unknown",
				"nativeCoverage": "none"
			},
			"entryHash": "sha256:2222222222222222222222222222222222222222222222222222222222222222:len:20"
		}
	]`, seat))
}

func testDigest(hex string, length int) string {
	return fmt.Sprintf("sha256:%s:len:%d", strings.Repeat(hex, 64), length)
}

func recordResult(t *testing.T, room *RoomRuntime, seat protocol.Seat, seq int64, postHash string, nextSeat protocol.Seat, nextNetID string) {
	t.Helper()

	if err := room.RecordCommandResult(protocol.CommandResult{
		SchemaVersion:   protocol.SchemaVersion,
		MatchID:         "match-1",
		ServerSeq:       seq,
		Seat:            seat,
		PostHash:        postHash,
		PostSummary:     testPostSummary(postHash, seq),
		NextActiveSeat:  nextSeat,
		NextActiveNetID: nextNetID,
		Round:           1,
		Turn:            1,
	}); err != nil {
		t.Fatalf("record result for %s: %v", seat, err)
	}
}

func testPostSummary(digest string, serverSeq int64) json.RawMessage {
	return testPostSummaryForMatch("match-1", digest, serverSeq)
}

func testPostSummaryForMatch(matchID string, digest string, serverSeq int64) json.RawMessage {
	return json.RawMessage(fmt.Sprintf(`{"schemaVersion":1,"algorithm":"sha256-state-v1","matchId":"%s","lastServerSeq":%d,"canonical":"state-hash-v1|x","digest":"%s"}`, matchID, serverSeq, digest))
}

func testPostSummaryCanonical(digest string, serverSeq int64, canonical string) json.RawMessage {
	return json.RawMessage(fmt.Sprintf(`{"schemaVersion":1,"algorithm":"sha256-state-v1","matchId":"match-1","lastServerSeq":%d,"canonical":"%s","digest":"%s"}`, serverSeq, canonical, digest))
}

func assertMVPUnrankedResult(t *testing.T, result protocol.MatchResultEvent) {
	t.Helper()
	if err := protocol.ValidateMatchResultEvent(result); err != nil {
		t.Fatalf("match result violates MVP result contract: %+v err=%v", result, err)
	}
	if result.RankedEligible || result.RankedReason != protocol.MatchResultRankedReasonMVPDisabled {
		t.Fatalf("match result should be MVP-unranked: %+v", result)
	}
}

func assertCheckpointBinding(t *testing.T, result protocol.MatchResultEvent, startHash string, lastCompletedSeq int64, lastHash string, pendingServerSeq int64, terminal bool) {
	t.Helper()
	if err := protocol.ValidateMatchCheckpointBinding(result.CheckpointBinding); err != nil {
		t.Fatalf("checkpoint binding invalid: %+v err=%v", result.CheckpointBinding, err)
	}
	if result.CheckpointBinding.MatchID != result.MatchID ||
		result.CheckpointBinding.StartHash != startHash ||
		result.CheckpointBinding.LastCompletedSeq != lastCompletedSeq ||
		result.CheckpointBinding.LastHash != lastHash ||
		result.CheckpointBinding.PendingServerSeq != pendingServerSeq ||
		result.CheckpointBinding.TerminalCheckpoint != terminal {
		t.Fatalf("unexpected checkpoint binding: %+v", result.CheckpointBinding)
	}
}

func assertTimeoutPolicy(t *testing.T, policy protocol.MatchTimeoutPolicy, commandTimeout int64, abandonGrace int64, battleEndTimeout int64) {
	t.Helper()
	if err := protocol.ValidateMatchTimeoutPolicy(policy); err != nil {
		t.Fatalf("timeout policy invalid: %+v err=%v", policy, err)
	}
	if policy.TickRate != MatchTickRate ||
		policy.CommandResultTimeoutTicks != commandTimeout ||
		policy.AbandonGraceTicks != abandonGrace ||
		policy.BattleEndReportTimeoutTicks != battleEndTimeout {
		t.Fatalf("unexpected timeout policy: %+v", policy)
	}
}

func assertParticipantSnapshot(t *testing.T, participants protocol.MatchParticipants, matchID string) {
	t.Helper()
	if err := protocol.ValidateMatchParticipants(participants); err != nil {
		t.Fatalf("participant snapshot invalid: %+v err=%v", participants, err)
	}
	if participants.MatchID != matchID ||
		participants.Red.Seat != protocol.SeatRed ||
		participants.Red.UserID != "u-red" ||
		participants.Red.RosterHash != roster(protocol.SeatRed).RosterHash ||
		participants.Blue.Seat != protocol.SeatBlue ||
		participants.Blue.UserID != "u-blue" ||
		participants.Blue.RosterHash != roster(protocol.SeatBlue).RosterHash ||
		participants.GameSHA256 != strings.Repeat("a", 64) ||
		participants.GameVersion != "1.5.1.8" ||
		participants.ModVersion != "mvp" ||
		participants.RulesetHash != "rules" {
		t.Fatalf("unexpected participant snapshot: %+v", participants)
	}
}
