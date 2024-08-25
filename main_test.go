package main

import (
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
