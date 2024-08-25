package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestValidateFlags(t *testing.T) {
	tests := []struct {
		name        string
		high        int
		low         int
		highPercent float64
		lowPercent  float64
		wantErr     bool
	}{
		{"valid inputs", 5, 5, 0, 0, false},
		{"negative high", -1, 5, 0, 0, true},
		{"negative low", 5, -1, 0, 0, true},
		{"invalid high percent", 0, 0, 101, 0, true},
		{"invalid low percent", 0, 0, 0, 101, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFlags(tt.high, tt.low, tt.highPercent, tt.lowPercent)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProcessInput(t *testing.T) {
	input := strings.NewReader("line1\nline2\nline1\nLINE1\n")

	tests := []struct {
		name       string
		ignoreCase bool
		want       map[string]int
	}{
		{
			name:       "case sensitive",
			ignoreCase: false,
			want: map[string]int{
				"line1": 2,
				"line2": 1,
				"LINE1": 1,
			},
		},
		{
			name:       "case insensitive",
			ignoreCase: true,
			want: map[string]int{
				"line1": 3,
				"line2": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processInput(input, tt.ignoreCase)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processInput() = %v, want %v", got, tt.want)
			}
			input.Seek(0, 0) // Reset the reader for the next test
		})
	}
}

func TestSortItems(t *testing.T) {
	items := []item{
		{"apple", 3},
		{"banana", 2},
		{"cherry", 5},
		{"date", 1},
		{"elderberry", 4},
	}

	sortItems(items)

	want := []item{
		{"cherry", 5},
		{"elderberry", 4},
		{"apple", 3},
		{"banana", 2},
		{"date", 1},
	}

	if !reflect.DeepEqual(items, want) {
		t.Errorf("sortItems() = %v, want %v", items, want)
	}
}

func TestDetermineCounts(t *testing.T) {
	tests := []struct {
		name        string
		totalItems  int
		high        int
		low         int
		highPercent float64
		lowPercent  float64
		wantHigh    int
		wantLow     int
	}{
		{"absolute counts", 100, 10, 5, 0, 0, 10, 5},
		{"percentage counts", 100, 0, 0, 20, 10, 20, 10},
		{"mixed counts", 100, 15, 0, 0, 5, 15, 5},
		{"exceed total items", 50, 60, 60, 0, 0, 50, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHigh, gotLow := determineCounts(tt.totalItems, tt.high, tt.low, tt.highPercent, tt.lowPercent)
			if gotHigh != tt.wantHigh || gotLow != tt.wantLow {
				t.Errorf("determineCounts() = (%v, %v), want (%v, %v)", gotHigh, gotLow, tt.wantHigh, tt.wantLow)
			}
		})
	}
}

func TestCollectInputs(t *testing.T) {
	// Test with file input
	tmpfile, err := os.CreateTemp("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := []byte("test1\ntest2\ntest3\n")
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	inputs, inputNames, cleanup, err := collectInputs(tmpfile.Name())
	if err != nil {
		t.Errorf("collectInputs() error = %v", err)
	} else {
		defer cleanup()

		if len(inputs) != 1 || len(inputNames) != 1 {
			t.Errorf("collectInputs() got %d inputs, want 1", len(inputs))
		} else if inputNames[0] != tmpfile.Name() {
			t.Errorf("collectInputs() got input name %s, want %s", inputNames[0], tmpfile.Name())
		}
	}

	// Test with no input (should return an error)
	inputs, inputNames, cleanup, err = collectInputs("")
	if err == nil {
		t.Errorf("collectInputs() expected error for no input, got nil")
		defer cleanup()
	}

	// Test with non-existent file
	inputs, inputNames, cleanup, err = collectInputs("non_existent_file.txt")
	if err == nil {
		t.Errorf("collectInputs() expected error for non-existent file, got nil")
		defer cleanup()
	}
}

func TestProcessInputs(t *testing.T) {
	input1 := strings.NewReader("line1\nline2\nline1\n")
	input2 := strings.NewReader("line3\nline4\nline3\n")
	inputs := []io.Reader{input1, input2}
	inputNames := []string{"input1", "input2"}

	results := processInputs(inputs, inputNames, false, true, false)

	if len(results) != 3 { // 2 individual results + 1 aggregate
		t.Errorf("processInputs() got %d results, want 3", len(results))
	}

	// Check aggregate result
	aggregateResult := results[2]
	if aggregateResult.name != "Aggregate" {
		t.Errorf("Aggregate result name = %s, want Aggregate", aggregateResult.name)
	}
	if len(aggregateResult.counts) != 4 {
		t.Errorf("Aggregate result has %d counts, want 4", len(aggregateResult.counts))
	}
	if aggregateResult.counts["line1"] != 2 {
		t.Errorf("Aggregate count for 'line1' = %d, want 2", aggregateResult.counts["line1"])
	}
}

func TestTextOutput(t *testing.T) {
	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	results := []inputResult{
		{name: "test", counts: map[string]int{"line1": 2, "line2": 1}},
	}

	textOutput(results, 2, 0, 0, 0)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "=== Results for test:") {
		t.Errorf("textOutput() output doesn't contain expected header")
	}
	if !strings.Contains(output, "2 line1") {
		t.Errorf("textOutput() output doesn't contain expected count for line1")
	}
	if !strings.Contains(output, "1 line2") {
		t.Errorf("textOutput() output doesn't contain expected count for line2")
	}
}

func TestPrintSortedResults(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	counts := map[string]int{"line1": 3, "line2": 2, "line3": 1}
	printSortedResults(counts, 2, 0, 0, 0)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "3 line1") {
		t.Errorf("printSortedResults() output doesn't contain expected count for line1")
	}
	if !strings.Contains(output, "2 line2") {
		t.Errorf("printSortedResults() output doesn't contain expected count for line2")
	}
	if strings.Contains(output, "1 line3") {
		t.Errorf("printSortedResults() output contains unexpected count for line3")
	}
}

func TestMainFunction(t *testing.T) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Write some content to the file
	content := []byte("line1\nline2\nline1\nline3\n")
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run main with arguments
	os.Args = []string{"cmd", "-high", "2", "-file", tmpfile.Name()}
	main()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "2 line1") {
		t.Errorf("main() output doesn't contain expected count for line1")
	}
	if !strings.Contains(output, "1 line2") {
		t.Errorf("main() output doesn't contain expected count for line2")
	}
}

func TestPrintAllItems(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	items := []item{
		{"line1", 3},
		{"line2", 2},
		{"line3", 1},
	}

	printAllItems(items)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	expectedLines := []string{"3 line1", "2 line2", "1 line3"}
	for _, line := range expectedLines {
		if !strings.Contains(output, line) {
			t.Errorf("printAllItems() output doesn't contain expected line: %s", line)
		}
	}
}

func TestExitWithError(t *testing.T) {
	oldOsExit := osExit
	oldStderr := os.Stderr
	defer func() {
		osExit = oldOsExit
		os.Stderr = oldStderr
	}()

	var exitCode int
	osExit = func(code int) {
		exitCode = code
	}

	r, w, _ := os.Pipe()
	os.Stderr = w

	expectedError := "test error"
	exitWithError(fmt.Errorf(expectedError))

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, expectedError) {
		t.Errorf("exitWithError() output doesn't contain expected error: %s", expectedError)
	}

	if exitCode != 1 {
		t.Errorf("exitWithError() exit code = %d, want 1", exitCode)
	}
}
