package ui2d

import (
	"bufio"
	"image/png"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/maxproske/games-with-go/29_fonts/game"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type ui struct {
	winWidth          int
	winHeight         int
	renderer          *sdl.Renderer
	window            *sdl.Window
	textureAtlas      *sdl.Texture             // Spritesheets called texture atlases
	textureIndex      map[game.Tile][]sdl.Rect // Go map from a tile to rect
	prevKeyboardState []uint8
	keyboardState     []uint8
	centerX           int // Keep camera centered around player
	centerY           int
	r                 *rand.Rand       // RNG should not be shared aross UIs
	levelChan         chan *game.Level // What level it's getting data from
	inputChan         chan *game.Input
	fontSmall         *ttf.Font
	fontMedium        *ttf.Font
	fontLarge         *ttf.Font
	str2TexSmall      map[string]*sdl.Texture // String/texture cache
	str2TexMedium     map[string]*sdl.Texture // TODO(max): map string for size to eliminate redundancy
	str2TexLarge      map[string]*sdl.Texture
}

// NewUI creates our UI struct
func NewUI(inputChan chan *game.Input, levelChan chan *game.Level) *ui {
	ui := &ui{}
	ui.str2TexSmall = make(map[string]*sdl.Texture)
	ui.str2TexMedium = make(map[string]*sdl.Texture)
	ui.str2TexLarge = make(map[string]*sdl.Texture)
	ui.inputChan = inputChan
	ui.levelChan = levelChan
	ui.r = rand.New(rand.NewSource(1)) // Each UI has its own random starting with the same seed
	ui.winHeight = 720
	ui.winWidth = 1280

	// Create a window.
	window, err := sdl.CreateWindow("RPG", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, int32(ui.winWidth), int32(ui.winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	ui.window = window

	// Create renderer.
	ui.renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}

	// Set hints.
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	// Create texture.
	ui.textureAtlas = ui.imgFileToTexture("../22_texture_index/ui2d/assets/tiles.png")
	ui.loadTextureIndex()

	// Update keyboard state
	ui.keyboardState = sdl.GetKeyboardState() // Updates by sdl
	ui.prevKeyboardState = make([]uint8, len(ui.keyboardState))
	for i, v := range ui.keyboardState {
		ui.prevKeyboardState[i] = v
	}

	// Uninitialize center pos
	ui.centerX = -1
	ui.centerY = -1

	// Get the font sizes
	ui.fontSmall, err = ttf.OpenFont("ui2d/assets/gothic.ttf", 24)
	if err != nil {
		panic(err)
	}
	ui.fontMedium, err = ttf.OpenFont("ui2d/assets/gothic.ttf", 32)
	if err != nil {
		panic(err)
	}
	ui.fontLarge, err = ttf.OpenFont("ui2d/assets/gothic.ttf", 64)
	if err != nil {
		panic(err)
	}

	return ui
}

// FontSize ...
type FontSize int

const (
	// FontSmall ...
	FontSmall FontSize = iota
	// FontMedium ...
	FontMedium
	// FontLarge ...
	FontLarge
)

func (ui *ui) stringToTexture(s string, color sdl.Color, size FontSize) *sdl.Texture {
	// Check if string exists in cache
	var font *ttf.Font
	switch size {
	case FontSmall:
		font = ui.fontSmall
		tex, exists := ui.str2TexSmall[s]
		if exists {
			return tex
		}
	case FontMedium:
		font = ui.fontMedium
		tex, exists := ui.str2TexMedium[s]
		if exists {
			return tex
		}
	case FontLarge:
		font = ui.fontLarge
		tex, exists := ui.str2TexLarge[s]
		if exists {
			return tex
		}
	}

	// Create font surface
	fontSurface, err := font.RenderUTF8Blended(s, color)
	if err != nil {
		panic(err)
	}

	// Create font texture
	tex, err := ui.renderer.CreateTextureFromSurface(fontSurface)
	if err != nil {
		panic(err)
	}

	// Put texture in cache
	switch size {
	case FontSmall:
		ui.str2TexSmall[s] = tex
	case FontMedium:
		ui.str2TexMedium[s] = tex
	case FontLarge:
		ui.str2TexLarge[s] = tex
	}

	//tex.Destroy() // Always destroy texture or it will stay in memory indefinitely
	return tex
}

func (ui *ui) loadTextureIndex() {
	ui.textureIndex = make(map[game.Tile][]sdl.Rect)
	infile, err := os.Open("ui2d/assets/atlas-index.txt")
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	// Read from scanner
	scanner := bufio.NewScanner(infile) // *File satisfies io.Reader interface
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line) // Remove extra spaces
		tileRune := game.Tile(line[0]) // Get first rune from the string
		xy := line[1:]                 // Get ButFirst
		splitXYC := strings.Split(xy, ",")
		x, err := strconv.ParseInt(strings.TrimSpace(splitXYC[0]), 10, 64) // base10, bit size 64
		if err != nil {
			panic(err)
		}
		y, err := strconv.ParseInt(strings.TrimSpace(splitXYC[1]), 10, 64)
		if err != nil {
			panic(err)
		}
		// Tile variation
		variationCount, err := strconv.ParseInt(strings.TrimSpace(splitXYC[2]), 10, 64)
		if err != nil {
			panic(err)
		}
		var rects []sdl.Rect
		for i := int64(0); i < variationCount; i++ {
			rects = append(rects, sdl.Rect{int32(x * 32), int32(y * 32), 32, 32})
			// Wrap around if varied images continue on a new line
			x++
			if x > 62 {
				x = 0
				y++
			}
		}
		ui.textureIndex[tileRune] = rects
	}
}

func (ui *ui) imgFileToTexture(filename string) *sdl.Texture {
	// Open
	infile, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	// Decode
	img, err := png.Decode(infile)
	if err != nil {
		panic(err)
	}

	// Extract w/h
	w := img.Bounds().Max.X
	h := img.Bounds().Max.Y

	pixels := make([]byte, w*h*4)
	bIndex := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixels[bIndex] = byte(r / 256)
			bIndex++
			pixels[bIndex] = byte(g / 256)
			bIndex++
			pixels[bIndex] = byte(b / 256)
			bIndex++
			pixels[bIndex] = byte(a / 256)
			bIndex++
		}
	}

	// Make an SDL2 texture out of pixels
	// AGBR is backwards from way we will be filling in out bytes
	tex, err := ui.renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	tex.Update(nil, pixels, w*4) // Can't provide a rectangle, pitch = 4 bytes per pixel

	// Set blend mode to alpha blending
	err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		panic(err)
	}
	return tex
}

// Init callback runs before anything else
func init() {
	// Initialize SDL2.
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		panic(err)
	}
	// Initialize SDL2_ttf.
	err = ttf.Init()
	if err != nil {
		panic(err)
	}
}

// Draw generates a random (but reproducable) tile variety
func (ui *ui) Draw(level *game.Level) {
	// Recent camera when player is 5 units away from center
	if ui.centerX == -1 && ui.centerY == -1 {
		ui.centerX = level.Player.X
		ui.centerY = level.Player.Y
	}
	limit := 5
	if level.Player.X > ui.centerX+limit {
		ui.centerX++
	} else if level.Player.X < ui.centerX-limit {
		ui.centerX--
	} else if level.Player.Y > ui.centerY+limit {
		ui.centerY++
	} else if level.Player.Y < ui.centerY-limit {
		ui.centerY--
	}

	// Center based on width and height of screen
	offsetX := int32((ui.winWidth / 2) - ui.centerX*32) // Cast int to int32 since we will always use it as int32
	offsetY := int32((ui.winHeight / 2) - ui.centerY*32)

	// Clear before drawing tiles
	ui.renderer.Clear()

	// Set reproducable seed
	ui.r.Seed(1)
	for y, row := range level.Map {
		for x, tile := range row {
			if tile != game.Blank {
				srcRects := ui.textureIndex[tile]
				srcRect := srcRects[ui.r.Intn(len(srcRects))] // Random number between 1 and length of variations
				dstRect := sdl.Rect{int32(x*32) + offsetX, int32(y*32) + offsetY, 32, 32}

				// If debug map contains position we are about to draw, set color
				pos := game.Pos{x, y}
				if level.Debug[pos] {
					ui.textureAtlas.SetColorMod(128, 0, 0) // Multiply color we set on top of it
				} else {
					ui.textureAtlas.SetColorMod(255, 255, 255) // No longer any changes to the texture
				}

				ui.renderer.Copy(ui.textureAtlas, &srcRect, &dstRect)
			}
		}
	}
	// Draw monsters
	for pos, monster := range level.Monsters {
		monsterSrcRect := ui.textureIndex[game.Tile(monster.Rune)][0]
		ui.renderer.Copy(ui.textureAtlas, &monsterSrcRect, &sdl.Rect{int32(pos.X)*32 + offsetX, int32(pos.Y)*32 + offsetY, 32, 32})
	}

	// Draw player
	playerSrcRect := ui.textureIndex['@'][0]
	ui.renderer.Copy(ui.textureAtlas, &playerSrcRect, &sdl.Rect{int32(level.Player.X)*32 + offsetX, int32(level.Player.Y)*32 + offsetY, 32, 32})

	// Draw events
	for i, event := range level.Events {
		tex := ui.stringToTexture(event, sdl.Color{255, 0, 0, 0}, FontLarge)
		_, _, w, h, _ := tex.Query()
		ui.renderer.Copy(tex, nil, &sdl.Rect{0, int32(i * 64), w, h})
	}

	ui.renderer.Present()
}

// GetInput polls for events, and quits when event is nil
func (ui *ui) Run() {
	// Keep waiting for user input
	for {
		// Poll for events
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				// Instead of returning, put inputn into channel
				ui.inputChan <- &game.Input{Typ: game.QuitGame}
			case *sdl.WindowEvent:
				if e.Event == sdl.WINDOWEVENT_CLOSE {
					ui.inputChan <- &game.Input{Typ: game.CloseWindow, LevelChannel: ui.levelChan} // Let game close that level channel
				}
			}
		}

		// Check if we have a new game state to draw
		select {
		// Don't wait on the channel
		case newLevel, ok := <-ui.levelChan:
			if ok {
				ui.Draw(newLevel)
			}
		default:
		}

		// Handle keypresses if window is in focus
		// Or else will crash because we are trying to send x3 input to all 3 windows at the same time
		if sdl.GetKeyboardFocus() == ui.window && sdl.GetMouseFocus() == ui.window {
			var input game.Input
			if ui.keyboardState[sdl.SCANCODE_UP] == 1 && ui.prevKeyboardState[sdl.SCANCODE_UP] == 0 {
				input.Typ = game.Up
			}
			if ui.keyboardState[sdl.SCANCODE_DOWN] == 1 && ui.prevKeyboardState[sdl.SCANCODE_DOWN] == 0 {
				input.Typ = game.Down
			}
			if ui.keyboardState[sdl.SCANCODE_LEFT] == 1 && ui.prevKeyboardState[sdl.SCANCODE_LEFT] == 0 {
				input.Typ = game.Left
			}
			if ui.keyboardState[sdl.SCANCODE_RIGHT] == 1 && ui.prevKeyboardState[sdl.SCANCODE_RIGHT] == 0 {
				input.Typ = game.Right
			}
			if ui.keyboardState[sdl.SCANCODE_S] == 1 && ui.prevKeyboardState[sdl.SCANCODE_S] == 0 {
				// Do a search
				input.Typ = game.Search
			}

			// Update previous keyboard state
			for i, v := range ui.keyboardState {
				ui.prevKeyboardState[i] = v
			}

			if input.Typ != game.None {
				ui.inputChan <- &input
			}
		}
		sdl.Delay(10) // Don't eat cpu waiting for inputs
	}
}
