// push constant 1
@1
D=A
@SP
A=M
M=D
@SP
M=M+1
// push local 2
@2
D=A
@LCL
A=M+D
D=M
@SP
A=M
M=D
@SP
M=M+1
// and
@SP
AM=M-1
D=M
@5
M=D
@SP
A=M-1
D=M
@5
D=D&M
@SP
A=M-1
M=D
(END)
@END
0;JMP
