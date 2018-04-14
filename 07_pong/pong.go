package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl" // go get github.com/veandco/go-sdl2/sdl
)

// Initialize constants.
const winWidth, winHeight int = 800, 600

// Initialize structs.
type color struct {
	r, g, b byte
}

type pos struct {
	x, y float32
}

// Inheritence in go is achieved through composition
type ball struct {
	pos    // remove name. no we can access ball.x instead of ball.pos.x
	radius int
	xv     float32
	yv     float32
	color  color
}

type paddle struct {
	pos
	w     int
	h     int
	color color
}

// We don't need to create a copy for our function
// Default to a pointer, it's less confusing!
func (paddle *paddle) draw(pixels []byte) {
	// Position is at center, so draw from top left corner
	startX := int(paddle.x) - paddle.w/2
	startY := int(paddle.y) - paddle.h/2

	// Draw a rectangle
	// Start with y, so we are going through memory cache in order
	for y := 0; y < paddle.h; y++ {
		for x := 0; x < paddle.w; x++ {
			setPixel(startX+x, startY+y, paddle.color, pixels)
		}
	}
}

// Don't make pixels a global variable, because if a function doesn't
// modify anything outside itself, that makes it a pure function.
func (ball *ball) draw(pixels []byte) {
	// Draw a circle that's filled in
	// Draw rectangle, and fill in if it's within the radius
	// YAGNI - Ya Aint Gonna Need It
	for y := -ball.radius; y < ball.radius; y++ {
		for x := -ball.radius; x < ball.radius; x++ {
			// Square root without expensive
			if x*x+y*y < ball.radius*ball.radius {
				setPixel(int(ball.x)+x, int(ball.y)+y, ball.color, pixels)
			}
		}
	}
}

func main() {
	// Added after EP06 to address macosx issues
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sdl.Quit()

	// Create a window.
	window, err := sdl.CreateWindow(
		"Pong",                  // title string
		sdl.WINDOWPOS_UNDEFINED, // x int32
		sdl.WINDOWPOS_UNDEFINED, // y int32
		int32(winWidth),         // width int32 (cast to int32)
		int32(winHeight),        // height int32
		sdl.WINDOW_SHOWN)        // flags (window is visible)

	// Check if an error happened.
	if err != nil {
		fmt.Println(err)
		return
	}

	// A defer runs once the function ends, instead of placing it at the end.
	// Destroy and clear up all the resources it was using.
	defer window.Destroy()

	// Create a renderer.
	renderer, err := sdl.CreateRenderer(
		window, // Associated with a *Window
		-1,     // index int
		sdl.RENDERER_ACCELERATED) // flags uint32 (hardware accelerated)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer renderer.Destroy()

	// Create a texture (RGB, where 255,0,0 is red).
	tex, err := renderer.CreateTexture(
		sdl.PIXELFORMAT_ABGR8888,    // Pixel format (one byte for each channel)
		sdl.TEXTUREACCESS_STREAMING, // Texture access (?)
		int32(winWidth),
		int32(winHeight))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tex.Destroy()

	// Create a bytearray (slice) of pixels
	pixels := make([]byte, winWidth*winHeight*4) // 4 bytes for each channel (ARGB)

	// Make a paddle and ball
	player1 := paddle{pos{100, 100}, 20, 100, color{255, 255, 255}}
	ball := ball{pos{300, 300}, 20, 0, 0, color{255, 255, 255}}

	// Poll for window events
	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}

		// Draw paddle and ball
		player1.draw(pixels)
		ball.draw(pixels)

		// Update SDL2 texture
		tex.Update(
			nil,        // rect *Rect
			pixels,     // pixels []byte
			winWidth*4, // pitch int
		)
		// Copy it to the renderer
		renderer.Copy(tex, nil, nil) // Copy
		renderer.Present()           // Present

		sdl.Delay(16)
	}
}

func setPixel(x, y int, c color, pixels []byte) {
	// Convert some x,y position to index of bytearray
	index := (y*winWidth + x) * 4 // jump to row, add column

	// Array index out of bounds check
	if index < len(pixels) && index >= 0 {
		// Set the colour
		pixels[index] = c.r // red component
		pixels[index+1] = c.g
		pixels[index+2] = c.b
	}
}
