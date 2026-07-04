package protocol

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const SchemaVersion = 1
const ProtocolVersion = SchemaVersion
const ServerVersion = "online-championships-server-mvp-v1"
const RuntimeVersion = "room-runtime-v1"
const RosterHashAlgorithmSHA256 = "sha256-canonical-v1"
const StateHashAlgorithmSHA256 = "sha256-state-v1"
const ResultContractVersion = "match-result-mvp-unranked-v1"
const RankedPolicyVersion = "ranked-policy-disabled-v1"
const RankedModeDisabled = "disabled"
const TimeoutPolicyVersion = "timeout-policy-mvp-v1"
const TimeoutClockSourceMatchLoopTick = "nakama-match-loop-tick"
const TimeoutActionDesync = "desync"
const TimeoutActionMatchResultAbandon = "match_result_abandon"

const (
	OpCodeRoomState int64 = iota + 300
	OpCodeSeatAssigned
	OpCodeRosterLocked
	OpCodeMatchStart
	OpCodeMatchDesync
	OpCodeMatchResult
	OpCodeCmdWaitTurn
	OpCodeCmdEndTurn
	OpCodeCmdResult
	OpCodeProtocolError
	OpCodeMatchForfeitRequest
	OpCodeMatchBattleEndReport
)

type Seat string

const (
	SeatAuto      Seat = "auto"
	SeatRed       Seat = "red"
	SeatBlue      Seat = "blue"
	SeatSpectator Seat = "spectator"
)

type RoomStatus string

const (
	RoomStatusOpen       RoomStatus = "Open"
	RoomStatusReadyCheck RoomStatus = "ReadyCheck"
	RoomStatusStarting   RoomStatus = "Starting"
	RoomStatusInBattle   RoomStatus = "InBattle"
	RoomStatusPaused     RoomStatus = "Paused"
	RoomStatusDesynced   RoomStatus = "Desynced"
	RoomStatusEnded      RoomStatus = "Ended"
	RoomStatusClosed     RoomStatus = "Closed"
)

type CommandOp string

const (
	CommandWaitTurn CommandOp = "CMD_WAIT_TURN"
	CommandEndTurn  CommandOp = "CMD_END_TURN"
)

type MatchResultReason string

const (
	MatchResultReasonForfeit   MatchResultReason = "forfeit"
	MatchResultReasonAbandon   MatchResultReason = "abandon"
	MatchResultReasonBattleEnd MatchResultReason = "battle_end"
)

const MatchResultRankedReasonMVPDisabled = "mvp-ranked-resolution-disabled"

type ErrorCode string

const (
	ErrorSchemaMismatch ErrorCode = "schema_mismatch"
	ErrorUnknownOpcode  ErrorCode = "unknown_opcode"
	ErrorInvalidState   ErrorCode = "invalid_state"
	ErrorInvalidSeat    ErrorCode = "invalid_seat"
	ErrorInvalidCommand ErrorCode = "invalid_command"
	ErrorDesync         ErrorCode = "desync"
)

type ProtocolError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e ProtocolError) Error() string {
	if e.Message == "" {
		return string(e.Code)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

type CommandEnvelope struct {
	SchemaVersion        int             `json:"schemaVersion"`
	MatchID              string          `json:"matchId"`
	ClientCommandID      string          `json:"clientCommandId,omitempty"`
	ServerSeq            int64           `json:"serverSeq"`
	Seat                 Seat            `json:"seat"`
	ActiveNetID          string          `json:"activeNetId"`
	Op                   CommandOp       `json:"op"`
	Payload              json.RawMessage `json:"payload"`
	PreHash              string          `json:"preHash"`
	Replay               bool            `json:"replay,omitempty"`
	ReplayRequiresResult bool            `json:"replayRequiresResult,omitempty"`
}

type CommandResult struct {
	SchemaVersion   int             `json:"schemaVersion"`
	MatchID         string          `json:"matchId"`
	ServerSeq       int64           `json:"serverSeq"`
	Seat            Seat            `json:"seat"`
	PostHash        string          `json:"postHash"`
	PostSummary     json.RawMessage `json:"postSummary,omitempty"`
	NextActiveNetID string          `json:"nextActiveNetId"`
	NextActiveSeat  Seat            `json:"nextActiveSeat"`
	Round           int             `json:"round"`
	Turn            int             `json:"turn"`
	Terminal        bool            `json:"terminal,omitempty"`
	Replay          bool            `json:"replay,omitempty"`
}

type CommandPostSummary struct {
	SchemaVersion int    `json:"schemaVersion"`
	Algorithm     string `json:"algorithm"`
	MatchID       string `json:"matchId"`
	LastServerSeq int64  `json:"lastServerSeq"`
	Canonical     string `json:"canonical"`
	Digest        string `json:"digest"`
}

type MatchForfeitRequest struct {
	SchemaVersion   int               `json:"schemaVersion"`
	MatchID         string            `json:"matchId"`
	Seat            Seat              `json:"seat"`
	Reason          MatchResultReason `json:"reason,omitempty"`
	ClientRequestID string            `json:"clientRequestId,omitempty"`
}

type MatchBattleEndReport struct {
	SchemaVersion    int    `json:"schemaVersion"`
	MatchID          string `json:"matchId"`
	Seat             Seat   `json:"seat"`
	WinnerSeat       Seat   `json:"winnerSeat"`
	LoserSeat        Seat   `json:"loserSeat"`
	ClientRequestID  string `json:"clientRequestId,omitempty"`
	LastCompletedSeq int64  `json:"lastCompletedSeq"`
	LastHash         string `json:"lastHash,omitempty"`
}

type MatchResultEvent struct {
	SchemaVersion          int                    `json:"schemaVersion"`
	ResultID               string                 `json:"resultId"`
	ResultContractVersion  string                 `json:"resultContractVersion"`
	MatchID                string                 `json:"matchId"`
	Status                 RoomStatus             `json:"status"`
	WinnerSeat             Seat                   `json:"winnerSeat"`
	LoserSeat              Seat                   `json:"loserSeat"`
	Reason                 MatchResultReason      `json:"reason"`
	SourceSeat             Seat                   `json:"sourceSeat"`
	SourceUserID           string                 `json:"sourceUserId,omitempty"`
	ClientRequestID        string                 `json:"clientRequestId,omitempty"`
	ReportClientRequestIDs map[Seat]string        `json:"reportClientRequestIds,omitempty"`
	ResolvedAtTick         int64                  `json:"resolvedAtTick"`
	Replay                 bool                   `json:"replay,omitempty"`
	LastCompletedSeq       int64                  `json:"lastCompletedSeq"`
	PendingServerSeq       int64                  `json:"pendingServerSeq,omitempty"`
	CheckpointBinding      MatchCheckpointBinding `json:"checkpointBinding"`
	Participants           MatchParticipants      `json:"participants"`
	TimeoutPolicy          MatchTimeoutPolicy     `json:"timeoutPolicy"`
	RankedEligible         bool                   `json:"rankedEligible"`
	RankedReason           string                 `json:"rankedReason"`
	RankedPolicyVersion    string                 `json:"rankedPolicyVersion"`
	RankedMode             string                 `json:"rankedMode"`
}

type MatchCheckpointBinding struct {
	SchemaVersion      int    `json:"schemaVersion"`
	MatchID            string `json:"matchId"`
	StartHash          string `json:"startHash"`
	LastCompletedSeq   int64  `json:"lastCompletedSeq"`
	LastHash           string `json:"lastHash"`
	PendingServerSeq   int64  `json:"pendingServerSeq,omitempty"`
	TerminalCheckpoint bool   `json:"terminalCheckpoint"`
}

type MatchParticipant struct {
	Seat                Seat   `json:"seat"`
	UserID              string `json:"userId"`
	Username            string `json:"username,omitempty"`
	RosterHashAlgorithm string `json:"rosterHashAlgorithm"`
	RosterHash          string `json:"rosterHash"`
	VisualStatusSummary string `json:"visualStatusSummary"`
}

type MatchParticipants struct {
	SchemaVersion int              `json:"schemaVersion"`
	MatchID       string           `json:"matchId"`
	Red           MatchParticipant `json:"red"`
	Blue          MatchParticipant `json:"blue"`
	GameSHA256    string           `json:"gameSha256"`
	GameVersion   string           `json:"gameVersion"`
	ModVersion    string           `json:"modVersion"`
	RulesetHash   string           `json:"rulesetHash"`
	Scenario      int              `json:"scenario"`
	MapSeed       int              `json:"mapSeed"`
	CombatSeed    int              `json:"combatSeed"`
}

type MatchTimeoutPolicy struct {
	SchemaVersion                int    `json:"schemaVersion"`
	PolicyVersion                string `json:"policyVersion"`
	TickRate                     int    `json:"tickRate"`
	ClockSource                  string `json:"clockSource"`
	CommandResultTimeoutTicks    int64  `json:"commandResultTimeoutTicks"`
	CommandResultTimeoutAction   string `json:"commandResultTimeoutAction"`
	AbandonGraceTicks            int64  `json:"abandonGraceTicks"`
	AbandonGraceExpiredAction    string `json:"abandonGraceExpiredAction"`
	BattleEndReportTimeoutTicks  int64  `json:"battleEndReportTimeoutTicks"`
	BattleEndReportTimeoutAction string `json:"battleEndReportTimeoutAction"`
}

type RosterSnapshot struct {
	SchemaVersion       int             `json:"schemaVersion"`
	CodecVersion        int             `json:"codecVersion"`
	GameSHA256          string          `json:"gameSha256"`
	GameVersion         string          `json:"gameVersion"`
	ModVersion          string          `json:"modVersion"`
	RulesetHash         string          `json:"rulesetHash"`
	OwnerUserID         string          `json:"ownerUserId"`
	Seat                Seat            `json:"seat"`
	Entries             json.RawMessage `json:"entries"`
	HashAlgorithm       string          `json:"hashAlgorithm"`
	RosterHash          string          `json:"rosterHash"`
	VisualStatusSummary string          `json:"visualStatusSummary"`
}

type RosterEntry struct {
	NetID            string            `json:"netId"`
	OwnerSeat        Seat              `json:"ownerSeat"`
	EntryIndex       int               `json:"entryIndex"`
	ScriptName       string            `json:"scriptName"`
	PlaceInFormation int               `json:"placeInFormation"`
	Faction          int               `json:"faction"`
	LogicState       RosterLogicState  `json:"logicState"`
	VisualState      RosterVisualState `json:"visualState"`
	EntryHash        string            `json:"entryHash"`
}

type RosterLogicState struct {
	Source         string `json:"source"`
	SchemaVersion  int    `json:"schemaVersion"`
	BufferEncoding string `json:"bufferEncoding"`
	BufferSize     int    `json:"bufferSize"`
	Buffer         string `json:"buffer"`
	BufferChecksum string `json:"bufferChecksum"`
	LogicHash      string `json:"logicHash"`
	DecodeStatus   string `json:"decodeStatus"`
	DecodePolicy   string `json:"decodePolicy"`
}

type RosterVisualState struct {
	Source              string `json:"source"`
	Status              string `json:"status"`
	NativeCoverage      string `json:"nativeCoverage"`
	PreCombatVisualHash string `json:"preCombatVisualHash,omitempty"`
	PostLoadVisualHash  string `json:"postLoadVisualHash,omitempty"`
}

type RoomMetadata struct {
	SchemaVersion         int                 `json:"schemaVersion"`
	ProtocolVersion       int                 `json:"protocolVersion"`
	ServerVersion         string              `json:"serverVersion"`
	RuntimeVersion        string              `json:"runtimeVersion"`
	ResultContractVersion string              `json:"resultContractVersion"`
	RankedPolicyVersion   string              `json:"rankedPolicyVersion"`
	RankedMode            string              `json:"rankedMode"`
	Name                  string              `json:"name"`
	Description           string              `json:"description,omitempty"`
	Scenario              int                 `json:"scenario"`
	MapSeed               int                 `json:"mapSeed"`
	CombatSeed            int                 `json:"combatSeed"`
	CreatedAt             string              `json:"createdAt"`
	Spectators            RoomSpectatorConfig `json:"spectators"`
	TimeoutPolicy         MatchTimeoutPolicy  `json:"timeoutPolicy"`
}

type RoomSpectatorConfig struct {
	Enabled bool `json:"enabled"`
	Limit   int  `json:"limit"`
}

type RoomState struct {
	SchemaVersion        int               `json:"schemaVersion"`
	RoomID               string            `json:"roomId"`
	Status               RoomStatus        `json:"status"`
	Metadata             RoomMetadata      `json:"metadata"`
	Result               *MatchResultEvent `json:"result,omitempty"`
	RedUserID            string            `json:"redUserId,omitempty"`
	RedConnected         bool              `json:"redConnected,omitempty"`
	BlueUserID           string            `json:"blueUserId,omitempty"`
	BlueConnected        bool              `json:"blueConnected,omitempty"`
	Spectators           []string          `json:"spectators,omitempty"`
	LastAcceptedHash     string            `json:"lastAcceptedHash,omitempty"`
	LastCompletedSeq     int64             `json:"lastCompletedSeq"`
	NextServerSeq        int64             `json:"nextServerSeq"`
	ExpectedActiveSeat   Seat              `json:"expectedActiveSeat,omitempty"`
	ExpectedActiveNetID  string            `json:"expectedActiveNetId,omitempty"`
	PendingServerSeq     int64             `json:"pendingServerSeq,omitempty"`
	CommandLogLength     int               `json:"commandLogLength"`
	CompletedResultCount int               `json:"completedResultCount"`
}

type SeatAssignedEvent struct {
	SchemaVersion int    `json:"schemaVersion"`
	Seat          Seat   `json:"seat"`
	UserID        string `json:"userId"`
	SessionID     string `json:"sessionId"`
}

type RosterLockedEvent struct {
	SchemaVersion int    `json:"schemaVersion"`
	Seat          Seat   `json:"seat"`
	RosterHash    string `json:"rosterHash"`
}

type MatchStartSnapshot struct {
	SchemaVersion       int            `json:"schemaVersion"`
	RoomID              string         `json:"roomId,omitempty"`
	StartHash           string         `json:"startHash"`
	MapSeed             int            `json:"mapSeed"`
	CombatSeed          int            `json:"combatSeed"`
	RulesetHash         string         `json:"rulesetHash"`
	ExpectedActiveSeat  Seat           `json:"expectedActiveSeat"`
	ExpectedActiveNetID string         `json:"expectedActiveNetId"`
	RedRoster           RosterSnapshot `json:"redRoster"`
	BlueRoster          RosterSnapshot `json:"blueRoster"`
	StrategicProperties any            `json:"strategicProperties"`
}

type MatchDesyncEvent struct {
	SchemaVersion int                 `json:"schemaVersion"`
	ServerSeq     int64               `json:"serverSeq,omitempty"`
	Code          ErrorCode           `json:"code"`
	Message       string              `json:"message"`
	Checkpoint    *CheckpointMismatch `json:"checkpoint,omitempty"`
}

type ProtocolErrorEvent struct {
	SchemaVersion   int       `json:"schemaVersion"`
	ServerSeq       int64     `json:"serverSeq,omitempty"`
	Code            ErrorCode `json:"code"`
	Message         string    `json:"message"`
	OpCode          int64     `json:"opCode,omitempty"`
	ClientCommandID string    `json:"clientCommandId,omitempty"`
}

type CheckpointMismatch struct {
	ServerSeq  int64         `json:"serverSeq"`
	RedResult  CommandResult `json:"redResult"`
	BlueResult CommandResult `json:"blueResult"`
}

func IsKnownOpcode(opCode int64) bool {
	switch opCode {
	case OpCodeRoomState,
		OpCodeSeatAssigned,
		OpCodeRosterLocked,
		OpCodeMatchStart,
		OpCodeMatchDesync,
		OpCodeMatchResult,
		OpCodeCmdWaitTurn,
		OpCodeCmdEndTurn,
		OpCodeCmdResult,
		OpCodeProtocolError,
		OpCodeMatchForfeitRequest,
		OpCodeMatchBattleEndReport:
		return true
	default:
		return false
	}
}

func OpcodeForCommand(op CommandOp) (int64, bool) {
	switch op {
	case CommandWaitTurn:
		return OpCodeCmdWaitTurn, true
	case CommandEndTurn:
		return OpCodeCmdEndTurn, true
	default:
		return 0, false
	}
}

func ValidateSchema(version int) error {
	if version != SchemaVersion {
		return ProtocolError{
			Code:    ErrorSchemaMismatch,
			Message: fmt.Sprintf("got %d, want %d", version, SchemaVersion),
		}
	}
	return nil
}

func IsPlayerSeat(seat Seat) bool {
	return seat == SeatRed || seat == SeatBlue
}

func OppositeSeat(seat Seat) (Seat, bool) {
	if seat == SeatRed {
		return SeatBlue, true
	}
	if seat == SeatBlue {
		return SeatRed, true
	}
	return "", false
}

func IsJoinSeat(seat Seat) bool {
	return IsPlayerSeat(seat) || seat == SeatAuto || seat == SeatSpectator
}

func IsSupportedCommand(op CommandOp) bool {
	return op == CommandWaitTurn || op == CommandEndTurn
}

func ValidateClientCommand(cmd CommandEnvelope) error {
	if err := ValidateSchema(cmd.SchemaVersion); err != nil {
		return err
	}
	if cmd.MatchID == "" {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "matchId is required"}
	}
	if cmd.ClientCommandID == "" {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "clientCommandId is required"}
	}
	if !IsPlayerSeat(cmd.Seat) {
		return ProtocolError{Code: ErrorInvalidSeat, Message: "command sender must be red or blue"}
	}
	if cmd.ActiveNetID == "" {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "activeNetId is required"}
	}
	if !IsSupportedCommand(cmd.Op) {
		return ProtocolError{Code: ErrorUnknownOpcode, Message: string(cmd.Op)}
	}
	if cmd.PreHash == "" {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "preHash is required"}
	}
	if cmd.ServerSeq != 0 {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "serverSeq must be assigned by server"}
	}
	if cmd.Replay {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "client command must not set replay"}
	}
	if cmd.ReplayRequiresResult {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "client command must not set replayRequiresResult"}
	}
	return nil
}

func ValidateCommandResult(result CommandResult) error {
	if err := ValidateSchema(result.SchemaVersion); err != nil {
		return err
	}
	if result.MatchID == "" {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "matchId is required"}
	}
	if result.ServerSeq <= 0 {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "serverSeq must be positive"}
	}
	if !IsPlayerSeat(result.Seat) {
		return ProtocolError{Code: ErrorInvalidSeat, Message: "result sender must be red or blue"}
	}
	if err := ValidateSHA256Digest("postHash", result.PostHash); err != nil {
		return err
	}
	if err := ValidateCommandPostSummary(result); err != nil {
		return err
	}
	if result.Terminal {
		if result.NextActiveNetID != "" {
			return ProtocolError{Code: ErrorInvalidCommand, Message: "terminal result must not set nextActiveNetId"}
		}
		if result.NextActiveSeat != "" {
			return ProtocolError{Code: ErrorInvalidSeat, Message: "terminal result must not set nextActiveSeat"}
		}
	} else {
		if result.NextActiveNetID == "" {
			return ProtocolError{Code: ErrorInvalidCommand, Message: "nextActiveNetId is required"}
		}
		if !IsPlayerSeat(result.NextActiveSeat) {
			return ProtocolError{Code: ErrorInvalidSeat, Message: "nextActiveSeat must be red or blue"}
		}
	}
	if result.Replay {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "client result must not set replay"}
	}
	return nil
}

func ValidateCommandPostSummary(result CommandResult) error {
	if len(result.PostSummary) == 0 {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "postSummary is required"}
	}
	if !json.Valid(result.PostSummary) {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "postSummary must be valid JSON"}
	}
	var summary CommandPostSummary
	if err := json.Unmarshal(result.PostSummary, &summary); err != nil {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "postSummary must be a JSON object"}
	}
	if summary.SchemaVersion != SchemaVersion {
		return ProtocolError{Code: ErrorSchemaMismatch, Message: "postSummary schemaVersion mismatch"}
	}
	if summary.Algorithm != StateHashAlgorithmSHA256 {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "unsupported postSummary algorithm"}
	}
	if summary.MatchID != result.MatchID {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "postSummary matchId mismatch"}
	}
	if summary.LastServerSeq != result.ServerSeq {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "postSummary lastServerSeq mismatch"}
	}
	if summary.Canonical == "" {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "postSummary canonical is required"}
	}
	if summary.Digest != result.PostHash {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "postSummary digest mismatch"}
	}
	if err := ValidateSHA256Digest("postSummary.digest", summary.Digest); err != nil {
		return err
	}
	return nil
}

func ValidateMatchForfeitRequest(request MatchForfeitRequest) error {
	if err := ValidateSchema(request.SchemaVersion); err != nil {
		return err
	}
	if request.MatchID == "" {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "matchId is required"}
	}
	if !IsPlayerSeat(request.Seat) {
		return ProtocolError{Code: ErrorInvalidSeat, Message: "forfeit sender must be red or blue"}
	}
	if request.Reason == "" {
		return nil
	}
	if request.Reason != MatchResultReasonForfeit {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "forfeit request reason must be forfeit"}
	}
	return nil
}

func ValidateMatchBattleEndReport(report MatchBattleEndReport) error {
	if err := ValidateSchema(report.SchemaVersion); err != nil {
		return err
	}
	if report.MatchID == "" {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "matchId is required"}
	}
	if !IsPlayerSeat(report.Seat) {
		return ProtocolError{Code: ErrorInvalidSeat, Message: "battle end reporter must be red or blue"}
	}
	if !IsPlayerSeat(report.WinnerSeat) {
		return ProtocolError{Code: ErrorInvalidSeat, Message: "winnerSeat must be red or blue"}
	}
	if !IsPlayerSeat(report.LoserSeat) {
		return ProtocolError{Code: ErrorInvalidSeat, Message: "loserSeat must be red or blue"}
	}
	if report.WinnerSeat == report.LoserSeat {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "winnerSeat and loserSeat must differ"}
	}
	if report.Seat != report.WinnerSeat && report.Seat != report.LoserSeat {
		return ProtocolError{Code: ErrorInvalidSeat, Message: "reporter seat must be winnerSeat or loserSeat"}
	}
	return nil
}

func ValidateMatchResultEvent(result MatchResultEvent) error {
	if err := ValidateSchema(result.SchemaVersion); err != nil {
		return err
	}
	if result.MatchID == "" {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "matchId is required"}
	}
	if result.ResultContractVersion != ResultContractVersion {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "unsupported result contract version"}
	}
	if result.Status != RoomStatusEnded {
		return ProtocolError{Code: ErrorInvalidState, Message: string(result.Status)}
	}
	if !IsPlayerSeat(result.WinnerSeat) {
		return ProtocolError{Code: ErrorInvalidSeat, Message: "winnerSeat must be red or blue"}
	}
	if !IsPlayerSeat(result.LoserSeat) {
		return ProtocolError{Code: ErrorInvalidSeat, Message: "loserSeat must be red or blue"}
	}
	if result.WinnerSeat == result.LoserSeat {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "winnerSeat and loserSeat must differ"}
	}
	if result.Reason != MatchResultReasonForfeit &&
		result.Reason != MatchResultReasonAbandon &&
		result.Reason != MatchResultReasonBattleEnd {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "unsupported match result reason"}
	}
	if err := ValidateMatchCheckpointBinding(result.CheckpointBinding); err != nil {
		return err
	}
	if result.CheckpointBinding.MatchID != result.MatchID {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "checkpoint binding matchId mismatch"}
	}
	if result.CheckpointBinding.LastCompletedSeq != result.LastCompletedSeq {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "checkpoint binding lastCompletedSeq mismatch"}
	}
	if result.CheckpointBinding.PendingServerSeq != result.PendingServerSeq {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "checkpoint binding pendingServerSeq mismatch"}
	}
	if err := ValidateMatchParticipants(result.Participants); err != nil {
		return err
	}
	if result.Participants.MatchID != result.MatchID {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "participants matchId mismatch"}
	}
	if err := ValidateMatchTimeoutPolicy(result.TimeoutPolicy); err != nil {
		return err
	}
	if result.RankedEligible {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "ranked persistence is disabled for MVP"}
	}
	if result.RankedReason == "" {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "rankedReason is required when rankedEligible is false"}
	}
	if result.RankedPolicyVersion != RankedPolicyVersion {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "unsupported ranked policy version"}
	}
	if result.RankedMode != RankedModeDisabled {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "unsupported ranked mode"}
	}
	if result.ResultID == "" {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "resultId is required"}
	}
	if expected := ComputeMatchResultID(result); result.ResultID != expected {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "resultId does not match result payload"}
	}
	return nil
}

func ValidateMatchCheckpointBinding(binding MatchCheckpointBinding) error {
	if binding.SchemaVersion != SchemaVersion {
		return ProtocolError{Code: ErrorSchemaMismatch, Message: "checkpoint binding schemaVersion mismatch"}
	}
	if binding.MatchID == "" {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "checkpoint binding matchId is required"}
	}
	if binding.StartHash == "" {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "checkpoint binding startHash is required"}
	}
	if binding.LastHash == "" {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "checkpoint binding lastHash is required"}
	}
	if binding.LastCompletedSeq < 0 {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "checkpoint binding lastCompletedSeq must not be negative"}
	}
	if binding.PendingServerSeq < 0 {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "checkpoint binding pendingServerSeq must not be negative"}
	}
	if binding.TerminalCheckpoint && binding.PendingServerSeq != 0 {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "terminal checkpoint binding must not have pendingServerSeq"}
	}
	return nil
}

func ValidateMatchParticipants(participants MatchParticipants) error {
	if participants.SchemaVersion != SchemaVersion {
		return ProtocolError{Code: ErrorSchemaMismatch, Message: "participants schemaVersion mismatch"}
	}
	if participants.MatchID == "" {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "participants matchId is required"}
	}
	if err := validateMatchParticipant(SeatRed, participants.Red); err != nil {
		return err
	}
	if err := validateMatchParticipant(SeatBlue, participants.Blue); err != nil {
		return err
	}
	if len(participants.GameSHA256) != 64 || !isLowerHex(participants.GameSHA256) {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "participants gameSha256 must be 64 lower hex characters"}
	}
	if participants.GameVersion == "" || participants.ModVersion == "" || participants.RulesetHash == "" {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "participants game/mod/ruleset identity is required"}
	}
	return nil
}

func validateMatchParticipant(wantSeat Seat, participant MatchParticipant) error {
	if participant.Seat != wantSeat {
		return ProtocolError{Code: ErrorInvalidSeat, Message: fmt.Sprintf("%s participant seat mismatch", wantSeat)}
	}
	if participant.UserID == "" {
		return ProtocolError{Code: ErrorInvalidCommand, Message: fmt.Sprintf("%s participant userId is required", wantSeat)}
	}
	if participant.RosterHashAlgorithm != RosterHashAlgorithmSHA256 {
		return ProtocolError{Code: ErrorInvalidCommand, Message: fmt.Sprintf("%s participant rosterHashAlgorithm is unsupported", wantSeat)}
	}
	if err := ValidateSHA256Digest(fmt.Sprintf("%s participant rosterHash", wantSeat), participant.RosterHash); err != nil {
		return err
	}
	if participant.VisualStatusSummary == "" {
		return ProtocolError{Code: ErrorInvalidCommand, Message: fmt.Sprintf("%s participant visualStatusSummary is required", wantSeat)}
	}
	return nil
}

func ValidateMatchTimeoutPolicy(policy MatchTimeoutPolicy) error {
	if policy.SchemaVersion != SchemaVersion {
		return ProtocolError{Code: ErrorSchemaMismatch, Message: "timeout policy schemaVersion mismatch"}
	}
	if policy.PolicyVersion != TimeoutPolicyVersion {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "unsupported timeout policy version"}
	}
	if policy.TickRate <= 0 {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "tickRate must be positive"}
	}
	if policy.ClockSource != TimeoutClockSourceMatchLoopTick {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "unsupported timeout policy clockSource"}
	}
	if policy.CommandResultTimeoutTicks <= 0 {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "commandResultTimeoutTicks must be positive"}
	}
	if policy.CommandResultTimeoutAction != TimeoutActionDesync {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "unsupported command result timeout action"}
	}
	if policy.AbandonGraceTicks <= 0 {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "abandonGraceTicks must be positive"}
	}
	if policy.AbandonGraceExpiredAction != TimeoutActionMatchResultAbandon {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "unsupported abandon grace expired action"}
	}
	if policy.BattleEndReportTimeoutTicks <= 0 {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "battleEndReportTimeoutTicks must be positive"}
	}
	if policy.BattleEndReportTimeoutAction != TimeoutActionDesync {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "unsupported battle end report timeout action"}
	}
	return nil
}

func ComputeMatchResultID(result MatchResultEvent) string {
	input := struct {
		SchemaVersion          int                    `json:"schemaVersion"`
		ResultContractVersion  string                 `json:"resultContractVersion"`
		MatchID                string                 `json:"matchId"`
		WinnerSeat             Seat                   `json:"winnerSeat"`
		LoserSeat              Seat                   `json:"loserSeat"`
		Reason                 MatchResultReason      `json:"reason"`
		SourceSeat             Seat                   `json:"sourceSeat"`
		SourceUserID           string                 `json:"sourceUserId,omitempty"`
		ClientRequestID        string                 `json:"clientRequestId,omitempty"`
		ReportClientRequestIDs map[Seat]string        `json:"reportClientRequestIds,omitempty"`
		LastCompletedSeq       int64                  `json:"lastCompletedSeq"`
		PendingServerSeq       int64                  `json:"pendingServerSeq,omitempty"`
		CheckpointBinding      MatchCheckpointBinding `json:"checkpointBinding"`
		Participants           MatchParticipants      `json:"participants"`
		TimeoutPolicy          MatchTimeoutPolicy     `json:"timeoutPolicy"`
		RankedPolicyVersion    string                 `json:"rankedPolicyVersion"`
		RankedMode             string                 `json:"rankedMode"`
	}{
		SchemaVersion:          result.SchemaVersion,
		ResultContractVersion:  result.ResultContractVersion,
		MatchID:                result.MatchID,
		WinnerSeat:             result.WinnerSeat,
		LoserSeat:              result.LoserSeat,
		Reason:                 result.Reason,
		SourceSeat:             result.SourceSeat,
		SourceUserID:           result.SourceUserID,
		ClientRequestID:        result.ClientRequestID,
		ReportClientRequestIDs: result.ReportClientRequestIDs,
		LastCompletedSeq:       result.LastCompletedSeq,
		PendingServerSeq:       result.PendingServerSeq,
		CheckpointBinding:      result.CheckpointBinding,
		Participants:           result.Participants,
		TimeoutPolicy:          result.TimeoutPolicy,
		RankedPolicyVersion:    result.RankedPolicyVersion,
		RankedMode:             result.RankedMode,
	}
	data, _ := json.Marshal(input)
	return FormatSHA256Digest(data)
}

func ValidateRosterSnapshot(roster RosterSnapshot) error {
	if err := ValidateSchema(roster.SchemaVersion); err != nil {
		return err
	}
	if roster.CodecVersion != 1 {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "unsupported roster codec version"}
	}
	if !IsPlayerSeat(roster.Seat) {
		return ProtocolError{Code: ErrorInvalidSeat, Message: "roster seat must be red or blue"}
	}
	if len(roster.GameSHA256) != 64 || !isLowerHex(roster.GameSHA256) {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "roster gameSha256 must be 64 lower hex characters"}
	}
	if roster.GameVersion == "" || roster.ModVersion == "" || roster.RulesetHash == "" {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "roster game/mod/ruleset identity is required"}
	}
	if roster.HashAlgorithm != RosterHashAlgorithmSHA256 {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "unsupported hashAlgorithm"}
	}
	if err := ValidateSHA256Digest("rosterHash", roster.RosterHash); err != nil {
		return err
	}
	if roster.VisualStatusSummary == "" {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "visualStatusSummary is required"}
	}
	entries, err := DecodeRosterEntries(roster)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return ProtocolError{Code: ErrorInvalidCommand, Message: "entries must not be empty"}
	}
	seenNetIDs := map[string]struct{}{}
	for i, entry := range entries {
		if entry.NetID == "" {
			return ProtocolError{Code: ErrorInvalidCommand, Message: "entry netId is required"}
		}
		if _, exists := seenNetIDs[entry.NetID]; exists {
			return ProtocolError{Code: ErrorInvalidCommand, Message: "duplicate entry netId"}
		}
		seenNetIDs[entry.NetID] = struct{}{}
		if entry.OwnerSeat != roster.Seat {
			return ProtocolError{Code: ErrorInvalidSeat, Message: "entry ownerSeat mismatch"}
		}
		if entry.EntryIndex != i {
			return ProtocolError{Code: ErrorInvalidCommand, Message: "entryIndex must match roster order"}
		}
		if entry.NetID != fmt.Sprintf("%s:%d", roster.Seat, i) {
			return ProtocolError{Code: ErrorInvalidCommand, Message: "entry netId must match seat:index"}
		}
		if entry.ScriptName == "" {
			return ProtocolError{Code: ErrorInvalidCommand, Message: "entry scriptName is required"}
		}
		if entry.EntryHash == "" {
			return ProtocolError{Code: ErrorInvalidCommand, Message: "entryHash is required"}
		}
		if entry.LogicState.Source == "" ||
			entry.LogicState.BufferEncoding == "" ||
			entry.LogicState.BufferChecksum == "" ||
			entry.LogicState.LogicHash == "" ||
			entry.LogicState.DecodePolicy == "" {
			return ProtocolError{Code: ErrorInvalidCommand, Message: "entry logicState is incomplete"}
		}
		if err := ValidateSHA256Digest("entryHash", entry.EntryHash); err != nil {
			return err
		}
		if err := ValidateSHA256Digest("logicState.bufferChecksum", entry.LogicState.BufferChecksum); err != nil {
			return err
		}
		if err := ValidateSHA256Digest("logicState.logicHash", entry.LogicState.LogicHash); err != nil {
			return err
		}
		if entry.LogicState.SchemaVersion != 1 {
			return ProtocolError{Code: ErrorInvalidCommand, Message: "entry logicState schemaVersion mismatch"}
		}
		if entry.LogicState.DecodeStatus == "failed" {
			return ProtocolError{Code: ErrorInvalidCommand, Message: "entry logicState decode failed"}
		}
		if entry.VisualState.Source == "" || entry.VisualState.Status == "" || entry.VisualState.NativeCoverage == "" {
			return ProtocolError{Code: ErrorInvalidCommand, Message: "entry visualState is incomplete"}
		}
		if entry.VisualState.Status == "mismatch" {
			return ProtocolError{Code: ErrorInvalidCommand, Message: "entry visualState mismatch"}
		}
	}
	return nil
}

func ValidateSHA256Digest(fieldName, value string) error {
	parts := strings.Split(value, ":")
	if len(parts) != 4 || parts[0] != "sha256" || parts[2] != "len" {
		return ProtocolError{Code: ErrorInvalidCommand, Message: fmt.Sprintf("%s must use sha256:<hex>:len:<n>", fieldName)}
	}
	if len(parts[1]) != 64 || !isLowerHex(parts[1]) {
		return ProtocolError{Code: ErrorInvalidCommand, Message: fmt.Sprintf("%s sha256 hex is invalid", fieldName)}
	}
	if parts[3] == "" {
		return ProtocolError{Code: ErrorInvalidCommand, Message: fmt.Sprintf("%s length is required", fieldName)}
	}
	if _, err := strconv.ParseUint(parts[3], 10, 64); err != nil {
		return ProtocolError{Code: ErrorInvalidCommand, Message: fmt.Sprintf("%s length is invalid", fieldName)}
	}
	return nil
}

func FormatSHA256Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("sha256:%s:len:%d", hex.EncodeToString(sum[:]), len(data))
}

func isLowerHex(value string) bool {
	for _, c := range value {
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') {
			continue
		}
		return false
	}
	return true
}

func DecodeRosterEntries(roster RosterSnapshot) ([]RosterEntry, error) {
	var entries []RosterEntry
	if err := json.Unmarshal(roster.Entries, &entries); err != nil {
		return nil, ProtocolError{Code: ErrorInvalidCommand, Message: "entries must be a JSON array"}
	}
	return entries, nil
}

func FirstRosterNetID(roster RosterSnapshot) (string, error) {
	entries, err := DecodeRosterEntries(roster)
	if err != nil {
		return "", err
	}
	if len(entries) == 0 || entries[0].NetID == "" {
		return "", ProtocolError{Code: ErrorInvalidCommand, Message: "roster has no first netId"}
	}
	return entries[0].NetID, nil
}

func ErrorCodeOf(err error) ErrorCode {
	var protocolErr ProtocolError
	if errors.As(err, &protocolErr) {
		return protocolErr.Code
	}
	return ""
}
