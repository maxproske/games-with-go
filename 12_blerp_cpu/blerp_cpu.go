package main

// Assets from opengameart.org
// http://gameswithgo.org/balloons/balloons.zip

// A blit is copying the image from one surface to another -- possibly cropped and shifted

import (
	"fmt"
	"image/png" // access to functions that can load png files
	"os"        // for loading files
	"time"

	noise "github.com/maxproske/games-with-go/10_package_noise"
	"github.com/veandco/go-sdl2/sdl"
)

// Initialize constants.
const winWidth, winHeight int = 800, 600

// Initialize structs.
type rgba struct {
	r, g, b byte
}
type texture struct {
	pos
	pixels      []byte
	w, h, pitch int // pitch is swith * size of each pixel
	scale       float32
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

	// Create a renderer.
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer renderer.Destroy()

	// Create a texture.
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(winWidth), int32(winHeight))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tex.Destroy()

	// Make some noise
	cloudNoise, min, max := noise.MakeNoise(noise.FBM, 0.009, 0.5, 3, 3, winWidth, winHeight)
	cloudGradient := getGradient(rgba{0, 0, 255}, rgba{255, 255, 255})
	cloudPixels := rescaleAndDraw(cloudNoise, min, max, cloudGradient, winWidth, winHeight)
	cloudTexture := texture{pos{0, 0}, cloudPixels, winWidth, winHeight, winWidth * 4, 1}

	// Create a slice of pixels
	pixels := make([]byte, winWidth*winHeight*4)

	// Blitting (copy texture to screen bufferm, without having to calculate with math)
	balloonTextures := loadBalloons()
	// Direction
	dir := 1

	for {

		frameStart := time.Now()

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}

		// Draw background
		cloudTexture.draw(pixels)

		// Draw balloons
		for _, tex := range balloonTextures {
			tex.drawBilinearScaled(tex.scale, tex.scale, pixels)
		}

		// Move balloons
		balloonTextures[1].x += float32(dir)
		if balloonTextures[1].x > 400 || balloonTextures[1].x < 0 {
			dir *= -1
		}

		// Associate it with our texture
		tex.Update(nil, pixels, winWidth*4)
		renderer.Copy(tex, nil, nil)
		renderer.Present()

		// Frame rate independence (in milliseconds)
		elapsedTime := float32(time.Since(frameStart).Seconds() * 1000) // shortcut for newTime = time.Now - frameStart;

		fmt.Println("ms per frame:", elapsedTime)

		// Limit framerate to 200fps
		if elapsedTime < 5 {
			sdl.Delay(5 - uint32(elapsedTime))
		}
		sdl.Delay(16)
	}
}

func loadBalloons() []texture {

	balloonStrs := []string{"balloon_red.png", "balloon_green.png", "balloon_blue.png"}
	balloonTextures := make([]texture, len(balloonStrs))

	for i, bstr := range balloonStrs {
		// Open each balloon string
		infile, err := os.Open(bstr)
		if err != nil {
			panic(err) // stop immediately and print the error
		}
		defer infile.Close()

		// Pass to png package
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

		// Pitch is width * 4
		balloonTextures[i] = texture{pos{float32(i * 60), float32(i * 60)}, balloonPixels, w, h, w * 4, float32(1 + i)}
	}
	return balloonTextures
}

// Takes a position and the screen buffer
func (tex *texture) draw(pixels []byte) {
	for y := 0; y < tex.h; y++ {
		// Compute screen positions
		screenY := y + int(tex.y) // margin-top (loop invariance)
		for x := 0; x < tex.w; x++ {
			screenX := x + int(tex.x) // margin-left
			// Only draw pixels on the screen
			if screenX >= 0 && screenX < winWidth && screenY >= 0 && screenY < winHeight {
				// Convert texture x,y to a texture index
				texIndex := y*tex.pitch + x*4                 // index to read data from texture
				screenIndex := screenY*winWidth*4 + screenX*4 // index to place data on the screen

				pixels[screenIndex] = tex.pixels[texIndex]
				pixels[screenIndex+1] = tex.pixels[texIndex+1]
				pixels[screenIndex+2] = tex.pixels[texIndex+2]
				pixels[screenIndex+3] = tex.pixels[texIndex+3]
			}
		}
	}
}

// More expensive version of draw that blends alpha
func (tex *texture) drawAlpha(pixels []byte) {
	for y := 0; y < tex.h; y++ {
		// Compute screen positions
		screenY := y + int(tex.y)
		for x := 0; x < tex.w; x++ {
			screenX := x + int(tex.x)
			if screenX >= 0 && screenX < winWidth && screenY >= 0 && screenY < winHeight {
				texIndex := y*tex.pitch + x*4                 // index to read data from texture
				screenIndex := screenY*winWidth*4 + screenX*4 // index to place data on the screen

				// Most people use opengl or let sdl do alpha blending for it
				// https://en.wikipedia.org/wiki/Alpha_compositing
				srcR := int(tex.pixels[texIndex]) // get source values. make larger as index and we will convert back later.
				srcG := int(tex.pixels[texIndex+1])
				srcB := int(tex.pixels[texIndex+2])
				srcA := int(tex.pixels[texIndex+3])

				dstR := int(pixels[screenIndex]) // set destination values. ignore alpha for now
				dstG := int(pixels[screenIndex+1])
				dstB := int(pixels[screenIndex+2])

				// Alpha blending by hand
				rstR := (srcR*255 + dstR*(255-srcA)) / 255 // scale it back down to be between 0-255
				rstG := (srcG*255 + dstG*(255-srcA)) / 255
				rstB := (srcB*255 + dstB*(255-srcA)) / 255

				pixels[screenIndex] = byte(rstR) // cast back to a byte
				pixels[screenIndex+1] = byte(rstG)
				pixels[screenIndex+2] = byte(rstB)
				// no alpha (pixels[screenIndex+3])
			}
		}
	}
}

// Nearest neighbour
func (tex *texture) drawScaled(scaleX, scaleY float32, pixels []byte) {
	newWidth := int(float32(tex.w) * scaleX)
	newHeight := int(float32(tex.h) * scaleY)
	// Precompute the pitch
	texW4 := tex.w * 4

	for y := 0; y < newHeight; y++ {
		fy := float32(y) / float32(newHeight) * float32(tex.h-1) // percentage along y axis we are on the new texture
		fyi := int(fy)                                           // integer approximation
		screenY := int(fy*scaleY) + int(tex.y)                   // coordinate to draw on screen
		screenIndex := screenY*winWidth*4 + int(tex.x)*4         // index in the pixels array
		for x := 0; x < newWidth; x++ {
			fx := float32(x) / float32(newWidth) * float32(tex.w-1)
			screenX := int(fx*scaleX) + int(tex.x)
			if screenX >= 0 && screenX < winWidth && screenY >= 0 && screenY < winHeight {
				fxi4 := int(fx) * 4
				pixels[screenIndex] = tex.pixels[fyi*texW4+fxi4] // Set screen pixels red
				screenIndex++
				pixels[screenIndex] = tex.pixels[fyi*texW4+fxi4+1] // Set screen pixels green
				screenIndex++
				pixels[screenIndex] = tex.pixels[fyi*texW4+fxi4+2] // Set screen pixels blue
				screenIndex++
				screenIndex++ // skip alpha
			}
		}
	}
}

// Bilinear interpolation. Not worthwhile to do in software.
func (tex *texture) drawBilinearScaled(scaleX, scaleY float32, pixels []byte) {
	newWidth := int(float32(tex.w) * scaleX)
	newHeight := int(float32(tex.h) * scaleY)
	// Precompute the pitch
	texW4 := tex.w * 4

	for y := 0; y < newHeight; y++ {
		fy := float32(y) / float32(newHeight) * float32(tex.h-1) // percentage along y axis we are on the new texture
		fyi := int(fy)                                           // integer approximation
		ty := fy - float32(fyi)                                  // percentage between two bottom pixels
		screenY := int(fy*scaleY) + int(tex.y)                   // coordinate to draw on screen
		screenIndex := screenY*winWidth*4 + int(tex.x)*4         // index in the pixels array

		for x := 0; x < newWidth; x++ {
			fx := float32(x) / float32(newWidth) * float32(tex.w-1)
			screenX := int(fx*scaleX) + int(tex.x)

			if screenX >= 0 && screenX < winWidth && screenY >= 0 && screenY < winHeight {
				fxi := int(fx)
				// Get index, for four pixels around the one we want
				c00i := fyi*texW4 + fxi*4     // top left
				c10i := fyi*texW4 + (fxi+1)*4 // reverse these
				c01i := (fyi+1)*texW4 + fxi*4 // reverse these
				c11i := (fyi+1)*texW4 + (fxi+1)*4

				tx := fx - float32(fxi) // percentage between top two pixels

				for i := 0; i < 4; i++ {
					// Get rgb, without having to repeat oruselves
					c00 := float32(tex.pixels[c00i+i])
					c10 := float32(tex.pixels[c10i+i]) // reverse these
					c01 := float32(tex.pixels[c01i+i]) // reverse these
					c11 := float32(tex.pixels[c11i+i])

					pixels[screenIndex] = byte(blerp(c00, c10, c01, c11, tx, ty))
					screenIndex++
				}
			}
		}
	}
}

// Linear interpolations on float
func flerp(a, b, pct float32) float32 {
	return a + (b-a)*pct
}

// Bilinear interpolation
func blerp(c00, c10, c01, c11, tx, ty float32) float32 {
	// lerp top row, lerp bottom row, lerp column
	return flerp(flerp(c00, c10, tx), flerp(c01, c11, tx), ty)
}

func clear(pixels []byte) {
	// Goes through memory in order. So it's still fast without having to clear only unchanged pixels
	for i := range pixels {
		pixels[i] = 0
	}
}

func setPixel(x, y int, c rgba, pixels []byte) {
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

// Go from color 1 to 2, then suddenly switch to 3
func getDualGradient(c1, c2, c3, c4 rgba) []rgba {
	result := make([]rgba, 256)
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
