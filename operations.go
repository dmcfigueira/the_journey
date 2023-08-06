package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

// Compute the dependencies for a given cell.
// Updates the cell expression if necessary.
func computeDependencies(cell *Cell, wl *Workload) {

	// check for the '^^' usages - first thing to be done since it is the only
	// operator that might requires us to change the expression of the cell
	if cell.expr == "=^^" {
		cell.expr = wl.getCellExpr(cell.row-1, cell.col)

		// update 'incFrom' usages
		cell.expr = incExprNumber(cell.expr, `incFrom\((\d+)\)`)

		// update cell references
		cell.expr = incExprNumber(cell.expr, `[A-Z]+(\d+)`)
	}

	// check for direct cell references
	for _, match := range findExprMatches(cell.expr, `[A-Z]+\d+`) {
		cell.addDependency(newCellRefFromString(match))
	}

	// check for 'E^' usages (note: problem description says 'A^' instead of 'E^')
	for range findExprMatchesNotFollowedBy(cell.expr, `E\^`, "v") {
		cell.addDependency(newCellRef(cell.row-1, cell.col))
	}

	// check for 'E^v' usages (note: problem description says 'A^' instead of 'E^')
	for range findExprMatches(cell.expr, `E\^v`) {
		// not implemented
	}

	// check for '@label<n>' usages
	for range findExprMatches(cell.expr, `@\w+<\d+>`) {
		// not implemented
	}

	if VERBOSE >= 3 && len(cell.dependencies) > 0 {
		cell.printDependencies()
	}
}

// Increment all matching numbers in an expression. The number to be
// incremented must be a group/submatch in the regex passed.
func incExprNumber(expr, regex string) string {

	pattern := regexp.MustCompile(regex)
	matches := pattern.FindAllStringSubmatchIndex(expr, -1)

	if len(matches) == 0 {
		return expr // no changes
	}

	i := 0
	newExpr := ""
	for _, match := range matches {
		newExpr += expr[i:match[2]]
		num, err := strconv.Atoi(expr[match[2]:match[3]])
		if err != nil {
			fmt.Println(err)
			os.Exit(INVALID_EXPRESSION_ERROR)
		}
		newExpr += strconv.Itoa(num + 1)
		i = match[3]
	}
	newExpr += expr[i:]

	return newExpr
}

// Returns all the matches in an expression.
func findExprMatches(expr, regex string) []string {
	pattern := regexp.MustCompile(regex)
	return pattern.FindAllString(expr, -1)
}

// Returns all the matches in an expression that are not followed by a given string.
// Necessary since Golang does not support regex negative lookaheads.
func findExprMatchesNotFollowedBy(expr, regex, suffix string) (result []string) {
	pattern := regexp.MustCompile(regex)
	for _, match := range pattern.FindAllStringIndex(expr, -1) {
		if expr[match[1]:match[1]+len(suffix)] != suffix {
			result = append(result, expr[match[0]:match[1]])
		}
	}
	return
}

// Tries evaluate a cell expression (probably a naive approach).
// Returns true if the cell was successfully evaluated.
func evaluateCell(cell *Cell, wl *Workload) bool {

	// check for direct cell references
	for _, match := range findExprMatches(cell.expr, `[A-Z]+\d+`) {
		ref := newCellRefFromString(match)
		value := wl.getCellValue(ref.row, ref.col)
		cell.expr = replaceExprMatches(cell.expr, `[A-Z]+\d+`, value)
	}

	// check for 'E^' usages (note: problem description says 'A^' instead of 'E^')
	for range findExprMatchesNotFollowedBy(cell.expr, `E\^`, "v") {
		value := wl.getCellValue(cell.row-1, cell.col)
		cell.expr = replaceExprMatchesNotFollowedBy(cell.expr, `E\^`, "v", value)
	}

	// check for 'E^v' usages (note: problem description says 'A^' instead of 'E^')
	for range findExprMatches(cell.expr, `E\^v`) {
		return false // not implemented
	}

	// check for '@label<n>' usages
	for range findExprMatches(cell.expr, `@\w+<\d+>`) {
		return false // not implemented
	}

	// check for other operations
	for range findExprMatches(cell.expr, `\w+`) {
		return false // not implemented
	}

	return true
}

// Replaces all matches in an expression.
func replaceExprMatches(expr, regex, value string) string {
	pattern := regexp.MustCompile(regex)
	return pattern.ReplaceAllString(expr, value)
}

// Replaces all matches in an expression that are not followed by a given string.
// Necessary since Golang does not support regex negative lookaheads.
func replaceExprMatchesNotFollowedBy(expr, regex, suffix, value string) (result string) {
	pattern := regexp.MustCompile(regex)
	i := 0
	for _, match := range pattern.FindAllStringIndex(expr, -1) {
		if expr[match[1]:match[1]+len(suffix)] != suffix {
			result += expr[i:match[0]]
			result += value
			i = match[1]
		}
	}
	result += expr[i:]
	return
}
