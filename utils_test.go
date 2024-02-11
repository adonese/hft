package main

import (
	"reflect"
	"testing"
)

func TestOrderBook_executeOperation(t *testing.T) {
	type fields struct {
	}

	tests := []struct {
		name   string
		fields fields
		args   []*Order
	}{
		{"test happy path", fields{}, sample},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			ob := &OrderBook{}

			for _, or := range sample {
				ob.executeOperation(*or)
			}

			if len(ob.orders) != 2 {
				t.Fatalf("error in the output shape: %#v", ob.orders)
			}

			if ob.orders[0].OrderID != 6 || ob.orders[1].OrderID != 3 {
				t.Fatalf("there is an error")
			}
		})
	}

}

var sample = []*Order{
	{OrderID: 1, Side: "BUY", Volume: 1, Price: 122},
	{OrderID: 2, Side: "BUY", Volume: 12, Price: 121},
	{OrderID: 3, Side: "BUY", Volume: 7, Price: 120},
	{OrderID: 4, Side: "BUY", Volume: 23, Price: 121},
	{OrderID: 5, Side: "SELL", Volume: 16, Price: 120},
	{OrderID: 6, Side: "SELL", Volume: 24, Price: 121},
}

func TestSortBuyers(t *testing.T) {
	ob := &OrderBook{
		buyers: []*Order{
			{OrderID: 1, Side: "BUY", Volume: 1, Price: 122},
			{OrderID: 2, Side: "BUY", Volume: 12, Price: 121},
			{OrderID: 3, Side: "BUY", Volume: 7, Price: 120},
			{OrderID: 4, Side: "BUY", Volume: 23, Price: 121},
		},
	}

	ob.sortBuyers()

	expected := []*Order{
		{OrderID: 1, Side: "BUY", Volume: 1, Price: 122},
		{OrderID: 2, Side: "BUY", Volume: 12, Price: 121},
		{OrderID: 4, Side: "BUY", Volume: 23, Price: 121},
		{OrderID: 3, Side: "BUY", Volume: 7, Price: 120},
	}

	if !reflect.DeepEqual(ob.buyers, expected) {
		t.Errorf("sortBuyers() = %v, want %v", ob.buyers, expected)
	}
}
