package gui

import (
	"github.com/veandco/go-sdl2/sdl"
)

type MouseState struct {
	LeftButton      bool
	RightButton     bool
	PrevLeftButton  bool
	PrevRightButton bool
	PrevX, PrevY    int
	X, Y            int
}

func GetMouseState() *MouseState {
	mouseX, mouseY, mouseButtonState := sdl.GetMouseState()
	// Extract data from bitmask
	leftButton := mouseButtonState & sdl.ButtonLMask()  // 1
	rightButton := mouseButtonState & sdl.ButtonRMask() // 4
	var result MouseState
	result.X = int(mouseX)
	result.Y = int(mouseY)
	result.LeftButton = !(leftButton == 0)
	result.RightButton = !(rightButton == 0)
	return &result
}

func (mouseState *MouseState) Update() {
	// Set previous state
	mouseState.PrevX = mouseState.X
	mouseState.PrevY = mouseState.Y
	mouseState.PrevLeftButton = mouseState.LeftButton
	mouseState.PrevRightButton = mouseState.RightButton

	// Get current state from SDL
	X, Y, mouseButtonState := sdl.GetMouseState()
	mouseState.X = int(X)
	mouseState.Y = int(Y)
	mouseState.LeftButton = !((mouseButtonState * sdl.ButtonLMask()) == 0)
	mouseState.RightButton = !((mouseButtonState * sdl.ButtonRMask()) == 0)
}

type ImageButton struct {
	Image *sdl.Texture // Use sdltexture for image
	Rect  sdl.Rect     // place on screen we will darw it
	//onclicked callback // TODO
	WasLeftClicked  bool
	WasRightClicked bool
	IsSelected      bool
	SelectedTex     *sdl.Texture
}

// Consutrctor
func NewImageButton(renderer *sdl.Renderer, image *sdl.Texture, rect sdl.Rect, selectedColor sdl.Color) *ImageButton {
	// Make a 1px picture
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, 1, 1)
	if err != nil {
		panic(err)
	}
	pixels := make([]byte, 4)
	pixels[0] = selectedColor.R
	pixels[1] = selectedColor.G
	pixels[2] = selectedColor.B
	pixels[3] = selectedColor.A

	tex.Update(nil, pixels, 4)
	return &ImageButton{image, rect, false, false, false, tex}
}

func (button *ImageButton) Update(mouseState *MouseState) {
	// If the mouse cursor is located on top of the image
	if button.Rect.HasIntersection(&sdl.Rect{int32(mouseState.X), int32(mouseState.Y), 1, 1}) {
		button.WasLeftClicked = mouseState.PrevLeftButton && !mouseState.LeftButton // If on the prev frame, the button was held down. And the current isn't, consider that a click
		button.WasRightClicked = mouseState.PrevRightButton && !mouseState.RightButton
	} else {
		button.WasLeftClicked = false
		button.WasRightClicked = false
	}
}

// Image button needs to draw itself
func (button *ImageButton) Draw(renderer *sdl.Renderer) {
	if button.IsSelected {
		// Make a border rectangle
		borderRect := button.Rect
		borderThickness := int32(float32(borderRect.W) * 0.01) // 1% of the button width
		borderRect.W = button.Rect.W + borderThickness*2       // Draw square with extra 1% behind (left and right for *2)
		borderRect.H = button.Rect.H + borderThickness*2       // Draw square with extra 1% behind (left and right for *2)
		borderRect.X -= borderThickness
		borderRect.Y -= borderThickness
		renderer.Copy(button.SelectedTex, nil, &borderRect)
	}
	renderer.Copy(button.Image, nil, &button.Rect) // Copy button image to button rectangle
}
