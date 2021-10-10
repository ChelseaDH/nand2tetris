package main

import (
	"bufio"
	"log"
	"os"
	"path"
	"strings"

	"github.com/ChelseaDH/VMTranslator/parser"
	"github.com/ChelseaDH/VMTranslator/translator"
)

func main() {
	args := os.Args
	if len(args) != 2  {
		 log.Fatal("Incorrect number of command line arguments provided")
	}

	inputFilename := args[1]
	inputFile, err := os.Open(inputFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer inputFile.Close()
	inputFileRef := strings.Replace(path.Base(inputFilename), ".vm", "", 1)

	outputFile, err := os.OpenFile(strings.Replace(inputFilename, ".vm", ".asm", 1), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	scanner := bufio.NewScanner(inputFile)
	t := translator.Translator{
		Namespace: inputFileRef,
		Output:    outputFile,
	}
	for scanner.Scan() {
		command, err := parser.Parse(scanner.Text())
		if err != nil {
			log.Fatal(err)
		}

		if command == nil {
			continue
		}

		err = t.Translate(command)
		if err != nil {
			log.Fatal(err)
		}
	}

	t.Terminate()
}