package main

import (
	"bufio"
	"fmt"
	"os"
)

// SplitByDelimiter returns a customized split function instance. When used
// by a Scanner, the split function returned will advance the scanner to the
// next matching delimiter or end-of-line (whichever comes first) and set the
// value of 'found' to true if a delimiter was found (or false otherwise).
func SplitByDelimiter(del byte, found *bool) bufio.SplitFunc {

	// return a closure (which will capture the enclosing context)
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		*found = false

		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		for i, b := range data {

			if b == del { // delimeter
				*found = true
				return i + 1, data[0:i], nil
			}
			if b == '\n' { // end-of-line
				if i > 0 && data[i-1] == '\r' {
					return i + 1, data[0 : i-1], nil // drop CR
				}
				return i + 1, data[0:i], nil
			}
		}

		return 0, nil, nil
	}
}

// Read the contents of file and return a 2 dimensional slice where the first
// dimension matches the lines of the file and the second dimension contains
// the tokens obtained by splitting each line with the given delimiter.
func readTokensFromFile(file string, del byte) (res [][]string) {

	// open the file
	fp, err := os.Open(file)

	if err != nil {
		fmt.Println(err)
		os.Exit(INPUT_FILE_OPEN_ERROR)
	}

	// close the file upon function exit
	defer fp.Close()

	// split the file by delimiter
	var delWasFound bool
	scanner := bufio.NewScanner(fp)
	scanner.Split(SplitByDelimiter(del, &delWasFound))

	for line, newLine := -1, true; scanner.Scan(); {

		if newLine { // starting a new line
			res = append(res, []string{})
			line++
		}

		res[line] = append(res[line], scanner.Text())
		newLine = !delWasFound
	}

	// check for errors
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		os.Exit(INPUT_FILE_SCANNER_ERROR)
	}

	return
}
