package game

// Monster is an enemy entity
type Monster struct {
	Pos       // Struct embedding
	Rune      rune
	Name      string
	Hitpoints int
	Strength  int
	Speed     float64
}

// NewRat spawns a slow monster
// Why a map? Can iterate over maps fast, and access values by key
//   level.monsters[pos]
//   for key, value := range level.Monster { }
func NewRat(p Pos) *Monster {
	return &Monster{p, 'R', "Rat", 5, 5, 2.0}
}

// NewSpider spawns a fast monster
func NewSpider(p Pos) *Monster {
	return &Monster{p, 'S', "Spider", 10, 10, 1.0}
}

// Update searches for player position
func (m *Monster) Update(level *Level) {
	playerPos := level.Player.Pos
	positions := level.astar(m.Pos, playerPos)
	//fmt.Println(positions)

	// Move 1 position closer if we have a path, and we're not on top of the player (>1)
	if len(positions) > 1 {
		m.Move(positions[1], level)
	}
}

// Move moves towards the player position
func (m *Monster) Move(to Pos, level *Level) {
	_, exists := level.Monsters[to] // Is there something at the position we want to move to?
	if !exists && to != level.Player.Pos {
		delete(level.Monsters, m.Pos) // Delete current, add new
		level.Monsters[to] = m
		m.Pos = to
	}
	//
}
