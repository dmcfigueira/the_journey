package main

import (
	"fmt"
	"os"
	"time"
)

// Application configs
const OUTPUT_FILE = "results.csv"
const PARALLELIZE = true
const VERBOSE = 0

// Error codes
const (
	NO_ERROR = iota
	INVALID_NUM_ARGS_ERROR
	INPUT_FILE_OPEN_ERROR
	INPUT_FILE_SCANNER_ERROR
	OUTPUT_FILE_WRITE_ERROR
	INVALID_CELL_REFERENCE_ERROR
	INVALID_EXPRESSION_ERROR
)

var goroutinesLaunched int

func main() {

	startTime := time.Now()
	args := os.Args[1:]

	// check the number of arguments
	if len(args) < 1 {
		fmt.Println("usage:", os.Args[0], "<filename>")
		os.Exit(INVALID_NUM_ARGS_ERROR)
	}

	// read the input file contents to memory
	tokens := readTokensFromFile(args[0], '|')

	// create a new workload from the contents read
	wl := newWorkload(tokens)
	if VERBOSE >= 2 {
		wl.printDetails()
	}

	// process the workload sequentially or split it among multiple goroutines
	if PARALLELIZE {
		wl.parallelize()
	} else {
		wl.runSequencially()
	}

	// format results for display
	result := resultBoardToString(wl.result)
	if VERBOSE >= 1 {
		fmt.Println("ResultBoard:")
		fmt.Print(resultBoardToString(wl.result))
	}

	// write results to the output file
	err := os.WriteFile(OUTPUT_FILE, []byte(result), 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(OUTPUT_FILE_WRITE_ERROR)
	}

	// print execution details
	if VERBOSE >= 0 {
		fmt.Printf("Execution ended. Goroutines Launched: %v. Total time: %v.\n", goroutinesLaunched, time.Since(startTime))
		fmt.Printf("Results written to: %v.\n", OUTPUT_FILE)
	}
}
