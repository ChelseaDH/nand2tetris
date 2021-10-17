package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/ChelseaDH/VMTranslator/parser"
	"github.com/ChelseaDH/VMTranslator/translator"
)

func main() {
	args := os.Args
	if len(args) != 2 {
		log.Fatal("Incorrect number of command line arguments provided")
	}

	name := args[1]
	fileInfo, err := os.Stat(name)
	if err != nil {
		log.Fatal(err)
	}

	switch mode := fileInfo.Mode(); {
	case mode.IsRegular():
		outputFile, err := os.OpenFile(strings.Replace(name, ".vm", ".asm", 1), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer outputFile.Close()

		namespace := strings.Replace(path.Base(name), ".vm", "", 1)
		t := translator.Translator{
			Namespace: namespace,
			Output:    outputFile,
		}

		translateFile(name, t)
		t.Terminate()

	case mode.IsDir():
		outputFile, err := os.OpenFile(path.Join(name, fmt.Sprintf("%s.asm", path.Base(name))), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer outputFile.Close()

		files, err := os.ReadDir(name)
		if err != nil {
			log.Fatal(err)
		}

		t := translator.Translator{
			Output: outputFile,
		}
		err = t.Initialise()
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			if path.Ext(file.Name()) == ".vm" {
				t.Namespace = strings.Replace(file.Name(), ".vm", "", 1)
				translateFile(path.Join(name, file.Name()), t)
			}
		}
		t.Terminate()

	default:
		log.Fatal("Second command line argument must be a .vm file or directory containing one or more .vm files")
	}
}

func translateFile(path string, translator translator.Translator) {
	inputFile, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer inputFile.Close()

	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		command, err := parser.Parse(scanner.Text())
		if err != nil {
			log.Fatal(err)
		}

		if command == nil {
			continue
		}

		err = translator.Translate(command)
		if err != nil {
			log.Fatal(err)
		}
	}
}
