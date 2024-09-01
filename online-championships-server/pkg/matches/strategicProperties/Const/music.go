package Const

import (
	"encoding/json"
	"log"
)

var Music struct {
	ArenaTracks                    []string   `json:"ArenaTracks"`
	BanditTracks                   []string   `json:"BanditTracks"`
	BarbarianTracks                []string   `json:"BarbarianTracks"`
	BattleTracks                   [][]string `json:"BattleTracks"`
	BeastsTracks                   []string   `json:"BeastsTracks"`
	BeastsTracksSouth              []string   `json:"BeastsTracksSouth"`
	CampfireTracks                 []string   `json:"CampfireTracks"`
	CityStateTracks                []string   `json:"CityStateTracks"`
	CityTracks                     []string   `json:"CityTracks"`
	CivilianTracks                 []string   `json:"CivilianTracks"`
	CreditsTracks                  []string   `json:"CreditsTracks"`
	CrossFadeTime                  int        `json:"CrossFadeTime"`
	DefeatTracks                   []string   `json:"DefeatTracks"`
	GoblinsTracks                  []string   `json:"GoblinsTracks"`
	IntroTracks                    []string   `json:"IntroTracks"`
	MenuTracks                     []string   `json:"MenuTracks"`
	NobleTracks                    []string   `json:"NobleTracks"`
	OrcsTracks                     []string   `json:"OrcsTracks"`
	OrientalBanditTracks           []string   `json:"OrientalBanditTracks"`
	OrientalCityStateTracks        []string   `json:"OrientalCityStateTracks"`
	Retirement1Tracks              []string   `json:"Retirement1Tracks"`
	Retirement2Tracks              []string   `json:"Retirement2Tracks"`
	Retirement3Tracks              []string   `json:"Retirement3Tracks"`
	Retirement4Tracks              []string   `json:"Retirement4Tracks"`
	SouthernIntroTracks            []string   `json:"SouthernIntroTracks"`
	StrongholdTracks               []string   `json:"StrongholdTracks"`
	UndeadTracks                   []string   `json:"UndeadTracks"`
	VictoryTracks                  []string   `json:"VictoryTracks"`
	VillageTracks                  []string   `json:"VillageTracks"`
	WorldmapTracks                 []string   `json:"WorldmapTracks"`
	WorldmapTracksGreaterEvil      []string   `json:"WorldmapTracksGreaterEvil"`
	WorldmapTracksGreaterEvilSouth []string   `json:"WorldmapTracksGreaterEvilSouth"`
	WorldmapTracksSouth            []string   `json:"WorldmapTracksSouth"`
}

func init() {
	data := `{"ArenaTracks":["music/beasts_02.ogg","music/beasts_04.ogg"],"BanditTracks":["music/bandits_01.ogg","music/bandits_02.ogg"],"BarbarianTracks":["music/barbarians_01.ogg","music/barbarians_02.ogg"],"BattleTracks":[[],[],[],[],["music/civilians_01.ogg"],["music/noble_01.ogg","music/noble_02.ogg"],["music/bandits_01.ogg","music/bandits_02.ogg"],["music/beasts_01.ogg","music/beasts_02.ogg","music/beasts_03.ogg"],["music/undead_01.ogg","music/undead_02.ogg","music/undead_03.ogg"],["music/undead_01.ogg","music/undead_02.ogg"],["music/orcs_01.ogg","music/orcs_02.ogg","music/orcs_03.ogg"],["music/goblins_01.ogg","music/goblins_02.ogg"],["music/barbarians_01.ogg","music/barbarians_02.ogg"],["music/barbarians_01.ogg","music/barbarians_02.ogg"],["music/barbarians_01.ogg","music/barbarians_02.ogg"]],"BeastsTracks":["music/beasts_01.ogg","music/beasts_02.ogg","music/beasts_03.ogg"],"BeastsTracksSouth":["music/beasts_01.ogg","music/beasts_02.ogg","music/beasts_04.ogg"],"CampfireTracks":["music/retirement_01.ogg"],"CityStateTracks":["music/city_03.ogg"],"CityTracks":["music/city_01.ogg","music/city_02.ogg"],"CivilianTracks":["music/civilians_01.ogg"],"CreditsTracks":["music/credits.ogg"],"CrossFadeTime":4000,"DefeatTracks":["music/defeat_01.ogg"],"GoblinsTracks":["music/goblins_01.ogg","music/goblins_02.ogg"],"IntroTracks":["music/worldmap_02.ogg"],"MenuTracks":["music/title.ogg"],"NobleTracks":["music/noble_01.ogg","music/noble_02.ogg"],"OrcsTracks":["music/orcs_01.ogg","music/orcs_02.ogg","music/orcs_03.ogg"],"OrientalBanditTracks":["music/gilded_01.ogg","music/beasts_04.ogg"],"OrientalCityStateTracks":["music/gilded_01.ogg","music/gilded_02.ogg"],"Retirement1Tracks":["music/retirement_01.ogg"],"Retirement2Tracks":["music/retirement_02.ogg"],"Retirement3Tracks":["music/credits.ogg"],"Retirement4Tracks":["music/credits.ogg"],"SouthernIntroTracks":["music/worldmap_02.ogg"],"StrongholdTracks":["music/stronghold_01.ogg"],"UndeadTracks":["music/undead_01.ogg","music/undead_02.ogg","music/undead_03.ogg"],"VictoryTracks":["music/victory_01.ogg"],"VillageTracks":["music/village_01.ogg","music/retirement_01.ogg"],"WorldmapTracks":["music/worldmap_03.ogg","music/worldmap_04.ogg","music/worldmap_05.ogg","music/worldmap_06.ogg","music/worldmap_07.ogg","music/worldmap_08.ogg","music/worldmap_09.ogg","music/worldmap_10.ogg"],"WorldmapTracksGreaterEvil":["music/worldmap_03.ogg","music/worldmap_04.ogg","music/worldmap_05.ogg","music/worldmap_06.ogg","music/worldmap_07.ogg","music/worldmap_08.ogg","music/worldmap_09.ogg","music/worldmap_10.ogg"],"WorldmapTracksGreaterEvilSouth":["music/worldmap_02.ogg","music/worldmap_03.ogg","music/worldmap_04.ogg","music/worldmap_05.ogg","music/worldmap_06.ogg","music/worldmap_11.ogg"],"WorldmapTracksSouth":["music/worldmap_02.ogg","music/worldmap_03.ogg","music/worldmap_04.ogg","music/worldmap_05.ogg","music/worldmap_06.ogg","music/worldmap_11.ogg"]}`
	if err := json.Unmarshal([]byte(data), &Music); err != nil {
		log.Fatalf("Unmarshal Music error: %v", err)
	}
}
