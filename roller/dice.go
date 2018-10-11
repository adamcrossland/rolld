package roller

import (
	"math/rand"
	"time"
)

// DoRolls takes a RollSpec and perform all of the random number generation
// and math to fulfill the request.
func DoRolls(spec RollSpec) RollResults {
	var localResults RollResults

	rand.Seed(time.Now().Unix())
	var eachSet int64
	for eachSet = 0; eachSet < spec.Times; eachSet++ {
		localResults.Rolls = append(localResults.Rolls, roll(spec))
		localResults.Count++
	}

	return localResults
}

func roll(spec RollSpec) SetResult {
	var result SetResult

	var eachDie int64
	var sidesInt = int(spec.Sides)
	for eachDie = 0; eachDie < spec.DieCount; eachDie++ {
		die := rand.Intn(sidesInt) + 1
		result.Count++
		if len(result.Dies) == 0 || die >= result.Dies[0] {
			result.Dies = append(result.Dies, die)
		} else {
			// We always want to have the smallest value at the
			// zero index
			result.Dies = append([]int{die}, result.Dies...)
		}
	}

	// Calculate the total, accounting for best-of modifiers
	var countIdx int64
	if spec.BestOf > 0 {
		countIdx = spec.DieCount - spec.BestOf
	}
	for ; countIdx < spec.DieCount; countIdx++ {
		result.Total += result.Dies[countIdx]
	}

	result.Total += int(spec.Modifier)

	return result
}
