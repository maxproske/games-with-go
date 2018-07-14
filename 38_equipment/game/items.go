package game

// ItemType is a tagged union/discriminating union/sum type
type ItemType int

const (
	// Weapon item type
	Weapon ItemType = iota
	Helmet
	Other
)

// Item is an entity
type Item struct {
	Typ ItemType
	Entity
	power float64
}

// NewSword is an instance of a sword
func NewSword(p Pos) *Item {
	return &Item{
		Typ: Weapon,
		Entity: Entity{
			Pos:  p,
			Name: "Sword",
			Rune: 's',
		},
		power: 2.0,
	}
}

// NewHelmet is an instance of a helmet
func NewHelmet(p Pos) *Item {
	return &Item{
		Typ: Helmet,
		Entity: Entity{
			Pos:  p,
			Name: "Helmet",
			Rune: 'h',
		},
		power: 0.50,
	}
}
