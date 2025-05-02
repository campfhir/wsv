package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/campfhir/wsv/document"
	"github.com/campfhir/wsv/internal"
	"github.com/campfhir/wsv/reader"
	"github.com/campfhir/wsv/xpaths"
)

func main() {
	var (
		input       string
		output      string
		inputFile   *os.File
		outputFile  *os.File
		verify      bool
		tabular     bool
		sorting     string
		showVersion bool
	)
	flag.StringVar(&input, "input", "-", "input file, use `-` for stdin (default stdin)")
	flag.StringVar(&input, "i", "-", "input file, use `-` for stdin (default stdin)")
	flag.StringVar(&input, "file", "-", "input file, use `-` for stdin (default stdin)")
	flag.StringVar(&input, "f", "-", "input file, use `-` for stdin (default stdin)")
	flag.StringVar(&output, "output", "-", "output file, use `-` for stdout (default stdout)")
	flag.StringVar(&output, "o", "-", "output file, use `-` for stdout (default stdout)")
	flag.StringVar(&sorting, "sort", "", "sort by column(s) seperated by `;` will be sorted in the order provided, can use `::` modifier followed by asc or desc to specify direction (defaults asc)")
	flag.StringVar(&sorting, "s", "", "sort by column(s) seperated by `;` will be sorted in the order provided, can use `::` modifier followed by asc or desc to specify direction (defaults asc)")
	flag.BoolVar(&tabular, "tabular", true, "specify if a document is tabular or not")
	flag.BoolVar(&verify, "verify", false, "verify that input is valid wsv")
	flag.BoolVar(&verify, "v", false, "verify that input is valid wsv")
	flag.BoolVar(&showVersion, "version", false, "print the version")
	flag.Parse()

	if showVersion {
		os.Stdout.WriteString(version())
		os.Exit(0)
		return
	}

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
		defer inputFile.Close()
	} else {
		inputFile = os.Stdin
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
			return
		}
		defer outputFile.Close()
	} else {
		outputFile = os.Stdout
	}

	r := reader.NewReader(inputFile)
	r.IsTabular = tabular
	if r.IsTabular {
		r.NullTrailingColumns = false
	}

	if verify {
		_, err := r.ReadAll()
		if err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(1)
			return
		}
		return
	}

	doc, err := r.ToDocument()
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("unable to convert to WSV document due to %s\n", err))
		os.Exit(2)
		return
	}
	if sorting != "" {
		columnsModifiers := strings.Split(sorting, ";")
		sortingModifier := internal.Map(columnsModifiers, func(e string, _ int, _ []string) *internal.SortOption {
			if e == "" {
				return nil
			}
			var (
				column              string
				typeModifier        string
				formatModifier      string
				orderModifier       string
				foundFormatModifier bool
				foundTypeModifier   bool
			)
			column, typeModifier, foundTypeModifier = strings.Cut(e, "||")
			if foundTypeModifier {
				typeModifier, formatModifier, foundFormatModifier = strings.Cut(typeModifier, "|")
				if foundFormatModifier {
					formatModifier, orderModifier, _ = strings.Cut(formatModifier, "::")
				} else {
					typeModifier, orderModifier, _ = strings.Cut(typeModifier, "::")
				}
			} else {
				column, orderModifier, _ = strings.Cut(column, "::")
			}
			if typeModifier == "number" {
				r := 10
				if formatModifier != "" {
					r64, err := strconv.ParseInt(formatModifier, 10, internal.PtrSize())
					if err != nil {
						os.Stderr.WriteString(fmt.Sprintf("could not parse the value [%s] into a base number for the column [%s]", formatModifier, column))
						os.Exit(1)
						return nil
					}
					r = int(r64)
				}
				if orderModifier == "desc" {
					return document.SortNumberBaseDesc(column, r)
				} else if orderModifier == "" || orderModifier == "asc" {
					return document.SortNumberBase(column, r)
				}
				os.Stderr.WriteString(fmt.Sprintf("the modifier for order [%s] on the number column [%s] is invalid", orderModifier, column))
				os.Exit(1)
				return nil
			}
			if typeModifier == "date" {
				format := time.DateOnly
				if formatModifier != "" {
					s := time.Now().Format(formatModifier)
					if s == "" {
						os.Stderr.WriteString(fmt.Sprintf("the date format [%s] could not parse date into something meaningful for the column [%s]", formatModifier, column))
						os.Exit(1)
						return nil
					}
					format = formatModifier
				}

				if orderModifier == "desc" {
					return document.SortTimeDesc(column, format)
				} else if orderModifier == "" || orderModifier == "asc" {
					return document.SortTime(column, format)
				}
				os.Stderr.WriteString(fmt.Sprintf("the modifier for order [%s] on the date column [%s] is invalid", orderModifier, column))
				os.Exit(1)
				return nil
			}

			if typeModifier != "" && typeModifier != "string" {
				os.Stderr.WriteString(fmt.Sprintf("the type modifier [%s] for the column [%s] is not valid, can only be date, number or string", typeModifier, column))
				os.Exit(1)
				return nil
			}

			if orderModifier == "desc" {
				return document.SortDesc(column)
			} else if orderModifier == "" || orderModifier == "asc" {
				return document.Sort(column)
			}
			os.Stderr.WriteString(fmt.Sprintf("the modifier [%s] for the string column [%s] is not valid, can only be asc or desc", orderModifier, column))
			os.Exit(1)
			return nil
		})

		doc.SortBy(sortingModifier...)
	}
	err = doc.WriteAllTo(outputFile)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("unable to write document to %s due to %s\n", outputFile.Name(), err))
		os.Exit(3)
	}

	return

}
