package game

// Monster is an enemy entity
type Monster struct {
	Character
}

// NewRat spawns a slow monster
// Why a map? Can iterate over maps fast, and access values by key
//   level.monsters[pos]
//   for key, value := range level.Monster { }
func NewRat(p Pos) *Monster {
	return &Monster{
		Character: Character{
			Entity: Entity{
				Pos:  p,
				Name: "Rat",
				Rune: 'R',
			},
			Hitpoints:    20,
			Strength:     0,
			Speed:        1.5,
			ActionPoints: 0.0,
			SightRange:   10.0,
			Items:        []*Item{NewSword(Pos{})},
		},
	}
}

// NewSpider spawns a fast monster
func NewSpider(p Pos) *Monster {
	return &Monster{
		Character: Character{
			Entity: Entity{
				Pos:  p,
				Name: "Spider",
				Rune: 'S',
			},
			Hitpoints:    200,
			Strength:     0,
			Speed:        1.0,
			ActionPoints: 0.0,
			SightRange:   10.0,
		},
	}
}

// Update searches for player position
func (m *Monster) Update(level *Level) {
	m.ActionPoints += m.Speed
	playerPos := level.Player.Pos
	apInt := int(m.ActionPoints)
	positions := level.astar(m.Pos, playerPos)
	if len(positions) == 0 {
		// Nothing we can do, pass turn
		m.Pass()
		return
	}
	moveIndex := 1 // Move 1 position closer if we have a path, and we're not on top of the player (>1)
	for i := 0; i < apInt; i++ {
		if moveIndex < len(positions) {
			m.Move(positions[moveIndex], level)
			moveIndex++
			m.ActionPoints--
		}
	}
}

// Pass prevents monsters from building up large sums of action points
func (m *Monster) Pass() {
	m.ActionPoints -= m.Speed
}

// Kill drops monster's items onto its current position
func (m *Monster) Kill(level *Level) {
	// Remove a monster from the map when it is dead.
	// It is safe to delete from a map while iterative over it. (cool!)
	delete(level.Monsters, m.Pos)

	groundItems := level.Items[m.Pos]
	for _, item := range m.Items {
		item.Pos = m.Pos
		groundItems = append(groundItems, item)
	}
	// TODO(max): will overwrite items on that tile
	level.Items[m.Pos] = groundItems
}

// Move moves towards the player position
func (m *Monster) Move(to Pos, level *Level) {
	_, exists := level.Monsters[to] // Is there something at the position we want to move to?
	if !exists && to != level.Player.Pos {
		delete(level.Monsters, m.Pos) // Delete current, add new
		level.Monsters[to] = m
		m.Pos = to
		return
	}
	// If there is another monster in the way, don't attack the player
	if to == level.Player.Pos {
		level.Attack(&m.Character, &level.Player.Character)
		if m.Hitpoints <= 0 {
			// Kill monster and drop any items
			m.Kill(level)
		}
		if level.Player.Hitpoints <= 0 {
			panic("ded")
		}
	}
}
