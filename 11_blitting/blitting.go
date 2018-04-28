package main

// Assets from opengameart.org
// http://gameswithgo.org/balloons/balloons.zip

import (
	"fmt"
	"image/png" // access to functions that can load png files
	"os"        // for loading files

	"github.com/veandco/go-sdl2/sdl"
)

// Initialize constants.
const winWidth, winHeight int = 800, 600

// Initialize structs.
type rgba struct {
	r, g, b byte
}
type texture struct {
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
	pixels := make([]byte, winWidth*winHeight*4) // 4 bytes for each channel (ARGB)

	// Blitting (copy texture to screen bufferm, without having to calculate with math)
	balloonTex := loadBalloons()
	balloonTex.draw(pos{0, 0}, pixels) // give it the screen and position

	// for i := range balloonPixels {
	// 	pixels[i] = balloonPixels[i]
	// }

	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}

			// Associate it with our texture
			tex.Update(nil, pixels, winWidth*4)
			renderer.Copy(tex, nil, nil)
			renderer.Present()
			sdl.Delay(16)
		}
	}
}

func loadBalloons() *texture {
	// Give it a file name
	infile, err := os.Open("balloon_blue.png")
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
	return &texture{balloonPixels, w, h, w * 4}
}

// Takes a position and the screen buffer
func (tex *texture) draw(p pos, pixels []byte) {
	for y := 0; y < tex.h; y++ {
		for x := 0; x < tex.w; x++ {
			// Compute screen positions
			screenY := y + int(p.y) // margin-top
			screenX := x + int(p.x) // margin-left
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
