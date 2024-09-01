package matches

const (
	OpCodeMatchTerminate = iota + 100
	OpCodeMatchFactionDispatch
	OpCodeBattleEnd
)

const (
	OpCodeOpponentReady = iota + 200
	OpCodeMatchStart
)

const (
	CombatResultNone = iota
	CombatResultRedSideWin
	CombatResultBlueSideWin
	CombatResultDraw
)

const (
	FeatureCustomMatchName = iota
	FeatureCustomMatchDescription
)

func convertMatchResult(result, faction int) int {
	if faction == FactionRedSide {
		if result == CombatResultRedSideWin {
			return CombatResultDraw
		}
		return CombatResultBlueSideWin
	}
	if faction == FactionBlueSide {
		if result == CombatResultBlueSideWin {
			return CombatResultDraw
		}
		return CombatResultRedSideWin
	}
	return CombatResultNone
}
