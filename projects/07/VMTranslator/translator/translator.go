package translator

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/ChelseaDH/VMTranslator/command"
)

const tempIndex = 5

type Translator struct {
	Output      io.Writer
	Namespace   string
	currentFunc string
	jumpCount   int
	returnCount int
}

func (t *Translator) Translate(c command.Command) error {
	t.write(fmt.Sprintf("// %s\n", c.String()))

	switch c.Type() {
	case command.Add:
		t.translateBinaryExpression("+", "")
		return nil

	case command.Sub:
		t.translateBinaryExpression("-", "")
		return nil

	case command.Eq:
		t.translateBinaryExpression("-", "JEQ")
		return nil

	case command.Gt:
		t.translateBinaryExpression("-", "JGT")
		return nil

	case command.Lt:
		t.translateBinaryExpression("-", "JLT")
		return nil

	case command.And:
		t.translateBinaryExpression("&", "")
		return nil

	case command.Or:
		t.translateBinaryExpression("|", "")
		return nil

	case command.Neg:
		t.translateUnaryExpression("-")
		return nil

	case command.Not:
		t.translateUnaryExpression("!")
		return nil

	case command.Pop:
		mac := c.(*command.MemoryAccessCommand)
		return t.translatePop(mac)

	case command.Push:
		mac := c.(*command.MemoryAccessCommand)
		return t.translatePush(mac)

	case command.Label:
		bc := c.(*command.BranchingCommand)
		t.addLabel(bc)
		return nil

	case command.Goto:
		bc := c.(*command.BranchingCommand)
		t.unconditionalGoto(bc)
		return nil

	case command.IfGoto:
		bc := c.(*command.BranchingCommand)
		t.conditionalGoto(bc)
		return nil

	case command.Function:
		fc := c.(*command.FunctionCommand)
		t.currentFunc = fc.Name
		t.defineFunction(fc)
		return nil

	case command.Call:
		fc := c.(*command.FunctionCommand)
		t.callFunction(fc)
		return nil

	case command.Return:
		t.translateReturn()
		return nil

	default:
		return fmt.Errorf("translation not yet implemented for command of type %s", c.Type())
	}
}

func (t *Translator) write(input string) {
	_, err := fmt.Fprintf(t.Output, "%s", input)
	if err != nil {
		log.Fatal(err)
	}
}

func (t *Translator) Terminate() {
	t.write("(END)\n@END\n0;JMP\n")
}

func (t *Translator) Initialise() error {
	t.write("// Initialising stack pointer\n@256\nD=A\n@SP\nM=D\n")
	return t.Translate(&command.FunctionCommand{
		RawCommand: command.RawCommand{Typ: command.Call},
		Name:       "Sys.init",
		Args:       0,
	})
}

func (t *Translator) translateBinaryExpression(operator string, jump string) {
	t.popStackIntoD()
	t.write("@temp\nM=D\n")

	t.write(fmt.Sprintf("@SP\nA=M-1\nD=M\n@temp\nD=D%sM\n", operator))
	if jump != "" {
		t.jump(jump)
	}
	t.write("@SP\nA=M-1\nM=D\n")
}

func (t *Translator) translateUnaryExpression(operator string) {
	t.write(fmt.Sprintf("@SP\nA=M-1\nM=%sM\n", operator))
}

func (t *Translator) translatePop(c *command.MemoryAccessCommand) error {
	var loc string

	switch c.Segment {
	case command.Local, command.Argument, command.This, command.That:
		t.write(fmt.Sprintf("@%d\nD=A\n@%s\nD=D+M\n@%d\nM=D\n", c.Index, c.Segment.Label(), tempIndex))
		loc = fmt.Sprintf("%d\nA=M", tempIndex)
		break

	case command.Static:
		loc = fmt.Sprintf("%s.%d", t.Namespace, c.Index)
		break

	case command.Temp:
		loc = fmt.Sprintf("%d", tempIndex+c.Index)
		break

	case command.Pointer:
		if c.Index == 0 {
			loc = "THIS"
		} else {
			loc = "THAT"
		}
		break

	default:
		return fmt.Errorf("%s is not a valid segment type for pop", c.Segment.String())
	}

	t.popStackIntoD()
	t.write(fmt.Sprintf("@%s\nM=D\n", loc))
	return nil
}

func (t *Translator) translatePush(c *command.MemoryAccessCommand) error {
	var loc, d string

	switch c.Segment {
	case command.Local, command.Argument, command.This, command.That:
		loc = fmt.Sprintf("%d\nD=A\n@%s\nA=M+D", c.Index, c.Segment.Label())
		d = "M"
		break

	case command.Constant:
		loc = fmt.Sprintf("%d", c.Index)
		d = "A"
		break

	case command.Static:
		loc = fmt.Sprintf("%s.%d", t.Namespace, c.Index)
		d = "M"
		break

	case command.Temp:
		loc = fmt.Sprintf("%d", tempIndex+c.Index)
		d = "M"
		break

	case command.Pointer:
		if c.Index == 0 {
			loc = "THIS"
		} else {
			loc = "THAT"
		}

		d = "M"
		break

	default:
		return fmt.Errorf("%s is not a valid segment type for push", c.Segment.String())
	}

	t.write(fmt.Sprintf("@%s\nD=%s\n", loc, d))
	t.pushDOntoStack()
	return nil
}

func (t *Translator) incrementSP() {
	t.write("@SP\nM=M+1\n")
}

func (t *Translator) jump(jumpType string) {
	t.write(fmt.Sprintf("@TRUE%d\nD;%s\nD=0\n@FALSE%d\n0;JMP\n(TRUE%d)\nD=-1\n(FALSE%d)\n", t.jumpCount, jumpType, t.jumpCount, t.jumpCount, t.jumpCount))
	t.jumpCount++
}

func (t *Translator) addLabel(bc *command.BranchingCommand) {
	t.write(fmt.Sprintf("(%s$%s)\n", t.currentFunc, strings.ToUpper(bc.Label)))
}

func (t *Translator) unconditionalGoto(bc *command.BranchingCommand) {
	t.write(fmt.Sprintf("@%s$%s\n0;JMP\n", t.currentFunc, strings.ToUpper(bc.Label)))
}

func (t *Translator) conditionalGoto(bc *command.BranchingCommand) {
	t.popStackIntoD()
	t.write(fmt.Sprintf("@%s$%s\nD;JNE\n", t.currentFunc, strings.ToUpper(bc.Label)))
}

func (t *Translator) defineFunction(fc *command.FunctionCommand) {
	// Function label
	t.write(fmt.Sprintf("(%s)\n", fc.Name))
	// Initialise local variables to 0
	for i := 0; i < fc.Args; i++ {
		t.write("@SP\nA=M\nM=0\n")
		t.incrementSP()
	}
}

func (t *Translator) callFunction(fc *command.FunctionCommand) {
	returnLabel := fmt.Sprintf("%s$ret.%d", t.Namespace, t.returnCount)

	// Push return address of caller to stack
	t.write(fmt.Sprintf("@%s\nD=A\n", returnLabel))
	t.pushDOntoStack()
	// Save state of caller
	t.saveCallerSegments()
	// ARG = SP - 5 - fc.Args && LCL = SP
	t.write(fmt.Sprintf("@SP\nD=M\n@%d\nD=D-A\n@%d\nD=D-A\n@%s\nM=D\n", 5, fc.Args, command.Argument.Label()))
	// LCL = AP
	t.write(fmt.Sprintf("@SP\nD=M\n@%s\nM=D\n", command.Local.Label()))
	// Jump to target function
	t.write(fmt.Sprintf("@%s\n0;JMP\n", fc.Name))
	// Write return address label
	t.write(fmt.Sprintf("(%s)\n", returnLabel))
	t.returnCount = t.returnCount + 1
}

func (t *Translator) saveSingleSegment(segment command.Segment) {
	t.write(fmt.Sprintf("@%s\nD=M\n", segment.Label()))
	t.pushDOntoStack()
}

func (t *Translator) saveCallerSegments() {
	t.saveSingleSegment(command.Local)
	t.saveSingleSegment(command.Argument)
	t.saveSingleSegment(command.This)
	t.saveSingleSegment(command.That)
}

func (t *Translator) translateReturn() {
	// Set temp endFrame var
	t.write(fmt.Sprintf("@%s\nD=M\n@%s\nM=D\n", command.Local.Label(), "endFrame"))
	// Get return address of caller
	t.write(fmt.Sprintf("@%d\nA=D-A\nD=M\n@%s\nM=D\n", 5, "retAddr"))
	// *ARG = pop()
	t.popStackIntoD()
	t.write(fmt.Sprintf("@%s\nA=M\nM=D\n", command.Argument.Label()))
	// SP = ARG + 1
	t.write(fmt.Sprintf("@%s\nD=M+1\n@SP\nM=D\n", command.Argument.Label()))
	// Restore state of caller
	t.restoreCallerSegments()
	// Goto return address
	t.write(fmt.Sprintf("@%s\nA=M\n0;JMP\n", "retAddr"))
}

func (t *Translator) restoreSingleSegment(segment command.Segment) {
	t.write(fmt.Sprintf("@%s\nAM=M-1\nD=M\n@%s\nM=D\n", "endFrame", segment.Label()))
}

func (t *Translator) restoreCallerSegments() {
	t.restoreSingleSegment(command.That)
	t.restoreSingleSegment(command.This)
	t.restoreSingleSegment(command.Argument)
	t.restoreSingleSegment(command.Local)
}

func (t *Translator) popStackIntoD() {
	t.write("@SP\nAM=M-1\nD=M\n")
}

func (t *Translator) pushDOntoStack() {
	t.write("@SP\nA=M\nM=D\n")
	t.incrementSP()
}
