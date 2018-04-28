package main

import (
	"fmt"
	"image/png"
	"os"
	"time"

	noise "github.com/PrawnSkunk/games-with-go/10_package_noise"
	sdl "github.com/veandco/go-sdl2/sdl"
)

const winWidth, winHeight int = 800, 600

type rgba struct {
	r, g, b byte
}

// We're using sdl2 texutres, rename texture->balooon
type balloon struct {
	tex *sdl.Texture
	pos
	scale float32
	w, h  int // hard to get a width and height out of a texture
}
type pos struct {
	x, y float32
}

func main() {

	// Initialize SDL2
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sdl.Quit()

	// Create a window.
	window, err := sdl.CreateWindow("SDL2 Images Demo", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, int32(winWidth), int32(winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer window.Destroy()

	// Create GPU renderer.
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED) // sdl.RENDERER_SOFTWARE loses blerp, and alpha blending is expensive
	if err != nil {
		fmt.Println(err)
		return
	}
	defer renderer.Destroy()
	// Set blerp globally
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1") // hint: it will TRY

	// Create a texture.
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(winWidth), int32(winHeight))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tex.Destroy()

	// Make noise
	cloudNoise, min, max := noise.MakeNoise(noise.FBM, 0.009, 0.5, 3, 3, winWidth, winHeight)
	cloudGradient := getGradient(rgba{0, 0, 255}, rgba{255, 255, 255})
	cloudPixels := rescaleAndDraw(cloudNoise, min, max, cloudGradient, winWidth, winHeight)
	cloudTexture := pixelsToTexture(renderer, cloudPixels, winWidth, winHeight) // Make sdl texture

	balloons := loadBalloons(renderer)
	dir := 1

	for {

		frameStart := time.Now()

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}

		// Copy background
		renderer.Copy(cloudTexture, nil, nil) // nil = draw entire source to entire destination

		// Draw sdl texture
		for _, balloon := range balloons {
			balloon.draw(renderer)
		}

		// Move balloons
		balloons[1].x += float32(dir)
		if balloons[1].x > 400 || balloons[1].x < 0 {
			dir *= -1
		}

		renderer.Present()

		// Advance frames
		elapsedTime := float32(time.Since(frameStart).Seconds() * 1000)
		fmt.Println("ms per frame:", elapsedTime)
		if elapsedTime < 5 {
			sdl.Delay(5 - uint32(elapsedTime))
		}
		sdl.Delay(16)
	}
}

func loadBalloons(renderer *sdl.Renderer) []balloon {

	balloonStrs := []string{"balloon_red.png", "balloon_green.png", "balloon_blue.png"}
	balloons := make([]balloon, len(balloonStrs))

	for i, bstr := range balloonStrs {
		// Open each balloon string
		infile, err := os.Open(bstr)
		if err != nil {
			panic(err)
		}
		defer infile.Close()

		// Decode pngs
		img, err := png.Decode(infile)
		if err != nil {
			panic(err)
		}

		// Extract width and height
		w := img.Bounds().Max.X
		h := img.Bounds().Max.Y

		// Get a bytearray
		balloonPixels := make([]byte, w*h*4)
		bIndex := 0
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				// If there is a, rgb will be dimmed according to the alpha
				r, g, b, a := img.At(x, y).RGBA()
				// Cast int32 too byte (16bit values)
				balloonPixels[bIndex] = byte(r / 256)
				bIndex++
				balloonPixels[bIndex] = byte(g / 256)
				bIndex++
				balloonPixels[bIndex] = byte(b / 256)
				bIndex++
				balloonPixels[bIndex] = byte(a / 256)
				bIndex++
			}
		}

		tex := pixelsToTexture(renderer, balloonPixels, w, h)

		// Set blend mode. Tell sdl we want to do alpha blending
		err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
		if err != nil {
			panic(err) // if we are on hardware that doesn't support alpha blending
		}

		balloons[i] = balloon{tex, pos{float32(i * 120), float32(i * 120)}, float32(1+i) / 2, w, h}
	}
	return balloons
}

func clear(pixels []byte) {
	// Goes through memory in order. So it's still fast without having to clear only unchanged pixels
	for i := range pixels {
		pixels[i] = 0
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
func rescaleAndDraw(noise []float32, min, max float32, gradient []rgba, w, h int) []byte {
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

func colorLerp(c1, c2 rgba, pct float32) rgba {
	return rgba{lerp(c1.r, c2.r, pct), lerp(c1.g, c2.g, pct), lerp(c1.b, c2.b, pct)}
}

// Linear interpretation (lerp) between two bytes
func lerp(b1, b2 byte, pct float32) byte {
	return byte(float32(b1) + pct*(float32(b2)-float32(b1)))
}

func getGradient(c1, c2 rgba) []rgba {
	result := make([]rgba, 256)
	for i := range result {
		// Get the current percentage
		pct := float32(i) / float32(255)
		result[i] = colorLerp(c1, c2, pct)
	}
	return result
}

// Take pixels and make an SDL2 texture out of them
// Needs copy of an sdl2 renderer
func pixelsToTexture(renderer *sdl.Renderer, pixels []byte, w, h int) *sdl.Texture {
	// AGBR is backwards from way we will be filling in out bytes
	// This is just because some gpus take in the byte in backwards order
	// sdl.TEXTUREACCESS_STATIC: not making changes often
	// sdl.TEXTUREACCESS_STREAMING: making changes often
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	tex.Update(nil, pixels, w*4) // can't provide a rectangle, pitch = 4 bytes per pixel
	return tex
}

func (balloon *balloon) draw(renderer *sdl.Renderer) {
	// Balloon has a scale
	newW := int32(float32(balloon.w) * balloon.scale)
	newH := int32(float32(balloon.h) * balloon.scale)
	// Position as the center (translate to top left and start drawing)
	x := int32(balloon.x - float32(newW)/2)
	y := int32(balloon.y - float32(newH)/2)
	rect := &sdl.Rect{x, y, newW, newH}
	renderer.Copy(balloon.tex, nil, rect) // destination rect will change
}
