package strategicProperties

import (
	"online-championships/pkg/matches/strategicProperties/Const"
)

func getArenaStrategicProperties() StrategicProperties {
	p := getDefaultStrategicProperties()
	p.CombatID = "Arena"
	p.LocationTemplate = getDefaultLocationTemplate()
	p.TerrainTemplate = "tactical.arena"
	p.LocationTemplate.Template[0] = "tactical.arena_floor"
	p.Music = Const.Music.ArenaTracks
	p.Ambience[0] = Const.SoundAmbience.ArenaBack
	p.Ambience[1] = Const.SoundAmbience.ArenaFront
	p.AmbienceMinDelay[0] = 0
	p.IsFleeingProhibited = false
	p.IsLootingProhibited = true
	p.IsWithoutAmbience = true
	p.IsFogOfWarVisible = false
	p.IsArenaMode = true
	p.IsAutoAssigningBases = false
	p.RedSideDeploymentType = Const.DeploymentType.Arena
	p.BlueSideDeploymentType = Const.DeploymentType.Arena

	return p
}
