package game

// Item is an entity
type Item struct {
	Entity
	UISize float32
}

// NewSword is an instance of a sword
func NewSword(p Pos) *Item {
	return &Item{
		Entity: Entity{
			Pos:  p,
			Name: "Sword",
			Rune: 's',
		},
	}
}

// NewHelmet is an instance of a helmet
func NewHelmet(p Pos) *Item {
	return &Item{
		Entity: Entity{
			Pos:  p,
			Name: "Helmet",
			Rune: 'h',
		},
	}
}
