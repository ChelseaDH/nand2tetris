package parser

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/ChelseaDH/VMTranslator/command"
)

func Parse(line string) (command.Command, error) {
	comment, err := regexp.Compile(`//.*`)
	if err != nil {
		log.Fatal(err)
	}

	parts := strings.Split(strings.TrimSpace(comment.ReplaceAllString(line, "")), " ")
	commandName := parts[0]

	switch commandName {
	case "":
		return nil, nil

	case "push", "pop":
		if len(parts) != 3 {
			return nil, fmt.Errorf("%s command must have 2 arguments", commandName)
		}

		rawCommand := command.RawCommand{Typ: command.ToCommandType(commandName)}

		segment := command.ToSegment(parts[1])
		if segment == -1 {
			return nil, fmt.Errorf("invalid value passed to segment argument of %s command: %s", commandName, parts[1])
		}

		index, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("the index argument of a %s command must be an integer", commandName)
		}

		err = checkArgumentValidity(segment, index, commandName)
		if err != nil {
			return nil, err
		}

		return &command.MemoryAccessCommand{
			RawCommand: rawCommand,
			Segment:    segment,
			Index:      index,
		}, nil

	case "label", "goto", "if-goto":
		if len(parts) != 2 {
			return nil, fmt.Errorf("%s command must have 1 arguments", commandName)
		}

		rawCommand := command.RawCommand{Typ: command.ToCommandType(commandName)}

		return &command.BranchingCommand{
			RawCommand: rawCommand,
			Label:      parts[1],
		}, nil

	default:
		commandType := command.ToCommandType(commandName)
		if commandType == -1 {
			return nil, fmt.Errorf("invalid command type provided: %s", commandName)
		}

		return &command.RawCommand{Typ: commandType}, nil
	}
}

func checkArgumentValidity(segment command.Segment, index int, commandName string) error {
	if index < 0 {
		return fmt.Errorf("the index argument of a %s command must be greater than or equal to 0", commandName)
	}

	if segment == command.Temp && index > 7 {
		return fmt.Errorf("index %d is out of bounds for the temp memory segment", index)
	}

	if segment == command.Pointer && (index < 0 || index > 1) {
		return fmt.Errorf("only values 0 and 1 are valid for the index of a pointer command, %d provided", index)
	}

	return nil
}
