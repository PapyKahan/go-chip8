package main

import (
	"fmt"
	"image/color"
	"log"

	"bitbucket.org/fajard_c/go-chip8/chip8"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

const (
	initScreenWidth  = 64
	initScreenHeight = 32
	initScreenScale  = 10
)

var pixel *ebiten.Image

var chip8System *chip8.System

func init() {

}

var (
	keyStates = map[ebiten.Key]int{
		// Chip8 keystates
		ebiten.Key1: 0x1,
		ebiten.Key2: 0x2,
		ebiten.Key3: 0x3,
		ebiten.Key4: 0xC,
		ebiten.KeyQ: 0x4,
		ebiten.KeyW: 0x5,
		ebiten.KeyE: 0x6,
		ebiten.KeyR: 0xD,
		ebiten.KeyA: 0x7,
		ebiten.KeyS: 0x8,
		ebiten.KeyD: 0x9,
		ebiten.KeyF: 0xE,
		ebiten.KeyZ: 0xA,
		ebiten.KeyX: 0x0,
		ebiten.KeyC: 0xB,
		ebiten.KeyV: 0xF,
	}
	debugMode bool
)

var currentFrame = [2048]byte{}

func update(screen *ebiten.Image) error {
	for key, value := range keyStates {
		if ebiten.IsKeyPressed(key) {
			err := chip8System.Input.SetKeyState(value, 1)
			if err != nil {
				return err
			}
			continue
		}
		chip8System.Input.SetKeyState(value, 0)
	}

	chip8System.Cpu.EmulateCycle()
	if chip8System.Gpu.Redraw {
		for x := 0; x < 2048; x++ {
			currentFrame[x] = chip8System.Gpu.FrameBuffer[x]
		}
		chip8System.Gpu.Redraw = false
	}

	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			if currentFrame[(y*64)+x] != 0 {
				options := &ebiten.DrawImageOptions{}
				options.GeoM.Translate(float64(x*initScreenScale), float64(y*initScreenScale))
				screen.DrawImage(pixel, options)
			}
		}
	}

	if keyStates[ebiten.KeyF5] == 1 {
		debugMode = !debugMode
	}

	if debugMode {
		x, y := ebiten.CursorPosition()
		msg := fmt.Sprintf(`Cursor: (%d, %d)
FPS: %0.2f`, x, y, ebiten.CurrentFPS())
		ebitenutil.DebugPrint(screen, msg)
	}
	return nil
}

func createPixel() error {
	p, err := ebiten.NewImage(initScreenScale, initScreenScale, ebiten.FilterNearest)
	if err != nil {
		return err
	}
	p.Fill(color.White)
	pixel = p
	return nil
}

func main() {
	if err := createPixel(); err != nil {
		panic(fmt.Errorf("Fail to create pixel image"))
	}
	chip8System = chip8.New()
	chip8System.LoadRom("roms/PONG")
	if err := ebiten.Run(update, initScreenWidth*initScreenScale, initScreenHeight*initScreenScale, 1, "Go-Chip8 Ebiten"); err != nil {
		log.Fatal(err)
	}
}
