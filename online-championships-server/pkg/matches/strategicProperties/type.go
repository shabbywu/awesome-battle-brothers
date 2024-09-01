package strategicProperties

import "online-championships/pkg/matches/strategicProperties/Const"

type LocationTemplate struct {
	AdditionalRadius int   `json:"AdditionalRadius"`
	CutDownTrees     bool  `json:"CutDownTrees"`
	ForceLineBattle  bool  `json:"ForceLineBattle"`
	Fortification    int   `json:"Fortification"`
	OwnedByFaction   int   `json:"OwnedByFaction"`
	ShiftX           int   `json:"ShiftX"`
	ShiftY           int   `json:"ShiftY"`
	Template         []any `json:"Template"`
}

type StrategicProperties struct {
	AfterDeploymentCallback  any                `json:"AfterDeploymentCallback"`
	AllyBanners              []string           `json:"AllyBanners"`
	Ambience                 [][]Const.Ambience `json:"Ambience"`
	AmbienceMinDelay         []int              `json:"AmbienceMinDelay"`
	BeforeDeploymentCallback any                `json:"BeforeDeploymentCallback"`
	BlueSideDeploymentType   int                `json:"BlueSideDeploymentType"`
	BlueSideEntities         []any              `json:"BlueSideEntities"`
	CombatID                 string             `json:"CombatID"`
	CombatSeed               int                `json:"CombatSeed"`
	EnemyBanners             []any              `json:"EnemyBanners"`
	EnemyDeploymentCallback  any                `json:"EnemyDeploymentCallback"`
	Entities                 []any              `json:"Entities"`
	InCombatAlready          bool               `json:"InCombatAlready"`
	IsArenaMode              bool               `json:"IsArenaMode"`
	IsAttackingLocation      bool               `json:"IsAttackingLocation"`
	IsAutoAssigningBases     bool               `json:"IsAutoAssigningBases"`
	IsFleeingProhibited      bool               `json:"IsFleeingProhibited"`
	IsFogOfWarVisible        bool               `json:"IsFogOfWarVisible"`
	IsLootingProhibited      bool               `json:"IsLootingProhibited"`
	IsUsingSetEntities       bool               `json:"IsUsingSetEntities"`
	IsWithoutAmbience        bool               `json:"IsWithoutAmbience"`
	LocationTemplate         LocationTemplate   `json:"LocationTemplate"`
	Loot                     []any              `json:"Loot"`
	MapSeed                  int                `json:"MapSeed"`
	Music                    []string           `json:"Music"`
	Parties                  []any              `json:"Parties"`
	PlayerDeploymentCallback any                `json:"PlayerDeploymentCallback"`
	Players                  []any              `json:"Players"`
	RedSideDeploymentType    int                `json:"RedSideDeploymentType"`
	RedSideEntities          []any              `json:"RedSideEntities"`
	TemporaryEnemies         []any              `json:"TemporaryEnemies"`
	TerrainTemplate          string             `json:"TerrainTemplate"`
	Tile                     any                `json:"Tile"`
}
