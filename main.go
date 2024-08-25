// Copyright (c) 2024 Derek Jenkins
// SPDX-License-Identifier: MIT

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
)

const (
	version       = "1.0.0"
	maxLineLength = 1000
	usageExample  = `
Example usage:
  sus -high 5 -file input.txt
  cat input.txt | sus -low 10 -i
  sus -hp 10 -lp 5 -file file1.txt,file2.txt -aggregate`
)

// item represents a line of text and its frequency count.
type item struct {
	value string
	count int
}

// inputResult holds the results of counting lines from a particular input source.
type inputResult struct {
	name   string
	counts map[string]int
}

var osExit = os.Exit

// Package main provides a command-line tool for analyzing line frequency in text (file, stdin, or both).
func main() {
	high := flag.Int("high", 0, "Show N most frequent results")
	low := flag.Int("low", 0, "Show N least frequent results")
	highPercent := flag.Float64("hp", 0, "Show top N percent of most frequent results")
	lowPercent := flag.Float64("lp", 0, "Show bottom N percent of least frequent results")
	aggregate := flag.Bool("aggregate", false, "Show aggregate results across all input sources")
	ignoreCase := flag.Bool("i", false, "Ignore case when counting lines")
	flagFiles := flag.String("file", "", "Input files separated by commas (optional)")
	showVersion := flag.Bool("version", false, "Show version information")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, usageExample)
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("sus version %s\n", version)
		os.Exit(0)
	}

	if *verbose {
		fmt.Println("Verbose mode enabled")
	}

	if err := validateFlags(*high, *low, *highPercent, *lowPercent); err != nil {
		exitWithError(err)
	}

	inputs, inputNames, cleanup, err := collectInputs(*flagFiles)
	if err != nil {
		exitWithError(err)
	}
	defer cleanup()

	results := processInputs(inputs, inputNames, *ignoreCase, *aggregate, *verbose)
	textOutput(results, *high, *low, *highPercent, *lowPercent)
}

// validateFlags checks if the provided flag values are valid.
func validateFlags(high, low int, highPercent, lowPercent float64) error {
	if high < 0 || low < 0 {
		return fmt.Errorf("Error: -high and -low must be non-negative integers")
	}
	if highPercent < 0 || highPercent > 100 || lowPercent < 0 || lowPercent > 100 {
		return fmt.Errorf("Error: -hp and -lp must be between 0 and 100")
	}
	return nil
}

// collectInputs gathers input from specified files or stdin.
// It returns slices of io.Readers and their names, a cleanup function, and any error encountered.
func collectInputs(flagFiles string) ([]io.Reader, []string, func(), error) {
	var inputs []io.Reader
	var inputNames []string
	var files []*os.File
	var once sync.Once

	cleanup := func() {
		once.Do(func() {
			for _, f := range files {
				f.Close()
			}
		})
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		inputs = append(inputs, os.Stdin)
		inputNames = append(inputNames, "stdin")
	}

	if flagFiles != "" {
		for _, file := range strings.Split(flagFiles, ",") {
			f, err := os.Open(file)
			if err != nil {
				cleanup()
				return nil, nil, nil, fmt.Errorf("Error opening file %s: %v", file, err)
			}
			if fi, _ := f.Stat(); fi.Size() == 0 {
				cleanup()
				return nil, nil, nil, fmt.Errorf("Error: file %s is empty", file)
			}
			files = append(files, f)
			inputs = append(inputs, f)
			inputNames = append(inputNames, file)
		}
	}

	if len(inputs) == 0 {
		return nil, nil, nil, fmt.Errorf("Error: No input provided. Use stdin or specify files.")
	}

	return inputs, inputNames, cleanup, nil
}

// processInputs processes multiple input sources, optionally aggregating results.
// It returns a slice of inputResult structures containing the processed data.
func processInputs(inputs []io.Reader, inputNames []string, ignoreCase, aggregate bool, verbose bool) []inputResult {
	var wg sync.WaitGroup
	resultsChan := make(chan inputResult, len(inputs))
	aggregateCounts := make(map[string]int)
	var aggregateMutex sync.Mutex

	for idx, input := range inputs {
		wg.Add(1)
		go func(index int, input io.Reader, name string) {
			defer wg.Done()

			if verbose {
				fmt.Printf("Processing input: %s\n", name)
			}

			fileCounts := processInput(input, ignoreCase)
			resultsChan <- inputResult{name, fileCounts}

			if aggregate {
				aggregateMutex.Lock()
				for line, count := range fileCounts {
					aggregateCounts[line] += count
				}
				aggregateMutex.Unlock()
			}

			if verbose {
				fmt.Printf("Finished processing input: %s\n", name)
			}
		}(idx, input, inputNames[idx])
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var allResults []inputResult
	for result := range resultsChan {
		allResults = append(allResults, result)
	}

	if aggregate && len(inputs) > 1 {
		allResults = append(allResults, inputResult{"Aggregate", aggregateCounts})
	}

	return allResults
}

// processInput reads from the provided input and counts line frequencies.
func processInput(input io.Reader, ignoreCase bool) map[string]int {
	counts := make(map[string]int)
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, maxLineLength), maxLineLength)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > maxLineLength {
			exitWithError(fmt.Errorf("Line exceeds maximum length of %d characters", maxLineLength))
		}
		if ignoreCase {
			line = strings.ToLower(line)
		}
		counts[line]++
	}

	if err := scanner.Err(); err != nil {
		exitWithError(fmt.Errorf("Error reading input: %v", err))
	}
	return counts
}

// textOutput prints the processed results to stdout in a human-readable format.
// It handles both individual and aggregated results based on the specified options.
func textOutput(results []inputResult, high, low int, highPercent, lowPercent float64) {
	for i, result := range results {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("=== Results for %s:\n", result.name)
		printSortedResults(result.counts, high, low, highPercent, lowPercent)
	}
}

// printSortedResults sorts and prints the line frequency results based on the specified criteria.
// It can show the most frequent, least frequent, or a percentage-based selection of results.
func printSortedResults(counts map[string]int, high, low int, highPercent, lowPercent float64) {
	items := make([]item, 0, len(counts))
	for value, count := range counts {
		items = append(items, item{value, count})
	}

	sortItems(items)

	totalItems := len(items)
	highCount, lowCount := determineCounts(totalItems, high, low, highPercent, lowPercent)

	// Ensure highCount and lowCount do not exceed available items
	highCount = min(highCount, totalItems)
	lowCount = min(lowCount, totalItems)

	if highCount > 0 {
		printFrequencyItems(items[:highCount], highCount, highPercent, "Highest")
	}
	if lowCount > 0 {
		if highCount > 0 {
			fmt.Println()
		}
		printFrequencyItems(items[len(items)-lowCount:], lowCount, lowPercent, "Lowest")
	}
	if highCount == 0 && lowCount == 0 {
		printAllItems(items)
	}
}

// sortItems sorts a slice of items in descending order of count.
// If counts are equal, it sorts lexicographically by value.
func sortItems(items []item) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].count == items[j].count {
			return items[i].value < items[j].value
		}
		return items[i].count > items[j].count
	})
}

// determineCounts calculates the number of items to display based on absolute counts or percentages.
// It returns the number of high and low frequency items to show.
func determineCounts(totalItems, high, low int, highPercent, lowPercent float64) (int, int) {
	highCount := high
	lowCount := low

	if highPercent > 0 {
		highCount = int(math.Ceil(float64(totalItems) * highPercent / 100))
	}
	if lowPercent > 0 {
		lowCount = int(math.Ceil(float64(totalItems) * lowPercent / 100))
	}

	// Cap the counts at the total number of items
	highCount = min(highCount, totalItems)
	lowCount = min(lowCount, totalItems)

	return highCount, lowCount
}

// printFrequencyItems prints a subset of items based on their frequency.
// It's used to display either the highest or lowest frequency items.
func printFrequencyItems(items []item, count int, percent float64, label string) {
	if percent > 0 {
		fmt.Printf("%s %.2f%% (%d) frequency items:\n", label, percent, count)
	} else {
		fmt.Printf("%s %d frequency items:\n", label, count)
	}
	for _, item := range items {
		fmt.Printf("%d %s\n", item.count, item.value)
	}
}

// printAllItems prints all items in the order they appear in the slice.
// This is used when no specific high or low count is requested.
func printAllItems(items []item) {
	for _, item := range items {
		fmt.Printf("%d %s\n", item.count, item.value)
	}
}

// exitWithError prints an error message to stderr and exits the program with status code 1.
func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, err)
	osExit(1)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
