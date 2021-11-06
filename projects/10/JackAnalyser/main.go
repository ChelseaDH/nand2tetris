package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/ChelseaDH/JackAnalyser/lexer"
	"github.com/ChelseaDH/JackAnalyser/parser"
)

const inputFileExt = ".jack"
const outputFileExt = ".xml"

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

	var filePaths []string

	switch mode := fileInfo.Mode(); {
	case mode.IsRegular():
		filePaths = append(filePaths, name)

	case mode.IsDir():
		files, err := os.ReadDir(name)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			if path.Ext(file.Name()) == inputFileExt {
				filePaths = append(filePaths, path.Join(name, file.Name()))
			}
		}

	default:
		log.Fatal(fmt.Sprintf("Second command line argument must be a %s file or directory containing one or more %s files", inputFileExt, inputFileExt))
	}

	for _, filePath := range filePaths {
		handleFile(filePath)
	}
}

func handleFile(filePath string) {
	inputFile, err := os.Open(filePath)
	if err != nil {
		log.Print(err)
		return
	}
	defer inputFile.Close()

	p := parser.NewParser(lexer.NewLexer(inputFile))
	class, err := p.Parse()
	if err != nil {
		log.Print(err)
		return
	}

	log.Printf("%#v", class)
}