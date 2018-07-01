package game

import (
	"bufio"
	"os"
)

// GameUI provides draw function
type GameUI interface {
	Draw(*Level)
}

// Tile enum is just an alias for a rune (a character in Go)
type Tile rune

const (
	// StoneWall represented by a character
	StoneWall Tile = '#'
	// DirtFloor represented by a character
	DirtFloor Tile = '.'
	// Door represented by a character
	Door Tile = '|'
	// Blank represented by zero
	Blank Tile = 0
)

// Level holds the 2D array that represents the map
type Level struct {
	Map [][]Tile
}

// loadLevelFromFile opens and prints a map
func loadLevelFromFile(filename string) *Level {
	// Open file
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Read from scanner
	scanner := bufio.NewScanner(file) // *File satisfies io.Reader interface
	levelLines := make([]string, 0)
	longestRow := 0 // Map width (length)
	index := 0      // Map height (rows)

	for scanner.Scan() {
		levelLines = append(levelLines, scanner.Text()) // String for each row of our map
		// Keep track of longest line
		if len(levelLines[index]) > longestRow {
			longestRow = len(levelLines[index])
		}
		index++
	}

	level := &Level{}
	level.Map = make([][]Tile, len(levelLines))

	for i := range level.Map {
		level.Map[i] = make([]Tile, longestRow) // Make each row the same length of the longest row (non-jagged slice)
	}

	for y := 0; y < len(level.Map); y++ {
		line := levelLines[y]
		for x, c := range line {
			var t Tile
			switch c {
			case ' ', '\t', '\n', '\r':
				t = Blank
			case '#':
				t = StoneWall
			case '|':
				t = Door
			case '.':
				t = DirtFloor
			default:
				panic("Invalid character in map!")
			}
			level.Map[y][x] = t
		}
	}

	return level
}

// Run loads the level from file
func Run(ui GameUI) {
	level := loadLevelFromFile("game/maps/level1.map")
	ui.Draw(level)
}
