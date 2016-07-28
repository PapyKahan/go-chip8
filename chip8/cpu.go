package chip8

import (
	"fmt"
	"math/rand"
)

// CPU Chip8 CPU
type CPU struct {
	system         *System
	ProgramCounter uint16     // 0x200
	IndexRegister  uint16     // 0
	StackPointer   uint16     // 0
	VRegister      [16]byte   // 16
	Stack          [16]uint16 // 16
	DelayTimer     byte       // 60
}

func (cpu *CPU) initialize(system *System) {
	cpu.system = system
	cpu.ProgramCounter = 0x200
	cpu.DelayTimer = 60
}

// EmulateCycle emulate a Chip8's CPU Cycle
func (cpu *CPU) EmulateCycle() error {
	opcode := uint16(cpu.system.Mmu.memory[cpu.ProgramCounter])<<8 | uint16(cpu.system.Mmu.memory[cpu.ProgramCounter+1])

	switch opcode & 0xF000 {
	case 0x0000:
		switch opcode & 0x000F {
		case 0x0000: // 0x00E0: Clears the screen
			cpu.system.Gpu.clearScreen()
			cpu.ProgramCounter += 2
		case 0x000E: // 0x00EE: Returns from subroutine
			cpu.StackPointer--                               // 16 levels of stack, decrease stack pointer to prevent overwrite
			cpu.ProgramCounter = cpu.Stack[cpu.StackPointer] // Put the stored return address from the cpu.Stack back into the program counter
			cpu.ProgramCounter += 2                          // Don't forget to increase the program counter!
		default:
			return fmt.Errorf("Unknown opcode [0x0000]: 0x%X\n", opcode)
		}
	case 0x1000: // 0x1NNN: Jumps to address NNN
		cpu.ProgramCounter = opcode & 0x0FFF
	case 0x2000: // 0x2NNN: Calls subroutine at NNN.
		cpu.Stack[cpu.StackPointer] = cpu.ProgramCounter // Store current address in stack
		cpu.StackPointer++                               // Increment stack pointer
		cpu.ProgramCounter = opcode & 0x0FFF             // Set the program counter to the address at NNN
	case 0x3000: // 0x3XNN: Skips the next instruction if VX equals NN
		if cpu.VRegister[(opcode&0x0F00)>>8] == byte(opcode&0x00FF) {
			cpu.ProgramCounter += 4
		} else {
			cpu.ProgramCounter += 2
		}
	case 0x4000: // 0x4XNN: Skips the next instruction if VX doesn't equal NN
		if cpu.VRegister[(opcode&0x0F00)>>8] != byte(opcode&0x00FF) {
			cpu.ProgramCounter += 4
		} else {
			cpu.ProgramCounter += 2
		}
	case 0x5000: // 0x5XY0: Skips the next instruction if VX equals VY.
		if cpu.VRegister[(opcode&0x0F00)>>8] == cpu.VRegister[(opcode&0x00F0)>>4] {
			cpu.ProgramCounter += 4
		} else {
			cpu.ProgramCounter += 2
		}
	case 0x6000: // 0x6XNN: Sets VX to NN.
		cpu.VRegister[(opcode&0x0F00)>>8] = byte(opcode & 0x00FF)
		cpu.ProgramCounter += 2
	case 0x7000: // 0x7XNN: Adds NN to VX.
		cpu.VRegister[(opcode&0x0F00)>>8] += byte(opcode & 0x00FF)
		cpu.ProgramCounter += 2
	case 0x8000:
		switch opcode & 0x000F {
		case 0x0000: // 0x8XY0: Sets VX to the value of VY
			cpu.VRegister[(opcode&0x0F00)>>8] = cpu.VRegister[(opcode&0x00F0)>>4]
			cpu.ProgramCounter += 2
		case 0x0001: // 0x8XY1: Sets VX to "VX OR VY"
			cpu.VRegister[(opcode&0x0F00)>>8] |= cpu.VRegister[(opcode&0x00F0)>>4]
			cpu.ProgramCounter += 2
		case 0x0002: // 0x8XY2: Sets VX to "VX AND VY"
			cpu.VRegister[(opcode&0x0F00)>>8] &= cpu.VRegister[(opcode&0x00F0)>>4]
			cpu.ProgramCounter += 2
		case 0x0003: // 0x8XY3: Sets VX to "VX XOR VY"
			cpu.VRegister[(opcode&0x0F00)>>8] ^= cpu.VRegister[(opcode&0x00F0)>>4]
			cpu.ProgramCounter += 2
		case 0x0004: // 0x8XY4: Adds VY to VX. VF is set to 1 when there's a carry, and to 0 when there isn't
			if cpu.VRegister[(opcode&0x00F0)>>4] > byte(0xFF-int(cpu.VRegister[(opcode&0x0F00)>>8])) {
				cpu.VRegister[0xF] = 1 //carry
			} else {
				cpu.VRegister[0xF] = 0
			}
			cpu.VRegister[(opcode&0x0F00)>>8] += cpu.VRegister[(opcode&0x00F0)>>4]
			cpu.ProgramCounter += 2
		case 0x0005: // 0x8XY5: VY is subtracted from VX. VF is set to 0 when there's a borrow, and 1 when there isn't
			if cpu.VRegister[(opcode&0x00F0)>>4] > cpu.VRegister[(opcode&0x0F00)>>8] {
				cpu.VRegister[0xF] = 0 // there is a borrow
			} else {
				cpu.VRegister[0xF] = 1
			}
			cpu.VRegister[(opcode&0x0F00)>>8] -= cpu.VRegister[(opcode&0x00F0)>>4]
			cpu.ProgramCounter += 2
		case 0x0006: // 0x8XY6: Shifts VX right by one. VF is set to the value of the least significant bit of VX before the shift
			cpu.VRegister[0xF] = cpu.VRegister[(opcode&0x0F00)>>8] & 0x1
			cpu.VRegister[(opcode&0x0F00)>>8] >>= 1
			cpu.ProgramCounter += 2
		case 0x0007: // 0x8XY7: Sets VX to VY minus VX. VF is set to 0 when there's a borrow, and 1 when there isn't
			if cpu.VRegister[(opcode&0x0F00)>>8] > cpu.VRegister[(opcode&0x00F0)>>4] { // VY-VX
				cpu.VRegister[0xF] = 0 // there is a borrow
			} else {
				cpu.VRegister[0xF] = 1
			}
			cpu.VRegister[(opcode&0x0F00)>>8] = cpu.VRegister[(opcode&0x00F0)>>4] - cpu.VRegister[(opcode&0x0F00)>>8]
			cpu.ProgramCounter += 2
		case 0x000E: // 0x8XYE: Shifts VX left by one. VF is set to the value of the most significant bit of VX before the shift
			cpu.VRegister[0xF] = cpu.VRegister[(opcode&0x0F00)>>8] >> 7
			cpu.VRegister[(opcode&0x0F00)>>8] <<= 1
			cpu.ProgramCounter += 2
		default:
			return fmt.Errorf("Unknown opcode [0x8000]: 0x%X\n", opcode)
		}
	case 0x9000: // 0x9XY0: Skips the next instruction if VX doesn't equal VY
		if cpu.VRegister[(opcode&0x0F00)>>8] != cpu.VRegister[(opcode&0x00F0)>>4] {
			cpu.ProgramCounter += 4
		} else {
			cpu.ProgramCounter += 2
		}
	case 0xA000: // ANNN: Sets I to the address NNN
		cpu.IndexRegister = opcode & 0x0FFF
		cpu.ProgramCounter += 2
	case 0xB000: // BNNN: Jumps to the addres NNN plus V0
		cpu.ProgramCounter = (opcode & 0x0FFF) + uint16(cpu.VRegister[0])
	case 0xC000: // CXNN: Sets VX to a random number and NN
		random := byte(rand.Intn(0xFF + 1))
		cpu.VRegister[(opcode&0x0F00)>>8] = random & byte(opcode&0x00FF)
		cpu.ProgramCounter += 2
	case 0xD000: // DXYN: Draws a sprite at coordinate (VX, VY) that has a width of 8 pixels and a height of N pixels.
		// Each row of 8 pixels is read as bit-coded starting from memory location I;
		// I value doesn't change after the execution of this instruction.
		// VF is set to 1 if any screen pixels are flipped from set to unset when the sprite is drawn,
		// and to 0 if that doesn't happen
		{
			var x, y, height uint16
			x = uint16(cpu.VRegister[(opcode&0x0F00)>>8])
			y = uint16(cpu.VRegister[(opcode&0x00F0)>>4])
			height = opcode & 0x000F
			cpu.VRegister[0xF] = 0
			for yline := uint16(0); yline < height; yline++ {
				pixel := cpu.system.Mmu.memory[cpu.IndexRegister+yline]
				for xline := uint16(0); xline < 8; xline++ {
					if pixel&(0x80>>xline) != 0 {
						offset := x + xline + ((y + yline) * 64)
						if len(cpu.system.Gpu.FrameBuffer) > int(offset) {
							if cpu.system.Gpu.FrameBuffer[offset] == 1 {
								cpu.VRegister[0xF] = 1
							}
							cpu.system.Gpu.FrameBuffer[offset] ^= 1
						}
					}
				}
			}

			cpu.system.Gpu.Redraw = true
			cpu.ProgramCounter += 2
		}
	case 0xE000:
		switch opcode & 0x00FF {
		case 0x009E: // EX9E: Skips the next instruction if the key stored in VX is pressed
			if cpu.system.Input.KeyStates[cpu.VRegister[(opcode&0x0F00)>>8]] != 0 {
				cpu.ProgramCounter += 4
			} else {
				cpu.ProgramCounter += 2
			}
		case 0x00A1: // EXA1: Skips the next instruction if the key stored in VX isn't pressed
			if cpu.system.Input.KeyStates[cpu.VRegister[(opcode&0x0F00)>>8]] == 0 {
				cpu.ProgramCounter += 4
			} else {
				cpu.ProgramCounter += 2
			}
		default:
			return fmt.Errorf("Unknown opcode [0xE000]: 0x%X\n", opcode)
		}
	case 0xF000:
		switch opcode & 0x00FF {
		case 0x0007: // FX07: Sets VX to the value of the delay timer
			cpu.VRegister[(opcode&0x0F00)>>8] = cpu.DelayTimer
			cpu.ProgramCounter += 2
		case 0x000A: // FX0A: A key press is awaited, and then stored in VX
			keyPress := false
			var i byte
			for i = 0; i < 16; i++ {
				if cpu.system.Input.KeyStates[i] != 0 {
					cpu.VRegister[(opcode&0x0F00)>>8] = i
					keyPress = true
				}
			}

			// If we didn't received a keypress, skip this cycle and try again.
			if !keyPress {
				return nil
			}

			cpu.ProgramCounter += 2

		case 0x0015: // FX15: Sets the delay timer to VX
			cpu.DelayTimer = cpu.VRegister[(opcode&0x0F00)>>8]
			cpu.ProgramCounter += 2
		case 0x0018: // FX18: Sets the sound timer to VX
			cpu.system.Spu.SoundTimer = cpu.VRegister[(opcode&0x0F00)>>8]
			cpu.ProgramCounter += 2
		case 0x001E: // FX1E: Adds VX to I
			if cpu.IndexRegister+uint16(cpu.VRegister[(opcode&0x0F00)>>8]) > 0xFFF { // VF is set to 1 when range overflow (I+VX>0xFFF), and 0 when there isn't.
				cpu.VRegister[0xF] = 1
			} else {
				cpu.VRegister[0xF] = 0
			}
			cpu.IndexRegister += uint16(cpu.VRegister[(opcode&0x0F00)>>8])
			cpu.ProgramCounter += 2
		case 0x0029: // FX29: Sets I to the location of the sprite for the character in VX. Characters 0-F (in hexadecimal) are represented by a 4x5 font
			cpu.IndexRegister = uint16(cpu.VRegister[(opcode&0x0F00)>>8]) * 0x5
			cpu.ProgramCounter += 2
			break

		case 0x0033: // FX33: Stores the Binary-coded decimal representation of VX at the addresses I, I plus 1, and I plus 2
			cpu.system.Mmu.memory[cpu.IndexRegister] = cpu.VRegister[(opcode&0x0F00)>>8] / 100
			cpu.system.Mmu.memory[cpu.IndexRegister+1] = (cpu.VRegister[(opcode&0x0F00)>>8] / 10) % 10
			cpu.system.Mmu.memory[cpu.IndexRegister+2] = (cpu.VRegister[(opcode&0x0F00)>>8] % 100) % 10
			cpu.ProgramCounter += 2
		case 0x0055: // FX55: Stores V0 to VX in memory starting at address I
			var i uint16
			for i = 0; i <= ((opcode & 0x0F00) >> 8); i++ {
				cpu.system.Mmu.memory[cpu.IndexRegister+i] = cpu.VRegister[i]
			}

			// On the original interpreter, when the operation is done, I = I + X + 1.
			cpu.IndexRegister += ((opcode & 0x0F00) >> 8) + 1
			cpu.ProgramCounter += 2
		case 0x0065: // FX65: Fills V0 to VX with values from memory starting at address I
			var i uint16
			for i = 0; i < ((opcode & 0x0F00) >> 8); i++ {
				cpu.VRegister[i] = cpu.system.Mmu.memory[cpu.IndexRegister+i]
			}

			// On the original interpreter, when the operation is done, I = I + X + 1.
			cpu.IndexRegister += ((opcode & 0x0F00) >> 8) + 1
			cpu.ProgramCounter += 2
		default:
			return fmt.Errorf("Unknown opcode [0xF000]: 0x%X\n", opcode)
		}

		// Update timers
		if cpu.DelayTimer > 0 {
			cpu.DelayTimer--
		}

		if cpu.system.Spu.SoundTimer > 0 {
			if cpu.system.Spu.SoundTimer == 1 {
				fmt.Printf("BEEP!\n")
			}
			cpu.system.Spu.SoundTimer--
		}
	}

	return nil
}
