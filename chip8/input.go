package chip8

import "fmt"

// Input Chip8's input
type Input struct {
	KeyStates [16]byte
}

// SetKeyState activate a specific key
func (input *Input) SetKeyState(keyindex int, value byte) error {
	if keyindex < 0 && keyindex > 16 {
		return fmt.Errorf("Invalid key")
	}
	if value < 0 && value > 1 {
		return fmt.Errorf("Invalid keypress value")
	}
	input.KeyStates[keyindex] = value
	return nil
}
