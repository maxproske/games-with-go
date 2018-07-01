package main

// 1. `go run lexing.go` to run app
// 2. Right click iamge to zoom in
// 3. Press 's' key on zoomed image to save picture as 1.apt
// 4. `go run 1.apt` to generate output

// Genetic algorithms are a type of optimization algorithm, called a metaheuristic
//   A metaheuristic will find an answer you hope will be good, but not the optimal solution
//   (Eg. The solution space is optimal amount of air to add to fuel for combustion.)
//   Hill-climbing algorithm find a random place to start, and changes one thing. If it improves, make that change again.
import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	noise "github.com/maxproske/games-with-go/10_package_noise"
	. "github.com/maxproske/games-with-go/20_lexing/apt"
	. "github.com/maxproske/games-with-go/20_lexing/gui"
	sdl "github.com/veandco/go-sdl2/sdl"
)

var winWidth, winHeight int = 600, 600
var rows, cols, numPics int = 3, 3, rows * cols

type pixelResult struct {
	pixels []byte
	index  int
}

type guiState struct {
	zoom      bool
	zoomImage *sdl.Texture
	zoomTree  *picture
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
	num := rand.Intn(25) + 5 // up to 20 nodes
	for i := 0; i < num; i++ {
		p.r.AddRandom(GetRandomNode()) // no leaves yet
	}
	num = rand.Intn(25) + 5 // up to 20 nodes
	for i := 0; i < num; i++ {
		p.g.AddRandom(GetRandomNode()) // no leaves yet
	}
	num = rand.Intn(25) + 5 // up to 20 nodes
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

// Helper function
func cross(a *picture, b *picture) *picture {
	// Get copy of A
	aCopy := &picture{CopyTree(a.r, nil), CopyTree(a.g, nil), CopyTree(a.b, nil)} // nil = root of tree
	// Choose random node of three start nodes
	aColor := aCopy.pickRandomColor()
	bColor := b.pickRandomColor()
	// Crossover point for each
	aIndex := rand.Intn(aColor.NodeCount())   // Index to a random point in the A tree
	aNode, _ := GetNthNode(aColor, aIndex, 0) // Get that node

	bIndex := rand.Intn(bColor.NodeCount())
	bNode, _ := GetNthNode(bColor, bIndex, 0)

	bNodeCopy := CopyTree(bNode, bNode.GetParent()) // Both are copies
	ReplaceNode(aNode, bNodeCopy)                   // Replace node in A tree
	return aCopy
}

// Get new generation of picture trees given a slice of survivors
func evolve(survivors []*picture) []*picture {
	newPics := make([]*picture, numPics)
	// Some part of every tree ends up
	i := 0
	for i < len(survivors) {
		// Iterate through survovrs at least 1 time
		a := survivors[i]                         // Pick our two trees to combine
		b := survivors[rand.Intn(len(survivors))] // A and b could end up being the same tree
		// New tree = cross(a,b)
		newPics[i] = cross(a, b)
		i++
	}

	// If user selects 3 pictures, we will only generate 3 new images. So keep going.
	for i < len(newPics) {
		a := survivors[rand.Intn(len(survivors))]
		b := survivors[rand.Intn(len(survivors))]
		newPics[i] = cross(a, b)
		i++
	}

	// Tune the rate of mutation. Mutate every picture 0-4 times
	for _, pic := range newPics {
		r := rand.Intn(4)
		for i := 0; i < r; i++ {
			pic.mutate()
		}
	}

	return newPics
}

// Operate on picture
func (p *picture) pickRandomColor() Node {
	r := rand.Intn(3)
	switch r {
	case 0:
		return p.r
	case 1:
		return p.g
	case 2:
		return p.b
	default:
		panic("pickRandomColor failed.")
	}

}

func (p *picture) mutate() {
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

	// Get command line arguments
	args := os.Args
	if len(args) > 1 { // First argument is always program's name itself
		fileBytes, err := ioutil.ReadFile(args[1])
		if err != nil {
			panic(err)
		}
		fileStr := string(fileBytes)
		_ = BeginLexing(fileStr) // Begin lexing our string
		return
	}

	// Seed rand
	rand.Seed(time.Now().UTC().UnixNano())

	// Get tree
	picTrees := make([]*picture, numPics)
	for i := range picTrees {
		picTrees[i] = NewPicture()
	}

	picWidth := int(float32(winWidth/cols) * float32(0.9))
	picHeight := int(float32(winHeight/rows) * float32(0.8)) // 0.8 instead of 0.9 to msake room at the bottom of the screen

	// Use channel and pull stuff out as soon as it's ready
	pixelsChannel := make(chan pixelResult, numPics) // have room up to numPics

	// Get textures (slower)
	buttons := make([]*ImageButton, numPics)

	// Make evolve button
	evolveButtonTex := GetSinglePixelTex(renderer, &sdl.Color{255, 255, 255, 0})
	evolveRect := sdl.Rect{int32(float32(winWidth)/2 - float32(picWidth)/2), int32(float32(winHeight) - (float32(winHeight) * 0.1)), int32(picWidth), int32(float32(winHeight) * 0.08)} // Define a rectangle
	evolveButton := NewImageButton(renderer, evolveButtonTex, evolveRect, sdl.Color{255, 255, 255, 0})

	for i := range picTrees {
		// Pass in an integer, so each go routine hs its own copy of i
		go func(i int) {
			pixels := aptToPixels(picTrees[i], picWidth*2, picHeight*2)
			pixelsChannel <- pixelResult{pixels, i} // Put into thread-safe queue
		}(i)
	}

	keyboardState := sdl.GetKeyboardState()
	prevKeyboardState := make([]uint8, len(keyboardState))

	mouseState := GetMouseState()
	state := guiState{false, nil, nil} // Not zoomed in

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

		// Only render button if in non-zoom state
		if !state.zoom {

			// Get stuff out of the channel as soon as it's available
			// Is there anything in it?
			select {
			case pixelsAndIndex, ok := <-pixelsChannel: // Get a texture and an OK (or not) out of a channel
				if ok {
					// Fill texture array
					tex := pixelsToTexture(renderer, pixelsAndIndex.pixels, picWidth*2, picHeight*2)
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
			for i, button := range buttons {
				// Tex will be nil until we get it out of channel
				if button != nil {
					button.Update(mouseState) // Update it
					if button.WasLeftClicked {
						button.IsSelected = !button.IsSelected
					} else if button.WasRightClicked {
						// Make a larger version
						zoomPixels := aptToPixels(picTrees[i], winWidth*2, winHeight*2)
						zoomTex := pixelsToTexture(renderer, zoomPixels, winWidth*2, winHeight*2)
						state.zoomImage = zoomTex    // Set zoom image
						state.zoomTree = picTrees[i] // Equation
						state.zoom = true
					}
					button.Draw(renderer) // Draw it
				}
			}

			// Draw and update evolve button
			evolveButton.Update(mouseState)
			if evolveButton.WasLeftClicked {
				// Build up a list of all the images user has selected
				selectedPictures := make([]*picture, 0)
				for i, button := range buttons {
					if button.IsSelected {
						selectedPictures = append(selectedPictures, picTrees[i])
					}
				}
				if len(selectedPictures) != 0 {
					// Clear out buttons
					for i := range buttons {
						buttons[i] = nil
					}
					// Evolve. ":=" will be a new tree outside the scope, and won't do what we think
					picTrees = evolve(selectedPictures) // Replcae original picTrees with slice we get frome evolve
					for i := range picTrees {
						go func(i int) {
							pixels := aptToPixels(picTrees[i], picWidth*2, picHeight*2) // Double width and height of images to get free anti-aliasing (beacuse of rendering hint)
							pixelsChannel <- pixelResult{pixels, i}                     //  Channel is still setup ready to use
						}(i)
					}
				}
			}
			evolveButton.Draw(renderer)
		} else {
			// In zoomed in state
			if !mouseState.RightButton && mouseState.PrevRightButton {
				state.zoom = false
			}
			if keyboardState[sdl.SCANCODE_S] == 0 && prevKeyboardState[sdl.SCANCODE_S] != 0 {
				// Save
				saveTree(state.zoomTree)
			}
			renderer.Copy(state.zoomImage, nil, nil)
		}
		renderer.Present()

		// Update keyboard state
		for i, v := range keyboardState {
			prevKeyboardState[i] = v
		}

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
func aptToPixels(pic *picture, w, h int) []byte {
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

// Save to disk
func saveTree(p *picture) {
	// Check for largest apt, and save as +1 that
	files, err := ioutil.ReadDir("./")
	if err != nil {
		panic(err)
	}
	biggestNumber := 0
	for _, f := range files {
		name := f.Name()                     // Get file name
		if strings.HasSuffix(name, ".apt") { // Check for .apt file suffix
			numberStr := strings.TrimSuffix(name, ".apt") // Get name that preceeds file extension
			num, err := strconv.Atoi(numberStr)           // Parse integer
			if err == nil {
				if num > biggestNumber {
					biggestNumber = num
				}
			}
		}
	}
	saveName := strconv.Itoa(biggestNumber+1) + ".apt" // Save as...
	file, err := os.Create(saveName)                   // Create a file handle
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fmt.Fprintf(file, p.String()) // Write to the file instead of console
}
