package matches

import (
	"encoding/json"
	"online-championships/pkg/matches/strategicProperties"
	"online-championships/pkg/protocol"
)

type MatchCreateParams struct {
	Scenario    int                           `json:"scenario"`
	MapSeed     int                           `json:"mapSeed,omitempty"`
	CombatSeed  int                           `json:"combatSeed,omitempty"`
	Name        string                        `json:"name,omitempty"`
	Description string                        `json:"description,omitempty"`
	Spectators  *protocol.RoomSpectatorConfig `json:"spectators,omitempty"`
}

type MatchCreatedEvent struct {
	Id          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Metadata    protocol.RoomMetadata `json:"metadata"`
}

type MatchLabel struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Metadata    protocol.RoomMetadata `json:"metadata"`
}

const MatchSignalKindCommandLogQuery = "command-log-query"

const DefaultCommandLogQueryLimit = 256
const MaxCommandLogQueryLimit = 512

type MatchSignalRequest struct {
	SchemaVersion int                     `json:"schemaVersion"`
	Kind          string                  `json:"kind"`
	CommandLog    *CommandLogQueryRequest `json:"commandLog,omitempty"`
}

type CommandLogQueryRequest struct {
	MatchID                 string `json:"matchId"`
	FromServerSeq           int64  `json:"fromServerSeq,omitempty"`
	Limit                   int    `json:"limit,omitempty"`
	IncludeRoomState        bool   `json:"includeRoomState,omitempty"`
	IncludeStartSnapshot    bool   `json:"includeStartSnapshot,omitempty"`
	IncludeCompletedResults bool   `json:"includeCompletedResults,omitempty"`
	IncludePending          bool   `json:"includePending,omitempty"`
	IncludeTerminalResult   bool   `json:"includeTerminalResult,omitempty"`
}

type CommandLogQueryResponse struct {
	SchemaVersion         int                          `json:"schemaVersion"`
	Kind                  string                       `json:"kind"`
	MatchID               string                       `json:"matchId"`
	Status                string                       `json:"status"`
	RoomState             *protocol.RoomState          `json:"roomState,omitempty"`
	StartSnapshot         *protocol.MatchStartSnapshot `json:"startSnapshot,omitempty"`
	Commands              []protocol.CommandEnvelope   `json:"commands"`
	CompletedResults      []protocol.CommandResult     `json:"completedResults,omitempty"`
	PendingCommand        *protocol.CommandEnvelope    `json:"pendingCommand,omitempty"`
	TerminalResult        *protocol.MatchResultEvent   `json:"terminalResult,omitempty"`
	FromServerSeq         int64                        `json:"fromServerSeq"`
	NextFromServerSeq     int64                        `json:"nextFromServerSeq,omitempty"`
	HasMore               bool                         `json:"hasMore"`
	NextServerSeq         int64                        `json:"nextServerSeq"`
	LastCompletedSeq      int64                        `json:"lastCompletedSeq"`
	PendingServerSeq      int64                        `json:"pendingServerSeq,omitempty"`
	CommandLogLength      int                          `json:"commandLogLength"`
	CompletedResultCount  int                          `json:"completedResultCount"`
	MaxLimit              int                          `json:"maxLimit"`
	Persistence           string                       `json:"persistence"`
	CanPromoteMVP         bool                         `json:"canPromoteMvp"`
	CanPromotePersistence bool                         `json:"canPromotePersistence"`
}

type MatchSignalErrorResponse struct {
	SchemaVersion int                `json:"schemaVersion"`
	Kind          string             `json:"kind"`
	Status        string             `json:"status"`
	Code          protocol.ErrorCode `json:"code"`
	Message       string             `json:"message"`
}

type BattleEndEvent struct {
	Faction   int  `json:"faction"`
	IsVictory bool `json:"isVictory"`
}

type FactionDispatchEvent struct {
	Presence string `json:"presence"`
	Faction  int    `json:"faction"`
}

type StartMatchEvent struct {
	PresenceFactions    map[string]int                          `json:"presenceFactions"`
	BlueSideRoster      map[string]interface{}                  `json:"blueSideRoster"`
	RedSideRoster       map[string]interface{}                  `json:"redSideRoster"`
	StrategicProperties strategicProperties.StrategicProperties `json:"strategicProperties"`
}

func DumpEvent(v any) []byte {
	data, _ := json.Marshal(v)
	return data
}

func defaultRoomMetadata() protocol.RoomMetadata {
	return protocol.RoomMetadata{
		SchemaVersion:         protocol.SchemaVersion,
		ProtocolVersion:       protocol.ProtocolVersion,
		ServerVersion:         protocol.ServerVersion,
		RuntimeVersion:        protocol.RuntimeVersion,
		ResultContractVersion: protocol.ResultContractVersion,
		RankedPolicyVersion:   protocol.RankedPolicyVersion,
		RankedMode:            protocol.RankedModeDisabled,
		Name:                  "EmptyRoom",
		CreatedAt:             "unknown",
		Spectators: protocol.RoomSpectatorConfig{
			Enabled: true,
			Limit:   0,
		},
		TimeoutPolicy: defaultTimeoutPolicy(),
	}
}

func defaultTimeoutPolicy() protocol.MatchTimeoutPolicy {
	return protocol.MatchTimeoutPolicy{
		SchemaVersion:                protocol.SchemaVersion,
		PolicyVersion:                protocol.TimeoutPolicyVersion,
		TickRate:                     MatchTickRate,
		ClockSource:                  protocol.TimeoutClockSourceMatchLoopTick,
		CommandResultTimeoutTicks:    DefaultCommandResultTimeoutTicks,
		CommandResultTimeoutAction:   protocol.TimeoutActionDesync,
		AbandonGraceTicks:            DefaultAbandonGraceTicks,
		AbandonGraceExpiredAction:    protocol.TimeoutActionMatchResultAbandon,
		BattleEndReportTimeoutTicks:  DefaultBattleEndReportTimeoutTicks,
		BattleEndReportTimeoutAction: protocol.TimeoutActionDesync,
	}
}

func buildRoomMetadata(params MatchCreateParams, name string, description string, mapSeed int, combatSeed int, createdAt string) protocol.RoomMetadata {
	metadata := defaultRoomMetadata()
	metadata.Name = name
	metadata.Description = description
	metadata.Scenario = params.Scenario
	metadata.MapSeed = mapSeed
	metadata.CombatSeed = combatSeed
	if params.Spectators != nil {
		metadata.Spectators = *params.Spectators
		if metadata.Spectators.Limit < 0 {
			metadata.Spectators.Limit = 0
		}
	}
	if createdAt != "" {
		metadata.CreatedAt = createdAt
	}
	return metadata
}
