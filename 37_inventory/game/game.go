package game

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Game contains channels for game and UI threads
type Game struct {
	LevelChans   []chan *Level // Send level state to multiple UIs
	InputChan    chan *Input   // Receieve input from multiple UIs
	Levels       map[string]*Level
	CurrentLevel *Level
}

// NewGame needs to know how many channels to take in
func NewGame(numWindows int) *Game {
	levelChans := make([]chan *Level, numWindows) // 1 level channel for each window
	for i := range levelChans {
		levelChans[i] = make(chan *Level)
	}
	inputChan := make(chan *Input)
	levels := loadLevels()

	game := &Game{levelChans, inputChan, levels, nil}
	game.loadWorldFile()            // Load world file
	game.CurrentLevel.lineOfSight() // Draw visible tiles without moving

	return game
}

// InputType is a tagged union/discriminating union/sum type
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
	// Right input type
	Right
	// TakeAll input type
	TakeAll
	// QuitGame triggers on last window closed
	QuitGame
	// CloseWindow triggers on 1 of many windows closed
	CloseWindow
	// TakeItem input type
	TakeItem
	// DropItem input type
	DropItem
	// Search input type
	Search
)

// Input ...
type Input struct {
	Typ          InputType
	Item         *Item // Item will be the data, not the position of a click
	LevelChannel chan *Level
}

// Tile enum is just an alias for a rune (a character in Go)
type Tile struct {
	Rune        rune
	OverlayRune rune
	Visible     bool
	Seen        bool // Have you seen the tile before
}

const (
	// StoneWall represented by a character
	StoneWall rune = '#'
	// DirtFloor represented by a character
	DirtFloor = '.'
	// ClosedDoor represented by a character
	ClosedDoor = '|'
	// OpenDoor represented by a character
	OpenDoor = '/'
	// UpStair represented bu a character
	UpStair = 'u'
	// DownStair represented by a character
	DownStair = 'd'
	// Blank represented by zero
	Blank = 0
	// Pending represented by -1
	Pending = -1
)

// Pos ...
type Pos struct {
	X, Y int
}

// LevelPos ...
type LevelPos struct {
	*Level
	Pos
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
	SightRange   int
	Items        []*Item
}

// GameEvent provides visibility of events to UI2D
type GameEvent int

const (
	// Move ...
	Move GameEvent = iota
	DoorOpen
	Attack
	Hit
	Portal
	PickUp
	Drop
)

// Level holds the 2D array that represents the map
type Level struct {
	Map       [][]Tile
	Player    *Player
	Monsters  map[Pos]*Monster // Pos as key, get back monster
	Items     map[Pos][]*Item  // Allow multiple items per tile
	Portals   map[Pos]*LevelPos
	Events    []string
	EventPos  int
	Debug     map[Pos]bool // Map x/y positions to true/false
	LastEvent GameEvent    // Events not visible to the player
}

// DropItem ...
func (level *Level) DropItem(itemToDrop *Item, character *Character) {
	pos := character.Pos
	items := character.Items
	for i, item := range items {
		if item == itemToDrop {
			// Reverse order of MoveItem function
			character.Items = append(character.Items[:i], character.Items[i+1:]...) // Delete item from world
			level.Items[pos] = append(level.Items[pos], item)                       // Add to inventory
			level.AddEvent(character.Name + " dropped 1x " + item.Name)
			return
		}
	}
	panic("Tried to drop an item we don't have.")
}

// MoveItem moves an item to a character's inventory
func (level *Level) MoveItem(itemToMove *Item, character *Character) {
	pos := character.Pos
	items := level.Items[pos]
	for i, item := range items {
		// Check if they are same address in memory, so you can have multiple swords per tile
		if item == itemToMove {
			items = append(items[:i], items[i+1:]...)       // Delete item from world
			level.Items[pos] = items                        // Update the map
			character.Items = append(character.Items, item) // Add to inventory
			level.AddEvent(character.Name + " picked up 1x " + item.Name)
			return // Return early
		}
	}
	panic("Tried to move an item we're not on top of")
}

// Attack engages two attackables
func (level *Level) Attack(c1, c2 *Character) {
	// a1 attacking a2 first
	c1.ActionPoints--
	c1AttackPower := c1.Strength
	c2.Hitpoints -= c1AttackPower

	if c2.Hitpoints > 0 {
		level.AddEvent(c1.Name + " Attacked " + c2.Name + " for " + strconv.Itoa(c1AttackPower))
	} else {
		level.AddEvent(c1.Name + " Killed " + c2.Name)
	}
}

// AddEvent handles events list
func (level *Level) AddEvent(event string) {
	level.Events[level.EventPos] = event
	level.EventPos++
	if level.EventPos == len(level.Events) {
		level.EventPos = 0 // Loop around to overwrite stale events
	}

}

func (level *Level) lineOfSight() {
	pos := level.Player.Pos
	dist := level.Player.SightRange // Radius
	// Iterate over square the size of player sight range
	for y := pos.Y - dist; y <= pos.Y+dist; y++ {
		for x := pos.X - dist; x <= pos.X+dist; x++ {
			xDelta := pos.X - x
			yDelta := pos.Y - y
			d := math.Sqrt(float64(xDelta*xDelta + yDelta*yDelta))
			if d <= float64(dist) {
				level.bresenham(pos, Pos{x, y})
			}
		}
	}
}

// Draw a circle around the player and draw a line to each endpoint
func (level *Level) bresenham(start Pos, end Pos) {
	steep := math.Abs(float64(end.Y-start.Y)) > math.Abs(float64(end.X-start.X)) // Is the line steep or not?
	// Swap the x and y for start and end
	if steep {
		start.X, start.Y = start.Y, start.X
		end.X, end.Y = end.Y, end.X
	}

	deltaY := int(math.Abs(float64(end.Y - start.Y)))

	err := 0
	y := start.Y
	ystep := 1 // How far we are stepping when err is above threshold
	if start.Y >= end.Y {
		ystep = -1 // Reverse it when we step
	}

	// Are we on the left or right side of graph
	if start.X > end.X {
		deltaX := start.X - end.X // We know start.X will be larger than end.X
		// Count down so lines extend FROM the player, not TO
		for x := start.X; x > end.X; x-- {
			var pos Pos
			if steep {
				pos = Pos{y, x} // If we are steep, x and y will be swapped
			} else {
				pos = Pos{x, y}
			}
			level.Map[pos.Y][pos.X].Visible = true
			level.Map[pos.Y][pos.X].Seen = true // Stay true
			if !canSeeThrough(level, pos) {
				return
			}
			err += deltaY
			if 2*err >= deltaX {
				y += ystep // Go up or down depending on the direction of our line
				err -= deltaX
			}
		}
	} else {
		deltaX := end.X - start.X // We know start.X will be larger than end.X
		for x := start.X; x < end.X; x++ {
			var pos Pos
			if steep {
				pos = Pos{y, x} // If we are steep, x and y will be swapped
			} else {
				pos = Pos{x, y}
			}
			level.Map[pos.Y][pos.X].Visible = true
			level.Map[pos.Y][pos.X].Seen = true // Stay true
			if !canSeeThrough(level, pos) {
				return
			}
			err += deltaY
			if 2*err >= deltaX {
				y += ystep // Go up or down depending on the direction of our line
				err -= deltaX
			}
		}
	}
}

func (game *Game) loadWorldFile() {
	file, err := os.Open("game/maps/world.txt")
	if err != nil {
		panic(err)
	}
	csvReader := csv.NewReader(file)
	csvReader.FieldsPerRecord = -1 // Don't enforce each row to have same num columns
	csvReader.TrimLeadingSpace = true
	rows, err := csvReader.ReadAll() // Our files are not going to be very big, so we can use ReadAll instead of Read
	if err != nil {
		panic(err)
	}
	for rowIndex, row := range rows {
		// Set current level
		if rowIndex == 0 {
			game.CurrentLevel = game.Levels[row[0]] // Get the first item from the first row, and set the level
			if game.CurrentLevel == nil {
				fmt.Println("Couldn't find current level name in world file.")
				panic(nil)
			}
			continue
		}
		levelWithPortal := game.Levels[row[0]] // Level 1 name
		if levelWithPortal == nil {
			fmt.Println("Couldn't find level from name in world file.")
			panic(nil)
		}

		x, err := strconv.ParseInt(row[1], 10, 64)
		if err != nil {
			panic(err)
		}
		y, err := strconv.ParseInt(row[2], 10, 64)
		if err != nil {
			panic(err)
		}
		pos := Pos{int(x), int(y)} // Level 1 pos

		levelToTeleportTo := game.Levels[row[3]] // Level 2 name
		if levelToTeleportTo == nil {
			fmt.Println("Couldn't find level to name in world file.")
			panic(nil)
		}
		x, err = strconv.ParseInt(row[4], 10, 64)
		if err != nil {
			panic(err)
		}
		y, err = strconv.ParseInt(row[5], 10, 64)
		if err != nil {
			panic(err)
		}
		posToTeleportTo := Pos{int(x), int(y)} // Level 2 pos

		levelWithPortal.Portals[pos] = &LevelPos{levelToTeleportTo, posToTeleportTo} // Our position to teleport to
	}
}

// loadLevels opens and prints a map
func loadLevels() map[string]*Level {
	// Make player
	player := &Player{} // Player used to not be a pointer
	player.Strength = 20
	player.Hitpoints = 20
	player.Name = "GoMan"
	player.Rune = '@'
	player.Speed = 1.0
	player.ActionPoints = 0
	player.SightRange = 7
	levels := make(map[string]*Level)
	// Load level
	filenames, err := filepath.Glob("game/maps/*.map")
	if err != nil {
		panic(err)
	}
	for _, filename := range filenames {
		extIndex := strings.LastIndex(filename, ".map")
		lastSlashIndex := strings.LastIndex(filename, "\\")
		levelName := filename[lastSlashIndex+1 : extIndex]
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
		level.Debug = make(map[Pos]bool)
		level.Events = make([]string, 10)
		level.Player = player
		level.Map = make([][]Tile, len(levelLines))
		level.Monsters = make(map[Pos]*Monster)
		level.Items = make(map[Pos][]*Item)
		level.Portals = make(map[Pos]*LevelPos)

		for i := range level.Map {
			level.Map[i] = make([]Tile, longestRow) // Make each row the same length of the longest row (non-jagged slice)
		}

		for y := 0; y < len(level.Map); y++ {
			line := levelLines[y]
			for x, c := range line {
				pos := Pos{x, y}
				var t Tile
				t.OverlayRune = Blank // Most things will not have an overlay rune
				switch c {
				case ' ', '\t', '\n', '\r':
					t.Rune = Blank
				case '#':
					t.Rune = StoneWall
				case '|':
					t.OverlayRune = ClosedDoor
					t.Rune = Pending
				case '/':
					t.Rune = OpenDoor
				case 'u':
					t.OverlayRune = UpStair
					t.Rune = Pending
				case 'd':
					t.OverlayRune = DownStair
					t.Rune = Pending
				case 's':
					level.Items[pos] = append(level.Items[pos], NewSword(pos)) // Append item to slice of items, follow monster template
					level.Items[pos] = append(level.Items[pos], NewHelmet(pos))
					t.Rune = Pending
				case 'h':
					level.Items[pos] = append(level.Items[pos], NewHelmet(pos))
					t.Rune = Pending
				case '.':
					t.Rune = DirtFloor
				case '@':
					level.Player.X = x // Set player X,Y
					level.Player.Y = y
					t.Rune = Pending // Be a placeholder
				case 'R':
					// Rat
					level.Monsters[pos] = NewRat(pos)
					t.Rune = Pending
				case 'S':
					// Spider
					level.Monsters[pos] = NewSpider(pos)
					t.Rune = Pending
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
				if tile.Rune == Pending {
					level.Map[y][x].Rune = level.bfsFloor(Pos{x, y}) // Use bfs to find the nearest floor tile, and send it to it
				}
			}
		}
		// Append the current level to our level slice
		levels[levelName] = level
	}
	return levels
}

// Check if x,y is inbounds
func inRange(level *Level, pos Pos) bool {
	return pos.X < len(level.Map[0]) && pos.Y < len(level.Map) && pos.X >= 0 && pos.Y >= 0
}

func canWalk(level *Level, pos Pos) bool {
	if inRange(level, pos) {
		// Check tile for solid object
		t := level.Map[pos.Y][pos.X]
		switch t.Rune {
		case StoneWall, Blank:
			return false
		}
		switch t.OverlayRune {
		case ClosedDoor:
			return false
		}
		// Check to see if a monster is in the way
		_, exists := level.Monsters[pos]
		if exists {
			return false
		}
		return true
	}
	return false
}

// Is there line of sight/a window?
func canSeeThrough(level *Level, pos Pos) bool {
	if inRange(level, pos) {
		// Check tile for solid object
		t := level.Map[pos.Y][pos.X]
		switch t.Rune {
		case StoneWall, Blank:
			return false
		}
		switch t.OverlayRune {
		case ClosedDoor:
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
	if t.OverlayRune == ClosedDoor {
		level.Map[pos.Y][pos.X].OverlayRune = OpenDoor // Player has opened a door
		level.LastEvent = OpenDoor
		level.lineOfSight() // Check line of sight without moving a tile
	}
}

// Move moves the player unless a monster exists in that location
func (game *Game) Move(to Pos) {
	level := game.CurrentLevel
	player := level.Player

	// Check position we are moving to for portals
	levelAndPos := level.Portals[to]
	if levelAndPos != nil {
		fmt.Println("in portal!")
		game.CurrentLevel = levelAndPos.Level
		game.CurrentLevel.Player.Pos = levelAndPos.Pos
		game.CurrentLevel.lineOfSight()
	} else {
		player.Pos = to // Player has moved
		level.LastEvent = Move
		// Draw line of sight
		for y, row := range level.Map {
			for x := range row {
				level.Map[y][x].Visible = false
			}
		}
		level.lineOfSight()
	}
}

// Handle decisions about player movement
func (game *Game) resolveMovement(pos Pos) {
	level := game.CurrentLevel
	monster, exists := level.Monsters[pos]
	if exists {
		level.Attack(&level.Player.Character, &monster.Character) // Attacked
		level.LastEvent = Attack
		if monster.Hitpoints <= 0 {
			monster.Kill(level)
		}
		if level.Player.Hitpoints <= 0 {
			panic("ded")
		}
	} else if canWalk(level, pos) {
		game.Move(pos)
	} else {
		checkDoor(level, pos)
	}
}

// Returning a *Level is slow
func (game *Game) handleInput(input *Input) {
	level := game.CurrentLevel
	p := level.Player
	// Check if the place the player is going to is available
	switch input.Typ {
	case Up:
		newPos := Pos{p.X, p.Y - 1}
		game.resolveMovement(newPos)
	case Down:
		newPos := Pos{p.X, p.Y + 1}
		game.resolveMovement(newPos)
	case Left:
		newPos := Pos{p.X - 1, p.Y}
		game.resolveMovement(newPos)
	case Right:
		newPos := Pos{p.X + 1, p.Y}
		game.resolveMovement(newPos)
	case TakeItem:
		level.MoveItem(input.Item, &p.Character)
		level.LastEvent = PickUp
	case DropItem:
		level.DropItem(input.Item, &level.Player.Character)
		level.LastEvent = Drop // Update activity log
	case TakeAll:
		for _, item := range level.Items[p.Pos] {
			level.MoveItem(item, &p.Character)
		}
		level.LastEvent = PickUp
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
// Return just the rune so the overlay is not overwritten
func (level *Level) bfsFloor(start Pos) rune {
	frontier := make([]Pos, 0, 8)      // Start at 8 instead of growing/shrinking frontier
	frontier = append(frontier, start) // Put in front of queue
	visited := make(map[Pos]bool)
	visited[start] = true // We have already visited start, we are on it

	for len(frontier) > 0 {
		current := frontier[0]

		// If there are no surrounding tiles, search for the nearest floor tile
		currentTile := level.Map[current.Y][current.X]
		switch currentTile.Rune {
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
		lchan <- game.CurrentLevel
	}

	// Get an input out of our input channel
	for input := range game.InputChan {
		if input.Typ == QuitGame {
			return
		}

		//p := game.Level.Player.Pos
		//level.bresenham(p, Pos{p.X + 5, p.Y - 5})
		// for _, pos := range line {
		// 	fmt.Println(pos)
		// 	game.Level.Debug[pos] = true
		// }

		game.handleInput(input) // Pass along the input we got

		// Update monsters
		for _, monster := range game.CurrentLevel.Monsters {
			monster.Update(game.CurrentLevel)
		}

		if len(game.LevelChans) == 0 {
			// All the windows have been closed
			return
		}

		// Send game state updates
		for _, lchan := range game.LevelChans {
			lchan <- game.CurrentLevel
		}
	}
}
