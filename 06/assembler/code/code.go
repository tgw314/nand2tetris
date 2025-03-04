package code

import "fmt"

func Dest(s string) (string, error) {
	bin, ok := map[string]string{
		"":    "000",
		"M":   "001",
		"D":   "010",
		"DM":  "011",
		"MD":  "011", // PongL.asm に登場 (可換なんかい)
		"A":   "100",
		"AM":  "101",
		"AD":  "110",
		"ADM": "111",
	}[s]

	if !ok {
		return "", fmt.Errorf("invalid dest: %s", s)
	}

	return bin, nil
}

func Comp(s string) (string, error) {
	bin, ok := map[string]string{
		"0":   "0101010",
		"1":   "0111111",
		"-1":  "0111010",
		"D":   "0001100",
		"A":   "0110000",
		"M":   "1110000",
		"!D":  "0001101",
		"!A":  "0110001",
		"!M":  "1110001",
		"-D":  "0001111",
		"-A":  "0110011",
		"-M":  "1110011",
		"D+1": "0011111",
		"A+1": "0110111",
		"M+1": "1110111",
		"D-1": "0001110",
		"A-1": "0110010",
		"M-1": "1110010",
		"D+A": "0000010",
		"D+M": "1000010",
		"D-A": "0010011",
		"D-M": "1010011",
		"A-D": "0000111",
		"M-D": "1000111",
		"D&A": "0000000",
		"D&M": "1000000",
		"D|A": "0010101",
		"D|M": "1010101",
	}[s]

	if !ok {
		return "", fmt.Errorf("invalid comp: %s", s)
	}

	return bin, nil
}

func Jump(s string) (string, error) {
	bin, ok := map[string]string{
		"":    "000",
		"JGT": "001",
		"JEQ": "010",
		"JGE": "011",
		"JLT": "100",
		"JNE": "101",
		"JLE": "110",
		"JMP": "111",
	}[s]

	if !ok {
		return "", fmt.Errorf("invalid jump: %s", s)
	}

	return bin, nil
}
