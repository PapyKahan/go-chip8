package chip8

// SPU Chip8's sound processing unit
type SPU struct {
	SoundTimer byte
}

func (spu *SPU) initialize() {
	spu.SoundTimer = 60
}
