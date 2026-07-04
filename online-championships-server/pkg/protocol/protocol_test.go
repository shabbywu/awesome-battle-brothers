package protocol

import (
	"encoding/json"
	"strconv"
	"testing"
)

func TestOpcodeWhitelist(t *testing.T) {
	if !IsKnownOpcode(OpCodeCmdWaitTurn) {
		t.Fatalf("CMD_WAIT_TURN opcode should be known")
	}
	if !IsKnownOpcode(OpCodeProtocolError) {
		t.Fatalf("PROTOCOL_ERROR opcode should be known")
	}
	if IsKnownOpcode(9999) {
		t.Fatalf("unknown opcode should not be accepted")
	}
}

func TestValidateClientCommand(t *testing.T) {
	cmd := CommandEnvelope{
		SchemaVersion:   SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            SeatRed,
		ActiveNetID:     "red:0",
		Op:              CommandWaitTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         "start-hash",
	}
	if err := ValidateClientCommand(cmd); err != nil {
		t.Fatalf("valid command rejected: %v", err)
	}

	cmd.SchemaVersion = 0
	if err := ValidateClientCommand(cmd); ErrorCodeOf(err) != ErrorSchemaMismatch {
		t.Fatalf("schema mismatch should be rejected, got %v", err)
	}
}

func TestValidateMatchForfeitRequestRejectsClientAbandon(t *testing.T) {
	request := MatchForfeitRequest{
		SchemaVersion: SchemaVersion,
		MatchID:       "match-1",
		Seat:          SeatRed,
		Reason:        MatchResultReasonAbandon,
	}
	if err := ValidateMatchForfeitRequest(request); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("client abandon reason should be rejected, got %v", err)
	}
}

func TestValidateMatchBattleEndReport(t *testing.T) {
	report := MatchBattleEndReport{
		SchemaVersion: SchemaVersion,
		MatchID:       "match-1",
		Seat:          SeatRed,
		WinnerSeat:    SeatRed,
		LoserSeat:     SeatBlue,
	}
	if err := ValidateMatchBattleEndReport(report); err != nil {
		t.Fatalf("valid battle end report rejected: %v", err)
	}

	report.WinnerSeat = SeatBlue
	report.LoserSeat = SeatBlue
	if err := ValidateMatchBattleEndReport(report); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("matching winner/loser should be rejected, got %v", err)
	}
}

func TestValidateMatchResultEventRequiresMVPUnrankedContract(t *testing.T) {
	result := validMatchResultEvent()
	if err := ValidateMatchResultEvent(result); err != nil {
		t.Fatalf("valid MVP result rejected: %v", err)
	}

	result.RankedEligible = true
	if err := ValidateMatchResultEvent(result); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("ranked result should be rejected while MVP persistence is disabled, got %v", err)
	}

	result = validMatchResultEvent()
	result.RankedReason = ""
	if err := ValidateMatchResultEvent(result); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("unranked result without rankedReason should be rejected, got %v", err)
	}

	result = validMatchResultEvent()
	result.ResultID = "sha256:0000000000000000000000000000000000000000000000000000000000000000:len:1"
	if err := ValidateMatchResultEvent(result); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("mismatched resultId should be rejected, got %v", err)
	}

	result = validMatchResultEvent()
	result.Participants.Red.UserID = "u-red-2"
	if err := ValidateMatchResultEvent(result); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("resultId should bind participant snapshot, got %v", err)
	}

	result = validMatchResultEvent()
	result.TimeoutPolicy.AbandonGraceTicks = 999
	if err := ValidateMatchResultEvent(result); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("resultId should bind timeout policy snapshot, got %v", err)
	}

	result = validMatchResultEvent()
	result.CheckpointBinding.LastHash = "start-hash"
	if err := ValidateMatchResultEvent(result); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("resultId should bind checkpoint snapshot, got %v", err)
	}

	result = validMatchResultEvent()
	result.Participants.Red.UserID = ""
	result.ResultID = ComputeMatchResultID(result)
	if err := ValidateMatchResultEvent(result); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("missing participant snapshot should be rejected, got %v", err)
	}

	result = validMatchResultEvent()
	result.TimeoutPolicy.CommandResultTimeoutTicks = 0
	result.ResultID = ComputeMatchResultID(result)
	if err := ValidateMatchResultEvent(result); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("missing timeout policy snapshot should be rejected, got %v", err)
	}
}

func TestMatchResultReplayFlagDoesNotAffectResultID(t *testing.T) {
	result := validMatchResultEvent()
	resultID := result.ResultID
	result.Replay = true
	if result.ResultID != resultID {
		t.Fatalf("test mutated result id unexpectedly")
	}
	if got := ComputeMatchResultID(result); got != resultID {
		t.Fatalf("replay flag changed resultId: got %q want %q", got, resultID)
	}
	if err := ValidateMatchResultEvent(result); err != nil {
		t.Fatalf("replayed server result should keep valid result identity: %v", err)
	}
}

func TestValidateClientCommandRejectsClientServerSeq(t *testing.T) {
	cmd := CommandEnvelope{
		SchemaVersion:   SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		ServerSeq:       10,
		Seat:            SeatRed,
		ActiveNetID:     "red:0",
		Op:              CommandEndTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         "start-hash",
	}
	if err := ValidateClientCommand(cmd); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("client supplied serverSeq should be rejected, got %v", err)
	}
}

func TestValidateClientCommandRejectsReplayFlag(t *testing.T) {
	cmd := CommandEnvelope{
		SchemaVersion:   SchemaVersion,
		MatchID:         "match-1",
		ClientCommandID: "client-1",
		Seat:            SeatRed,
		ActiveNetID:     "red:0",
		Op:              CommandWaitTurn,
		Payload:         json.RawMessage(`{}`),
		PreHash:         "start-hash",
		Replay:          true,
	}
	if err := ValidateClientCommand(cmd); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("client replay command should be rejected, got %v", err)
	}
}

func TestValidateClientCommandRejectsReplayRequiresResultFlag(t *testing.T) {
	cmd := CommandEnvelope{
		SchemaVersion:        SchemaVersion,
		MatchID:              "match-1",
		ClientCommandID:      "client-1",
		Seat:                 SeatRed,
		ActiveNetID:          "red:0",
		Op:                   CommandWaitTurn,
		Payload:              json.RawMessage(`{}`),
		PreHash:              "start-hash",
		ReplayRequiresResult: true,
	}
	if err := ValidateClientCommand(cmd); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("client replayRequiresResult command should be rejected, got %v", err)
	}
}

func TestValidateCommandResultRejectsReplayFlag(t *testing.T) {
	result := validCommandResult()
	result.Replay = true
	if err := ValidateCommandResult(result); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("client replay result should be rejected, got %v", err)
	}
}

func TestValidateCommandResultRejectsInvalidPostSummary(t *testing.T) {
	result := validCommandResult()
	result.PostSummary = json.RawMessage(`{invalid`)
	if err := ValidateCommandResult(result); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("invalid postSummary should be rejected, got %v", err)
	}
}

func TestValidateCommandResultRejectsMissingPostSummary(t *testing.T) {
	result := validCommandResult()
	result.PostSummary = nil
	if err := ValidateCommandResult(result); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("missing postSummary should be rejected, got %v", err)
	}
}

func TestValidateCommandResultRejectsPostSummaryDigestMismatch(t *testing.T) {
	result := validCommandResult()
	result.PostSummary = validPostSummary("sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb:len:20", result.ServerSeq)
	if err := ValidateCommandResult(result); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("postSummary digest mismatch should be rejected, got %v", err)
	}
}

func TestValidateCommandResultAllowsTerminalWithoutNextActive(t *testing.T) {
	result := validCommandResult()
	result.Terminal = true
	result.NextActiveSeat = ""
	result.NextActiveNetID = ""
	if err := ValidateCommandResult(result); err != nil {
		t.Fatalf("terminal result without next active rejected: %v", err)
	}
}

func TestValidateCommandResultRejectsNonTerminalWithoutNextActive(t *testing.T) {
	result := validCommandResult()
	result.NextActiveSeat = ""
	result.NextActiveNetID = ""
	if err := ValidateCommandResult(result); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("non-terminal result without next active should be rejected, got %v", err)
	}
}

func TestValidateRosterSnapshot(t *testing.T) {
	roster := validRosterSnapshot()
	if err := ValidateRosterSnapshot(roster); err != nil {
		t.Fatalf("valid roster rejected: %v", err)
	}

	roster.HashAlgorithm = ""
	if err := ValidateRosterSnapshot(roster); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("missing hashAlgorithm should be rejected, got %v", err)
	}
}

func TestValidateRosterSnapshotRejectsUnknownGameSHA256(t *testing.T) {
	roster := validRosterSnapshot()
	roster.GameSHA256 = "unknown"
	if err := ValidateRosterSnapshot(roster); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("unknown gameSha256 should be rejected, got %v", err)
	}
}

func TestValidateRosterSnapshotRejectsFallbackHashAlgorithm(t *testing.T) {
	roster := validRosterSnapshot()
	roster.HashAlgorithm = "adler32-canonical-v1-fallback"
	if err := ValidateRosterSnapshot(roster); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("fallback hashAlgorithm should be rejected, got %v", err)
	}
}

func TestValidateRosterSnapshotRejectsInvalidDigestFormat(t *testing.T) {
	roster := validRosterSnapshot()
	roster.RosterHash = "adler32-fallback:00000000:len:4"
	if err := ValidateRosterSnapshot(roster); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("invalid rosterHash digest should be rejected, got %v", err)
	}
}

func TestValidateRosterSnapshotRejectsDuplicateNetID(t *testing.T) {
	roster := validRosterSnapshot()
	roster.Entries = json.RawMessage(`[
		{
			"netId": "red:0",
			"ownerSeat": "red",
			"entryIndex": 0,
			"scriptName": "scripts/entity/tactical/player",
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
		},
		{
			"netId": "red:0",
			"ownerSeat": "red",
			"entryIndex": 1,
			"scriptName": "scripts/entity/tactical/player",
			"logicState": {
				"source": "squirrel-onSerialize-bufferio",
				"schemaVersion": 1,
				"bufferEncoding": "base64",
				"bufferSize": 4,
				"buffer": "BBBB",
				"bufferChecksum": "sha256:3333333333333333333333333333333333333333333333333333333333333333:len:4",
				"logicHash": "sha256:4444444444444444444444444444444444444444444444444444444444444444:len:10",
				"decodeStatus": "not-validated",
				"decodePolicy": "reject-match-start-on-failure"
			},
			"visualState": {
				"source": "derived-v1",
				"status": "unknown",
				"nativeCoverage": "none"
			},
			"entryHash": "sha256:5555555555555555555555555555555555555555555555555555555555555555:len:20"
		}
	]`)
	if err := ValidateRosterSnapshot(roster); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("duplicate netId should be rejected, got %v", err)
	}
}

func TestValidateRosterSnapshotRejectsNetIDNotMatchingSeatIndex(t *testing.T) {
	roster := validRosterSnapshot()
	roster.Entries = json.RawMessage(`[
		{
			"netId": "red:1",
			"ownerSeat": "red",
			"entryIndex": 0,
			"scriptName": "scripts/entity/tactical/player",
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
	]`)
	if err := ValidateRosterSnapshot(roster); ErrorCodeOf(err) != ErrorInvalidCommand {
		t.Fatalf("netId not matching seat:index should be rejected, got %v", err)
	}
}

func TestFirstRosterNetID(t *testing.T) {
	netID, err := FirstRosterNetID(validRosterSnapshot())
	if err != nil {
		t.Fatalf("first roster netId rejected: %v", err)
	}
	if netID != "red:0" {
		t.Fatalf("first roster netId = %q, want red:0", netID)
	}
}

func validCommandResult() CommandResult {
	postHash := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:len:20"
	return CommandResult{
		SchemaVersion:   SchemaVersion,
		MatchID:         "match-1",
		ServerSeq:       1,
		Seat:            SeatRed,
		PostHash:        postHash,
		PostSummary:     validPostSummary(postHash, 1),
		NextActiveSeat:  SeatBlue,
		NextActiveNetID: "blue:0",
		Round:           1,
		Turn:            2,
	}
}

func validPostSummary(digest string, serverSeq int64) json.RawMessage {
	return json.RawMessage(`{"schemaVersion":1,"algorithm":"sha256-state-v1","matchId":"match-1","lastServerSeq":` + strconv.FormatInt(serverSeq, 10) + `,"canonical":"state-hash-v1|x","digest":"` + digest + `"}`)
}

func validMatchResultEvent() MatchResultEvent {
	result := MatchResultEvent{
		SchemaVersion:         SchemaVersion,
		ResultContractVersion: ResultContractVersion,
		MatchID:               "match-1",
		Status:                RoomStatusEnded,
		WinnerSeat:            SeatRed,
		LoserSeat:             SeatBlue,
		Reason:                MatchResultReasonBattleEnd,
		SourceSeat:            SeatRed,
		ResolvedAtTick:        10,
		LastCompletedSeq:      1,
		CheckpointBinding: MatchCheckpointBinding{
			SchemaVersion:      SchemaVersion,
			MatchID:            "match-1",
			StartHash:          "start-hash",
			LastCompletedSeq:   1,
			LastHash:           "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc:len:20",
			TerminalCheckpoint: true,
		},
		Participants: MatchParticipants{
			SchemaVersion: SchemaVersion,
			MatchID:       "match-1",
			Red: MatchParticipant{
				Seat:                SeatRed,
				UserID:              "u-red",
				Username:            "red",
				RosterHashAlgorithm: RosterHashAlgorithmSHA256,
				RosterHash:          "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:len:10",
				VisualStatusSummary: "unknown",
			},
			Blue: MatchParticipant{
				Seat:                SeatBlue,
				UserID:              "u-blue",
				Username:            "blue",
				RosterHashAlgorithm: RosterHashAlgorithmSHA256,
				RosterHash:          "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb:len:10",
				VisualStatusSummary: "unknown",
			},
			GameSHA256:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			GameVersion: "1.5.1.8",
			ModVersion:  "mvp",
			RulesetHash: "rules",
			Scenario:    1,
			MapSeed:     101,
			CombatSeed:  202,
		},
		TimeoutPolicy: MatchTimeoutPolicy{
			SchemaVersion:                SchemaVersion,
			PolicyVersion:                TimeoutPolicyVersion,
			TickRate:                     10,
			ClockSource:                  TimeoutClockSourceMatchLoopTick,
			CommandResultTimeoutTicks:    300,
			CommandResultTimeoutAction:   TimeoutActionDesync,
			AbandonGraceTicks:            600,
			AbandonGraceExpiredAction:    TimeoutActionMatchResultAbandon,
			BattleEndReportTimeoutTicks:  300,
			BattleEndReportTimeoutAction: TimeoutActionDesync,
		},
		RankedEligible:      false,
		RankedReason:        MatchResultRankedReasonMVPDisabled,
		RankedPolicyVersion: RankedPolicyVersion,
		RankedMode:          RankedModeDisabled,
	}
	result.ResultID = ComputeMatchResultID(result)
	return result
}

func validRosterSnapshot() RosterSnapshot {
	return RosterSnapshot{
		SchemaVersion: SchemaVersion,
		CodecVersion:  1,
		GameSHA256:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		GameVersion:   "1.5.1.8",
		ModVersion:    "0.0.1",
		RulesetHash:   "arena-v1",
		OwnerUserID:   "u-red",
		Seat:          SeatRed,
		Entries: json.RawMessage(`[
			{
				"netId": "red:0",
				"ownerSeat": "red",
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
		]`),
		HashAlgorithm:       RosterHashAlgorithmSHA256,
		RosterHash:          "sha256:3333333333333333333333333333333333333333333333333333333333333333:len:30",
		VisualStatusSummary: "unknown",
	}
}
