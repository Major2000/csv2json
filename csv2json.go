package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
		reader.Comma = ';'
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
		} else if err != nil {
			exitGracefully(err) // unexpected error
		}
		// process a CSV line
		record, err := processLine(headers, line)


		if err != nil {
			// error here means we got wrong number of columns
			fmt.Printf("Line: %sError: %s\n", line, err)
			continue
		}

		// otherwise, we send the processed record to writer channel
		writerChannel <- record
	}
}


func exitGracefully(err error)  {
	fmt.Fprint(os.Stderr, "error: %v\n", err)
	os.Exit(1)	
}


func check(e error)  {
	if e != nil {
		exitGracefully(e)
	}
	
}



func processLine(headers []string, dataList []string) (map[string]string, error)  {
	// check if same number of headers and columns
	if len(dataList) != len(headers) {
		return nil, errors.New("line doesn't match headers format. Skipping")
	}
	// creating the map to populate
	recordMap := make(map[string]string)
	
	// set new map key for each header
	for i, name := range headers {
		recordMap[name] = dataList[i]
	}

	// returning generated map
	return recordMap, nil
}



func writeJSONFile(csvPath string, writerChannel <-chan map[string]string, done chan<- bool, pretty bool) {
	writeString := createStringWriter(csvPath) // Instanciating a JSON writer function
	jsonFunc, breakLine := getJSONFunc(pretty) // Instanciating the JSON parse function and the breakline character
	 // Log for informing
	fmt.Println("Writing JSON file...")
	// Writing the first character of our JSON file. We always start with a "[" since we always generate array of record
	writeString("["+breakLine, false) 
	first := true
	for {
		// Waiting for pushed records into our writerChannel
		record, more := <-writerChannel
		if more {
			if !first { // If it's not the first record, we break the line
				writeString(","+breakLine, false)
			} else {
				first = false // If it's the first one, we don't break the line
			}

			jsonData := jsonFunc(record) // Parsing the record into JSON
			writeString(jsonData, false) // Writing the JSON string with our writer function
		} else { // If we get here, it means there aren't more record to parse. So we need to close the file
			writeString(breakLine+"]", true) // Writing the final character and closing the file
			fmt.Println("Completed!") // Logging that we're done
			done <- true // Sending the signal to the main function so it can correctly exit out.
			break // Stoping the for-loop
		}
	}
}



func createStringWriter(csvPath string) func(string, bool) {
	jsonDir := filepath.Dir(csvPath) // Getting the directory where the CSV file is
	jsonName := fmt.Sprintf("%s.json", strings.TrimSuffix(filepath.Base(csvPath), ".csv")) // Declaring the JSON filename, using the CSV file name as base
	finalLocation := filepath.Join(jsonDir, jsonName) // Declaring the JSON file location, using the previous variables as base
	// Opening the JSON file that we want to start writing
	f, err := os.Create(finalLocation) 
	check(err)
	// This is the function we want to return, we're going to use it to write the JSON file
	return func(data string, close bool) { // 2 arguments: The piece of text we want to write, and whether or not we should close the file
		_, err := f.WriteString(data) // Writing the data string into the file
		check(err)
		// If close is "true", it means there are no more data left to be written, so we close the file
		if close { 
			f.Close()
		}
	}
}


func getJSONFunc(pretty bool) (func(map[string]string) string, string) {
	// Declaring the variables we're going to return at the end
	var jsonFunc func(map[string]string) string
	var breakLine string
	if pretty { //Pretty is enabled, so we should return a well-formatted JSON file (multi-line)
		breakLine = "\n"
		jsonFunc = func(record map[string]string) string {
			jsonData, _ := json.MarshalIndent(record, "   ", "   ") // By doing this we're ensuring the JSON generated is indented and multi-line
			return "   " + string(jsonData) // Transforming from binary data to string and adding the indent characets to the front
		}
	} else { // Now pretty is disabled so we should return a compact JSON file (one single line)
		breakLine = "" // It's an empty string because we never break lines when adding a new JSON object
		jsonFunc = func(record map[string]string) string {
			jsonData, _ := json.Marshal(record) // Now we're using the standard Marshal function, which generates JSON without formating
			return string(jsonData) // Transforming from binary data to string
		}
	}

	return jsonFunc, breakLine // Returning everythinbg
}



// main function
func main() {
	// Showing useful information when the user enters the --help option
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] <csvFile>\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	// Getting the file data that was entered by the user
	fileData, err := getFileData()

	if err != nil {
		exitGracefully(err)
	}
	// Validating the file entered
	if _, err := checkIfValidFile(fileData.filepath); err != nil {
		exitGracefully(err)
	}
	// Declaring the channels that our go-routines are going to use
	writerChannel := make(chan map[string]string)
	done := make(chan bool) 
	// Running both of our go-routines, the first one responsible for reading and the second one for writing
	go processCsvFile(fileData, writerChannel) 
	go writeJSONFile(fileData.filepath, writerChannel, done, fileData.pretty)
	// Waiting for the done channel to receive a value, so that we can terminate the programn execution
	<-done 
}