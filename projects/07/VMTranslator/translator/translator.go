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
	Namespace string
	Output    io.Writer
	jumpCount int
}

func (t *Translator) Translate(c command.Command) error {
	t.write(fmt.Sprintf("// %s\n", c.String()))

	switch c.Type() {
	case command.Add:
		return t.translateBinaryExpression("+", "")

	case command.Sub:
		return t.translateBinaryExpression("-", "")

	case command.Eq:
		return t.translateBinaryExpression("-", "JEQ")

	case command.Gt:
		return t.translateBinaryExpression("-", "JGT")

	case command.Lt:
		return t.translateBinaryExpression("-", "JLT")

	case command.And:
		return t.translateBinaryExpression("&", "")

	case command.Or:
		return t.translateBinaryExpression("|", "")

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

func (t *Translator) translateBinaryExpression(operator string, jump string) error {
	err := t.translatePop(&command.MemoryAccessCommand{
		RawCommand: command.RawCommand{Typ: command.Pop},
		Segment:    command.Temp,
		Index:      0,
	})
	if err != nil {
		return err
	}

	t.write(fmt.Sprintf("@SP\nA=M-1\nD=M\n@%d\nD=D%sM\n", tempIndex, operator))
	if jump != "" {
		t.jump(jump)
	}
	t.write("@SP\nA=M-1\nM=D\n")
	return nil
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

	t.decrementSP()
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
	t.write("@SP\nA=M\nM=D\n")
	t.incrementSP()
	return nil
}

func (t *Translator) incrementSP() {
	t.write("@SP\nM=M+1\n")
}

func (t *Translator) decrementSP() {
	t.write("@SP\nAM=M-1\nD=M\n")
}

func (t *Translator) jump(jumpType string) {
	t.write(fmt.Sprintf("@TRUE%d\nD;%s\nD=0\n@FALSE%d\n0;JMP\n(TRUE%d)\nD=-1\n(FALSE%d)\n", t.jumpCount, jumpType, t.jumpCount, t.jumpCount, t.jumpCount))
	t.jumpCount++
}

func (t *Translator) addLabel(bc *command.BranchingCommand) {
	t.write(fmt.Sprintf("(%s)\n", strings.ToUpper(bc.Label)))
}

func (t *Translator) unconditionalGoto(bc *command.BranchingCommand) {
	t.write(fmt.Sprintf("@%s\n0;JMP\n", strings.ToUpper(bc.Label)))
}

func (t *Translator) conditionalGoto(bc *command.BranchingCommand) {
	t.write(fmt.Sprintf("@SP\nAM=M-1\nD=M\n@%s\nD;JNE\n", strings.ToUpper(bc.Label)))
}
