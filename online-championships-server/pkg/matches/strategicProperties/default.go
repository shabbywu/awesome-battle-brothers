package strategicProperties

import (
	"encoding/json"
)

const defaultCombatInfoTemplate = `
{
  "AfterDeploymentCallback": null,
  "AllyBanners": [],
  "Ambience": [
    [],
    []
  ],
  "AmbienceMinDelay": [
    2500,
    2500
  ],
  "BeforeDeploymentCallback": null,
  "BlueSideDeploymentType": 0,
  "BlueSideEntities": [],
  "CombatID": "Default",
  "CombatSeed": 0,
  "EnemyBanners": [],
  "EnemyDeploymentCallback": null,
  "Entities": [],
  "InCombatAlready": false,
  "IsArenaMode": false,
  "IsAttackingLocation": false,
  "IsAutoAssigningBases": true,
  "IsFleeingProhibited": false,
  "IsFogOfWarVisible": true,
  "IsLootingProhibited": false,
  "IsUsingSetEntities": false,
  "IsWithoutAmbience": false,
  "LocationTemplate": null,
  "Loot": [],
  "MapSeed": 0,
  "Music": [],
  "Parties": [],
  "PlayerDeploymentCallback": null,
  "Players": [],
  "RedSideDeploymentType": 0,
  "RedSideEntities": [],
  "TemporaryEnemies": [],
  "TerrainTemplate": null,
  "Tile": null
}
`

const DefaultLocationTemplate = `
{
  "AdditionalRadius": 0,
  "CutDownTrees": false,
  "ForceLineBattle": false,
  "Fortification": 0,
  "OwnedByFaction": 0,
  "ShiftX": 6,
  "ShiftY": 0,
  "Template": [
    "tactical.arena_floor",
    null
  ]
}
`

func getDefaultStrategicProperties() StrategicProperties {
	var p StrategicProperties
	_ = json.Unmarshal([]byte(defaultCombatInfoTemplate), &p)
	return p
}

func getDefaultLocationTemplate() LocationTemplate {
	var t LocationTemplate
	_ = json.Unmarshal([]byte(DefaultLocationTemplate), &t)
	return t
}
