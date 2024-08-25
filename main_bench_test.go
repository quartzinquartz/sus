package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func BenchmarkProcessInput(b *testing.B) {
	input := strings.NewReader("line1\nline2\nline1\nLINE1\n")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processInput(input, false)
		input.Seek(0, 0) // Reset the reader for the next iteration
	}
}

func BenchmarkSortItems(b *testing.B) {
	items := []item{
		{"apple", 3},
		{"banana", 2},
		{"cherry", 5},
		{"date", 1},
		{"elderberry", 4},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sortItems(items)
	}
}

func BenchmarkProcessLargeInput(b *testing.B) {
	// Create a larger input
	var sb strings.Builder
	for i := 0; i < 10000; i++ {
		sb.WriteString(fmt.Sprintf("line%d\n", i%100))
	}
	input := strings.NewReader(sb.String())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processInput(input, false)
		input.Seek(0, 0)
	}
}

func BenchmarkPrintSortedResults(b *testing.B) {
	counts := make(map[string]int)
	for i := 0; i < 1000; i++ {
		counts[fmt.Sprintf("line%d", i)] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		printSortedResults(counts, 10, 10, 0, 0)
	}
}

func BenchmarkDetermineCounts(b *testing.B) {
	for i := 0; i < b.N; i++ {
		determineCounts(1000, 10, 5, 1.0, 0.5)
	}
}

func BenchmarkPercentageBased(b *testing.B) {
	// Create a large dataset
	var sb strings.Builder
	for i := 0; i < 100000; i++ {
		sb.WriteString(fmt.Sprintf("line%d\n", i%1000))
	}
	input := strings.NewReader(sb.String())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		oldStdout := os.Stdout
		os.Stdout = nil
		counts := processInput(input, false)
		printSortedResults(counts, 0, 0, 1.0, 0) // Top 1%
		os.Stdout = oldStdout
		input.Seek(0, 0)
	}
}
