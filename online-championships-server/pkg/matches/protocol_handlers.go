package matches

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/heroiclabs/nakama-common/runtime"
	"online-championships/pkg/matches/strategicProperties"
	"online-championships/pkg/protocol"
)

var protocolHandlers = map[int64]func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, state interface{}, message runtime.MatchData, tick int64){
	protocol.OpCodeRosterLocked:         handleRosterLocked,
	protocol.OpCodeMatchForfeitRequest:  handleMatchForfeit,
	protocol.OpCodeMatchBattleEndReport: handleMatchBattleEndReport,
	protocol.OpCodeCmdWaitTurn:          handleCommand,
	protocol.OpCodeCmdEndTurn:           handleCommand,
	protocol.OpCodeCmdResult:            handleCommandResult,
}

func handleRosterLocked(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, state interface{}, message runtime.MatchData, tick int64) {
	mState, _ := state.(*MatchState)
	var roster protocol.RosterSnapshot
	if err := json.Unmarshal(message.GetData(), &roster); err != nil {
		logger.Error("unable to unmarshal roster snapshot: %v", err)
		handleProtocolError(logger, dispatcher, protocol.ProtocolError{
			Code:    protocol.ErrorInvalidCommand,
			Message: "roster snapshot must be valid JSON",
		}, 0, nil, message, message.GetOpCode(), "")
		return
	}
	seat, ok := mState.roomRuntime.SeatForSession(message.GetSessionId())
	if !ok || seat == protocol.SeatSpectator {
		logger.Info("roster lock rejected for unseated session %s", message.GetSessionId())
		handleProtocolError(logger, dispatcher, protocol.ProtocolError{
			Code:    protocol.ErrorInvalidSeat,
			Message: "roster lock requires a player seat",
		}, 0, nil, message, message.GetOpCode(), "")
		return
	}
	if roster.OwnerUserID == "" {
		roster.OwnerUserID = message.GetUserId()
	}
	if err := mState.roomRuntime.LockRoster(seat, roster); err != nil {
		logger.Info("roster lock rejected for %s: %v", message.GetSessionId(), err)
		handleProtocolError(logger, dispatcher, err, 0, nil, message, message.GetOpCode(), "")
		return
	}
	if err := dispatcher.BroadcastMessage(protocol.OpCodeRosterLocked, DumpEvent(protocol.RosterLockedEvent{
		SchemaVersion: protocol.SchemaVersion,
		Seat:          seat,
		RosterHash:    roster.RosterHash,
	}), nil, message, true); err != nil {
		logger.Error("broadcast roster lock error: %v", err)
	}
	if mState.roomRuntime.ReadyToStart() {
		startHash := computeStartHash(mState.strategicProperties, mState.roomRuntime.Rosters[protocol.SeatRed], mState.roomRuntime.Rosters[protocol.SeatBlue])
		initialNetID, err := protocol.FirstRosterNetID(mState.roomRuntime.Rosters[protocol.SeatRed])
		if err != nil {
			logger.Error("unable to resolve initial active netId: %v", err)
			return
		}
		if err := mState.roomRuntime.Start(startHash, protocol.SeatRed, initialNetID); err != nil {
			logger.Error("unable to start v1 match: %v", err)
			return
		}
		mState.battleStarted = true
		mState.totalFaction = 2
		snapshot := protocol.MatchStartSnapshot{
			SchemaVersion:       protocol.SchemaVersion,
			RoomID:              mState.roomID,
			StartHash:           startHash,
			MapSeed:             mState.strategicProperties.MapSeed,
			CombatSeed:          mState.strategicProperties.CombatSeed,
			RulesetHash:         mState.roomRuntime.Rosters[protocol.SeatRed].RulesetHash,
			ExpectedActiveSeat:  mState.roomRuntime.ExpectedActiveSeat,
			ExpectedActiveNetID: mState.roomRuntime.ExpectedActiveNetID,
			RedRoster:           mState.roomRuntime.Rosters[protocol.SeatRed],
			BlueRoster:          mState.roomRuntime.Rosters[protocol.SeatBlue],
			StrategicProperties: mState.strategicProperties,
		}
		mState.roomRuntime.SetStartSnapshot(snapshot)
		if err := dispatcher.BroadcastMessage(protocol.OpCodeMatchStart, DumpEvent(snapshot), nil, nil, true); err != nil {
			logger.Error("broadcast match start error: %v", err)
		}
		broadcastRoomState(logger, dispatcher, mState)
	}
}

func handleMatchForfeit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, state interface{}, message runtime.MatchData, tick int64) {
	mState, _ := state.(*MatchState)
	var request protocol.MatchForfeitRequest
	if err := json.Unmarshal(message.GetData(), &request); err != nil {
		logger.Error("unable to unmarshal match forfeit request: %v", err)
		handleProtocolError(logger, dispatcher, protocol.ProtocolError{
			Code:    protocol.ErrorInvalidCommand,
			Message: "match forfeit request must be valid JSON",
		}, 0, nil, message, message.GetOpCode(), "")
		return
	}
	if err := protocol.ValidateMatchForfeitRequest(request); err != nil {
		handleProtocolError(logger, dispatcher, err, 0, nil, message, message.GetOpCode(), request.ClientRequestID)
		return
	}
	if err := requireSenderSeat(mState, message, request.Seat); err != nil {
		logger.Info("match forfeit sender rejected: %v", err)
		handleProtocolError(logger, dispatcher, err, 0, nil, message, message.GetOpCode(), request.ClientRequestID)
		return
	}
	if err := requireMatchID(mState, request.MatchID); err != nil {
		logger.Info("match forfeit match rejected: %v", err)
		handleProtocolError(logger, dispatcher, err, 0, nil, message, message.GetOpCode(), request.ClientRequestID)
		return
	}
	result, err := mState.roomRuntime.ResolveMatchResult(request.MatchID, request.Seat, message.GetUserId(), request.Reason, request.ClientRequestID, tick)
	if err != nil {
		handleProtocolError(logger, dispatcher, err, 0, nil, message, message.GetOpCode(), request.ClientRequestID)
		return
	}
	if err := dispatcher.BroadcastMessage(protocol.OpCodeMatchResult, DumpEvent(result), nil, nil, true); err != nil {
		logger.Error("broadcast match result error: %v", err)
	}
	broadcastRoomState(logger, dispatcher, mState)
}

func handleMatchBattleEndReport(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, state interface{}, message runtime.MatchData, tick int64) {
	mState, _ := state.(*MatchState)
	var report protocol.MatchBattleEndReport
	if err := json.Unmarshal(message.GetData(), &report); err != nil {
		logger.Error("unable to unmarshal match battle end report: %v", err)
		handleProtocolError(logger, dispatcher, protocol.ProtocolError{
			Code:    protocol.ErrorInvalidCommand,
			Message: "match battle end report must be valid JSON",
		}, 0, nil, message, message.GetOpCode(), "")
		return
	}
	if err := protocol.ValidateMatchBattleEndReport(report); err != nil {
		handleProtocolError(logger, dispatcher, err, 0, nil, message, message.GetOpCode(), report.ClientRequestID)
		return
	}
	if err := requireSenderSeat(mState, message, report.Seat); err != nil {
		logger.Info("battle end report sender rejected: %v", err)
		handleProtocolError(logger, dispatcher, err, 0, nil, message, message.GetOpCode(), report.ClientRequestID)
		return
	}
	if err := requireMatchID(mState, report.MatchID); err != nil {
		logger.Info("battle end report match rejected: %v", err)
		handleProtocolError(logger, dispatcher, err, 0, nil, message, message.GetOpCode(), report.ClientRequestID)
		return
	}
	result, resolved, err := mState.roomRuntime.RecordBattleEndReport(report, tick)
	if err != nil {
		handleProtocolError(logger, dispatcher, err, 0, nil, message, message.GetOpCode(), report.ClientRequestID)
		if protocol.ErrorCodeOf(err) == protocol.ErrorDesync {
			broadcastRoomState(logger, dispatcher, mState)
		}
		return
	}
	if !resolved {
		broadcastRoomState(logger, dispatcher, mState)
		return
	}
	if err := dispatcher.BroadcastMessage(protocol.OpCodeMatchResult, DumpEvent(result), nil, nil, true); err != nil {
		logger.Error("broadcast battle end match result error: %v", err)
	}
	broadcastRoomState(logger, dispatcher, mState)
}

func handleCommand(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, state interface{}, message runtime.MatchData, tick int64) {
	mState, _ := state.(*MatchState)
	var cmd protocol.CommandEnvelope
	if err := json.Unmarshal(message.GetData(), &cmd); err != nil {
		logger.Error("unable to unmarshal command: %v", err)
		handleProtocolError(logger, dispatcher, protocol.ProtocolError{
			Code:    protocol.ErrorInvalidCommand,
			Message: "command envelope must be valid JSON",
		}, 0, nil, message, message.GetOpCode(), "")
		return
	}
	if expectedOp, ok := protocol.OpcodeForCommand(cmd.Op); !ok || expectedOp != message.GetOpCode() {
		logger.Info("command opcode mismatch: message=%d command=%s", message.GetOpCode(), cmd.Op)
		handleProtocolError(logger, dispatcher, protocol.ProtocolError{
			Code:    protocol.ErrorInvalidCommand,
			Message: "command opcode mismatch",
		}, 0, nil, message, message.GetOpCode(), cmd.ClientCommandID)
		return
	}
	if err := requireSenderSeat(mState, message, cmd.Seat); err != nil {
		logger.Info("command sender rejected: %v", err)
		handleProtocolError(logger, dispatcher, err, 0, nil, message, message.GetOpCode(), cmd.ClientCommandID)
		return
	}
	if err := requireMatchID(mState, cmd.MatchID); err != nil {
		logger.Info("command match rejected: %v", err)
		handleProtocolError(logger, dispatcher, err, 0, nil, message, message.GetOpCode(), cmd.ClientCommandID)
		return
	}
	accepted, err := mState.roomRuntime.AcceptCommandAt(cmd, tick)
	if err != nil {
		handleProtocolError(logger, dispatcher, err, 0, nil, message, message.GetOpCode(), cmd.ClientCommandID)
		if protocol.ErrorCodeOf(err) == protocol.ErrorDesync {
			broadcastRoomState(logger, dispatcher, mState)
		}
		return
	}
	if err := dispatcher.BroadcastMessage(message.GetOpCode(), DumpEvent(accepted), nil, message, true); err != nil {
		logger.Error("broadcast accepted command error: %v", err)
	}
	broadcastRoomState(logger, dispatcher, mState)
}

func handleCommandResult(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, state interface{}, message runtime.MatchData, tick int64) {
	mState, _ := state.(*MatchState)
	var result protocol.CommandResult
	if err := json.Unmarshal(message.GetData(), &result); err != nil {
		logger.Error("unable to unmarshal command result: %v", err)
		handleProtocolError(logger, dispatcher, protocol.ProtocolError{
			Code:    protocol.ErrorInvalidCommand,
			Message: "command result must be valid JSON",
		}, 0, nil, message, message.GetOpCode(), "")
		return
	}
	if err := requireSenderSeat(mState, message, result.Seat); err != nil {
		logger.Info("command result sender rejected: %v", err)
		handleProtocolError(logger, dispatcher, err, result.ServerSeq, nil, message, message.GetOpCode(), "")
		return
	}
	if err := requireMatchID(mState, result.MatchID); err != nil {
		logger.Info("command result match rejected: %v", err)
		handleProtocolError(logger, dispatcher, err, result.ServerSeq, nil, message, message.GetOpCode(), "")
		return
	}
	if err := mState.roomRuntime.RecordCommandResult(result); err != nil {
		handleProtocolError(logger, dispatcher, err, result.ServerSeq, mState.roomRuntime.LastMismatch, message, message.GetOpCode(), "")
		if protocol.ErrorCodeOf(err) == protocol.ErrorDesync {
			broadcastRoomState(logger, dispatcher, mState)
		}
		return
	}
	if completed, ok := mState.roomRuntime.CompletedResult(result.ServerSeq); ok {
		if err := dispatcher.BroadcastMessage(protocol.OpCodeCmdResult, DumpEvent(completed), nil, nil, true); err != nil {
			logger.Error("broadcast command result error: %v", err)
		}
		broadcastRoomState(logger, dispatcher, mState)
		resolvePendingBattleEndReports(logger, dispatcher, mState, tick)
	}
}

func resolvePendingBattleEndReports(logger runtime.Logger, dispatcher runtime.MatchDispatcher, mState *MatchState, tick int64) {
	if mState == nil || mState.roomRuntime == nil {
		return
	}
	result, resolved, err := mState.roomRuntime.ResolveBattleEndReports(mState.roomID, tick)
	if err != nil {
		handleProtocolError(logger, dispatcher, err, 0, nil, nil, protocol.OpCodeMatchBattleEndReport, "")
		broadcastRoomState(logger, dispatcher, mState)
		return
	}
	if !resolved {
		return
	}
	if err := dispatcher.BroadcastMessage(protocol.OpCodeMatchResult, DumpEvent(result), nil, nil, true); err != nil {
		logger.Error("broadcast pending battle end match result error: %v", err)
	}
	broadcastRoomState(logger, dispatcher, mState)
}

func requireSenderSeat(mState *MatchState, message runtime.MatchData, claimedSeat protocol.Seat) error {
	seat, ok := mState.roomRuntime.SeatForSession(message.GetSessionId())
	if !ok {
		return protocol.ProtocolError{Code: protocol.ErrorInvalidSeat, Message: "session is not seated"}
	}
	if seat != claimedSeat {
		return protocol.ProtocolError{Code: protocol.ErrorInvalidSeat, Message: fmt.Sprintf("claimed %s but assigned %s", claimedSeat, seat)}
	}
	return nil
}

func requireMatchID(mState *MatchState, matchID string) error {
	if mState.roomID == "" {
		return protocol.ProtocolError{Code: protocol.ErrorInvalidState, Message: "roomId is not initialized"}
	}
	if matchID != mState.roomID {
		return protocol.ProtocolError{Code: protocol.ErrorInvalidCommand, Message: "matchId does not match room"}
	}
	return nil
}

func handleProtocolError(logger runtime.Logger, dispatcher runtime.MatchDispatcher, err error, serverSeq int64, checkpoint *protocol.CheckpointMismatch, message runtime.MatchData, opCode int64, clientCommandID string) {
	code := protocol.ErrorCodeOf(err)
	if code == "" {
		code = protocol.ErrorInvalidCommand
	}
	if code == protocol.ErrorDesync {
		if broadcastErr := dispatcher.BroadcastMessage(protocol.OpCodeMatchDesync, DumpEvent(protocol.MatchDesyncEvent{
			SchemaVersion: protocol.SchemaVersion,
			ServerSeq:     serverSeq,
			Code:          code,
			Message:       err.Error(),
			Checkpoint:    checkpoint,
		}), nil, nil, true); broadcastErr != nil {
			logger.Error("broadcast desync error: %v", broadcastErr)
		}
		return
	}
	logger.Info("protocol message rejected: %v", err)
	if message == nil {
		return
	}
	if broadcastErr := dispatcher.BroadcastMessage(protocol.OpCodeProtocolError, DumpEvent(protocol.ProtocolErrorEvent{
		SchemaVersion:   protocol.SchemaVersion,
		ServerSeq:       serverSeq,
		Code:            code,
		Message:         err.Error(),
		OpCode:          opCode,
		ClientCommandID: clientCommandID,
	}), []runtime.Presence{message}, nil, true); broadcastErr != nil {
		logger.Error("broadcast protocol error response error: %v", broadcastErr)
	}
}

func computeStartHash(strategic strategicProperties.StrategicProperties, redRoster, blueRoster protocol.RosterSnapshot) string {
	input := struct {
		SchemaVersion       int                                     `json:"schemaVersion"`
		RedRosterHash       string                                  `json:"redRosterHash"`
		BlueRosterHash      string                                  `json:"blueRosterHash"`
		StrategicProperties strategicProperties.StrategicProperties `json:"strategicProperties"`
	}{
		SchemaVersion:       protocol.SchemaVersion,
		RedRosterHash:       redRoster.RosterHash,
		BlueRosterHash:      blueRoster.RosterHash,
		StrategicProperties: strategic,
	}
	data, _ := json.Marshal(input)
	return protocol.FormatSHA256Digest(data)
}
