package chip8

// GPU Chip8's graphical processing unit
type GPU struct {
	mmu 		*MMU
	FrameBuffer [2048]byte
	Redraw      bool
}

func (gpu *GPU) initialize() {
	gpu.Redraw = false
}

func (gpu *GPU) clearScreen() {
	for i := 0; i < 2048; i++ {
		gpu.FrameBuffer[i] = 0x0
	}
	gpu.Redraw = true
}