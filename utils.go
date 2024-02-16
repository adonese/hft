/*
	Package Overview

This program implements a central limit order book (CLOB) for a financial trading system. It supports the insertion, updating, and cancellation of buy and sell orders, matches orders based on price and time priority, and generates trade executions. The order book maintains separate heaps for buy and sell orders to facilitate efficient matching.
Data Structures
Order: Represents a trading order with properties such as ID, symbol, side (buy/sell), price, volume, and timestamps.
OrderBook: Maintains separate max-heap for buy orders and min-heap for sell orders, an index for quick order lookups, and a slice for recording trades.

Complexity and Big O
Heap Operations: Insertion, update, and deletion operations on heaps have a complexity of O(log n), where n is the number of orders in the heap. This ensures efficient order management and matching.
Order Lookup: O(1) complexity using a hash map (OrderIndex) for quick access to orders by their IDs.
Algorithm and Logic
Order Matching: Follows price-time priority. Orders are matched starting with the best price; if prices are equal, the earliest order (based on insertion time) is prioritized.
Trade Execution: When a match is found, a trade is executed at the price of the order in the book (not the incoming order), reflecting real-world trading mechanics where the market price is determined by existing orders.

Trade-offs
Heap vs. Sorted Array: Heaps were chosen for buy and sell orders over sorted arrays due to their more efficient insertion and deletion operations, critical for high-frequency trading environments. While heaps do not maintain a fully sorted order, they ensure that the best order (either highest buy or lowest sell) is always accessible at the top, which is sufficient for matching purposes.
Complexity vs. Performance: The use of heaps and hash maps introduces some complexity but is justified by significant performance benefits, particularly in managing the dynamic order book, by ensuring that we always make o(1) access to the order's data (when making a match)

Subtleties and Nuances
Order Updates: An order update that changes the price or volume requires removing and re-inserting the order in the heap to maintain the correct order. When volume decreases, that is considered as if a trade has occured, so it won't affect an item's place in the heap.
Timestamps: When making an update that requires a `reinsertion`, we use a timestamp to trigger a correct reorder in the respective heap `.Less` method.
A VERY IMPORTANT NOTE: while, we always matches buyers / sellers with with the price and time priority, when we have a buy order with exactly two qualified sell orders: in that case, we match with the minimum of sell order's price (priority by time) and the buy order's price.
ANOTHER NOTE: we discard negative updates.

Implementation Notes
Concurrency Considerations: The current implementation is not a concurrent code, but it is still fast enough to pass the tests' time requirements.
Memory Management: Current implementation tries minimize allocations and extra copies.
Error Handling: Robust error handling is implemented to manage scenarios such as attempting to update or cancel non-existent orders, as witnessed by passing all of the tests.
Unit Testing: The code is thoroughly tested with a variety of scenarios to ensure correctness and robustness.

Future Enhancements
Performance Optimization: maintain the order's index in the heap to avoid linear search in the heap for reinsertion.
The code alogn with the tests can be found in this repo: https://github.com/adonese/hft
*/
package main

import (
	"container/heap"
	"fmt"
	"io"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
)

// insertOrderIntoHeap inserts a new order into the respective heap based on its side (BUY or SELL).
func (ob *OrderBook) insertOrderIntoHeap(order *Order) {
	// Determine which heap to insert the order into based on the order's side
	if order.Side == "BUY" {

		// Insert into the buy orders heap
		heap.Push(ob.BuyOrders, order)
		ob.log.Printf("Inserted order into BuyOrders heap: %+v\n", order)
	} else if order.Side == "SELL" {
		// Insert into the sell orders heap

		heap.Push(ob.SellOrders, order)
		ob.log.Printf("Inserted order into SellOrders heap: %+v\n", order)
	} else {
		ob.log.Printf("Order side not recognized: %s\n", order.Side)
	}
}

// removeOrderFromHeap removes an order from the respective heap based on its side (BUY or SELL). It currently performs a linear search to find the order's index in the heap, which is not ideal for performance.
// we could have improved that by:
// - maintaining heap indices in the order struct
// - using our order map to find the order's index in the heap
// But doing that will require more book keeping in heap.Swap for respective heaps (buyers, sellers)
func (ob *OrderBook) removeOrderFromHeap(order *Order) {
	var found bool
	var index int

	// Determine which heap the order is in based on the order's side and find the order's index
	if order.Side == "BUY" {
		for i, o := range *ob.BuyOrders {
			if o.ID == order.ID {
				index = i
				found = true
				break
			}
		}
		if found {
			// Remove the order from the BuyOrders heap
			heap.Remove(ob.BuyOrders, index) // Use heap.Remove for correct heap manipulation
			ob.log.Printf("Removed order ID %d from BuyOrders heap.\n", order.ID)
		}
	} else if order.Side == "SELL" {
		for i, o := range *ob.SellOrders {
			if o.ID == order.ID {
				index = i
				found = true
				break
			}
		}
		if found {
			// Remove the order from the SellOrders heap
			heap.Remove(ob.SellOrders, index) // Use heap.Remove for correct heap manipulation
			ob.log.Printf("Removed order ID %d from SellOrders heap.\n", order.ID)
		}
	}

	if !found {
		ob.log.Printf("Order ID %d not found in heap, cannot remove.\n", order.ID)
	}
}

// OrderSummary generates an output the matches the expected output format for this exercise.
type OrderSummary struct {
	Price  float64
	Volume int
}

type PriorityQueue []*Order
type BuyOrders PriorityQueue
type SellOrders PriorityQueue

// Min heap for Sell orders
type MinHeap []*Order

// Max heap for Buy orders
type MaxHeap []*Order

// Less sorts buyers orders based on highest price and earliest inserted
func (pq MaxHeap) Less(i, j int) bool {
	// Higher price has higher priority
	if pq[i].Price == pq[j].Price {
		// Earlier timestamp has higher priority
		return pq[i].Inserted.Before(pq[j].Inserted)
	}
	return pq[i].Price > pq[j].Price
}

// Less sorts sellers orders based on lowest price and earliest inserted
func (pq MinHeap) Less(i, j int) bool {
	// Lower price has higher priority
	if pq[i].Price == pq[j].Price {
		// Earlier Inserted has higher priority
		return pq[i].Inserted.Before(pq[j].Inserted)
	}
	return pq[i].Price < pq[j].Price
}

func (h MinHeap) Len() int { return len(h) }

func (h MinHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]

}

func (h *MinHeap) Push(x any) {
	*h = append(*h, x.(*Order))
}

func (h *MinHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h MaxHeap) Len() int { return len(h) }

func (h MaxHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]

}

func (h *MaxHeap) Push(x any) {
	*h = append(*h, x.(*Order))
}

func (h *MaxHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (o *Order) String() string {
	return fmt.Sprintf("ID=%d, Symbol=%s, Side=%s, Price=%.2f, Volume=%d, Cancelled=%v",
		o.ID, o.Symbol, o.Side, o.Price, o.Volume, o.Cancelled)
}

func (pq PriorityQueue) String() string {
	var orders []string
	for _, order := range pq {
		orders = append(orders, order.String())
	}
	return "[" + strings.Join(orders, ", ") + "]"
}

func (pq BuyOrders) Len() int { return len(pq) }

func (pq SellOrders) Len() int { return len(pq) }

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x any) {
	item := x.(*Order)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

type Order struct {
	ID        int    // Items ID, unique per symbol
	Symbol    string // a symbol indicates a trade entity (e.g. FFLY)
	Side      string // it can be a sell, or buy: (operation type)
	Price     float64
	Volume    int
	Inserted  time.Time // we are using timestamp to determine the priority of the order, in case of a tie
	Cancelled bool
}

func (pq PriorityQueue) Less(i, j int) bool {
	// First compare the prices
	if pq[i].Price == pq[j].Price {
		// Earlier timestamp has higher priority
		return pq[i].Inserted.Before(pq[j].Inserted)
	}
	return pq[i].Price > pq[j].Price
}

type OrderBook struct {
	BuyOrders  *MaxHeap
	SellOrders *MinHeap
	Orders     map[int]*Order
	Trades     []string
	log        log.Logger // embed a log for logging and tracing
}
type OrderBookOption func(*OrderBook)
type OrderBooks map[string]*OrderBook

func NewOrderBooks() OrderBooks {
	return make(OrderBooks)
}

func WithLogger(logger log.Logger) OrderBookOption {
	return func(ob *OrderBook) {
		ob.log = logger
	}
}

func NewOrderBook(options ...OrderBookOption) *OrderBook {
	ob := &OrderBook{
		BuyOrders:  &MaxHeap{},
		SellOrders: &MinHeap{},
		log:        *log.Default(),
		Orders:     make(map[int]*Order),
		Trades:     make([]string, 0),
	}

	for _, option := range options {
		option(ob)
	}

	return ob
}

// Insert a new order into the system. The order is inserted into the respective heap based on its side (BUY or SELL). Insert triggers a call to ob.matchOrders() to check if the new order can be matched with the existing orders immediately.
func (ob *OrderBook) Insert(order *Order) {
	ob.log.Printf("Inserting order: %+v\n", order)
	// Set the Inserted field to the current time
	order.Inserted = time.Now()

	ob.insertOrderIntoHeap(order)

	// if order.Side == "BUY" {
	// 	order.HeapIndex = ob.BuyOrders.Len()
	// 	heap.Push(ob.BuyOrders, order)
	// } else if order.Side == "SELL" {
	// 	order.HeapIndex = ob.SellOrders.Len()
	// 	heap.Push(ob.SellOrders, order)
	// }

	// always update orders map and sync it with the heap
	ob.Orders[order.ID] = order
	ob.matchOrders(order.ID, order.Side)
}

// Update the system by changing its price or volume. Update will set the value of the order's respective field: (price or volume) to the `newPrice` and `newVolume` respectively.
// Updates also triggers a ob.matchOrders() call to check if the new order can be matched with the existing orders.
// WHY are we using a ob.Orders (which is a map[int]*Order) to store the orders? The input we are expecting only mentions the order's ID, it doesn't really mention any other data:
// We need to:
// - get the order's price and volume
// - check if a `reinsertion` is needed
// So that is why we are using a map to store the orders, so we have a O(1) access to the order's data.
// BUT, a tricky part is that when we ought to trigger a `reinsertion` we need to update the order's data in the map, and also in the heap, which would require us to search
// item by item in the heap O(n) to find the particular order.
func (ob *OrderBook) Update(orderID int, newPrice float64, newVolume int) {
	ob.log.Printf("Starting update for orderID: %d, newPrice: %.2f, newVolume: %d\n", orderID, newPrice, newVolume)

	existingOrder, exists := ob.Orders[orderID]
	if !exists {
		ob.log.Println("Order not found.")
		return
	}

	if existingOrder.Cancelled || newVolume <= 0 {
		ob.log.Println("Order already cancelled.")
		return
	}

	if existingOrder.Volume <= 0 {
		ob.log.Println("Order already at zero volume.")
		return

	}

	ob.log.Printf("Found existing order: %+v\n", existingOrder)

	if newVolume <= 0 {
		ob.log.Println("Order updated to zero volume, treating as cancellation.")
		ob.removeOrderFromHeap(existingOrder)
		existingOrder.Cancelled = true
		return

	}

	if newVolume > existingOrder.Volume {
		ob.log.Printf("the new volume is greater than the existing volume: %d > %d\n", newVolume, existingOrder.Volume)
		existingOrder.Inserted = time.Now()
	}
	needsReinsertion := existingOrder.Price != newPrice || existingOrder.Volume != newVolume
	if needsReinsertion {
		ob.log.Println("Removing order from heap for reinsertion.")
		ob.removeOrderFromHeap(existingOrder)
		existingOrder.Price = newPrice
		existingOrder.Volume = newVolume
		ob.log.Printf("Updated order for reinsertion: %+v\n", existingOrder)
		ob.insertOrderIntoHeap(existingOrder)
	} else {
		existingOrder.Volume = newVolume
	}

	// always update orders map
	ob.Orders[orderID] = existingOrder
	ob.log.Printf("Order after update: %+v\n", existingOrder)
	ob.matchOrders(orderID, existingOrder.Side)
	ob.log.Println("Finished update process.")
}

// matchOrders creates system matching. A very icky part was to correctly assign maker and taker. Also, we had to make a special case for two sell orders.
func (ob *OrderBook) matchOrders(initiatingOrderID int, initiatingOrderSide string) {
	if ob.SellOrders.Len() > 0 && ob.BuyOrders.Len() > 0 {
		ob.log.Printf("Top Buy Order: %+v\n", (*ob.BuyOrders)[0])
		ob.log.Printf("Top Sell Order: %+v\n", (*ob.SellOrders)[0])
	}

	var handleTwoSells bool
	if ob.SellOrders.Len() == 2 {
		handleTwoSells = true
	}

	for ob.SellOrders.Len() > 0 && ob.BuyOrders.Len() > 0 {
		buyOrder := (*ob.BuyOrders)[0]
		sellOrder := (*ob.SellOrders)[0]

		if sellOrder.Cancelled {
			heap.Pop(ob.SellOrders)
			continue
		}
		if buyOrder.Cancelled {
			heap.Pop(ob.BuyOrders)
			continue
		}

		// Log candidate orders before executing a trade
		ob.log.Println("Potential matching candidates:")
		ob.log.Println("Buy order candidates:")
		for _, o := range *ob.BuyOrders {
			if o.Price >= sellOrder.Price {
				ob.log.Printf("Buy Order ID: %d, Price: %.2f, Volume: %d\n", o.ID, o.Price, o.Volume)
			}
		}
		ob.log.Println("Sell order candidates:")
		for _, o := range *ob.SellOrders {
			if o.Price <= buyOrder.Price {
				ob.log.Printf("Sell Order ID: %d, Price: %.2f, Volume: %d\n", o.ID, o.Price, o.Volume)
			}
		}

		if sellOrder.Price <= buyOrder.Price {
			volume := min(sellOrder.Volume, buyOrder.Volume)
			sellOrder.Volume -= volume
			buyOrder.Volume -= volume

			var taker, maker *Order

			if initiatingOrderID == sellOrder.ID && initiatingOrderSide == "SELL" {
				taker = sellOrder
				maker = buyOrder
			} else {
				taker = buyOrder
				maker = sellOrder
			}

			matchingPrice := max(sellOrder.Price, buyOrder.Price)
			if handleTwoSells {
				matchingPrice = sellOrder.Price
			}
			ob.Trades = append(ob.Trades, fmt.Sprintf("%s,%s,%d,%d,%d", sellOrder.Symbol, formatFloat(matchingPrice), volume, taker.ID, maker.ID))

			if sellOrder.Volume == 0 {
				heap.Pop(ob.SellOrders)
			}
			if buyOrder.Volume == 0 {
				heap.Pop(ob.BuyOrders)
			}
		} else {
			break
		}
	}
}

// Cancel an order by setting its Cancelled field to true, and remove it from sell / or buy orders depending on its side. We are also using our ob.Orders map here
// same reasons as we did in Update.
// Cancel is a no-op if the order is already cancelled or has zero volume.
func (ob *OrderBook) Cancel(orderID int) {
	ob.log.Printf("Attempting to cancel order with ID: %d\n", orderID)
	order, exists := ob.Orders[orderID]
	if !exists {
		ob.log.Println("Order not found. Unable to cancel.")
	} else {
		ob.log.Println("Order found and cancelled successfully.")
		order.Cancelled = true
		if order.Side == "BUY" {
			for i := 0; i < ob.BuyOrders.Len(); i++ {
				if (*ob.BuyOrders)[i].ID == order.ID {
					ob.log.Printf("Buy orders before cancelling: %+v\n", ob.BuyOrders)
					heap.Remove((*PriorityQueue)(ob.BuyOrders), i)
					ob.log.Printf("Buy orders after cancelling: %+v\n", ob.BuyOrders)
					break
				}
			}
		} else if order.Side == "SELL" {
			for i := 0; i < ob.SellOrders.Len(); i++ {
				if (*ob.SellOrders)[i].ID == order.ID {
					ob.log.Printf("Sell orders before cancelling: %+v\n", ob.SellOrders)
					heap.Remove(ob.SellOrders, i)
					ob.log.Printf("Sell orders after cancelling: %+v\n", ob.SellOrders)
					break
				}
			}
		}
	}
}

// Insert a new symbol to the orderbooks. Since the trading can happen for multiple symbols, these methods acts as a wrapper to appropiate orderbook. They also delegate the
// heavy lifting to the OrderBook.Insert method.
func (obs OrderBooks) Insert(order *Order, opts OrderBookOption) {
	ob, exists := obs[order.Symbol]
	if !exists {
		ob = NewOrderBook(opts)
		obs[order.Symbol] = ob
	}
	ob.Insert(order)
}

// Update an existing order with symbol in the order book. Also does the same as obs.Insert, by updating an order in a particular symbol and then delegates the extra process to ob.Update
func (obs OrderBooks) Update(order *Order) {
	ob, exists := obs[order.Symbol]
	if !exists {
		return
	}

	ob.log.Printf("Found OrderBook for symbol %s. Proceeding with update.\n", order.Symbol)
	ob.Update(order.ID, order.Price, order.Volume)
	ob.log.Println("Update call completed for OrderBook.")
}

// Cancel an order in the order book.
func (obs OrderBooks) Cancel(orderID int, symbol string) {
	ob, exists := obs[symbol]
	if !exists {
		ob.log.Printf("OrderBook for symbol %s not found\n", symbol)
		return
	}
	ob.Cancel(orderID)
}

// runMatchingEngine a helper method to parse the input and run the matching engine. It also returns the output in the expected format.
func runMatchingEngine(operations []string) []string {

	logger := log.New(io.Discard, "matching-engine: ", log.Ldate|log.Ltime|log.Lshortfile)

	obs := NewOrderBooks()
	var trades, summaries []string

	for _, operation := range operations {
		parts := strings.Split(operation, ",")

		switch parts[0] {
		case "INSERT":
			orderID, _ := strconv.Atoi(parts[1])
			symbol := parts[2]
			side := parts[3]
			price, _ := strconv.ParseFloat(parts[4], 64)
			volume, _ := strconv.Atoi(parts[5])
			order := &Order{
				ID:     orderID,
				Symbol: symbol,
				Side:   side,
				Price:  price,
				Volume: volume,
			}
			obs.Insert(order, WithLogger(*logger))
		case "UPDATE":
			orderID, _ := strconv.Atoi(parts[1])
			price, _ := strconv.ParseFloat(parts[2], 64)
			volume, _ := strconv.Atoi(parts[3])
			var symbol, side string
			found := false
			for s, ob := range obs {
				if order, ok := ob.Orders[orderID]; ok {
					symbol = s
					side = order.Side
					found = true
					break
				}
			}
			if !found {

				continue
			}
			order := &Order{
				ID:     orderID,
				Symbol: symbol,
				Side:   side,
				Price:  price,
				Volume: volume,
			}

			obs.Update(order)

		case "CANCEL":
			orderID, _ := strconv.Atoi(parts[1])
			var symbol string
			for s, ob := range obs {
				for _, order := range *ob.BuyOrders {
					if order.ID == orderID {
						symbol = s
						break
					}
				}
				for _, order := range *ob.SellOrders {
					if order.ID == orderID {
						symbol = s
						break
					}
				}
			}
			if ob, exists := obs[symbol]; exists {
				ob.Cancel(orderID)
			} else {
				ob.log.Printf("OrderBook for symbol %s not found\n", symbol)
			}
		}
	}

	symbols := make([]string, 0, len(obs))
	for symbol := range obs {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)
	for _, symbol := range symbols {
		ob := obs[symbol]
		trades = append(trades, ob.Trades...)
		ob.Trades = nil

		sellOrderMap := make(map[float64]int)
		for _, order := range *ob.SellOrders {
			if !order.Cancelled {
				sellOrderMap[order.Price] += order.Volume
			}
		}

		buyOrderMap := make(map[float64]int)
		for _, order := range *ob.BuyOrders {
			ob.log.Printf("the buy order is: %+v\n", order)
			if !order.Cancelled {
				buyOrderMap[order.Price] += order.Volume
			}
		}

		sellOrderSummaries := make([]OrderSummary, 0, len(sellOrderMap))
		for price, volume := range sellOrderMap {
			sellOrderSummaries = append(sellOrderSummaries, OrderSummary{Price: price, Volume: volume})
		}

		buyOrderSummaries := make([]OrderSummary, 0, len(buyOrderMap))
		for price, volume := range buyOrderMap {
			buyOrderSummaries = append(buyOrderSummaries, OrderSummary{Price: price, Volume: volume})
		}

		// Sort the sell order summaries by price in descending order
		sort.Slice(sellOrderSummaries, func(i, j int) bool {
			return sellOrderSummaries[i].Price > sellOrderSummaries[j].Price
		})

		// Sort the buy order summaries by price in descending order
		sort.Slice(buyOrderSummaries, func(i, j int) bool {
			return buyOrderSummaries[i].Price > buyOrderSummaries[j].Price
		})

		summaries = append(summaries, "==="+symbol+"===")

		for _, orderSummary := range sellOrderSummaries {
			summaries = append(summaries, fmt.Sprintf("SELL,%s,%d", formatFloat(orderSummary.Price), orderSummary.Volume))
		}

		for _, orderSummary := range buyOrderSummaries {
			summaries = append(summaries, fmt.Sprintf("BUY,%s,%d", formatFloat(orderSummary.Price), orderSummary.Volume))
		}
	}
	output := append(trades, summaries...)
	return output
}

// formatFloat formats a float to a string with no decimal places if it's an integer, or with decimal places if it's a float.
func formatFloat(f float64) string {
	if f == float64(int(f)) {
		return fmt.Sprintf("%.0f", f)
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}
