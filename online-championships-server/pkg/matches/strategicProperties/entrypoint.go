package strategicProperties

import "math/rand"

const (
	ScenarioAuto = iota
	ScenarioArena
)

func BuildStrategicProperties(Scenario, MapSeed, CombatSeed int) StrategicProperties {
	var p StrategicProperties
	switch Scenario {
	case ScenarioAuto:
		fallthrough
	case ScenarioArena:
		p = getArenaStrategicProperties()
	}
	if MapSeed == 0 {
		MapSeed = rand.Int()
	}
	if CombatSeed == 0 {
		CombatSeed = rand.Int()
	}
	p.MapSeed = MapSeed
	p.CombatSeed = CombatSeed
	return p
}
