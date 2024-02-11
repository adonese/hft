package main

import (
	"reflect"
	"testing"
)

func TestRunMatchingEngine(t *testing.T) {
	testCases := []struct {
		name     string
		input    []string
		expected []string
	}{

		{
			name:     "Test Case 1", // this is the happy path, just inserting and updating
			input:    []string{"INSERT,1,FFLY,BUY,23.45,12", "INSERT,2,FFLY,SELL,23.50,10", "UPDATE,1,23.48,12", "CANCEL,2"},
			expected: []string{"FFLY,23.50,10,1,2", "===FFLY===", "BUY,23.48,2"},
		},

		{
			name:     "Test Case 2",
			input:    []string{"INSERT,1,FFLY,BUY,23.45,12"},
			expected: []string{"===FFLY===", "BUY,23.45,12"},
		},

		{
			name:     "Test Case 3",
			input:    []string{"INSERT,1,FFLY,SELL,23.45,12"},
			expected: []string{"===FFLY===", "SELL,23.45,12"},
		},

		{
			name:     "Test Case 4",
			input:    []string{"INSERT,1,FFLY,BUY,23.45,12", "INSERT,2,FFLY,SELL,23.45,10", "INSERT,3,FFLY,BUY,23.45,5", "INSERT,4,FFLY,SELL,23.45,5"},
			expected: []string{"FFLY,23.45,10,1,2", "FFLY,23.45,5,3,4", "===FFLY===", "BUY,23.45,2"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := runMatchingEngine(tc.input)
			if !reflect.DeepEqual(output, tc.expected) {
				t.Errorf("Expected %v, but got %v", tc.expected, output)
			}
		})
	}
}
