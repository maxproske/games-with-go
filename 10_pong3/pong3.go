package main

import (
	"fmt"
	"math"
	"time"

	noise "github.com/maxproske/games-with-go/10_package_noise"
	"github.com/veandco/go-sdl2/sdl" // go get github.com/veandco/go-sdl2/sdl
)

// Initialize constants.
const winWidth, winHeight int = 800, 600

// Enumerator with crazy go syntax. Cheaper than strings.
type gameState int

const (
	start gameState = iota // 0
	play                   // 1
)

var state = start

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
	radius float32
	xv     float32
	yv     float32
	color  color
}

type paddle struct {
	pos
	w     int
	h     int
	speed float32
	score int
	color color
}

var nums = [][]byte{
	{
		1, 1, 1,
		1, 0, 1,
		1, 0, 1,
		1, 0, 1,
		1, 1, 1,
	},
	{
		1, 1, 0,
		0, 1, 0,
		0, 1, 0,
		0, 1, 0,
		1, 1, 1,
	},
	{
		1, 1, 1,
		0, 0, 1,
		1, 1, 1,
		1, 0, 0,
		1, 1, 1,
	},
	{
		1, 1, 1,
		0, 0, 1,
		1, 1, 1,
		0, 0, 1,
		1, 1, 1}}

// We don't need to create a copy for our function
// Default to a pointer, it's less confusing!
func (paddle *paddle) draw(pixels []byte) {
	// Position is at center, so draw from top left corner
	startX := int(paddle.x - float32(paddle.w/2))
	startY := int(paddle.y - float32(paddle.h/2))

	// Draw a rectangle
	// Start with y, so we are going through memory cache in order
	for y := 0; y < int(paddle.h); y++ {
		for x := 0; x < int(paddle.w); x++ {
			setPixel(startX+x, startY+y, paddle.color, pixels)
		}
	}

	// Draw score in a good place using lerp
	numX := flerp(paddle.x, getCenter().x, 0.2) // A little towards the center (20%)
	drawNumber(pos{numX, 35}, paddle.color, 10, paddle.score, pixels)
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
				setPixel(int(ball.x+x), int(ball.y+y), ball.color, pixels)
			}
		}
	}
}

// Utility function for resetting ball position
func getCenter() pos {
	return pos{float32(winWidth) / 2, float32(winHeight) / 2}
}

func (ball *ball) update(leftPaddle, rightPaddle *paddle, elapsedTime float32) {
	// Update position, scale by elapsedTime
	ball.x += ball.xv * elapsedTime
	ball.y += ball.yv * elapsedTime

	// Handle collisions with boundary
	if ball.y-ball.radius <= 0 || ball.y+ball.radius > float32(winHeight) {
		ball.yv = -ball.yv
	}

	// Set the score for the appropriate player, then place ball back in middle of screen
	if ball.x < 0 {
		rightPaddle.score++
		ball.pos = getCenter()
		// Wait for spacebar
		state = start
	} else if int(ball.x) > winWidth {
		leftPaddle.score++
		ball.pos = getCenter()
		state = start
	}

	// Collisiosn with ball
	if ball.x-ball.radius < leftPaddle.x+float32(leftPaddle.w/2) {
		// Ball is at same x position as right edge of left paddle
		if ball.y > leftPaddle.y-float32(leftPaddle.h/2) && ball.y < leftPaddle.y+float32(leftPaddle.h/2) {
			// ball is lower than top of the paddle, and ball is higher than bottom of paddle
			ball.xv = -ball.xv
			// Minimum translation vector, resolve wiggle when blal is stuck in paddle
			ball.x = leftPaddle.x + float32(leftPaddle.w/2.0) + ball.radius
		}

	}

	if ball.x+ball.radius > rightPaddle.x-float32(rightPaddle.w/2) {
		if ball.y > rightPaddle.y-float32(rightPaddle.h/2) && ball.y < rightPaddle.y+float32(rightPaddle.h/2) {
			ball.xv = -ball.xv
			// MTV again
			ball.x = rightPaddle.x - float32(rightPaddle.w/2.0) - ball.radius
		}
	}
}

func (paddle *paddle) update(keyState []uint8, controllerAxis int16, elapsedTime float32) {
	// Respond to keyboard input
	if keyState[sdl.SCANCODE_UP] != 0 {
		paddle.y -= paddle.speed * elapsedTime
	}
	if keyState[sdl.SCANCODE_DOWN] != 0 {
		paddle.y += paddle.speed * elapsedTime
	}
	// Respond to controller input
	if (math.Abs(float64(controllerAxis))) > 1500 { // don't count on axis being 0 for an analog controller, set a low minimum of 1500
		pct := float32(controllerAxis) / 32767.0 // see how much controller has moved. divide by maximum speed (all the way down/up)
		paddle.y += paddle.speed * pct * elapsedTime
	}
}

func clear(pixels []byte) {
	// Goes through memory in order. So it's still fast without having to clear only unchanged pixels
	for i := range pixels {
		pixels[i] = 0
	}
}

func (paddle *paddle) aiUpdate(ball *ball, elapsedTime float32) {
	paddle.y = ball.y
}

// Linear interpretation
// [1				10] --> for 50%, lerp will give you 5
func flerp(a float32, b float32, pct float32) float32 {
	return a + pct*(b-a)
}

func drawNumber(pos pos, color color, size int, num int, pixels []byte) {
	startX := int(pos.x) - (size*3)/2 // width is 3 elements wide
	startY := int(pos.y) - (size*5)/2 // height is 5 elements high

	// for index and value, iterate over inner array
	for i, v := range nums[num] {
		if v == 1 {
			// draw our squyare
			for y := startY; y < startY+size; y++ {
				for x := startX; x < startX+size; x++ {
					setPixel(x, y, color, pixels)
				}
			}
		}
		startX += size
		// Go down after first line
		if (i+1)%3 == 0 {
			startY += size
			startX -= size * 3
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
		sdl.TEXTUREACCESS_STREAMING, // Texture access
		int32(winWidth),
		int32(winHeight))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tex.Destroy()

	// Analog input, does not allow for controllers plugged in during gameplay however
	var controllerHandlers []*sdl.GameController
	for i := 0; i < sdl.NumJoysticks(); i++ {
		controllerHandlers = append(controllerHandlers, sdl.GameControllerOpen(i)) // add a struct to get funcs and properties
		defer controllerHandlers[i].Close()                                        // close later
	}

	// Create a bytearray (slice) of pixels
	pixels := make([]byte, winWidth*winHeight*4) // 4 bytes for each channel (ARGB)

	// Make a paddle and ball
	player1 := paddle{pos{50, 100}, 20, 100, 300, 0, color{255, 255, 255}}
	player2 := paddle{pos{float32(winWidth) - 50, 100}, 20, 100, 300, 0, color{255, 255, 255}}
	ball := ball{pos{300, 300}, 20, 400, 400, color{255, 255, 255}}

	// Get keyboard state
	keyState := sdl.GetKeyboardState() // pointer with representation of every key. Updated by PollEvent

	// Call custom package
	noise, min, max := noise.MakeNoise(noise.FBM, .01, 0.5, 2, 3, winWidth, winHeight)
	// Draw colours
	gradient := getGradient(color{40, 40, 40}, color{10, 10, 10})
	noisePixels := rescaleAndDraw(noise, min, max, gradient, winWidth, winHeight)

	var frameStart time.Time
	var elapsedTime float32
	var controllerAxis int16

	// Poll for window events
	for {
		frameStart = time.Now()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			// Type switch (Special switch. Type of sdl.PollEvent() isn't totally decided)
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}

		// Loop over every controller in the controllerHandlers slice
		for _, controller := range controllerHandlers {
			if controller != nil {
				controllerAxis = controller.Axis(sdl.CONTROLLER_AXIS_LEFTY)
			}
		}

		// If we are playing
		if state == play {
			player1.update(keyState, controllerAxis, elapsedTime)
			player2.aiUpdate(&ball, elapsedTime)
			ball.update(&player1, &player2, elapsedTime)
		} else if state == start {
			if keyState[sdl.SCANCODE_SPACE] != 0 {
				// Space has been pressed
				if player1.score == 3 || player2.score == 3 {
					player1.score = 0
					player2.score = 0
				}
				state = play
			}
		}

		// Update
		drawNumber(getCenter(), color{255, 255, 255}, 20, 2, pixels)

		// Copy noise pixels into screen output
		for i := range noisePixels {
			pixels[i] = noisePixels[i]
		}

		// Draw
		player1.draw(pixels)
		player2.draw(pixels)
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

		// Frame rate independence (in milliseconds)
		elapsedTime = float32(time.Since(frameStart).Seconds()) // shortcut for newTime = time.Now - frameStart;

		// Limit framerate to 200fps
		if elapsedTime < .005 {
			sdl.Delay(5 - uint32(elapsedTime*1000.0))
			elapsedTime = float32(time.Since(frameStart).Seconds()) // update again after the delay so update has the right value
		}
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

func clamp(min, max, v int) int {
	if v < min {
		v = min
	} else if v > max {
		v = max
	}
	return v
}

// Rescale values we get from noise to be between 0-255
func rescaleAndDraw(noise []float32, min, max float32, gradient []color, w, h int) []byte {
	result := make([]byte, w*h*4)

	// Rescale noise
	scale := 255.0 / (max - min)
	offset := min * scale

	// Turn it into bytes
	for i := range noise {
		noise[i] = noise[i]*scale - offset

		// Take in a gradient
		c := gradient[clamp(0, 255, int(noise[i]))]

		p := i * 4 // pixel index
		//b := byte(noise[i]) // Make an integer
		result[p] = c.r
		result[p+1] = c.g
		result[p+2] = c.b
	}
	return result
}

// Fractional Brownian motion
func fbm2(x, y, frequency, lacunarity, gain float32, octaves int) float32 {
	var sum float32
	amplitude := float32(1.0)
	for i := 0; i < octaves; i++ {
		sum += noise.Snoise2(x*frequency, y*frequency) * amplitude // x * our frequency
		frequency = frequency * lacunarity                         // frequency will change every iteration
		amplitude = amplitude * gain                               // amplitude will change every iteration
	}
	return sum
}

func colorLerp(c1, c2 color, pct float32) color {
	return color{lerp(c1.r, c2.r, pct), lerp(c1.g, c2.g, pct), lerp(c1.b, c2.b, pct)}
}

// Linear interpretation (lerp) between two bytes
func lerp(b1, b2 byte, pct float32) byte {
	return byte(float32(b1) + pct*(float32(b2)-float32(b1)))
}

func getGradient(c1, c2 color) []color {
	result := make([]color, 256)
	for i := range result {
		// Get the current percentage
		pct := float32(i) / float32(255)
		result[i] = colorLerp(c1, c2, pct)
	}
	return result
}

// Go from color 1 to 2, then suddenly switch to 3
func getDualGradient(c1, c2, c3, c4 color) []color {
	result := make([]color, 256)
	for i := range result {
		// Get the current percentage
		pct := float32(i) / float32(255)
		// Do the same as getGradient()
		if pct < 0.5 {
			result[i] = colorLerp(c1, c2, pct*float32(2))
		} else {
			result[i] = colorLerp(c3, c4, pct*float32(1.5)-float32(0.5)) // keep between 0-1 range
		}

	}
	return result
}
