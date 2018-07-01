package main

// Genetic algorithms are a type of optimization algorithm, called a metaheuristic
//   A metaheuristic will find an answer you hope will be good, but not the optimal solution
//   (Eg. The solution space is optimal amount of air to add to fuel for combustion.)
//   Hill-climbing algorithm find a random place to start, and changes one thing. If it improves, make that change again.
import (
	"fmt"
	"math/rand"
	"time"

	noise "github.com/maxproske/games-with-go/10_package_noise"
	. "github.com/maxproske/games-with-go/17_mutation/apt_random"
	sdl "github.com/veandco/go-sdl2/sdl"
)

const winWidth, winHeight, winDepth int = 800, 600, 100

type rgba struct {
	r, g, b byte
}

type mouseState struct {
	leftButton  bool
	rightButton bool
	x, y        int
}

type audioState struct {
	explosionBytes []byte
	deviceID       sdl.AudioDeviceID
	audioSpec      *sdl.AudioSpec
}

type picture struct {
	r, g, b Node
}

func (p *picture) String() string {
	return "R" + p.r.String() + "\n" + "G" + p.g.String() + "\n" + "B" + p.b.String()
}

func NewPicture() *picture {
	p := &picture{}
	p.r = GetRandomNode()
	p.g = GetRandomNode()
	p.b = GetRandomNode()
	// How big to make it
	num := rand.Intn(4) // up to 20 nodes
	for i := 0; i < num; i++ {
		p.r.AddRandom(GetRandomNode()) // no leaves yet
	}
	num = rand.Intn(4) // up to 20 nodes
	for i := 0; i < num; i++ {
		p.g.AddRandom(GetRandomNode()) // no leaves yet
	}
	num = rand.Intn(4) // up to 20 nodes
	for i := 0; i < num; i++ {
		p.b.AddRandom(GetRandomNode()) // no leaves yet
	}
	// Add leaf nodes. Will continue until it doesn't have anywhere to put a leaf
	for p.r.AddLeaf(GetRandomLeaf()) {
	}
	for p.g.AddLeaf(GetRandomLeaf()) {
	}
	for p.b.AddLeaf(GetRandomLeaf()) {
	}

	return p
}

func (p *picture) Mutate() {
	// Mutate r, g, or b?
	r := rand.Intn(3)
	var nodeToMutate Node
	switch r {
	case 0:
		nodeToMutate = p.r
	case 1:
		nodeToMutate = p.g
	case 2:
		nodeToMutate = p.b
	}

	// Grab the node
	count := nodeToMutate.NodeCount()
	r = rand.Intn(count)
	nodeToMutate, count = GetNthNode(nodeToMutate, r, 0)
	mutation := Mutate(nodeToMutate) // The now mutated node
	if nodeToMutate == p.r {
		// Look at nodeToMutate, and if it is equal to the top-level node, p.r
		p.r = mutation // Then set it to the mutation
	} else if nodeToMutate == p.g {
		p.g = mutation
	} else if nodeToMutate == p.b {
		p.b = mutation
	}
}

func main() {

	// Check what best performing, most reliable renderer SDL is using
	// (direct3d on windows, opengl on linux, or software for just sdl)
	sdl.LogSetAllPriority(sdl.LOG_PRIORITY_VERBOSE)

	// Initialize SDL2.
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sdl.Quit()

	// Create a window.
	window, err := sdl.CreateWindow("Evolving Pictures", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, int32(winWidth), int32(winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer window.Destroy()

	// Create renderer.
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer renderer.Destroy()

	// Set hints.
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	var elapsedTime float32

	currentMouseState := getMouseState()
	prevMouseState := currentMouseState

	// Seed rand
	rand.Seed(time.Now().UTC().UnixNano())

	// Create a simple abstract syntax tree
	pic := NewPicture()
	tex := aptToTexture(pic, winWidth, winHeight, renderer)

	for {

		frameStart := time.Now()

		currentMouseState = getMouseState()

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				return
			case *sdl.TouchFingerEvent:
				if e.Type == sdl.FINGERDOWN {
					touchX := int(e.X * float32(winWidth))
					touchY := int(e.Y * float32(winHeight))
					currentMouseState.x = touchX
					currentMouseState.y = touchY
					currentMouseState.leftButton = true
				}
			}
		}

		// Mutate and update the texture
		if prevMouseState.leftButton && !currentMouseState.leftButton {
			pic.Mutate()
			tex = aptToTexture(pic, winWidth, winHeight, renderer)
		}

		// Turn a tree into a texture

		// Copy background
		// renderer.Copy(cloudTexture, nil, nil) // nil = draw entire source to entire destination
		renderer.Copy(tex, nil, nil)

		renderer.Present()

		elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		//fmt.Println("ms per frame:", elapsedTime)
		if elapsedTime < 5 {
			sdl.Delay(5 - uint32(elapsedTime))
			elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		}

		prevMouseState = currentMouseState
	}
}

func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
}

// FMB noise
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

// Make an SDL2 texture out of pixels
func pixelsToTexture(renderer *sdl.Renderer, pixels []byte, w, h int) *sdl.Texture {
	// AGBR is backwards from way we will be filling in out bytes
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	tex.Update(nil, pixels, w*4) // Can't provide a rectangle, pitch = 4 bytes per pixel
	return tex
}

func getMouseState() mouseState {
	mouseX, mouseY, mouseButtonState := sdl.GetMouseState()
	// Extract data from bitmask
	leftButton := mouseButtonState & sdl.ButtonLMask()  // 1
	rightButton := mouseButtonState & sdl.ButtonRMask() // 4
	var result mouseState
	result.x = int(mouseX)
	result.y = int(mouseY)
	result.leftButton = !(leftButton == 0)
	result.rightButton = !(rightButton == 0)
	return result
}

func aptToTexture(pic *picture, w, h int, renderer *sdl.Renderer) *sdl.Texture {
	// Keep between -1.0 and 1.0
	scale := float32(255 / 2)
	offset := float32(-1.0 * scale)

	pixels := make([]byte, w*h*4)
	pixelIndex := 0
	for yi := 0; yi < h; yi++ {
		y := float32(yi)/float32(h)*2 - 1 // scale y by height of texture
		for xi := 0; xi < w; xi++ {
			x := float32(xi)/float32(w)*2 - 1
			r := pic.r.Eval(x, y)
			g := pic.g.Eval(x, y)
			b := pic.b.Eval(x, y)
			pixels[pixelIndex] = byte(r*scale - offset)
			pixelIndex++
			pixels[pixelIndex] = byte(g*scale - offset)
			pixelIndex++
			pixels[pixelIndex] = byte(b*scale - offset)
			pixelIndex++
			pixelIndex++
		}
	}
	return pixelsToTexture(renderer, pixels, w, h)
}
