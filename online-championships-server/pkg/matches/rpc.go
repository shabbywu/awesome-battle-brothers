package matches

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/heroiclabs/nakama-common/runtime"
	"online-championships/pkg/protocol"
)

func RpcCreateMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var createdParams MatchCreateParams
	if err := json.Unmarshal([]byte(payload), &createdParams); err != nil {
		return "", runtime.NewError("unable to unmarshal payload", 13)
	}

	matchId, err := nk.MatchCreate(ctx, "standard_match", map[string]interface{}{
		"createdParams": payload,
	})
	if err != nil {
		logger.Error("MatchCreate err: %v", err)
		return "", err
	}
	match, err := nk.MatchGet(ctx, matchId)
	if err != nil {
		logger.Error("MatchGet err: %v", err)
		return "", err
	}
	var label MatchLabel
	_ = json.Unmarshal([]byte(match.GetLabel().String()), &label)
	return string(DumpEvent(MatchCreatedEvent{
		Id:          matchId,
		Name:        label.Name,
		Description: label.Description,
		Metadata:    label.Metadata,
	})), nil
}

func RpcGetCommandLog(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var query CommandLogQueryRequest
	if err := json.Unmarshal([]byte(payload), &query); err != nil {
		return "", runtime.NewError("unable to unmarshal payload", 13)
	}
	if query.MatchID == "" {
		return "", runtime.NewError("matchId is required", 13)
	}
	signal := MatchSignalRequest{
		SchemaVersion: protocol.SchemaVersion,
		Kind:          MatchSignalKindCommandLogQuery,
		CommandLog:    &query,
	}
	response, err := nk.MatchSignal(ctx, query.MatchID, string(DumpEvent(signal)))
	if err != nil {
		logger.Error("MatchSignal command log query err: %v", err)
		return "", err
	}
	return response, nil
}
