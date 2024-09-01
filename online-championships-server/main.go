package main

import (
	"context"
	"database/sql"
	"github.com/heroiclabs/nakama-common/runtime"
	"online-championships/pkg/matches"
)

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	if err := initializer.RegisterMatch("standard_match", matches.NewMatch); err != nil {
		logger.Error("[RegisterMatch] error: ", err.Error())
		return err
	}
	if err := initializer.RegisterRpc("createMatch", matches.RpcCreateMatch); err != nil {
		logger.Error("[RegisterRpc] error: ", err.Error())
		return err
	}
	return nil
}
