package game

import (
	"bufio"
	"math"
	"os"
	"time"
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

// Entity ...
type Entity struct {
	Pos
}

// Player ...
type Player struct {
	Entity
}

// Level holds the 2D array that represents the map
type Level struct {
	Map    [][]Tile
	Player Player
	Debug  map[Pos]bool // Map x/y positions to true/false
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
				t = ClosedDoor
			case '/':
				t = OpenDoor
			case '.':
				t = DirtFloor
			case 'P':
				level.Player.X = x // Set player X,Y
				level.Player.Y = y
				t = Pending // Be a placeholder
			default:
				panic("Invalid character in map!")
			}
			level.Map[y][x] = t
		}
	}

	// Go over the map again
	for y, row := range level.Map {
		for x, tile := range row {
			if tile == Pending {
			SearchLoop:
				// Search adjacent squares for floor tile
				for searchX := x - 1; searchX <= x+1; searchX++ {
					for searchY := y - 1; searchY <= y+1; searchY++ {
						searchTile := level.Map[searchY][searchX]
						switch searchTile {
						case DirtFloor:
							level.Map[y][x] = DirtFloor
							break SearchLoop // label break
						default:
							panic("Error in searchTile")
						}
					}
				}
			}
		}
	}

	return level
}

func canWalk(level *Level, pos Pos) bool {
	// Check tile for solid object
	t := level.Map[pos.Y][pos.X]
	switch t {
	case StoneWall, ClosedDoor, Blank:
		return false
	default:
		return true
	}
}

func checkDoor(level *Level, pos Pos) {
	// Check tile for closed door
	t := level.Map[pos.Y][pos.X]
	if t == ClosedDoor {
		level.Map[pos.Y][pos.X] = OpenDoor
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
		if canWalk(level, Pos{p.X, p.Y - 1}) {
			level.Player.Y--
		} else {
			checkDoor(level, Pos{p.X, p.Y - 1})
		}
	case Down:
		if canWalk(level, Pos{p.X, p.Y + 1}) {
			level.Player.Y++
		} else {
			checkDoor(level, Pos{p.X, p.Y + 1})
		}
	case Left:
		if canWalk(level, Pos{p.X - 1, p.Y}) {
			level.Player.X--
		} else {
			checkDoor(level, Pos{p.X - 1, p.Y})
		}
	case Right:
		if canWalk(level, Pos{p.X + 1, p.Y}) {
			level.Player.X++
		} else {
			checkDoor(level, Pos{p.X + 1, p.Y})
		}
	case Search:
		//bfs(ui, level, level.Player.Pos)
		game.astar(level.Player.Pos, Pos{2, 2})
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
func (game *Game) bfs(start Pos) {
	level := game.Level
	frontier := make([]Pos, 0, 8)      // Start at 8 instead of growing/shrinking frontier
	frontier = append(frontier, start) // Put in front of queue
	visited := make(map[Pos]bool)
	visited[start] = true // We have already visited start, we are on it
	level.Debug = visited

	for len(frontier) > 0 {
		current := frontier[0]
		frontier = frontier[1:] // But first
		for _, next := range getNeighbors(level, current) {
			// Check if it is already visited
			if !visited[next] {
				frontier = append(frontier, next)
				visited[next] = true
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
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

func (game *Game) astar(start, goal Pos) []Pos {
	level := game.Level
	frontier := make(pqueue, 0, 8) // Start at 8 instead of growing/shrinking frontier
	frontier = frontier.push(start, 1)

	cameFrom := make(map[Pos]Pos) // Keep a map of where we came from
	cameFrom[start] = start       // Start didn't come from anywhere

	costSoFar := make(map[Pos]int) // Read to handle varying costs of travel
	costSoFar[start] = 0

	level.Debug = make(map[Pos]bool)

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

			level.Debug = make(map[Pos]bool) // Clear debug tiles
			for _, pos := range path {
				level.Debug[pos] = true
				time.Sleep(100 * time.Millisecond)
			}
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
				level.Debug[next] = true
				time.Sleep(100 * time.Millisecond)
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
