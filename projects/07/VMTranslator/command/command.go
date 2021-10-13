package command

import (
	"fmt"
	"strings"
)

type Command interface {
	Type() CommandType
	String() string
}

type CommandType int

const (
	Add CommandType = iota
	Sub
	Neg
	Eq
	Gt
	Lt
	And
	Or
	Not
	Pop
	Push
	Label
	Goto
	IfGoto
)

func (t CommandType) String() string {
	names := []string{"add", "sub", "neg", "eq", "gt", "lt", "and", "or", "not", "pop", "push", "label", "goto", "if-goto"}
	if int(t) >= len(names) || int(t) < 0 {
		return ""
	}
	return names[t]
}

type RawCommand struct {
	Typ CommandType
}

func (c RawCommand) Type() CommandType {
	return c.Typ
}
func (c RawCommand) String() string {
	return c.Typ.String()
}

type Segment int

const (
	Local Segment = iota
	Argument
	This
	That
	Constant
	Static
	Pointer
	Temp
)

func (s Segment) String() string {
	names := []string{"local", "argument", "this", "that", "constant", "static", "pointer", "temp"}
	if int(s) >= len(names) || int(s) < 0 {
		return ""
	}
	return names[s]
}

func (s Segment) Label() string {
	switch s {
	case Local:
		return "LCL"
	case Argument:
		return "ARG"
	case This:
		return "THIS"
	case That:
		return "THAT"
	default:
		return ""
	}
}

type MemoryAccessCommand struct {
	RawCommand
	Segment Segment
	Index   int
}

func (m MemoryAccessCommand) String() string {
	return fmt.Sprintf("%s %s %d", m.Typ.String(), m.Segment.String(), m.Index)
}

func ToCommandType(s string) CommandType {
	names := []string{"add", "sub", "neg", "eq", "gt", "lt", "and", "or", "not", "pop", "push", "label", "goto", "if-goto"}
	for i := range names {
		if names[i] == s {
			return CommandType(i)
		}
	}

	return -1
}

func ToSegment(s string) Segment {
	segments := []string{"local", "argument", "this", "that", "constant", "static", "pointer", "temp"}
	for i := range segments {
		if segments[i] == s {
			return Segment(i)
		}
	}

	return -1
}

type BranchingCommand struct {
	RawCommand
	Label string
}

func (b *BranchingCommand) String() string {
	return fmt.Sprintf("%s %s", b.RawCommand.String(), strings.ToUpper(b.Label))
}
