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
			name: "Test Case 5",
			input: []string{
				"INSERT,1,FFLY,BUY,45.95,5",
				"INSERT,2,FFLY,BUY,45.95,6",
				"INSERT,3,FFLY,BUY,45.95,12",
				"INSERT,4,FFLY,SELL,46,8",
				"UPDATE,2,46,3",
				"INSERT,5,FFLY,SELL,45.95,1",
				"UPDATE,1,45.95,3",
				"INSERT,6,FFLY,SELL,45.95,1",
				"UPDATE,1,45.95,5",
				"INSERT,7,FFLY,SELL,45.95,1",
			},
			expected: []string{
				"FFLY,46,3,2,4",
				"FFLY,45.95,1,5,1",
				"FFLY,45.95,1,6,1",
				"FFLY,45.95,1,7,3",
				"===FFLY===",
				"SELL,46,5",
				"BUY,45.95,16",
			},
		},

		{name: "Test Case 4",
			input: []string{
				"INSERT,1,FFLY,BUY,14.235,5",
				"INSERT,2,FFLY,BUY,14.235,6",
				"INSERT,3,FFLY,BUY,14.235,12",
				"INSERT,4,FFLY,BUY,14.234,5",
				"INSERT,5,FFLY,BUY,14.23,3",
				"INSERT,6,FFLY,SELL,14.237,8",
				"INSERT,7,FFLY,SELL,14.24,9",
				"CANCEL,1",
				"INSERT,8,FFLY,SELL,14.234,25",
			}, expected: []string{
				"FFLY,14.235,6,8,2",
				"FFLY,14.235,12,8,3",
				"FFLY,14.234,5,8,4",
				"===FFLY===",
				"SELL,14.24,9",
				"SELL,14.237,8",
				"SELL,14.234,2",
				"BUY,14.23,3"}},

		{
			name: "Test Case 1",
			input: []string{
				"INSERT,1,FFLY,BUY,0.3854,5",
				"INSERT,2,ETH,BUY,412,31",
				"INSERT,3,ETH,BUY,410.5,27",
				"INSERT,4,DOT,SELL,21,8",
				"INSERT,11,FFLY,SELL,0.3854,4",
				"INSERT,13,FFLY,SELL,0.3853,6",
			},
			expected: []string{
				"FFLY,0.3854,4,11,1",
				"FFLY,0.3854,1,13,1",
				"===DOT===",
				"SELL,21,8",
				"===ETH===",
				"BUY,412,31",
				"BUY,410.5,27",
				"===FFLY===",
				"SELL,0.3853,5",
			},
		},
		{
			name: "Test case 2",
			input: []string{
				"INSERT,1,FFLY,BUY,12.2,5",
				"INSERT,2,FFLY,SELL,12.3,5",
				"INSERT,3,FFLY,SELL,12.3,5",
				"CANCEL,2",
			},
			expected: []string{
				"===FFLY===",
				"SELL,12.3,5",
				"BUY,12.2,5",
			},
		},
		{name: "Test case 6",
			input: []string{
				"INSERT,1,FFLY,SELL,12.2,5",
				"INSERT,2,FFLY,SELL,12.1,8",
				"INSERT,3,FFLY,BUY,12.5,10",
			},

			expected: []string{
				"FFLY,12.5,8,3,2",
				"FFLY,12.5,2,3,1",
				"===FFLY===",
				"SELL,12.2,3",
			},
		},
		{
			name: "test case 10",
			input: []string{
				"INSERT,1,FFLY,BUY,47,5",
				"INSERT,2,FFLY,BUY,47,6",
				"INSERT,3,FFLY,SELL,47,9",
				"UPDATE,2,47,-1",
			},
			expected: []string{
				"FFLY,47,5,3,1",
				"FFLY,47,4,3,2",
				"===FFLY===",
			},
		},
		{
			name: "test case 11",
			input: []string{
				"INSERT,1,FFLY,BUY,47,5",
				"INSERT,2,FFLY,BUY,47,6",
				"INSERT,3,FFLY,SELL,47,9", // this should be zero
				"UPDATE,1,45,2",           // one is gone, no-op
				"UPDATE,5,45,2",           // five is no-existent, no-op
			},
			expected: []string{
				"FFLY,47,5,3,1",
				"FFLY,47,4,3,2",
				"===FFLY===",
				"BUY,47,2",
			},
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
