package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/ChelseaDH/JackAnalyser/lexer"
	"github.com/ChelseaDH/JackAnalyser/token"
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

	outputFile, err := os.OpenFile(strings.Replace(filePath, inputFileExt, outputFileExt, 1), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Print(err)
		return
	}
	defer outputFile.Close()

	outputFile.WriteString("<tokens>\n")

	l := lexer.NewLexer(inputFile)
	for {
		tok, _, err := l.Next()
		if err != nil {
			log.Print(err)
			break
		}

		if tok != token.End {
			break
		}

		outputFile.WriteString(fmt.Sprintf("%s\n", tok.String()))
	}
	outputFile.WriteString("</tokens>\n")
}
