package game

import (
	"bufio"
	"fmt"
	"math"
	"os"
)

// Game contains channels for game and UI threads
type Game struct {
	LevelChans []chan *Level // Send level state to multiple UIs
	InputChan  chan *Input   // Receieve input from multiple UIs
	Level      *Level
}

// NewGame needs to know how many channels to take in
func NewGame(numWindows int, levelPath string) *Game {
	levelChans := make([]chan *Level, numWindows) // 1 level channel for each window
	for i := range levelChans {
		levelChans[i] = make(chan *Level)
	}
	inputChan := make(chan *Input)

	return &Game{levelChans, inputChan, loadLevelFromFile(levelPath)}
}

// InputType ...
type InputType int

const (
	// None input type
	None InputType = iota
	// Up input type
	Up
	// Down input type
	Down
	// Left input type
	Left
	//Right input type
	Right
	// QuitGame triggers on last window closed
	QuitGame
	// CloseWindow triggers on 1 of many windows closed
	CloseWindow
	// Search input type
	Search
)

// Input ...
type Input struct {
	Typ          InputType
	LevelChannel chan *Level
}

// Tile enum is just an alias for a rune (a character in Go)
type Tile rune

const (
	// StoneWall represented by a character
	StoneWall Tile = '#'
	// DirtFloor represented by a character
	DirtFloor Tile = '.'
	// ClosedDoor represented by a character
	ClosedDoor Tile = '|'
	// OpenDoor represented by a character
	OpenDoor Tile = '/'
	// Blank represented by zero
	Blank Tile = 0
	// Pending represented by -1
	Pending Tile = -1
)

// Pos ...
type Pos struct {
	X, Y int
}

// Entity can be items or characters
type Entity struct {
	Pos
	Name string
	Rune rune
}

// Player ...
type Player struct {
	Character
}

// Character ...
type Character struct {
	Entity
	Hitpoints    int
	Strength     int
	Speed        float64
	ActionPoints float64
}

// Level holds the 2D array that represents the map
type Level struct {
	Map      [][]Tile
	Player   *Player
	Monsters map[Pos]*Monster // Pos as key, get back monster
	Events   []string
	EventPos int
	Debug    map[Pos]bool // Map x/y positions to true/false
}

// Attackable is an interface that allows attacking
type Attackable interface {
	GetActionPoints() float64
	SetActionPoints(float64)
	GetHitPoints() int
	SetHitPoints(int)
	GetAttackPower() int
}

// You can use en embedded struct in receiver, but not in parameters
func (c *Character) GetActionPoints() float64   { return c.ActionPoints }
func (c *Character) SetActionPoints(ap float64) { c.ActionPoints = ap }
func (c *Character) GetHitPoints() int          { return c.Hitpoints }
func (c *Character) SetHitPoints(hp int)        { c.Hitpoints = hp }
func (c *Character) GetAttackPower() int        { return c.Strength }

// Attack engages two attackables
func Attack(a1, a2 Attackable) {
	// a1 attacking a2 first
	a1.SetActionPoints(a1.GetActionPoints() - 1)
	a2.SetHitPoints(a2.GetHitPoints() - a1.GetAttackPower())

	// If a2 is alive, a2 attacks back
	// if a2.GetHitPoints() > 0 {
	// 	a2.SetActionPoints(a2.GetActionPoints() - 1)
	// 	a1.SetHitPoints(a1.GetHitPoints() - a2.GetAttackPower())
	// }
}

// AddEvent handles events list
func (level *Level) AddEvent(event string) {
	level.Events[level.EventPos] = event
	level.EventPos++
	if level.EventPos == len(level.Events) {
		level.EventPos = 0 // Loop around to overwrite stale events
	}

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
	level.Events = make([]string, 10)
	level.Player = &Player{} // Player used to not be a pointer

	level.Player.Strength = 20
	level.Player.Hitpoints = 20
	level.Player.Name = "GoMan"
	level.Player.Rune = '@'
	level.Player.Speed = 1.0
	level.Player.ActionPoints = 0

	level.Map = make([][]Tile, len(levelLines))
	level.Monsters = make(map[Pos]*Monster)

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
				t = ClosedDoor
			case '/':
				t = OpenDoor
			case '.':
				t = DirtFloor
			case '@':
				level.Player.X = x // Set player X,Y
				level.Player.Y = y
				t = Pending // Be a placeholder
			case 'R':
				// Rat
				level.Monsters[Pos{x, y}] = NewRat(Pos{x, y})
				t = Pending
			case 'S':
				// Spider
				level.Monsters[Pos{x, y}] = NewSpider(Pos{x, y})
				t = Pending
			default:
				panic("Invalid character in map!")
			}
			level.Map[y][x] = t
		}
	}

	// Go over the map again
	// TODO(max): Use bfs to find first floor tile
	for y, row := range level.Map {
		for x, tile := range row {
			if tile == Pending {
				level.Map[y][x] = level.bfsFloor(Pos{x, y}) // Use bfs to find the nearest floor tile, and send it to it
			}
		}
	}

	return level
}

// Check if x,y is inbounds
func inRange(level *Level, pos Pos) bool {
	return pos.X < len(level.Map[0]) && pos.Y < len(level.Map) && pos.X >= 0 && pos.Y >= 0
}

func canWalk(level *Level, pos Pos) bool {
	if inRange(level, pos) {
		// Check tile for solid object
		t := level.Map[pos.Y][pos.X]
		switch t {
		case StoneWall, ClosedDoor, Blank:
			return false
		default:
			return true
		}
	}
	return false
}

func checkDoor(level *Level, pos Pos) {
	// Check tile for closed door
	t := level.Map[pos.Y][pos.X]
	if t == ClosedDoor {
		level.Map[pos.Y][pos.X] = OpenDoor
	}
}

// Move moves the player unless a monster exists in that location
func (player *Player) Move(to Pos, level *Level) {
	monster, exists := level.Monsters[to] // Is there something at the position we want to move to?
	if !exists {
		player.Pos = to
	} else {
		Attack(level.Player, monster)
		level.AddEvent("Player Attacked " + monster.Name)
		if monster.Hitpoints <= 0 {
			level.AddEvent("You killed the " + monster.Name)
			delete(level.Monsters, monster.Pos) // It is safe to delete from an array while iterating over it
		}

		if level.Player.Hitpoints <= 0 {
			fmt.Println("YOU DIED")
			panic("You Died") // TODO(max): exit properly
		}
	}
}

// Returning a *Level is slow
func (game *Game) handleInput(input *Input) {
	level := game.Level
	p := level.Player
	// Check if the place the player is going to is available
	//fmt.Println(input.Typ)
	switch input.Typ {
	case Up:
		newPos := Pos{p.X, p.Y - 1}
		if canWalk(level, newPos) {
			level.Player.Move(newPos, level)
		} else {
			checkDoor(level, newPos)
		}
	case Down:
		newPos := Pos{p.X, p.Y + 1}
		if canWalk(level, newPos) {
			level.Player.Move(newPos, level)
		} else {
			checkDoor(level, newPos)
		}
	case Left:
		newPos := Pos{p.X - 1, p.Y}
		if canWalk(level, newPos) {
			level.Player.Move(newPos, level)
		} else {
			checkDoor(level, newPos)
		}
	case Right:
		newPos := Pos{p.X + 1, p.Y}
		if canWalk(level, newPos) {
			level.Player.Move(newPos, level)
		} else {
			checkDoor(level, newPos)
		}
	case Search:
		level.astar(level.Player.Pos, Pos{2, 2})
	case CloseWindow:
		close(input.LevelChannel) // Close level input game from
		chanIndex := 0
		for i, c := range game.LevelChans {
			if c == input.LevelChannel {
				chanIndex = i
				break
			}
		}
		// Remove an item from a slice
		game.LevelChans = append(game.LevelChans[:chanIndex], game.LevelChans[chanIndex+1:]...)
	}
}

// Breath-first search for a* algorithm
func (level *Level) bfsFloor(start Pos) Tile {
	frontier := make([]Pos, 0, 8)      // Start at 8 instead of growing/shrinking frontier
	frontier = append(frontier, start) // Put in front of queue
	visited := make(map[Pos]bool)
	visited[start] = true // We have already visited start, we are on it
	//level.Debug = visited

	for len(frontier) > 0 {
		current := frontier[0]

		// If there are no surrounding tiles, search for the nearest floor tile
		currentTile := level.Map[current.Y][current.X]
		switch currentTile {
		case DirtFloor:
			return DirtFloor
		default:
		}

		frontier = frontier[1:] // But first
		for _, next := range getNeighbors(level, current) {
			// Check if it is already visited
			if !visited[next] {
				frontier = append(frontier, next)
				visited[next] = true
			}
		}
	}

	return DirtFloor
}

// Return slice of positions that are adjacent
func getNeighbors(level *Level, pos Pos) []Pos {
	neighbors := make([]Pos, 0, 4)
	dirs := make([]Pos, 0, 4)
	dirs = append(dirs, Pos{pos.X - 1, pos.Y})
	dirs = append(dirs, Pos{pos.X + 1, pos.Y})
	dirs = append(dirs, Pos{pos.X, pos.Y - 1})
	dirs = append(dirs, Pos{pos.X, pos.Y + 1})

	for _, dir := range dirs {
		if canWalk(level, dir) {
			neighbors = append(neighbors, dir)
		}
	}

	return neighbors
}

func (level *Level) astar(start, goal Pos) []Pos {
	frontier := make(pqueue, 0, 8) // Start at 8 instead of growing/shrinking frontier
	frontier = frontier.push(start, 1)

	cameFrom := make(map[Pos]Pos) // Keep a map of where we came from
	cameFrom[start] = start       // Start didn't come from anywhere

	costSoFar := make(map[Pos]int) // Read to handle varying costs of travel
	costSoFar[start] = 0

	//level.Debug = make(map[Pos]bool)

	var current Pos
	for len(frontier) > 0 {
		frontier, current = frontier.pop() // Get starting node from the beginning of it

		if current == goal {
			// Reverse slice
			path := make([]Pos, 0)

			// We've found our path
			p := current
			for p != start {
				path = append(path, p)
				p = cameFrom[p]
			}
			path = append(path, p)

			// Reverse slice
			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
				path[i], path[j] = path[j], path[i]
			}

			//level.Debug = make(map[Pos]bool) // Clear debug tiles
			// for _, pos := range path {
			// 	level.Debug[pos] = true
			// }
			return path
		}

		for _, next := range getNeighbors(level, current) {
			newCost := costSoFar[current] + 1 // Always 1 for now
			_, exists := costSoFar[next]
			if !exists || newCost < costSoFar[next] {
				costSoFar[next] = newCost
				// Manhatten distance (how many nodes in a straight line)
				xDist := int(math.Abs(float64(goal.X - next.X)))
				yDist := int(math.Abs(float64(goal.Y - next.Y)))
				priority := newCost + xDist + yDist
				frontier = frontier.push(next, priority) // Stick a new priority onto the queue
				//level.Debug[next] = true
				cameFrom[next] = current // Update where we came from
			}
		}
	}

	return nil
}

// Run loads the level from file
func (game *Game) Run() {

	// Send level state to all level channels
	for _, lchan := range game.LevelChans {
		lchan <- game.Level
	}

	// Get an input out of our input channel
	for input := range game.InputChan {
		if input.Typ == QuitGame {
			return
		}
		game.handleInput(input) // Pass along the input we got

		// Update monsters
		for _, monster := range game.Level.Monsters {
			monster.Update(game.Level)
		}

		if len(game.LevelChans) == 0 {
			// All the windows have been closed
			return
		}

		// Send game state updates
		for _, lchan := range game.LevelChans {
			lchan <- game.Level
		}
	}
}
