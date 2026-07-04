package matches

import (
	"fmt"
	"online-championships/pkg/protocol"
)

const DefaultCommandResultTimeoutTicks int64 = 300
const DefaultAbandonGraceTicks int64 = 600
const DefaultBattleEndReportTimeoutTicks int64 = 300

type RoomRuntime struct {
	Status                      protocol.RoomStatus
	Metadata                    protocol.RoomMetadata
	Seats                       map[protocol.Seat]SeatPresence
	Spectators                  map[string]SeatPresence
	Rosters                     map[protocol.Seat]protocol.RosterSnapshot
	StartSnapshot               *protocol.MatchStartSnapshot
	StartHash                   string
	LastAcceptedHash            string
	ExpectedActiveSeat          protocol.Seat
	ExpectedActiveNetID         string
	NextServerSeq               int64
	CommandLog                  []protocol.CommandEnvelope
	PendingCheckpoints          map[int64]PendingCheckpoint
	AcceptedCommands            map[string]protocol.CommandEnvelope
	CompletedResults            map[int64]protocol.CommandResult
	CompletedCheckpoints        map[int64]CompletedCheckpoint
	CommandResultTimeoutTicks   int64
	AbandonGraceTicks           int64
	AbandonDeadlines            map[protocol.Seat]AbandonDeadline
	BattleEndReports            map[protocol.Seat]protocol.MatchBattleEndReport
	BattleEndReportDeadlines    map[protocol.Seat]BattleEndReportDeadline
	BattleEndReportTimeoutTicks int64
	LastMismatch                *protocol.CheckpointMismatch
	TerminalResult              *protocol.MatchResultEvent
}

type SeatPresence struct {
	UserID    string
	SessionID string
	Username  string
}

type PendingCheckpoint struct {
	Command      protocol.CommandEnvelope
	Results      map[protocol.Seat]protocol.CommandResult
	AcceptedTick int64
	DeadlineTick int64
}

type CompletedCheckpoint struct {
	Command protocol.CommandEnvelope
	Results map[protocol.Seat]protocol.CommandResult
}

type AbandonDeadline struct {
	Seat         protocol.Seat
	UserID       string
	AcceptedTick int64
	DeadlineTick int64
}

type BattleEndReportDeadline struct {
	Seat         protocol.Seat
	AcceptedTick int64
	DeadlineTick int64
}

func NewRoomRuntime(startHash string) *RoomRuntime {
	return &RoomRuntime{
		Status:                      protocol.RoomStatusOpen,
		Metadata:                    defaultRoomMetadata(),
		Seats:                       map[protocol.Seat]SeatPresence{},
		Spectators:                  map[string]SeatPresence{},
		Rosters:                     map[protocol.Seat]protocol.RosterSnapshot{},
		StartHash:                   startHash,
		LastAcceptedHash:            startHash,
		NextServerSeq:               1,
		PendingCheckpoints:          map[int64]PendingCheckpoint{},
		AcceptedCommands:            map[string]protocol.CommandEnvelope{},
		CompletedResults:            map[int64]protocol.CommandResult{},
		CompletedCheckpoints:        map[int64]CompletedCheckpoint{},
		CommandResultTimeoutTicks:   DefaultCommandResultTimeoutTicks,
		AbandonGraceTicks:           DefaultAbandonGraceTicks,
		AbandonDeadlines:            map[protocol.Seat]AbandonDeadline{},
		BattleEndReports:            map[protocol.Seat]protocol.MatchBattleEndReport{},
		BattleEndReportDeadlines:    map[protocol.Seat]BattleEndReportDeadline{},
		BattleEndReportTimeoutTicks: DefaultBattleEndReportTimeoutTicks,
	}
}

func (r *RoomRuntime) AssignSeat(presence SeatPresence, requested protocol.Seat) (protocol.Seat, error) {
	if !protocol.IsJoinSeat(requested) {
		return "", protocol.ProtocolError{Code: protocol.ErrorInvalidSeat, Message: string(requested)}
	}
	if r.Status == protocol.RoomStatusClosed {
		return "", protocol.ProtocolError{Code: protocol.ErrorInvalidState, Message: string(r.Status)}
	}
	if existing, ok := r.findSeatByUserID(presence.UserID); ok {
		r.Seats[existing] = presence
		delete(r.AbandonDeadlines, existing)
		return existing, nil
	}
	if requested == protocol.SeatSpectator {
		if existingSessionID, ok := r.findSpectatorSessionByUserID(presence.UserID); ok {
			if existingSessionID != presence.SessionID {
				delete(r.Spectators, existingSessionID)
			}
			r.Spectators[presence.SessionID] = presence
			return protocol.SeatSpectator, nil
		}
		if !r.Metadata.Spectators.Enabled {
			return "", protocol.ProtocolError{Code: protocol.ErrorInvalidState, Message: "spectators disabled"}
		}
		if r.Metadata.Spectators.Limit > 0 && len(r.Spectators) >= r.Metadata.Spectators.Limit {
			return "", protocol.ProtocolError{Code: protocol.ErrorInvalidState, Message: "spectator limit reached"}
		}
		r.Spectators[presence.SessionID] = presence
		return protocol.SeatSpectator, nil
	}
	if r.Status == protocol.RoomStatusEnded {
		return "", protocol.ProtocolError{Code: protocol.ErrorInvalidState, Message: string(r.Status)}
	}
	if r.Status != protocol.RoomStatusOpen && r.Status != protocol.RoomStatusReadyCheck {
		return "", protocol.ProtocolError{Code: protocol.ErrorInvalidState, Message: string(r.Status)}
	}
	if requested == protocol.SeatAuto {
		if _, ok := r.Seats[protocol.SeatRed]; !ok {
			requested = protocol.SeatRed
		} else {
			requested = protocol.SeatBlue
		}
	}
	if !protocol.IsPlayerSeat(requested) {
		return "", protocol.ProtocolError{Code: protocol.ErrorInvalidSeat, Message: string(requested)}
	}
	if _, occupied := r.Seats[requested]; occupied {
		return "", protocol.ProtocolError{Code: protocol.ErrorInvalidSeat, Message: fmt.Sprintf("%s is occupied", requested)}
	}
	r.Seats[requested] = presence
	return requested, nil
}

func (r *RoomRuntime) LockRoster(seat protocol.Seat, roster protocol.RosterSnapshot) error {
	if r.Status != protocol.RoomStatusOpen && r.Status != protocol.RoomStatusReadyCheck {
		return protocol.ProtocolError{Code: protocol.ErrorInvalidState, Message: string(r.Status)}
	}
	if !protocol.IsPlayerSeat(seat) {
		return protocol.ProtocolError{Code: protocol.ErrorInvalidSeat, Message: string(seat)}
	}
	presence, ok := r.Seats[seat]
	if !ok {
		return protocol.ProtocolError{Code: protocol.ErrorInvalidSeat, Message: fmt.Sprintf("%s is not occupied", seat)}
	}
	if roster.OwnerUserID != "" && roster.OwnerUserID != presence.UserID {
		return protocol.ProtocolError{Code: protocol.ErrorInvalidCommand, Message: "roster ownerUserId mismatch"}
	}
	roster.OwnerUserID = presence.UserID
	if err := protocol.ValidateRosterSnapshot(roster); err != nil {
		return err
	}
	if roster.Seat != seat {
		return protocol.ProtocolError{Code: protocol.ErrorInvalidSeat, Message: "roster seat mismatch"}
	}
	for otherSeat, otherRoster := range r.Rosters {
		if otherSeat == seat {
			continue
		}
		if otherRoster.GameSHA256 != roster.GameSHA256 {
			return protocol.ProtocolError{Code: protocol.ErrorInvalidCommand, Message: "roster gameSha256 mismatch"}
		}
		if otherRoster.RulesetHash != roster.RulesetHash {
			return protocol.ProtocolError{Code: protocol.ErrorInvalidCommand, Message: "roster rulesetHash mismatch"}
		}
		if otherRoster.GameVersion != roster.GameVersion {
			return protocol.ProtocolError{Code: protocol.ErrorInvalidCommand, Message: "roster gameVersion mismatch"}
		}
		if otherRoster.ModVersion != roster.ModVersion {
			return protocol.ProtocolError{Code: protocol.ErrorInvalidCommand, Message: "roster modVersion mismatch"}
		}
	}
	r.Rosters[seat] = roster
	if r.ReadyToStart() {
		r.Status = protocol.RoomStatusReadyCheck
	}
	return nil
}

func (r *RoomRuntime) ReadyToStart() bool {
	_, redSeat := r.Seats[protocol.SeatRed]
	_, blueSeat := r.Seats[protocol.SeatBlue]
	_, redRoster := r.Rosters[protocol.SeatRed]
	_, blueRoster := r.Rosters[protocol.SeatBlue]
	return redSeat && blueSeat && redRoster && blueRoster
}

func (r *RoomRuntime) Start(startHash string, expectedSeat protocol.Seat, expectedNetID string) error {
	if !r.ReadyToStart() {
		return protocol.ProtocolError{Code: protocol.ErrorInvalidState, Message: "both player rosters must be locked"}
	}
	if startHash == "" {
		return protocol.ProtocolError{Code: protocol.ErrorInvalidCommand, Message: "startHash is required"}
	}
	if !protocol.IsPlayerSeat(expectedSeat) {
		return protocol.ProtocolError{Code: protocol.ErrorInvalidSeat, Message: string(expectedSeat)}
	}
	if expectedNetID == "" {
		return protocol.ProtocolError{Code: protocol.ErrorInvalidCommand, Message: "expected active netId is required"}
	}
	r.Status = protocol.RoomStatusInBattle
	r.StartHash = startHash
	r.LastAcceptedHash = startHash
	r.ExpectedActiveSeat = expectedSeat
	r.ExpectedActiveNetID = expectedNetID
	return nil
}

func (r *RoomRuntime) SetStartSnapshot(snapshot protocol.MatchStartSnapshot) {
	r.StartSnapshot = &snapshot
}

func (r *RoomRuntime) RemovePresence(sessionID string) (protocol.Seat, bool) {
	return r.RemovePresenceAt(sessionID, 0)
}

func (r *RoomRuntime) RemovePresenceAt(sessionID string, tick int64) (protocol.Seat, bool) {
	if _, ok := r.Spectators[sessionID]; ok {
		delete(r.Spectators, sessionID)
		return protocol.SeatSpectator, true
	}
	for seat, presence := range r.Seats {
		if presence.SessionID != sessionID {
			continue
		}
		if r.Status == protocol.RoomStatusOpen || r.Status == protocol.RoomStatusReadyCheck || r.Status == protocol.RoomStatusStarting {
			delete(r.Seats, seat)
			delete(r.Rosters, seat)
			if r.Status == protocol.RoomStatusReadyCheck && !r.ReadyToStart() {
				r.Status = protocol.RoomStatusOpen
			}
			return seat, true
		}
		presence.SessionID = ""
		r.Seats[seat] = presence
		if r.Status == protocol.RoomStatusInBattle {
			r.AbandonDeadlines[seat] = AbandonDeadline{
				Seat:         seat,
				UserID:       presence.UserID,
				AcceptedTick: tick,
				DeadlineTick: r.abandonDeadline(tick),
			}
		}
		return seat, true
	}
	return "", false
}

func (r *RoomRuntime) AcceptCommand(cmd protocol.CommandEnvelope) (protocol.CommandEnvelope, error) {
	return r.AcceptCommandAt(cmd, 0)
}

func (r *RoomRuntime) AcceptCommandAt(cmd protocol.CommandEnvelope, tick int64) (protocol.CommandEnvelope, error) {
	if err := protocol.ValidateClientCommand(cmd); err != nil {
		return protocol.CommandEnvelope{}, err
	}
	commandKey := commandIdempotencyKey(cmd.Seat, cmd.ClientCommandID)
	if accepted, ok := r.AcceptedCommands[commandKey]; ok {
		return accepted, nil
	}
	if r.Status != protocol.RoomStatusInBattle {
		return protocol.CommandEnvelope{}, protocol.ProtocolError{Code: protocol.ErrorInvalidState, Message: string(r.Status)}
	}
	if cmd.Seat != r.ExpectedActiveSeat {
		return protocol.CommandEnvelope{}, protocol.ProtocolError{Code: protocol.ErrorInvalidSeat, Message: "not active seat"}
	}
	if cmd.ActiveNetID != r.ExpectedActiveNetID {
		return protocol.CommandEnvelope{}, protocol.ProtocolError{Code: protocol.ErrorInvalidCommand, Message: "not active entity"}
	}
	if cmd.PreHash != r.LastAcceptedHash {
		r.Status = protocol.RoomStatusDesynced
		return protocol.CommandEnvelope{}, protocol.ProtocolError{Code: protocol.ErrorDesync, Message: "preHash mismatch"}
	}
	if len(r.PendingCheckpoints) > 0 {
		return protocol.CommandEnvelope{}, protocol.ProtocolError{Code: protocol.ErrorInvalidState, Message: "pending checkpoint exists"}
	}
	cmd.ServerSeq = r.NextServerSeq
	r.NextServerSeq++
	r.CommandLog = append(r.CommandLog, cmd)
	r.AcceptedCommands[commandKey] = cmd
	r.PendingCheckpoints[cmd.ServerSeq] = PendingCheckpoint{
		Command:      cmd,
		Results:      map[protocol.Seat]protocol.CommandResult{},
		AcceptedTick: tick,
		DeadlineTick: r.commandResultDeadline(tick),
	}
	return cmd, nil
}

func (r *RoomRuntime) RecordCommandResult(result protocol.CommandResult) error {
	if err := protocol.ValidateCommandResult(result); err != nil {
		return err
	}
	pending, ok := r.PendingCheckpoints[result.ServerSeq]
	if !ok {
		return protocol.ProtocolError{Code: protocol.ErrorInvalidCommand, Message: "unknown serverSeq"}
	}
	if result.MatchID != pending.Command.MatchID {
		return protocol.ProtocolError{Code: protocol.ErrorInvalidCommand, Message: "result matchId mismatch"}
	}
	if _, exists := pending.Results[result.Seat]; exists {
		return protocol.ProtocolError{Code: protocol.ErrorInvalidCommand, Message: "duplicate command result"}
	}
	pending.Results[result.Seat] = result
	r.PendingCheckpoints[result.ServerSeq] = pending
	if len(pending.Results) < 2 {
		return nil
	}
	red := pending.Results[protocol.SeatRed]
	blue := pending.Results[protocol.SeatBlue]
	if red.PostHash != blue.PostHash ||
		red.Terminal != blue.Terminal ||
		(!red.Terminal && red.NextActiveSeat != blue.NextActiveSeat) ||
		(!red.Terminal && red.NextActiveNetID != blue.NextActiveNetID) ||
		red.Round != blue.Round ||
		red.Turn != blue.Turn {
		r.Status = protocol.RoomStatusDesynced
		r.LastMismatch = &protocol.CheckpointMismatch{
			ServerSeq:  result.ServerSeq,
			RedResult:  red,
			BlueResult: blue,
		}
		return protocol.ProtocolError{Code: protocol.ErrorDesync, Message: "checkpoint mismatch"}
	}
	r.LastAcceptedHash = red.PostHash
	if red.Terminal {
		r.ExpectedActiveSeat = ""
		r.ExpectedActiveNetID = ""
	} else {
		r.ExpectedActiveSeat = red.NextActiveSeat
		r.ExpectedActiveNetID = red.NextActiveNetID
	}
	r.CompletedResults[result.ServerSeq] = red
	r.CompletedCheckpoints[result.ServerSeq] = CompletedCheckpoint{
		Command: pending.Command,
		Results: map[protocol.Seat]protocol.CommandResult{
			protocol.SeatRed:  red,
			protocol.SeatBlue: blue,
		},
	}
	delete(r.PendingCheckpoints, result.ServerSeq)
	return nil
}

func (r *RoomRuntime) CompletedResult(serverSeq int64) (protocol.CommandResult, bool) {
	result, ok := r.CompletedResults[serverSeq]
	return result, ok
}

func (r *RoomRuntime) CompletedCheckpoint(serverSeq int64) (CompletedCheckpoint, bool) {
	checkpoint, ok := r.CompletedCheckpoints[serverSeq]
	return checkpoint, ok
}

func (r *RoomRuntime) LastCompletedSeq() int64 {
	var last int64
	for _, cmd := range r.CommandLog {
		if _, ok := r.CompletedResults[cmd.ServerSeq]; !ok {
			break
		}
		last = cmd.ServerSeq
	}
	return last
}

func (r *RoomRuntime) PendingServerSeq() int64 {
	var pending int64
	for serverSeq := range r.PendingCheckpoints {
		if pending == 0 || serverSeq < pending {
			pending = serverSeq
		}
	}
	return pending
}

func (r *RoomRuntime) ExpirePendingCheckpoints(tick int64) (int64, error) {
	if r.Status != protocol.RoomStatusInBattle {
		return 0, nil
	}
	var expiredSeq int64
	var expired PendingCheckpoint
	for serverSeq, pending := range r.PendingCheckpoints {
		if pending.DeadlineTick == 0 || tick < pending.DeadlineTick {
			continue
		}
		if expiredSeq == 0 || serverSeq < expiredSeq {
			expiredSeq = serverSeq
			expired = pending
		}
	}
	if expiredSeq == 0 {
		return 0, nil
	}
	r.Status = protocol.RoomStatusDesynced
	return expiredSeq, protocol.ProtocolError{
		Code:    protocol.ErrorDesync,
		Message: fmt.Sprintf("checkpoint result timeout: serverSeq=%d acceptedTick=%d deadlineTick=%d tick=%d results=%d", expiredSeq, expired.AcceptedTick, expired.DeadlineTick, tick, len(expired.Results)),
	}
}

func (r *RoomRuntime) Forfeit(matchID string, loserSeat protocol.Seat, sourceUserID string, reason protocol.MatchResultReason) (protocol.MatchResultEvent, error) {
	return r.ResolveMatchResult(matchID, loserSeat, sourceUserID, reason, "", 0)
}

func (r *RoomRuntime) ResolveMatchResult(matchID string, loserSeat protocol.Seat, sourceUserID string, reason protocol.MatchResultReason, clientRequestID string, tick int64) (protocol.MatchResultEvent, error) {
	return r.resolveMatchResult(matchID, loserSeat, loserSeat, sourceUserID, reason, clientRequestID, nil, tick)
}

func (r *RoomRuntime) resolveMatchResult(matchID string, loserSeat protocol.Seat, sourceSeat protocol.Seat, sourceUserID string, reason protocol.MatchResultReason, clientRequestID string, reportClientRequestIDs map[protocol.Seat]string, tick int64) (protocol.MatchResultEvent, error) {
	if matchID == "" {
		return protocol.MatchResultEvent{}, protocol.ProtocolError{Code: protocol.ErrorInvalidCommand, Message: "matchId is required"}
	}
	if r.Status != protocol.RoomStatusInBattle {
		return protocol.MatchResultEvent{}, protocol.ProtocolError{Code: protocol.ErrorInvalidState, Message: string(r.Status)}
	}
	if !protocol.IsPlayerSeat(loserSeat) {
		return protocol.MatchResultEvent{}, protocol.ProtocolError{Code: protocol.ErrorInvalidSeat, Message: string(loserSeat)}
	}
	if reason == "" {
		reason = protocol.MatchResultReasonForfeit
	}
	if reason != protocol.MatchResultReasonForfeit &&
		reason != protocol.MatchResultReasonAbandon &&
		reason != protocol.MatchResultReasonBattleEnd {
		return protocol.MatchResultEvent{}, protocol.ProtocolError{Code: protocol.ErrorInvalidCommand, Message: "unsupported match result reason"}
	}
	winnerSeat, ok := protocol.OppositeSeat(loserSeat)
	if !ok {
		return protocol.MatchResultEvent{}, protocol.ProtocolError{Code: protocol.ErrorInvalidSeat, Message: string(loserSeat)}
	}
	resultContractVersion := r.Metadata.ResultContractVersion
	if resultContractVersion == "" {
		resultContractVersion = protocol.ResultContractVersion
	}
	rankedPolicyVersion := r.Metadata.RankedPolicyVersion
	if rankedPolicyVersion == "" {
		rankedPolicyVersion = protocol.RankedPolicyVersion
	}
	rankedMode := r.Metadata.RankedMode
	if rankedMode == "" {
		rankedMode = protocol.RankedModeDisabled
	}
	lastCompletedSeq := r.LastCompletedSeq()
	pendingServerSeq := r.PendingServerSeq()
	r.Status = protocol.RoomStatusEnded
	r.PendingCheckpoints = map[int64]PendingCheckpoint{}
	r.AbandonDeadlines = map[protocol.Seat]AbandonDeadline{}
	r.BattleEndReports = map[protocol.Seat]protocol.MatchBattleEndReport{}
	r.BattleEndReportDeadlines = map[protocol.Seat]BattleEndReportDeadline{}
	result := protocol.MatchResultEvent{
		SchemaVersion:          protocol.SchemaVersion,
		ResultContractVersion:  resultContractVersion,
		MatchID:                matchID,
		Status:                 r.Status,
		WinnerSeat:             winnerSeat,
		LoserSeat:              loserSeat,
		Reason:                 reason,
		SourceSeat:             sourceSeat,
		SourceUserID:           sourceUserID,
		ClientRequestID:        clientRequestID,
		ReportClientRequestIDs: reportClientRequestIDs,
		ResolvedAtTick:         tick,
		LastCompletedSeq:       lastCompletedSeq,
		PendingServerSeq:       pendingServerSeq,
		CheckpointBinding:      r.checkpointBindingSnapshot(matchID, lastCompletedSeq, pendingServerSeq),
		Participants:           r.participantsSnapshot(matchID),
		TimeoutPolicy:          r.timeoutPolicySnapshot(),
		RankedEligible:         false,
		RankedReason:           protocol.MatchResultRankedReasonMVPDisabled,
		RankedPolicyVersion:    rankedPolicyVersion,
		RankedMode:             rankedMode,
	}
	result.ResultID = protocol.ComputeMatchResultID(result)
	if err := protocol.ValidateMatchResultEvent(result); err != nil {
		return protocol.MatchResultEvent{}, err
	}
	r.TerminalResult = &result
	return result, nil
}

func (r *RoomRuntime) checkpointBindingSnapshot(matchID string, lastCompletedSeq int64, pendingServerSeq int64) protocol.MatchCheckpointBinding {
	startHash := r.StartHash
	if startHash == "" && r.StartSnapshot != nil {
		startHash = r.StartSnapshot.StartHash
	}
	if startHash == "" {
		startHash = r.LastAcceptedHash
	}
	terminalCheckpoint := false
	if lastCompletedSeq > 0 {
		if result, ok := r.CompletedResults[lastCompletedSeq]; ok {
			terminalCheckpoint = result.Terminal
		}
	}
	return protocol.MatchCheckpointBinding{
		SchemaVersion:      protocol.SchemaVersion,
		MatchID:            matchID,
		StartHash:          startHash,
		LastCompletedSeq:   lastCompletedSeq,
		LastHash:           r.LastAcceptedHash,
		PendingServerSeq:   pendingServerSeq,
		TerminalCheckpoint: terminalCheckpoint,
	}
}

func (r *RoomRuntime) participantsSnapshot(matchID string) protocol.MatchParticipants {
	redPresence := r.Seats[protocol.SeatRed]
	bluePresence := r.Seats[protocol.SeatBlue]
	redRoster := r.Rosters[protocol.SeatRed]
	blueRoster := r.Rosters[protocol.SeatBlue]
	return protocol.MatchParticipants{
		SchemaVersion: protocol.SchemaVersion,
		MatchID:       matchID,
		Red: protocol.MatchParticipant{
			Seat:                protocol.SeatRed,
			UserID:              redPresence.UserID,
			Username:            redPresence.Username,
			RosterHashAlgorithm: redRoster.HashAlgorithm,
			RosterHash:          redRoster.RosterHash,
			VisualStatusSummary: redRoster.VisualStatusSummary,
		},
		Blue: protocol.MatchParticipant{
			Seat:                protocol.SeatBlue,
			UserID:              bluePresence.UserID,
			Username:            bluePresence.Username,
			RosterHashAlgorithm: blueRoster.HashAlgorithm,
			RosterHash:          blueRoster.RosterHash,
			VisualStatusSummary: blueRoster.VisualStatusSummary,
		},
		GameSHA256:  firstNonEmpty(redRoster.GameSHA256, blueRoster.GameSHA256),
		GameVersion: firstNonEmpty(redRoster.GameVersion, blueRoster.GameVersion),
		ModVersion:  firstNonEmpty(redRoster.ModVersion, blueRoster.ModVersion),
		RulesetHash: firstNonEmpty(redRoster.RulesetHash, blueRoster.RulesetHash),
		Scenario:    r.Metadata.Scenario,
		MapSeed:     r.Metadata.MapSeed,
		CombatSeed:  r.Metadata.CombatSeed,
	}
}

func (r *RoomRuntime) timeoutPolicySnapshot() protocol.MatchTimeoutPolicy {
	policy := defaultTimeoutPolicy()
	policy.CommandResultTimeoutTicks = effectiveTimeout(r.CommandResultTimeoutTicks, DefaultCommandResultTimeoutTicks)
	policy.AbandonGraceTicks = effectiveTimeout(r.AbandonGraceTicks, DefaultAbandonGraceTicks)
	policy.BattleEndReportTimeoutTicks = effectiveTimeout(r.BattleEndReportTimeoutTicks, DefaultBattleEndReportTimeoutTicks)
	return policy
}

func (r *RoomRuntime) RecordBattleEndReport(report protocol.MatchBattleEndReport, tick int64) (protocol.MatchResultEvent, bool, error) {
	if err := protocol.ValidateMatchBattleEndReport(report); err != nil {
		return protocol.MatchResultEvent{}, false, err
	}
	if r.Status != protocol.RoomStatusInBattle {
		return protocol.MatchResultEvent{}, false, protocol.ProtocolError{Code: protocol.ErrorInvalidState, Message: string(r.Status)}
	}
	if existing, ok := r.BattleEndReports[report.Seat]; ok {
		if !sameBattleEndOutcome(existing, report) {
			r.Status = protocol.RoomStatusDesynced
			return protocol.MatchResultEvent{}, false, protocol.ProtocolError{Code: protocol.ErrorDesync, Message: fmt.Sprintf("conflicting battle end report for %s", report.Seat)}
		}
		return r.ResolveBattleEndReports(report.MatchID, tick)
	}
	if err := r.validateBattleEndReportWatermark(report); err != nil {
		return protocol.MatchResultEvent{}, false, err
	}
	r.BattleEndReports[report.Seat] = report
	r.BattleEndReportDeadlines[report.Seat] = BattleEndReportDeadline{
		Seat:         report.Seat,
		AcceptedTick: tick,
		DeadlineTick: r.battleEndReportDeadline(tick),
	}
	return r.ResolveBattleEndReports(report.MatchID, tick)
}

func (r *RoomRuntime) ResolveBattleEndReports(matchID string, tick int64) (protocol.MatchResultEvent, bool, error) {
	if r.Status != protocol.RoomStatusInBattle {
		return protocol.MatchResultEvent{}, false, nil
	}
	red, redOK := r.BattleEndReports[protocol.SeatRed]
	blue, blueOK := r.BattleEndReports[protocol.SeatBlue]
	if !redOK || !blueOK {
		return protocol.MatchResultEvent{}, false, nil
	}
	if !sameBattleEndOutcome(red, blue) {
		r.Status = protocol.RoomStatusDesynced
		return protocol.MatchResultEvent{}, false, protocol.ProtocolError{
			Code:    protocol.ErrorDesync,
			Message: fmt.Sprintf("conflicting battle end reports: red winner=%s loser=%s; blue winner=%s loser=%s", red.WinnerSeat, red.LoserSeat, blue.WinnerSeat, blue.LoserSeat),
		}
	}
	if r.PendingServerSeq() != 0 {
		return protocol.MatchResultEvent{}, false, nil
	}
	reportIDs := map[protocol.Seat]string{}
	if red.ClientRequestID != "" {
		reportIDs[protocol.SeatRed] = red.ClientRequestID
	}
	if blue.ClientRequestID != "" {
		reportIDs[protocol.SeatBlue] = blue.ClientRequestID
	}
	if len(reportIDs) == 0 {
		reportIDs = nil
	}
	result, err := r.resolveMatchResult(matchID, red.LoserSeat, red.WinnerSeat, "", protocol.MatchResultReasonBattleEnd, "", reportIDs, tick)
	if err != nil {
		return protocol.MatchResultEvent{}, false, err
	}
	return result, true, nil
}

func (r *RoomRuntime) validateBattleEndReportWatermark(report protocol.MatchBattleEndReport) error {
	if report.LastHash == "" {
		return protocol.ProtocolError{Code: protocol.ErrorInvalidCommand, Message: "battle end report lastHash is required"}
	}
	if report.LastHash != r.LastAcceptedHash {
		r.Status = protocol.RoomStatusDesynced
		return protocol.ProtocolError{
			Code:    protocol.ErrorDesync,
			Message: fmt.Sprintf("battle end report hash mismatch: seat=%s got=%s want=%s", report.Seat, report.LastHash, r.LastAcceptedHash),
		}
	}
	lastCompletedSeq := r.LastCompletedSeq()
	pendingServerSeq := r.PendingServerSeq()
	if report.LastCompletedSeq == lastCompletedSeq {
		return nil
	}
	if pendingServerSeq != 0 && report.LastCompletedSeq == pendingServerSeq {
		return nil
	}
	r.Status = protocol.RoomStatusDesynced
	return protocol.ProtocolError{
		Code:    protocol.ErrorDesync,
		Message: fmt.Sprintf("battle end report seq mismatch: seat=%s got=%d completed=%d pending=%d", report.Seat, report.LastCompletedSeq, lastCompletedSeq, pendingServerSeq),
	}
}

func (r *RoomRuntime) ExpireBattleEndReports(tick int64) error {
	if r.Status != protocol.RoomStatusInBattle {
		return nil
	}
	if len(r.BattleEndReports) == 0 || len(r.BattleEndReports) >= 2 {
		return nil
	}
	var expired BattleEndReportDeadline
	for seat, deadline := range r.BattleEndReportDeadlines {
		if deadline.DeadlineTick == 0 || tick < deadline.DeadlineTick {
			continue
		}
		if expired.Seat == "" ||
			deadline.DeadlineTick < expired.DeadlineTick ||
			(deadline.DeadlineTick == expired.DeadlineTick && seat < expired.Seat) {
			expired = deadline
		}
	}
	if expired.Seat == "" {
		return nil
	}
	r.Status = protocol.RoomStatusDesynced
	return protocol.ProtocolError{
		Code:    protocol.ErrorDesync,
		Message: fmt.Sprintf("battle end report timeout: seat=%s acceptedTick=%d deadlineTick=%d tick=%d reports=%d", expired.Seat, expired.AcceptedTick, expired.DeadlineTick, tick, len(r.BattleEndReports)),
	}
}

func (r *RoomRuntime) ExpireAbandonedSeats(matchID string, tick int64) (protocol.MatchResultEvent, bool, error) {
	if r.Status != protocol.RoomStatusInBattle {
		return protocol.MatchResultEvent{}, false, nil
	}
	var expired AbandonDeadline
	for seat, deadline := range r.AbandonDeadlines {
		if deadline.DeadlineTick == 0 || tick < deadline.DeadlineTick {
			continue
		}
		if expired.Seat == "" ||
			deadline.DeadlineTick < expired.DeadlineTick ||
			(deadline.DeadlineTick == expired.DeadlineTick && seat < expired.Seat) {
			expired = deadline
		}
	}
	if expired.Seat == "" {
		return protocol.MatchResultEvent{}, false, nil
	}
	result, err := r.ResolveMatchResult(matchID, expired.Seat, expired.UserID, protocol.MatchResultReasonAbandon, "", tick)
	if err != nil {
		return protocol.MatchResultEvent{}, false, err
	}
	return result, true, nil
}

func (r *RoomRuntime) RoomState(roomID string) protocol.RoomState {
	metadata := r.Metadata
	metadata.TimeoutPolicy = r.timeoutPolicySnapshot()
	state := protocol.RoomState{
		SchemaVersion:        protocol.SchemaVersion,
		RoomID:               roomID,
		Status:               r.Status,
		Metadata:             metadata,
		Result:               r.TerminalResult,
		LastAcceptedHash:     r.LastAcceptedHash,
		LastCompletedSeq:     r.LastCompletedSeq(),
		NextServerSeq:        r.NextServerSeq,
		ExpectedActiveSeat:   r.ExpectedActiveSeat,
		ExpectedActiveNetID:  r.ExpectedActiveNetID,
		PendingServerSeq:     r.PendingServerSeq(),
		CommandLogLength:     len(r.CommandLog),
		CompletedResultCount: len(r.CompletedResults),
	}
	if red, ok := r.Seats[protocol.SeatRed]; ok {
		state.RedUserID = red.UserID
		state.RedConnected = red.SessionID != ""
	}
	if blue, ok := r.Seats[protocol.SeatBlue]; ok {
		state.BlueUserID = blue.UserID
		state.BlueConnected = blue.SessionID != ""
	}
	for _, spectator := range r.Spectators {
		state.Spectators = append(state.Spectators, spectator.UserID)
	}
	return state
}

func (r *RoomRuntime) findSeatByUserID(userID string) (protocol.Seat, bool) {
	for seat, presence := range r.Seats {
		if presence.UserID == userID {
			return seat, true
		}
	}
	return "", false
}

func (r *RoomRuntime) findSpectatorSessionByUserID(userID string) (string, bool) {
	for sessionID, presence := range r.Spectators {
		if presence.UserID == userID {
			return sessionID, true
		}
	}
	return "", false
}

func commandIdempotencyKey(seat protocol.Seat, clientCommandID string) string {
	return fmt.Sprintf("%s:%s", seat, clientCommandID)
}

func sameBattleEndOutcome(a protocol.MatchBattleEndReport, b protocol.MatchBattleEndReport) bool {
	return a.WinnerSeat == b.WinnerSeat && a.LoserSeat == b.LoserSeat
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func effectiveTimeout(value int64, fallback int64) int64 {
	if value > 0 {
		return value
	}
	return fallback
}

func (r *RoomRuntime) commandResultDeadline(tick int64) int64 {
	return tick + effectiveTimeout(r.CommandResultTimeoutTicks, DefaultCommandResultTimeoutTicks)
}

func (r *RoomRuntime) abandonDeadline(tick int64) int64 {
	return tick + effectiveTimeout(r.AbandonGraceTicks, DefaultAbandonGraceTicks)
}

func (r *RoomRuntime) battleEndReportDeadline(tick int64) int64 {
	return tick + effectiveTimeout(r.BattleEndReportTimeoutTicks, DefaultBattleEndReportTimeoutTicks)
}

func (r *RoomRuntime) SeatForSession(sessionID string) (protocol.Seat, bool) {
	for seat, presence := range r.Seats {
		if presence.SessionID == sessionID {
			return seat, true
		}
	}
	if _, ok := r.Spectators[sessionID]; ok {
		return protocol.SeatSpectator, true
	}
	return "", false
}
