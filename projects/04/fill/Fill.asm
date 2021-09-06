// This file is part of www.nand2tetris.org
// and the book "The Elements of Computing Systems"
// by Nisan and Schocken, MIT Press.
// File name: projects/04/Fill.asm

// Runs an infinite loop that listens to the keyboard input.
// When a key is pressed (any key), the program blackens the screen,
// i.e. writes "black" in every pixel;
// the screen should remain fully black as long as the key is pressed. 
// When no key is pressed, the program clears the screen, i.e. writes
// "white" in every pixel;
// the screen should remain fully clear as long as no key is pressed.

(INIT)
  @SCREEN
  D=A
  @current
  M=D

  @24575
  D=A
  @max
  M=D

(LOOP)
  @KBD
  D=M

  @FILL
  D;JNE

  @CLEAR
  0;JEQ
  
(FILL)
  @current
  A=M
  M=-1

  // Are we at max
  // If so, restart loop
  @current
  D=M
  @max
  D=D-M
  @LOOP
  D;JEQ

  // Otherwise, increment @current
  @current
  M=M+1

  @LOOP
  0;JEQ

(CLEAR)
  @current
  A=M
  M=0

  // Are we at address of SCREEN
  // If so, restart loop
  @current
  D=M
  @SCREEN
  D=D-A
  @LOOP
  D;JEQ

  // Otherwise, decrement @current
  @current
  M=M-1

  @LOOP
  0;JEQ