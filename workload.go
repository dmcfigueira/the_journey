package main

import (
	"fmt"
	"os"
	"sync"
)

const PROCESSED = -1
const MISSING = -2

// A unit unit of work yet to be processed. Might contain dependencies to other cells.
type Cell struct {
	row, col     int
	expr         string
	dependencies []CellRef
}

// Create a new cell and compute its dependencies.
func newCell(row, col int, expr string, wl *Workload) (cell Cell) {
	cell = Cell{row, col, expr, nil}
	computeDependencies(&cell, wl)
	return cell
}

// Convert a cell into a formated string (ex: A1).
func (cell Cell) String() string {
	return fmt.Sprintf("%s%d", string('A'+cell.col), cell.row+1)
}

// Add a new dependency to a cell.
func (cell *Cell) addDependency(newRef CellRef) {
	for _, ref := range cell.dependencies {
		if newRef == ref {
			return // already present
		}
	}
	cell.dependencies = append(cell.dependencies, newRef)
}

// Print all the cell dependencies (for debbuging purposes).
func (cell *Cell) printDependencies() {
	fmt.Printf("%v: ", cell)
	for _, ref := range cell.dependencies {
		fmt.Printf("%s ", ref)
	}
	fmt.Println()
}

// A cell reference.
type CellRef struct {
	row, col int
}

// Create a new cell reference.
func newCellRef(row, col int) CellRef {
	return CellRef{row, col}
}

// Create a new cell reference from a formatted string (ex. A1).
func newCellRefFromString(s string) CellRef {
	var col rune
	var row int
	_, err := fmt.Sscanf(s, "%c%d", &col, &row)
	if err != nil {
		fmt.Println(err)
		os.Exit(INVALID_EXPRESSION_ERROR)
	}
	return CellRef{row - 1, int(col - 'A')}
}

// Convert a cell reference into a formated string (ex: A1).
func (ref CellRef) String() string {
	return fmt.Sprintf("%s%d", string('A'+ref.col), ref.row+1)
}

// A group of cells to be processed in sequence. Upon being picked up
// for processing the cells in this bucket should not be dependant on
// cells from other buckets - only on cells already computed or higher
// prio cells in this same bucket. To facilitate the execution the cells
// present in this bucket are kept in the order that they should be
// processed - such that no cell will have dependencies to the cells
// that come after.
type Bucket = []Cell

// Create a new bucket with a single element.
func newBucket(cell Cell) (bucket Bucket) {
	return Bucket{cell}
}

// Structure used to keep track of the values already computed.
type ResultBoard = [][]string

// Create an empty results board.
func newResultBoard(rows, cols int) (board ResultBoard) {
	board = make([][]string, rows)
	for row := range board {
		board[row] = make([]string, cols)
	}
	return
}

// Convert a result board into a formated string.
// Note: overwriting the String() method is not an option here as
// ResultBoard is just an alias for [][]string.
func resultBoardToString(board ResultBoard) (s string) {

	// compute padding for each column
	pads := make([]int, len(board[0]))
	for _, row := range board {
		for x, cell := range row {
			if len(cell) > pads[x] {
				pads[x] = len(cell)
			}
		}
	}

	// build horizontal legend
	s += "\t"
	for x, pad := range pads {
		s += fmt.Sprintf("%*s%*s", pad/2+(pad%2), string('A'+x), pad/2+2, "")
		if pad >= 1 {
			s += " "
		}
	}
	s += "\n"

	// build formatted grid
	for y, row := range board {
		s += fmt.Sprintf("%02d\t", y+1)
		for x, cell := range row {
			s += fmt.Sprintf("%*s", -pads[x], cell)
			if x < len(row)-1 {
				s += " | "
			}
		}
		s += "\n"
	}

	return
}

// Structure used to keep track of the status of each cell. Positive values
// indicate the index of the bucket in which the cell is currently waiting to
// be processed. Negative values will mean either that the cell was already
// processed (PROCESSED) or that the cell was not yet added to the workload (MISSING).
type RefBoard = [][]int

// Create a reference board with all values initialized to MISSING.
func newRefBoard(rows, cols int) (board RefBoard) {
	board = make([][]int, rows)
	for row := range board {
		board[row] = make([]int, cols)
		for col := range board[row] {
			board[row][col] = MISSING
		}
	}
	return
}

// Convert a reference board into a formated string.
// Note: overwriting the String() method is not an option here as
// RefBoard is just an alias for [][]int.
func refBoardToString(board RefBoard) (s string) {
	s += "RefBoard:\n"

	if len(board) == 0 {
		return
	}

	// build horizontal legend
	s += fmt.Sprintf("%*s", 6, "")
	for col := range board[0] {
		s += fmt.Sprintf("%2s ", string('A'+col))
	}
	s += "\n"

	// build formatted board
	for y, row := range board {
		s += fmt.Sprintf("%02d  [ ", y+1)
		for _, ref := range row {
			s += fmt.Sprintf("%2d ", ref)
		}
		s += "]\n"
	}
	return
}

// Contains all the work yet to be processed by the application
// (split into buckets) aswell as the results computed so far.
type Workload struct {
	buckets []Bucket
	result  ResultBoard
	ref     RefBoard
}

// Create a new workload from a 2 dimensional slice of tokens
func newWorkload(tokens [][]string) (wl *Workload) {

	// compute board dimensions
	rows, cols := calcBoardDimensions(tokens)

	// initialize a new workload
	wl = new(Workload)
	wl.result = newResultBoard(rows, cols)
	wl.ref = newRefBoard(rows, cols)

	// process static cells, compute dependencies,
	// assign dynamic cells to buckets
	wl.createWorkloadBuckets(tokens)

	return
}

// Calculate the dimensions that the result and reference boards will have
// based on the 2 dimensional slice of tokens received. Should the number
// of tokens per line differ, the maximum number of tokens per line found
// will be used as the board number of columns.
func calcBoardDimensions(tokens [][]string) (rows, cols int) {
	maxLen := 0
	for _, line := range tokens {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}
	return len(tokens), maxLen
}

// Compute all the work to be done from the tokens received and split the
// work into independent buckets according to their dependencies
func (wl *Workload) createWorkloadBuckets(tokens [][]string) {

	if VERBOSE >= 3 {
		fmt.Println("Dependencies:")
	}

	// process the tokens received
	for row, line := range tokens {
		for col, expr := range line {

			switch {
			case len(expr) == 0 || expr[0] != '=':
				// static cell - copy its value to the results board
				wl.result[row][col] = expr
				wl.ref[row][col] = PROCESSED
			default:
				// dynamic cell - put it in a new bucket
				wl.buckets = append(wl.buckets, newBucket(newCell(row, col, expr, wl)))
				wl.ref[row][col] = len(wl.buckets) - 1
			}
		}
	}

	if VERBOSE >= 3 {
		fmt.Println("Moves:")
	}

	// merge the buckets created according to their dependencies
	for i := range wl.buckets {

		// skip empty buckets
		if len(wl.buckets[i]) == 0 {
			continue
		}

		// move all childs of the current bucket onto himself
		wl.buckets[i] = wl.getDependencyTree(wl.buckets[i][0])

		// make list of all the buckets that are now deprecated
		oldBuckets := []int{}
		for _, cell := range wl.buckets[i][1:] {
			present := false
			for _, j := range oldBuckets {
				if j == wl.ref[cell.row][cell.col] {
					present = true
				}
			}
			if !(present) {
				oldBuckets = append(oldBuckets, wl.ref[cell.row][cell.col])
			}
		}

		// clear the deprecated buckets and update the reference board
		for _, j := range oldBuckets {
			// note: simply emptying the bucket is faster than actually removing it from the slice
			// since it would require several references in the reference board to be shifted, etc.
			for _, oldCell := range wl.buckets[j] {
				wl.ref[oldCell.row][oldCell.col] = i
			}
			if VERBOSE >= 3 {
				fmt.Println("- Moved", wl.buckets[j], "from bucket", j, "to", i)
			}
			wl.buckets[j] = nil
		}
	}

	// invert bucket order
	for i, bucket := range wl.buckets {
		temp := Bucket{}
		for j := len(bucket) - 1; j >= 0; j-- {
			temp = append(temp, bucket[j])
		}
		wl.buckets[i] = temp
	}
}

// Iterate through all the dependencies of a node (using a BFS search)
// and collect all the cells (including the root cell) into a single slice
// Note: does not detect circular dependencies
func (wl *Workload) getDependencyTree(cell Cell) (result []Cell) {
	result = []Cell{cell}
	for i := 0; i < len(result); i++ {
		for _, ref := range result[i].dependencies {
			ptr := wl.getCell(ref)
			if ptr != nil {
				result = append(result, *ptr)
			}
		}
	}
	return
}

// Retrieve a cell from one of the work buckets - O(1).
// Returns nil if the cell was already processed.
func (wl *Workload) getCell(ref CellRef) *Cell {

	wl.assertWithinBounds(ref.row, ref.col)

	idx := wl.ref[ref.row][ref.col]
	switch {
	case idx >= 0:
		for i, cell := range wl.buckets[idx] {
			if cell.row == ref.row && cell.col == ref.col {
				return &wl.buckets[idx][i]
			}
		}
	case idx == PROCESSED:
		return nil
	default:
		fmt.Printf("Error: Invalid cell reference (%d, %d).", ref.row, ref.col)
		os.Exit(INVALID_CELL_REFERENCE_ERROR)
	}
	return nil
}

// Retrieve a cell expression - O(1). The expression will be retrieved
// either from a cell in a bucket or from the results board.
func (wl *Workload) getCellExpr(row, col int) (expr string) {

	wl.assertWithinBounds(row, col)

	idx := wl.ref[row][col]
	switch {
	case idx >= 0:
		for _, cell := range wl.buckets[idx] {
			if cell.row == row && cell.col == col {
				expr = cell.expr
				break
			}
		}
	case idx == PROCESSED:
		expr = wl.result[row][col]
	default:
		fmt.Printf("Error: Invalid cell reference (%d, %d).", row, col)
		os.Exit(INVALID_CELL_REFERENCE_ERROR)
	}

	return
}

// Retrieve a cell value from the results board - O(1).
func (wl *Workload) getCellValue(row, col int) (expr string) {

	wl.assertWithinBounds(row, col)

	idx := wl.ref[row][col]
	switch {
	case idx >= 0:
		// Note: this should return an error, but since the solution
		// is not complete it just returns a tag with the cell name for now
		expr = newCellRef(row, col).String() + "_val"
	case idx == PROCESSED:
		expr = wl.result[row][col]
	default:
		fmt.Printf("Error: Invalid cell reference (%d, %d).", row, col)
		os.Exit(INVALID_CELL_REFERENCE_ERROR)
	}

	return
}

// Stops execution if the coordinates passed are out of bounds.
func (wl *Workload) assertWithinBounds(row, col int) {

	if row < 0 || col < 0 || row >= len(wl.result) || col >= len(wl.result[0]) {
		fmt.Printf("Error: Invalid cell reference (%d, %d).", row, col)
		os.Exit(INVALID_CELL_REFERENCE_ERROR)
	}
}

// Process a given workload bucket.
func (wl *Workload) processBucket(i int) {

	for j, cell := range wl.buckets[i] {
		if evaluateCell(&wl.buckets[i][j], wl) {
			wl.result[cell.row][cell.col] = cell.expr
			wl.ref[cell.row][cell.col] = PROCESSED
		}
	}

	// since the solution is not complete copy non-evaluated values
	// either way (but only after the initial processing loop)
	for _, cell := range wl.buckets[i] {
		if wl.ref[cell.row][cell.col] >= 0 {
			wl.result[cell.row][cell.col] = cell.expr
			wl.ref[cell.row][cell.col] = PROCESSED
		}
	}
}

// Process the workload in parallel (by assigning each bucket to a different
// goroutine). Let Go and the OS care about the parallelization details.
func (wl *Workload) parallelize() {

	var wg sync.WaitGroup

	for i := range wl.buckets {
		if len(wl.buckets[i]) > 0 {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				wl.processBucket(i)
			}(i)
			goroutinesLaunched++
		}
	}

	wg.Wait()
}

// Process the workload sequentially (on the main thread).
func (wl *Workload) runSequencially() {

	for i := range wl.buckets {
		wl.processBucket(i)
	}
}

// Prints some workload internal details (for debugging purposes).
func (wl *Workload) printDetails() {

	// print bucket distribution
	fmt.Println("Buckets:")
	for i, bucket := range wl.buckets {
		if len(bucket) > 0 {
			fmt.Printf("%d: ", i)
			for j, cell := range bucket {
				fmt.Printf("%s ", cell)
				if j < len(bucket)-1 {
					fmt.Print("-> ")
				}
			}
			fmt.Print("\n")
		}
	}

	// print reference board
	fmt.Print(refBoardToString(wl.ref))
}
