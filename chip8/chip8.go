package chip8

import (
	"fmt"
	"os"
)

type System struct {
	Cpu   *CPU
	Mmu   *MMU
	Gpu   *GPU
	Spu   *SPU
	Input *Input
}

func (system *System) initialize() {
	system.Mmu = &MMU{}
	system.Mmu.initialize()

	system.Gpu = &GPU{}
	system.Gpu.initialize()

	system.Spu = &SPU{}
	system.Spu.initialize()

	system.Input = &Input{}

	system.Cpu = &CPU{}
	system.Cpu.initialize(system)
}

// New Creates a new Chip8's CPU instance
func New() *System {
	system := &System{}
	system.initialize()
	return system
}

// LoadRom load rom filename into Chip8 memory
func (system *System) LoadRom(filename string) error {
	// TODO: this can be done ouside CPU by only passing to this method buffer array.
	// Open file
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Check file size
	stat, err := file.Stat()
	if err != nil {
		return err
	}

	lSize := stat.Size()
	fmt.Printf("Filesize: %d\n", int(lSize))

	// Allocate memory to contain the whole file
	buffer := make([]byte, lSize)
	if buffer == nil {
		return fmt.Errorf("Memory error")
	}

	// Copy the file into the buffer
	count, err := file.Read(buffer)
	if err != nil {
		return err
	}
	if count != int(lSize) {
		return fmt.Errorf("Reading error")
	}

	// Copy buffer to Chip8 memory
	if (4096 - 512) > lSize {
		for i := 0; i < int(lSize); i++ {
			system.Mmu.memory[i+512] = buffer[i]
		}
	} else {
		return fmt.Errorf("Error: ROM too big for memory")
	}

	// Close file, free buffer
	return nil
}
