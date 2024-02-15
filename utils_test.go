package main

import (
	"container/heap"
	"log"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func setupOrderBook() (*OrderBook, []*Order) {
	ob := NewOrderBook()
	orders := []*Order{
		{ID: 1, Price: 10.00, Volume: 5, Side: "BUY"},
		{ID: 2, Price: 9.50, Volume: 10, Side: "BUY"},
		{ID: 3, Price: 10.50, Volume: 5, Side: "SELL"},
		{ID: 4, Price: 11.00, Volume: 10, Side: "SELL"},
	}
	return ob, orders
}

func TestInsertOrderIntoHeap(t *testing.T) {
	ob, orders := setupOrderBook()

	// Insert buy and sell orders
	for _, order := range orders {
		ob.insertOrderIntoHeap(order)
	}

	// Verify heap properties and order priorities
	if (*ob.BuyOrders)[0].ID != 1 || (*ob.SellOrders)[0].ID != 3 {
		t.Errorf("InsertOrderIntoHeap did not insert orders correctly")
	}
}

func TestRemoveOrderFromHeap(t *testing.T) {
	ob, orders := setupOrderBook()

	// Insert orders first
	for _, order := range orders {
		ob.insertOrderIntoHeap(order)
	}

	// Now remove a buy and a sell order
	ob.removeOrderFromHeap(orders[0]) // Remove first BUY order
	ob.removeOrderFromHeap(orders[2]) // Remove first SELL order

	// Check if the orders were removed correctly
	for _, order := range *ob.BuyOrders {
		if order.ID == 1 {
			t.Errorf("RemoveOrderFromHeap did not remove the BUY order correctly")
		}
	}
	for _, order := range *ob.SellOrders {
		if order.ID == 3 {
			t.Errorf("RemoveOrderFromHeap did not remove the SELL order correctly")
		}
	}
}

// func TestMatchOrdersPricePriority(t *testing.T) {
// 	ob := NewOrderBook()

// 	// Insert buy and sell orders at different prices
// 	ob.Insert(&Order{ID: 1, Symbol: "FFLY", Side: "BUY", Price: 23.50, Volume: 10})
// 	ob.Insert(&Order{ID: 2, Symbol: "FFLY", Side: "SELL", Price: 23.45, Volume: 10})

// 	// Trigger matching
// 	ob.matchOrders()

// 	// Check if the orders were matched correctly
// 	if len(ob.Trades) != 1 {
// 		t.Errorf("Expected 1 trade, got %d", len(ob.Trades))
// 	}

// 	// Verify the trade details
// 	trade := ob.Trades[0]
// 	if trade != "FFLY,23.45,10,2,1" {
// 		t.Errorf("Trade did not match expected details, got %s - wanted: %s", trade, trade)
// 	}
// }

// func TestMatchOrdersTimePriority(t *testing.T) {
// 	ob := NewOrderBook()

// 	// Insert two buy orders at the same price but different times
// 	ob.Insert(&Order{ID: 1, Symbol: "FFLY", Side: "BUY", Price: 23.45, Volume: 5})
// 	ob.Insert(&Order{ID: 2, Symbol: "FFLY", Side: "BUY", Price: 23.45, Volume: 5})

// 	// Insert a sell order that can match with both buy orders
// 	ob.Insert(&Order{ID: 3, Symbol: "FFLY", Side: "SELL", Price: 23.45, Volume: 10})

// 	// Trigger matching
// 	ob.matchOrders()

// 	// Verify that the first buy order was matched first
// 	if len(ob.Trades) != 2 {
// 		t.Errorf("We are getting more extra trades")
// 	}

// 	if !reflect.DeepEqual(ob.Trades[0], "FFLY,23.45,5,3,1") {
// 		t.Errorf("First trade did not match expected details, got %s", ob.Trades[0])
// 	}
// }

// func TestMatchOrdersWithCancelAndUpdate(t *testing.T) {
// 	ob := NewOrderBook()

// 	// Setup orders and insert them
// 	ob.Insert(&Order{ID: 1, Symbol: "FFLY", Side: "BUY", Price: 23.45, Volume: 10})
// 	sellOrder := &Order{ID: 2, Symbol: "FFLY", Side: "SELL", Price: 23.50, Volume: 10}
// 	ob.Insert(sellOrder)

// 	// Cancel the sell order
// 	ob.Cancel(sellOrder.ID)

// 	// Update the buy order to match the sell order's price, then attempt a match
// 	ob.Update(1, 23.50, 10)
// 	ob.matchOrders()

// 	// Since the sell order was canceled, no trades should occur
// 	if len(ob.Trades) != 0 {
// 		t.Errorf("Expected no trades due to cancellation, got %d trades", len(ob.Trades))
// 	}
// }

func TestUpdateFullyMatchedOrder(t *testing.T) {
	ob := NewOrderBook()

	// Insert an order
	ob.Insert(&Order{ID: 1, Symbol: "FFLY", Side: "BUY", Price: 45.95, Volume: 5})

	// Simulate the order being fully matched by setting its volume to 0
	// This step might be replaced by actual matching logic if you prefer a more integrated test
	ob.Orders[1].Volume = 0

	// Attempt to update the fully matched order
	ob.Update(1, 45.95, 0) // This update should not reinstate the order in the heap

	// Check if the order is still in the heap or has been correctly handled
	for _, order := range *ob.BuyOrders {
		if order.ID == 1 {
			t.Errorf("Order with ID 1 should not be reinstated in the heap after being fully matched and updated with volume 0: the order is: %+v", order)
		}
	}

	// Optionally, verify the order is not present in the SellOrders heap as well
}

func TestMatchingEngineMakerTakerRoles(t *testing.T) {
	ob := NewOrderBook()

	// Insert initial orders
	ob.Insert(&Order{ID: 1, Symbol: "FFLY", Side: "BUY", Price: 45.95, Volume: 5})
	ob.Insert(&Order{ID: 2, Symbol: "FFLY", Side: "SELL", Price: 45.95, Volume: 5})

	// Update the first order to test if it affects the maker/taker roles
	ob.Update(1, 45.95, 10) // Assuming this increases volume, which could affect its priority

	// Insert a matching order to trigger a trade
	ob.Insert(&Order{ID: 3, Symbol: "FFLY", Side: "SELL", Price: 45.95, Volume: 5})

	// Check the trades to ensure correct maker/taker assignment
	// Expected: The original order (ID: 1) should still be the maker, and the new order (ID: 3) the taker
	if len(ob.Trades) != 1 {
		t.Fatalf("Expected 1 trade, got %d", len(ob.Trades))
	}

	// Extracting trade details
	tradeDetails := strings.Split(ob.Trades[0], ",")
	if tradeDetails[3] != "3" || tradeDetails[4] != "1" {
		t.Errorf("Expected maker/taker roles to be ID 1 (maker) and ID 3 (taker), got maker: %s, taker: %s", tradeDetails[4], tradeDetails[3])
	}
}

func TestOrderReinsertionAfterUpdate(t *testing.T) {
	ob := NewOrderBook()

	// Insert BUY orders at different prices
	ob.Insert(&Order{ID: 1, Symbol: "TEST", Side: "BUY", Price: 100.0, Volume: 10, Inserted: time.Now()})
	ob.Insert(&Order{ID: 2, Symbol: "TEST", Side: "BUY", Price: 101.0, Volume: 10, Inserted: time.Now()})
	ob.Insert(&Order{ID: 3, Symbol: "TEST", Side: "BUY", Price: 102.0, Volume: 10, Inserted: time.Now()})

	// Update the price of the first order to be higher than the rest, ensuring it should be re-inserted with highest priority
	ob.Update(1, 103.0, 10) // Increase price to 103.0

	// Verify that the updated order (ID: 1) is now the first order in the BuyOrders heap
	if (*ob.BuyOrders)[0].ID != 1 {
		t.Errorf("Expected order ID 1 to be the first in the BuyOrders heap after update, found ID %d", (*ob.BuyOrders)[0].ID)
	}

	// Further, verify that the heap maintains the correct order for all other orders
	expectedOrderIDs := []int{1, 3, 2} // After update, the order by priority should be 1, 3, 2 based on price
	for i, expectedID := range expectedOrderIDs {
		if (*ob.BuyOrders)[i].ID != expectedID {
			t.Errorf("At position %d, expected order ID %d, found ID %d", i, expectedID, (*ob.BuyOrders)[i].ID)
		}
	}

	// Optionally, verify that the heap size remains correct (no duplicate insertions)
	if len(*ob.BuyOrders) != 3 {
		t.Errorf("Expected BuyOrders heap size to be 3, found %d", len(*ob.BuyOrders))
	}
}

func TestOrderReinsertionAfterUpdateDetailed(t *testing.T) {
	ob := NewOrderBook()

	// Insert BUY orders at different prices
	ob.Insert(&Order{ID: 1, Symbol: "TEST", Side: "BUY", Price: 100.0, Volume: 10, Inserted: time.Now()})
	ob.Insert(&Order{ID: 2, Symbol: "TEST", Side: "BUY", Price: 101.0, Volume: 10, Inserted: time.Now()})
	ob.Insert(&Order{ID: 3, Symbol: "TEST", Side: "BUY", Price: 102.0, Volume: 10, Inserted: time.Now()})

	// Check initial heap order
	checkHeapOrder(t, ob.BuyOrders, []int{3, 2, 1}, "Initial")

	// Update the price of the first order to be higher than the rest
	ob.Update(1, 103.0, 10) // Increase price to 103.0

	// Check heap order immediately after update
	checkHeapOrder(t, ob.BuyOrders, []int{1, 3, 2}, "After Update")

	// Optionally, verify that the heap size remains correct (no duplicate insertions)
	if len(*ob.BuyOrders) != 3 {
		t.Errorf("Expected BuyOrders heap size to be 3, found %d", len(*ob.BuyOrders))
	}
}

// checkHeapOrder checks the order of orders in the heap against the expected order of IDs
func checkHeapOrder(t *testing.T, heap *MaxHeap, expectedOrder []int, step string) {
	for i, expectedID := range expectedOrder {
		if (*heap)[i].ID != expectedID {
			t.Errorf("%s heap check: At position %d, expected order ID %d, found ID %d", step, i, expectedID, (*heap)[i].ID)
		}
	}
}

func TestMatchingEngineTestCase5(t *testing.T) { // FAILING
	// Initialize a new order book
	ob := NewOrderBook()

	// Define the input operations
	inputs := []string{
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
	}

	// Execute each input operation
	for _, input := range inputs {
		parts := strings.Split(input, ",")
		switch parts[0] {
		case "INSERT":
			id, _ := strconv.Atoi(parts[1])
			price, _ := strconv.ParseFloat(parts[4], 64)
			volume, _ := strconv.Atoi(parts[5])
			ob.Insert(&Order{ID: id, Symbol: parts[2], Side: parts[3], Price: price, Volume: volume})
		case "UPDATE":
			id, _ := strconv.Atoi(parts[1])
			price, _ := strconv.ParseFloat(parts[2], 64)
			volume, _ := strconv.Atoi(parts[3])
			ob.Update(id, price, volume) // Assuming Update method signature matches
			// Add case for "CANCEL" if needed
		}
	}

	// Verify the resulting trades
	expectedTrades := []string{
		"FFLY,46,3,2,4",
		"FFLY,45.95,1,5,1",
		"FFLY,45.95,1,6,1",
		"FFLY,45.95,1,7,3",
	}
	for i, trade := range ob.Trades {
		if i >= len(expectedTrades) {
			t.Error("error number in trades")
		}

		if trade != expectedTrades[i] {
			t.Errorf("Expected trade %s, got %s", expectedTrades[i], trade)
		} else if trade == expectedTrades[i] {
			log.Printf("the matching trades are: %+v and found: %+v", trade, expectedTrades[i])
		}
	}

	// Verify the final state of the order book (simplified check)
	// This part needs to be adjusted based on how you can access and verify the order book's state.
	// For instance, you might want to check the remaining volumes and prices in the buy and sell heaps.
}

// TestVolumeDecreaseWithoutPriceChange ensures that an order that decreases in volume
// without a price change maintains its time priority in the order book and does not
// adversely affect the integrity of the order book.
func TestVolumeDecreaseWithoutPriceChange(t *testing.T) {
	ob := NewOrderBook()

	// Insert a buy order
	buyOrder := &Order{
		ID:       1,
		Symbol:   "TEST",
		Side:     "BUY",
		Price:    100.0,
		Volume:   10,
		Inserted: time.Now(),
	}
	ob.Insert(buyOrder)

	// Insert a sell order
	sellOrder := &Order{
		ID:       2,
		Symbol:   "TEST",
		Side:     "SELL",
		Price:    101.0,
		Volume:   5,
		Inserted: time.Now(),
	}
	ob.Insert(sellOrder)

	// Update the buy order with a decreased volume, keeping the price the same
	ob.Update(buyOrder.ID, buyOrder.Price, 5) // Decrease volume to 5

	// Verify the buy order's volume and position
	if len(*ob.BuyOrders) != 1 {
		t.Fatalf("Expected 1 buy order, found %d", len(*ob.BuyOrders))
	}

	updatedOrder := (*ob.BuyOrders)[0]
	if updatedOrder.Volume != 5 {
		t.Errorf("Expected volume of 5, got %d", updatedOrder.Volume)
	}

	if updatedOrder.ID != buyOrder.ID {
		t.Errorf("Expected buy order ID %d to maintain its position, found ID %d", buyOrder.ID, updatedOrder.ID)
	}

	// Verify the sell order remains unaffected
	if len(*ob.SellOrders) != 1 {
		t.Fatalf("Expected 1 sell order, found %d", len(*ob.SellOrders))
	}

	if (*ob.SellOrders)[0].ID != sellOrder.ID {
		t.Errorf("Expected sell order ID %d to remain unchanged, found ID %d", sellOrder.ID, (*ob.SellOrders)[0].ID)
	}

	// Ensure no trades were executed as a result of the update
	if len(ob.Trades) != 0 {
		t.Errorf("Expected no trades to be executed, found %d trades", len(ob.Trades))
	}

	// Ensure the overall integrity of the order book is maintained
	if len(ob.Orders) != 2 {
		t.Errorf("Expected total of 2 orders in the order book, found %d", len(ob.Orders))
	}
}

// TestMatchingLogicAfterUpdate verifies that after an order update leading to a trade,
// the matching logic correctly identifies the taker and maker in the trade.
func TestMatchingLogicAfterUpdate(t *testing.T) {
	ob := NewOrderBook()

	// Insert a sell order
	sellOrder := &Order{
		ID:       1,
		Symbol:   "TEST",
		Side:     "SELL",
		Price:    100.0,
		Volume:   10,
		Inserted: time.Now(),
	}
	ob.Insert(sellOrder)

	// Insert a buy order with a lower price, ensuring no immediate match
	buyOrder := &Order{
		ID:       2,
		Symbol:   "TEST",
		Side:     "BUY",
		Price:    99.0,
		Volume:   10,
		Inserted: time.Now(),
	}
	ob.Insert(buyOrder)

	// Update the buy order to match the sell order's price, triggering a trade
	ob.Update(buyOrder.ID, 100.0, buyOrder.Volume)

	// Assuming trades are recorded as "Symbol,Price,Volume,TakerID,MakerID"
	if len(ob.Trades) != 1 {
		t.Fatalf("Expected 1 trade to be executed, found %d", len(ob.Trades))
	}

	trade := ob.Trades[0]
	expectedTrade := "TEST,100,10,2,1" // Expecting the updated buy order as the taker

	if trade != expectedTrade {
		t.Errorf("Expected trade %s, got %s", expectedTrade, trade)
	}
}

// Helper function to create an order
func createTestOrder(id int, price float64, volume int, inserted string) *Order {
	t, _ := time.Parse(time.RFC3339, inserted)
	return &Order{
		ID:       id,
		Price:    price,
		Volume:   volume,
		Inserted: t,
	}
}

func TestMaxHeapLess(t *testing.T) {
	// Create test orders
	order1 := createTestOrder(1, 100.0, 10, "2023-01-01T00:00:00Z")
	order2 := createTestOrder(2, 100.0, 10, "2023-01-02T00:00:00Z")
	order3 := createTestOrder(3, 101.0, 10, "2023-01-03T00:00:00Z")

	// Simulate a small heap
	heap := MaxHeap{order1, order2, order3}

	// Test for price priority
	if !heap.Less(2, 0) {
		t.Errorf("Expected order3 with higher price to have higher priority than order1")
	}

	// Test for time priority with equal prices
	if !heap.Less(0, 1) {
		t.Errorf("Expected order1 (earlier) to have higher priority than order2 (later) when prices are equal")
	}
}

func TestMinHeapLess(t *testing.T) {
	// Create test orders with the same setup as MaxHeap for consistency
	order1 := createTestOrder(1, 100.0, 10, "2023-01-01T00:00:00Z")
	order2 := createTestOrder(2, 100.0, 10, "2023-01-02T00:00:00Z")
	order3 := createTestOrder(3, 99.0, 10, "2023-01-03T00:00:00Z")

	// Simulate a small heap
	heap := MinHeap{order1, order2, order3}

	// Test for price priority
	if !heap.Less(2, 0) {
		t.Errorf("Expected order3 with lower price to have higher priority than order1")
	}

	// Test for time priority with equal prices
	if !heap.Less(0, 1) {
		t.Errorf("Expected order1 (earlier) to have higher priority than order2 (later) when prices are equal")
	}
}

func TestComplexHeapOperations(t *testing.T) {
	ob := NewOrderBook()

	// Initial setup with three orders
	now := time.Now()
	ob.insertOrderIntoHeap(&Order{ID: 1, Symbol: "TEST", Side: "BUY", Price: 100.0, Volume: 10, Inserted: now.Add(-10 * time.Minute)})
	ob.insertOrderIntoHeap(&Order{ID: 2, Symbol: "TEST", Side: "BUY", Price: 105.0, Volume: 15, Inserted: now.Add(-5 * time.Minute)})
	ob.insertOrderIntoHeap(&Order{ID: 3, Symbol: "TEST", Side: "BUY", Price: 110.0, Volume: 5, Inserted: now})

	// Update order 1 to have the highest price, should move to top
	ob.Update(1, 115.0, 10) // Makes order 1 the top due to highest price

	// Decrease volume of order 2 without changing price, should not affect order
	ob.Update(2, 105.0, 5) // Volume decrease

	// Insert a new order with a price lower than the existing top but newer, should not become top
	ob.insertOrderIntoHeap(&Order{ID: 4, Symbol: "TEST", Side: "BUY", Price: 112.0, Volume: 10, Inserted: now.Add(1 * time.Minute)})

	// Remove order 3, the previously top order
	ob.removeOrderFromHeap(&Order{ID: 3})

	// Expected order in heap: ID 1 (Price 115), ID 4 (Price 112), ID 2 (Price 105) after removal and updates
	expectedOrderIDs := []int{1, 4, 2}
	for i, expectedID := range expectedOrderIDs {
		if (*ob.BuyOrders)[i].ID != expectedID {
			t.Errorf("After complex operations, expected order at position %d to have ID %d, got ID %d", i, expectedID, (*ob.BuyOrders)[i].ID)
		}
	}

	// Verify heap size to catch any potential issues with insertions or deletions not being handled correctly
	expectedHeapSize := 3
	if len(*ob.BuyOrders) != expectedHeapSize {
		t.Errorf("Expected BuyOrders heap size to be %d, found %d", expectedHeapSize, len(*ob.BuyOrders))
	}

	// Extra checks can be added here to verify specific scenarios or corner cases
}

func TestOrderUpdateScenario(t *testing.T) {
	ob := NewOrderBook()

	// Insert initial orders
	ob.Insert(&Order{ID: 1, Symbol: "FFLY", Side: "BUY", Price: 23.45, Volume: 10, Inserted: time.Now().Add(-10 * time.Minute)})
	ob.Insert(&Order{ID: 2, Symbol: "FFLY", Side: "SELL", Price: 23.50, Volume: 10, Inserted: time.Now().Add(-5 * time.Minute)})
	ob.Insert(&Order{ID: 3, Symbol: "FFLY", Side: "BUY", Price: 23.40, Volume: 5, Inserted: time.Now().Add(-15 * time.Minute)})
	ob.Insert(&Order{ID: 4, Symbol: "FFLY", Side: "SELL", Price: 23.55, Volume: 5, Inserted: time.Now()})

	// // check the order of the buy and sell orders (buy: highest price first, sell: lowest price first)
	// expectedBuyOrderIDs := []int{1, 3}
	// expectedSellOrderIDs := []int{4, 2}
	// for i, expectedID := range expectedBuyOrderIDs {
	// 	if (*ob.BuyOrders)[i].ID != expectedID {
	// 		t.Errorf("Expected buy order at position %d to have ID %d, got ID %d", i, expectedID, (*ob.BuyOrders)[i].ID)
	// 	}
	// }
	// for i, expectedID := range expectedSellOrderIDs {
	// 	if (*ob.SellOrders)[i].ID != expectedID {
	// 		t.Errorf("Expected sell order at position %d to have ID %d, got ID %d", i, expectedID, (*ob.SellOrders)[i].ID)
	// 	}
	// }

	// Update order to change price into a range where it can match, simulating a price drop in a SELL order
	ob.Update(2, 23.40, 10) // This should trigger a match with BUY order ID 1

	// Verify trades after the update
	expectedTrades := []string{"FFLY,23.45,10,2,1"}
	if !reflect.DeepEqual(ob.Trades, expectedTrades) {
		t.Errorf("Expected trades to match: %+v, got: %+v", expectedTrades, ob.Trades)
	}

	logOrderBookState(t, ob)                        // Custom function to log the order book content
	verifyOrderBookState(t, ob, []int{3}, []int{4}) // Custom function to verify the order book state

	// Update a BUY order to increase its price, potentially changing its position in the order book
	ob.Update(3, 23.50, 5) // No direct match since the best SELL is at 23.55 now

	// After this update, order ID 3 should be the highest priced BUY order.
	// Verify the new state of the order book
	logOrderBookState(t, ob) // Assuming this function logs the current state of the order book for debugging

	// Check that order ID 3 is now the top BUY order due to its increased price
	if (*ob.BuyOrders)[0].ID != 3 {
		t.Errorf("Expected top BUY order ID to be 3 after update, got %d", (*ob.BuyOrders)[0].ID)
	}

	// Since order ID 2 matched and was removed during the previous update, the only SELL order left should be ID 4
	if len(*ob.SellOrders) != 1 || (*ob.SellOrders)[0].ID != 4 {
		t.Errorf("Expected top SELL order ID to be 4, got %d", (*ob.SellOrders)[0].ID)
	}

	// Verify trades are still as expected after the second update
	expectedTradesAfterSecondUpdate := []string{"FFLY,23.45,10,2,1"}
	if !reflect.DeepEqual(ob.Trades, expectedTradesAfterSecondUpdate) {
		t.Errorf("Expected trades after second update to match: %+v, got: %+v", expectedTradesAfterSecondUpdate, ob.Trades)
	}

	// Insert a new SELL order with a price that could potentially match with the updated BUY order if the BUY order's price is increased further
	ob.Insert(&Order{ID: 5, Symbol: "FFLY", Side: "SELL", Price: 23.50, Volume: 5, Inserted: time.Now().Add(1 * time.Minute)})

	// Update the BUY order again, this time to a price that matches the new SELL order's price, triggering a match
	// (debug notes:) this one here means that this order should lose its priority and be placed at the end of the queue

	// ob.Update(4, 123.5, 5) // This should NOT trigger a match

	// Verify the new trades after the update
	expectedTradesAfterThirdUpdate := []string{
		"FFLY,23.45,10,2,1", // Only the initial trade
		"FFLY,23.5,5,5,3",
	}
	if !reflect.DeepEqual(ob.Trades, expectedTradesAfterThirdUpdate) {
		t.Errorf("Expected trades after third update to match: %+v, got: %+v", expectedTradesAfterThirdUpdate, ob.Trades)
	}

	// Verify the updated state of the order book after the match
	logOrderBookState(t, ob) // Assuming this function logs the current state of the order book for debugging
	// After the trade, the BUY side should only have order ID 3 removed (since it matched and was fully filled)
	// The SELL side should now only have order ID 4 remaining
	verifyOrderBookState(t, ob, []int{}, []int{4}) // Assuming this function verifies the current state of the order book

	// Insert another BUY order with a price higher than the remaining SELL order to test immediate matching
	ob.Insert(&Order{ID: 6, Symbol: "FFLY", Side: "BUY", Price: 23.60, Volume: 5, Inserted: time.Now().Add(2 * time.Minute)})

	// This new BUY order should immediately match with the remaining SELL order ID 4
	expectedTradesAfterInsert := []string{
		"FFLY,23.45,10,2,1",
		"FFLY,23.5,5,5,3",
		"FFLY,23.6,5,6,4", // This trade results from the immediate match of the new BUY order with the existing SELL order
	}
	if !reflect.DeepEqual(ob.Trades, expectedTradesAfterInsert) {
		t.Errorf("Expected trades after new BUY order insert to match: %+v, got: %+v", expectedTradesAfterInsert, ob.Trades)
	}

	// Finally, verify the order book is empty on both sides after all matching operations
	// verifyOrderBookIsEmpty(t, ob)
}

func TestComplexOrderFlowTestCase5(t *testing.T) { // failing
	ob := NewOrderBook()

	// Insert initial orders
	ob.Insert(&Order{ID: 1, Symbol: "FFLY", Side: "BUY", Price: 45.95, Volume: 5})
	ob.Insert(&Order{ID: 2, Symbol: "FFLY", Side: "BUY", Price: 45.95, Volume: 6})
	ob.Insert(&Order{ID: 3, Symbol: "FFLY", Side: "BUY", Price: 45.95, Volume: 12})
	ob.Insert(&Order{ID: 4, Symbol: "FFLY", Side: "SELL", Price: 46, Volume: 8})

	// Update order 2 to match sell order at price 46
	ob.Update(2, 46, 3) // This should trigger a match with sell order ID 4

	// Insert sell orders at 45.95
	ob.Insert(&Order{ID: 5, Symbol: "FFLY", Side: "SELL", Price: 45.95, Volume: 1})

	ob.Update(1, 45.95, 3) // Reduce volume of order 1. Safe update, shouldn't change anything. In-place update

	ob.Insert(&Order{ID: 6, Symbol: "FFLY", Side: "SELL", Price: 45.95, Volume: 1}) // this should trigger a match with order 1 and 6

	ob.Update(1, 45.95, 5) // Increase volume back of order 1, from 3 to 5 (5, 4, 3, 5). OrderID 1 will lose its priority

	// When Order 7 is inserted, it matches with an existing BUY order.
	// Order 3 should be the maker since it has the highest volume among the remaining BUY orders at the same price level (45.95).

	ob.Insert(&Order{ID: 7, Symbol: "FFLY", Side: "SELL", Price: 45.95, Volume: 1}) // the heap order should be 3, 1

	// Expected trades and order book state verification
	expectedTrades := []string{
		"FFLY,46,3,2,4",    // Correct, as Order 2 becomes the taker by updating to match Order 4's price.
		"FFLY,45.95,1,5,1", // Order 5 triggers the trade as a new order, making it the taker, and Order 1 is the maker.
		"FFLY,45.95,1,6,1", // Similar logic for Order 6 as a taker and Order 1 as a maker.
		"FFLY,45.95,1,7,3", // Order 7 triggers the trade as a new order, making it the taker, and Order 3 is the maker.
	}

	if !reflect.DeepEqual(ob.Trades, expectedTrades) {
		t.Errorf("Expected trades to match: %+v, got: %+v", expectedTrades, ob.Trades)
	}

	// Further verification steps for order book state can be added here
	// For example: verifyOrderBookState(t, ob, expectedBuyOrders, expectedSellOrders)
}

func TestDetailedOrderBookOps(t *testing.T) {
	ob := NewOrderBook()

	// Insert initial orders
	ob.Insert(&Order{ID: 1, Symbol: "FFLY", Side: "BUY", Price: 23.45, Volume: 10, Inserted: time.Now().Add(-10 * time.Minute)})
	ob.Insert(&Order{ID: 2, Symbol: "FFLY", Side: "SELL", Price: 23.50, Volume: 10, Inserted: time.Now().Add(-5 * time.Minute)})
	ob.Insert(&Order{ID: 3, Symbol: "FFLY", Side: "BUY", Price: 23.40, Volume: 5, Inserted: time.Now().Add(-15 * time.Minute)})
	ob.Insert(&Order{ID: 4, Symbol: "FFLY", Side: "SELL", Price: 23.55, Volume: 5, Inserted: time.Now()})

	// Log the order book
	ob.LogHeapContents(t)

	// check the order of the buy and sell orders (buy: highest price first, sell: lowest price first)

	expectedBuyOrderIDs := []int{1, 3}
	expectedSellOrderIDs := []int{4, 2}
	for i, expectedID := range expectedBuyOrderIDs {
		if (*ob.BuyOrders)[i].ID != expectedID {
			t.Errorf("Expected buy order at position %d to have ID %d, got ID %d", i, expectedID, (*ob.BuyOrders)[i].ID)
		}
	}
	for i, expectedID := range expectedSellOrderIDs {
		if (*ob.SellOrders)[i].ID != expectedID {
			t.Errorf("Expected sell order at position %d to have ID %d, got ID %d", i, expectedID, (*ob.SellOrders)[i].ID)
		}
	}
}

func (ob *OrderBook) LogHeapContents(t *testing.T) {
	// Make a copy of the buy orders heap to preserve the original order
	buyOrdersCopy := make(PriorityQueue, len(*ob.BuyOrders))
	copy(buyOrdersCopy, *ob.BuyOrders)
	heap.Init(&buyOrdersCopy) // Ensure the copy is a valid heap

	t.Log("Buy Orders (in priority order):")
	for buyOrdersCopy.Len() > 0 {
		order := heap.Pop(&buyOrdersCopy).(*Order)
		t.Logf("ID=%d, Price=%.2f, Volume=%d, Inserted=%v", order.ID, order.Price, order.Volume, order.Inserted)
	}

	// Repeat the process for sell orders
	sellOrdersCopy := make(PriorityQueue, len(*ob.SellOrders))
	copy(sellOrdersCopy, *ob.SellOrders)
	heap.Init(&sellOrdersCopy) // Ensure the copy is a valid heap

	t.Log("Sell Orders (in priority order):")
	for sellOrdersCopy.Len() > 0 {
		order := heap.Pop(&sellOrdersCopy).(*Order)
		t.Logf("ID=%d, Price=%.2f, Volume=%d, Inserted=%v", order.ID, order.Price, order.Volume, order.Inserted)
	}
}

func TestOrderInsertionAndMatching(t *testing.T) {
	ob := NewOrderBook()

	// Insert buy orders
	ob.Insert(&Order{ID: 1, Symbol: "FFLY", Side: "BUY", Price: 23.45, Volume: 10})
	ob.Insert(&Order{ID: 3, Symbol: "FFLY", Side: "BUY", Price: 23.40, Volume: 5})

	// Insert sell orders
	ob.Insert(&Order{ID: 2, Symbol: "FFLY", Side: "SELL", Price: 23.50, Volume: 10})
	ob.Insert(&Order{ID: 4, Symbol: "FFLY", Side: "SELL", Price: 23.55, Volume: 5})

	// Attempt to match orders
	// Assuming automatic matching occurs upon insertion

	// Log the state after all insertions
	ob.LogHeapContents(t) // Custom function to log the order book content

	// Expected trades should be empty since inserted sell orders have higher prices than buy orders
	if len(ob.Trades) != 0 {
		t.Errorf("Expected 0 trades, got %d", len(ob.Trades))
	}

	// Insert a sell order that matches the highest buy order
	ob.Insert(&Order{ID: 5, Symbol: "FFLY", Side: "SELL", Price: 23.45, Volume: 5})

	// Check for executed trade
	if len(ob.Trades) != 1 {
		t.Errorf("Expected 1 trade after matching sell order, got %d", len(ob.Trades))
	} else {
		expectedTradeDetail := "FFLY,23.45,5,5,1" // Format: Symbol,Price,Volume,TakerOrderID,MakerOrderID
		if ob.Trades[0] != expectedTradeDetail {
			t.Errorf("Expected trade detail %s, got %s", expectedTradeDetail, ob.Trades[0])
		}
	}

	// Log the state after matching
	ob.LogHeapContents(t)
}

func logOrderBookState(t *testing.T, ob *OrderBook) {
	t.Log("Order Book State after updates:")
	t.Log("Buy Orders:")
	for _, order := range *ob.BuyOrders {
		t.Logf("ID=%d, Symbol=%s, Side=%s, Price=%.2f, Volume=%d, Inserted=%v", order.ID, order.Symbol, order.Side, order.Price, order.Volume, order.Inserted)
	}
	t.Log("Sell Orders:")
	for _, order := range *ob.SellOrders {
		t.Logf("ID=%d, Symbol=%s, Side=%s, Price=%.2f, Volume=%d, Inserted=%v", order.ID, order.Symbol, order.Side, order.Price, order.Volume, order.Inserted)
	}
}

func verifyOrderBookState(t *testing.T, ob *OrderBook, expectedBuyOrderIDs, expectedSellOrderIDs []int) {
	// Verify Buy Orders
	if len(*ob.BuyOrders) != len(expectedBuyOrderIDs) {
		t.Errorf("Expected %d buy orders, found %d", len(expectedBuyOrderIDs), len(*ob.BuyOrders))
	} else {
		for i, expectedID := range expectedBuyOrderIDs {
			if (*ob.BuyOrders)[i].ID != expectedID {
				t.Errorf("Expected buy order at position %d to have ID %d, got ID %d", i, expectedID, (*ob.BuyOrders)[i].ID)
			}
		}
	}

	// Verify Sell Orders
	if len(*ob.SellOrders) != len(expectedSellOrderIDs) {
		t.Errorf("Expected %d sell orders, found %d", len(expectedSellOrderIDs), len(*ob.SellOrders))
	} else {
		for i, expectedID := range expectedSellOrderIDs {
			if (*ob.SellOrders)[i].ID != expectedID {
				t.Errorf("Expected sell order at position %d to have ID %d, got ID %d", i, expectedID, (*ob.SellOrders)[i].ID)
			}
		}
	}
}
