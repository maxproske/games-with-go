package game

import (
	"bufio"
	"math"
	"os"
	"sort"
	"time"
)

// GameUI provides draw function
type GameUI interface {
	// Seperate draw and input to draw multiple times before handling input
	Draw(*Level)
	GetInput() *Input
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
	// Quit input type
	Quit
	// Search input type
	Search
)

// Input ...
type Input struct {
	Typ InputType
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

type priorityPos struct {
	Pos
	priority int
}

// Implement go sort interface
type priorityArray []priorityPos

func (p priorityArray) Len() int           { return len(p) }
func (p priorityArray) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p priorityArray) Less(i, j int) bool { return p[i].priority < p[j].priority }

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
func handleInput(ui GameUI, level *Level, input *Input) {
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
		astar(ui, level, level.Player.Pos, Pos{2, 1})
	}
}

// Breath-first search for a* algorithm
func bfs(ui GameUI, level *Level, start Pos) {
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
				ui.Draw(level) //  Every time we update the visited map, we are updating the debug map
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

func astar(ui GameUI, level *Level, start, goal Pos) []Pos {
	frontier := make(priorityArray, 0, 8)
	frontier = append(frontier, priorityPos{start, 1})

	cameFrom := make(map[Pos]Pos) // Keep a map of where we came from
	cameFrom[start] = start       // Start didn't come from anywhere

	costSoFar := make(map[Pos]int) // Read to handle varying costs of travel
	costSoFar[start] = 0

	level.Debug = make(map[Pos]bool)

	for len(frontier) > 0 {
		// Constantly sort queue for highest priority
		sort.Stable(frontier)  // Slow priority queue, make a real one!
		current := frontier[0] // Get starting node from the beginning of it

		if current.Pos == goal {
			// Reverse slice
			path := make([]Pos, 0)

			// We've found our path
			p := current.Pos
			for p != start {
				path = append(path, p)
				p = cameFrom[p]
			}
			path = append(path, p)

			// Reverse slice
			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
				path[i], path[j] = path[j], path[i]
			}

			lastPos := Pos{0, 0}
			for _, pos := range path {

				level.Debug[pos] = true
				ui.Draw(level)
				time.Sleep(100 * time.Millisecond)
				lastPos = pos
			}
			if lastPos.X != 0 {
				level.Debug[lastPos] = true
			}

			//fmt.Println(p)
			//fmt.Println("Done!")
			break
		}

		frontier = frontier[1:] // Shrink by popping
		for _, next := range getNeighbors(level, current.Pos) {
			newCost := costSoFar[current.Pos] + 1 // Always 1 for now
			_, exists := costSoFar[next]
			if !exists || newCost < costSoFar[next] {
				costSoFar[next] = newCost
				// Manhatten distance (how many nodes in a straight line)
				xDist := int(math.Abs(float64(goal.X - next.X)))
				yDist := int(math.Abs(float64(goal.Y - next.Y)))
				priority := newCost + xDist + yDist
				frontier = append(frontier, priorityPos{next, priority}) // Stick a new priority onto the queue
				cameFrom[next] = current.Pos                             // Update where we came from
				//level.Debug[next] = true
			}
		}
	}

	return nil
}

// Run loads the level from file
func Run(ui GameUI) {
	level := loadLevelFromFile("game/maps/level1.map")
	for {
		ui.Draw(level)
		input := ui.GetInput()

		if input != nil && input.Typ == Quit {
			return
		}
		handleInput(ui, level, input)
	}
}
