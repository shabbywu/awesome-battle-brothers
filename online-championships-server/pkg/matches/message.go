package matches

import (
	"encoding/json"
	"online-championships/pkg/matches/strategicProperties"
)

type MatchCreateParams struct {
	Scenario    int    `json:"scenario"`
	MapSeed     int    `json:"mapSeed,omitempty"`
	CombatSeed  int    `json:"combatSeed,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type MatchCreatedEvent struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
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
