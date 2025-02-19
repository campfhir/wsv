package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/campfhir/wsv/reader"
	"github.com/campfhir/wsv/xpaths"
)

func main() {
	var (
		input      string
		output     string
		inputFile  *os.File
		outputFile *os.File
	)
	flag.StringVar(&input, "input", "-", "if omitted or `-` is specified the stdin will be used")
	flag.StringVar(&input, "i", "-", "if omitted or `-` is specified the stdin will be used")
	flag.StringVar(&output, "output", "-", "if omitted or `-` is specified the stdout will be used")
	flag.StringVar(&output, "o", "-", "if omitted or `-` is specified the stdout will be used")
	flag.Parse()
	if input != "-" {
		inputPath, err := xpaths.Resolve(input)
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("unable to resolve the input path %s due to %s\n", input, err))
			os.Exit(1)
			return
		}
		inputFile, err = os.Open(inputPath)
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("unable to open the input file %s due to %s\n", inputPath, err))
			os.Exit(1)
			return
		}
	}

	if output != "-" {
		outputPath, err := xpaths.Resolve(output)
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("unable to resolve the output path %s due to %s\n", output, err))
			os.Exit(1)
			return
		}
		outputFile, err = os.Create(outputPath)
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("unable to open the input file %s due to %s\n", outputPath, err))
			os.Exit(1)
		}
	} else {
		outputFile = os.Stdout
	}

	r := reader.NewReader(inputFile)
	doc, err := r.ToDocument()
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("unable to convert to WSV document due to %s\n", err))
		os.Exit(2)
		return
	}

	err = doc.WriteAllTo(outputFile)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("unable to write document to %s due to %s\n", outputFile.Name(), err))
		os.Exit(3)
	}

	return

}
