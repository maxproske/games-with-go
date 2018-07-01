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
	. "github.com/maxproske/games-with-go/16_genetic_algorithms/apt_random"
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

	// Create audio state.
	/*
		var audioSpec sdl.AudioSpec
		explosionBytes, _ := sdl.LoadWAV("explode.wav")
		audioID, err := sdl.OpenAudioDevice("", false, &audioSpec, nil, 0)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer sdl.FreeWAV(explosionBytes)
		audioState := audioState{explosionBytes, audioID, &audioSpec}
	*/

	// Set hints.
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	var elapsedTime float32

	currentMouseState := getMouseState()
	//prevMouseState := currentMouseState

	// Seed rand
	rand.Seed(time.Now().UTC().UnixNano())

	// Create a simple abstract syntax tree
	aptR := GetRandomNode()
	aptG := GetRandomNode()
	aptB := GetRandomNode()
	// How big to make it
	num := rand.Intn(20) // up to 20 nodes
	for i := 0; i < num; i++ {
		aptR.AddRandom(GetRandomNode()) // no leaves yet
	}
	num = rand.Intn(20) // up to 20 nodes
	for i := 0; i < num; i++ {
		aptG.AddRandom(GetRandomNode()) // no leaves yet
	}
	num = rand.Intn(20) // up to 20 nodes
	for i := 0; i < num; i++ {
		aptB.AddRandom(GetRandomNode()) // no leaves yet
	}

	// Fill out leaves
	for {
		_, nilCount := aptR.NodeCounts()
		if nilCount == 0 {
			break
		}
		aptR.AddRandom(GetRandomLeaf())
	}
	for {
		_, nilCount := aptG.NodeCounts()
		if nilCount == 0 {
			break
		}
		aptG.AddRandom(GetRandomLeaf())
	}
	for {
		_, nilCount := aptB.NodeCounts()
		if nilCount == 0 {
			break
		}
		aptB.AddRandom(GetRandomLeaf())
	}

	fmt.Println("R:", aptR)
	fmt.Println("G:", aptG)
	fmt.Println("B:", aptB)

	tex := aptToTexture(aptR, aptG, aptB, 800, 600, renderer)

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

		//prevMouseState = currentMouseState
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

func aptToTexture(redNode, greenNode, blueNode Node, w, h int, renderer *sdl.Renderer) *sdl.Texture {
	// Keep between -1.0 and 1.0
	scale := float32(255 / 2)
	offset := float32(-1.0 * scale)

	pixels := make([]byte, w*h*4)
	pixelIndex := 0
	for yi := 0; yi < h; yi++ {
		y := float32(yi)/float32(h)*2 - 1 // scale y by height of texture
		for xi := 0; xi < w; xi++ {
			x := float32(xi)/float32(w)*2 - 1
			r := redNode.Eval(x, y)
			g := greenNode.Eval(x, y)
			b := blueNode.Eval(x, y)
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
