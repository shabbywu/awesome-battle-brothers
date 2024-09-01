package matches

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/tidwall/gjson"
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
	label := match.GetLabel().String()
	return string(DumpEvent(MatchCreatedEvent{
		Id:          matchId,
		Name:        gjson.Get(label, "name").String(),
		Description: gjson.Get(label, "description").String(),
	})), nil
}
