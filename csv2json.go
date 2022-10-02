package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type inputFile struct {
	filepath  string // path to csv file
	separator string // the separator used in the file
	pretty    bool   // whether or not the generated JSON is well-formatted
}





func getFileData() (inputFile, error) {
	// validate that we're getting the correct number of arguments

	if len(os.Args) < 2 {
		return inputFile{}, errors.New("a file path argument is required")
	}

	// separator and pretty variables
	separator := flag.String("separator", "comma", "column separator")
	pretty := flag.Bool("pretty", false, "Generate pretty JSON")

	flag.Parse() // this will parse all arguments from the terminal

	fileLocation := flag.Arg(0) // The only argument (that is not a flag option) is the file location (CSV file)

	// validating whether on not the "comma" or "semicolon" is received
	// if not return error
	if !(*separator == "comma" || *separator == "semicolon") {
		return inputFile{}, errors.New("only comma or semicolon separators are allowed")
	}

	// at this point the arguments are validated
	// return the corresponding struct instance with all required data
	return inputFile{fileLocation, *separator, *pretty}, nil
}



func checkIfValidFile(filename string) (bool, error) {
	// Checking if the entered file is a csv
	if fileExtension := filepath.Ext(filename); fileExtension != ".csv" {
		return false, fmt.Errorf("file %s is not CSV", filename)
	}

	// checking if file path entered belongs to existing file
	if _, err := os.Stat(filename); err != nil && os.IsNotExist(err) {
		return false, fmt.Errorf("file %s does not exist", filename)
	}

	return true, nil
}





func processCsvFile(fileData inputFile, writerChannel chan<- map[string]string) {

	// open file for reading
	file, err := os.Open(fileData.filepath)

	//check for errors
	check(err)

	// close the file 
	defer file.Close()

	// defining a "header" and "line" slice
	var headers, line []string

	// initialize CSV reader
	reader := csv.NewReader(file)

	// change default separator from "," to ";"
	if fileData.separator == "semicolon" {
		read.comma = ';'
	}

	// reading the first line
	headers, err = reader.Read()
	check(err) // check error again

	// iterate over each line in CSV
	for {
		
		line, err = reader.Read()

		// if reached the end of the file
		if err == io.EOF {
			close(writerChannel)
			break
		}
	}
}




// main function
func main() {
	// fileData, err := getFileData()
}
