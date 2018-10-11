// roller
package roller

// RollSpec contains all of the information about a requested die roll
type RollSpec struct {
	Sides    int64
	BestOf   int64
	DieCount int64
	Modifier int64
	Times    int64
}

// SetResult contains all of the information about an individual set of
// die rolls. Any requested roll will have one SetResult unless the roll
// has a times modifier, in which case there will be one SetResult for each.
type SetResult struct {
	Total int
	Count int
	Dies  []int
}

// RollResults contains all of the data associated with a roll. For most
// rolls, this will be a single SetResult, but if the requested roll contains
// a times modifier, the Rolls array will contain multiple SetResults.
type RollResults struct {
	Count int
	Rolls []SetResult
}
