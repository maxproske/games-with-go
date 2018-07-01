package ui2d

import (
	"bufio"
	"fmt"
	"image/png"
	"os"
	"strconv"
	"strings"

	"github.com/maxproske/games-with-go/22_texture_index/game"
	"github.com/veandco/go-sdl2/sdl"
)

const winWidth, winHeight int = 1280, 720

var renderer *sdl.Renderer
var textureAtlas *sdl.Texture // Spritesheets called texture atlases

var textureIndex map[game.Tile]sdl.Rect // Go map from a tile to rect

func loadTextureIndex() {
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
		splitXy := strings.Split(xy, ",")
		x, err := strconv.ParseInt(strings.TrimSpace(splitXy[0]), 10, 64) // base10, bit size 64
		if err != nil {
			panic(err)
		}
		y, err := strconv.ParseInt(strings.TrimSpace(splitXy[1]), 10, 64)
		if err != nil {
			panic(err)
		}
		fmt.Println(tileRune, x, y)
	}
}

func imgFileToTexture(filename string) *sdl.Texture {
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
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, int32(w), int32(h))
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
	// Check what best performing, most reliable renderer SDL is using
	// (direct3d on windows, opengl on linux, or software for just sdl)
	sdl.LogSetAllPriority(sdl.LOG_PRIORITY_VERBOSE)

	// Initialize SDL2.
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create a window.
	window, err := sdl.CreateWindow("RPG", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, int32(winWidth), int32(winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create renderer.
	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Set hints.
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	// Create texture.
	textureAtlas = imgFileToTexture("ui2d/assets/tiles.png")

	loadTextureIndex()
}

// UI2d ...
type UI2d struct {
}

// Draw ...
func (ui *UI2d) Draw(level *game.Level) {
	renderer.Copy(textureAtlas, nil, nil)
	renderer.Present()
	sdl.Delay(5000)
}
