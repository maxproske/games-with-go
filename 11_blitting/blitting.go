package main

// Assets from opengameart.org
// http://gameswithgo.org/balloons/balloons.zip

// A blit is copying the image from one surface to another -- possibly cropped and shifted

import (
	"fmt"
	"image/png" // access to functions that can load png files
	"os"        // for loading files
	"time"

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

		// Clear everything
		clear(pixels)

		// Draw balloons
		for _, tex := range balloonTextures {
			tex.drawAlpha(pixels)
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
		balloonTextures[i] = texture{pos{0, 0}, balloonPixels, w, h, w * 4}
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

func clear(pixels []byte) {
	// Goes through memory in order. So it's still fast without having to clear only unchanged pixels
	for i := range pixels {
		pixels[i] = 0
	}
}
