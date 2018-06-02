package game

import (
	"bufio"
	"fmt"
	"os"
)

// Tile enum is just an alias for a rune (a character in Go)
type Tile rune

const (
	// StoneWall represented by a character
	StoneWall Tile = '#'
	// DirtFloor represented by a character
	DirtFloor Tile = '.'
	// Door represented by a character
	Door Tile = '|'
)

// Level holds the 2D array that represents the map
type Level struct {
	Map [][]Tile
}

// LoadLevelFromFile opens and prints a map
func LoadLevelFromFile(filename string) *Level {
	// Open file
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Read from scanner
	scanner := bufio.NewScanner(file) // *File satisfies io.Reader interface
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	return nil
}
