package main

// Genetic algorithms are a type of optimization algorithm, called a metaheuristic
//   A metaheuristic will find an answer you hope will be good, but not the optimal solution
//   (Eg. The solution space is optimal amount of air to add to fuel for combustion.)
//   Hill-climbing algorithm find a random place to start, and changes one thing. If it improves, make that change again.
import (
	"fmt"
	"math/rand"
	"time"

	noise "github.com/PrawnSkunk/games-with-go/10_package_noise"
	. "github.com/PrawnSkunk/games-with-go/18_ui/apt"
	. "github.com/PrawnSkunk/games-with-go/18_ui/gui"
	sdl "github.com/veandco/go-sdl2/sdl"
)

var winWidth, winHeight int = 600, 600
var rows, cols, numPics int = 2, 2, rows * cols

type pixelResult struct {
	pixels []byte
	index  int
}

type rgba struct {
	r, g, b byte
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
	num := rand.Intn(20) + 5 // up to 20 nodes
	for i := 0; i < num; i++ {
		p.r.AddRandom(GetRandomNode()) // no leaves yet
	}
	num = rand.Intn(20) + 5 // up to 20 nodes
	for i := 0; i < num; i++ {
		p.g.AddRandom(GetRandomNode()) // no leaves yet
	}
	num = rand.Intn(20) + 5 // up to 20 nodes
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

	// Seed rand
	rand.Seed(time.Now().UTC().UnixNano())

	// Get tree
	picTrees := make([]*picture, numPics)
	for i := range picTrees {
		picTrees[i] = NewPicture()
	}

	picWidth := int(float32(winWidth/cols) * float32(0.9))
	picHeight := int(float32(winHeight/rows) * float32(0.9))

	// Use channel and pull stuff out as soon as it's ready
	pixelsChannel := make(chan pixelResult, numPics) // have room up to numPics

	// Get textures (slower)
	buttons := make([]*ImageButton, numPics)
	for i := range picTrees {
		// Pass in an integer, so each go routine hs its own copy of i
		go func(i int) {
			pixels := aptToPixels(picTrees[i], picWidth, picHeight, renderer)
			pixelsChannel <- pixelResult{pixels, i} // Put into thread-safe queue
		}(i)
	}

	keyboardState := sdl.GetKeyboardState()
	mouseState := GetMouseState()

	for {

		frameStart := time.Now()

		// Update mouse state every frame
		mouseState.Update()

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				return
			case *sdl.TouchFingerEvent:
				if e.Type == sdl.FINGERDOWN {
					touchX := int(e.X * float32(winWidth))
					touchY := int(e.Y * float32(winHeight))
					mouseState.X = touchX
					mouseState.Y = touchY
					mouseState.LeftButton = true
				}
			}
		}

		// End the program using esc
		if keyboardState[sdl.SCANCODE_ESCAPE] != 0 {
			return
		}

		// Get stuff out of the channel as soon as it's available
		// Is there anything in it?
		select {
		case pixelsAndIndex, ok := <-pixelsChannel: // Get a texture and an OK (or not) out of a channel
			if ok {
				// Fill texture array
				tex := pixelsToTexture(renderer, pixelsAndIndex.pixels, picWidth, picHeight)
				xi := pixelsAndIndex.index % cols
				yi := (pixelsAndIndex.index - xi) / cols
				x := int32(xi * picWidth) // Figure out where to draw it
				y := int32(yi * picHeight)
				xPad := int32(float32(winWidth) * 0.1 / float32(cols+1))
				yPad := int32(float32(winHeight) * 0.1 / float32(rows+1))
				x += xPad * (int32(xi) + 1)
				y += yPad * (int32(yi) + 1)
				rect := sdl.Rect{x, y, int32(picWidth), int32(picHeight)}
				// Make image button
				button := NewImageButton(renderer, tex, rect, sdl.Color{255, 255, 255, 0})
				buttons[pixelsAndIndex.index] = button
			}
		default:
			// Default means nothing to do
		}

		renderer.Clear()

		// Copy background
		// renderer.Copy(cloudTexture, nil, nil) // nil = draw entire source to entire destination
		for _, button := range buttons {
			// Tex will be nil until we get it out of channel
			if button != nil {
				button.Update(mouseState) // Update it
				if button.WasLeftClicked {
					button.IsSelected = !button.IsSelected
				}
				button.Draw(renderer) // Draw it
			}
		}

		renderer.Present()

		elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		//fmt.Println("ms per frame:", elapsedTime)
		if elapsedTime < 5 {
			sdl.Delay(5 - uint32(elapsedTime))
			elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		}
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

// Return pixela rray not *sdl.Texture, because renderer can only be changed on the main thread
func aptToPixels(pic *picture, w, h int, renderer *sdl.Renderer) []byte {
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
	return pixels
}
