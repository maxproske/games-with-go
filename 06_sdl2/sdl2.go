// 1. Install GCC for Windows
//   Binding SDL2 in Go requires Windows to compile C code.
//   http://mingw-w64.org/doku.php/download/mingw-builds (During installation, choose x86_64)

// 2. Download SDL2 (MinGW 32/64-bit)
//   https://www.libsdl.org/download-2.0.php

// 3. Drag the contents of SDL2-2.0.8\x86_64-w64-mingw32 into Program Files\mingw-w64\mingw64\x86_64-w64-mingw32

// 4. Set Environment variables
//   For SDL2: New System variable > C:\Program Files\mingw-w64\x86_64-7.3.0-posix-seh-rt_v5-rev0\mingw64\x86_64-w64-mingw32\bin
//   For GCC: New System variable > C:\Program Files\mingw-w64\x86_64-7.3.0-posix-seh-rt_v5-rev0\mingw64\bin
//   Restart VS Code

package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl" // Run: go get github.com/veandco/go-sdl2/sdl
)

// Initialize constants.
const winWidth, winHeight int = 800, 600

// Initialize structs.
type color struct {
	r, g, b byte
}

func main() {
	// Create a window.
	// Can return more than one (result, error type).
	window, err := sdl.CreateWindow(
		"Testing SDL2",          // title string
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

	for y := 0; y < winHeight; y++ {
		for x := 0; x < winWidth; x++ {
			setPixel(x, y, color{byte(x % 255), byte(y % 255), 0}, pixels)
		}
	}

	// Associate it with our texture
	tex.Update(
		nil,        // rect *Rect
		pixels,     // pixels []byte
		winWidth*4, // pitch int
	)
	renderer.Copy(tex, nil, nil) // Copy
	renderer.Present()           // Present

	// Wait a specified number of milliseconds before returning.
	sdl.Delay(2000)
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
