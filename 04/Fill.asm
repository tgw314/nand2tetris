// This file is part of www.nand2tetris.org
// and the book "The Elements of Computing Systems"
// by Nisan and Schocken, MIT Press.
// File name: projects/4/Fill.asm

// Runs an infinite loop that listens to the keyboard input. 
// When a key is pressed (any key), the program blackens the screen,
// i.e. writes "black" in every pixel. When no key is pressed, 
// the screen should be cleared.

// LOOP:
//     if (KBD == 0) goto IF_0
//     goto IF_1
// IF_0:
//     if (SCREEN[0] == -1) goto FILL
//     goto LOOP
// IF_1:
//     if (SCREEN[0] == 0) goto FILL
//     goto LOOP
// FILL:
//     count = 0
// FILL_LOOP:
//     if (count >= 8192) goto LOOP
//     SCREEN[count] = !SCREEN[count]
//     count = count + 1
//     goto FILL_LOOP

(LOOP)
    // if (KBD == 0) goto IF_0
    @KBD
    D=M
    @IF_0
    D;JEQ

    // goto IF_1
    @IF_1
    0;JMP

(IF_0)
    // if (SCREEN[0] == -1) goto FILL
    @SCREEN
    D=M+1
    @FILL
    D;JEQ

    // goto LOOP
    @LOOP
    0;JMP

(IF_1)
    // if (SCREEN[0] == 0) goto FILL
    @SCREEN
    D=M
    @FILL
    D;JEQ

    // goto LOOP
    @LOOP
    0;JMP

(FILL)
    // count = 0
    @count
    M=0

(FILL_LOOP)
    // if (count >= 8192) goto LOOP
    @8192
    D=A
    @count
    D=M-D
    @LOOP
    D;JGE

    // SCREEN[count] = !SCREEN[count]
    @count
    D=M
    @SCREEN
    A=D+A
    D=M
    M=!D
    
    // count = count + 1
    @count
    M=M+1

    // goto FILL_LOOP
    @FILL_LOOP
    0;JMP
