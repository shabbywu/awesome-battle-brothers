package matches

import "github.com/heroiclabs/nakama-common/runtime"

const (
	FactionNone = iota
	FactionRedSide
	FactionBlueSide
)

type Presence struct {
	runtime.Presence
	Faction int
	Roster  map[string]interface{}
	IsReady bool
}

func spawnPlayer(presence runtime.Presence) *Presence {
	return &Presence{
		Presence: presence,
		Faction:  FactionNone,
	}
}
