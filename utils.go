package main

import (
	"container/heap"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (ob *OrderBook) insertOrderIntoHeap(order *Order) {
	// Determine which heap to insert the order into based on the order's side
	if order.Side == "BUY" {

		// Insert into the buy orders heap
		heap.Push(ob.BuyOrders, order)
		fmt.Printf("Inserted order into BuyOrders heap: %+v\n", order)
	} else if order.Side == "SELL" {
		// Insert into the sell orders heap
		heap.Push(ob.SellOrders, order)
		fmt.Printf("Inserted order into SellOrders heap: %+v\n", order)
	} else {
		fmt.Printf("Order side not recognized: %s\n", order.Side)
	}
}

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
			fmt.Printf("Removed order ID %d from BuyOrders heap.\n", order.ID)
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
			fmt.Printf("Removed order ID %d from SellOrders heap.\n", order.ID)
		}
	}

	if !found {
		fmt.Printf("Order ID %d not found in heap, cannot remove.\n", order.ID)
	}
}

type OrderSummary struct {
	Price  float64
	Volume int
}

type PriorityQueue []*Order
type BuyOrders PriorityQueue
type SellOrders PriorityQueue

// Min heap for Sell orders
type MinHeap []*Order

type MaxHeap PriorityQueue

func (pq MaxHeap) Less(i, j int) bool {
	// Higher price has higher priority
	if pq[i].Price == pq[j].Price {
		// Earlier timestamp has higher priority
		return pq[i].Inserted.Before(pq[j].Inserted)
	}
	return pq[i].Price > pq[j].Price
}

func (pq MinHeap) Less(i, j int) bool {
	// Lower price has higher priority
	if pq[i].Price == pq[j].Price {
		// Earlier Inserted has higher priority
		return pq[i].Inserted.Before(pq[j].Inserted)
	}
	return pq[i].Price < pq[j].Price
}

func (h MinHeap) Len() int { return len(h) }

func (h MinHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *MinHeap) Push(x interface{}) {
	*h = append(*h, x.(*Order))
}

func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h MaxHeap) Len() int { return len(h) }

func (h MaxHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *MaxHeap) Push(x interface{}) {
	*h = append(*h, x.(*Order))
}

func (h *MaxHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (pq BuyOrders) Less(i, j int) bool {
	// First compare the prices
	if pq[i].Price == pq[j].Price {
		// Earlier timestamp has higher priority
		return pq[i].Inserted.Before(pq[j].Inserted)
	}
	return pq[i].Price > pq[j].Price
}

func (pq SellOrders) Less(i, j int) bool {
	if pq[i].Price == pq[j].Price {
		// Earlier Inserted has higher priority
		return pq[i].Inserted.Before(pq[j].Inserted)
	}
	return pq[i].Price < pq[j].Price
}

func (o *Order) String() string {
	return fmt.Sprintf("ID=%d, Symbol=%s, Side=%s, Price=%.2f, Volume=%d, Updated=%v, Cancelled=%v",
		o.ID, o.Symbol, o.Side, o.Price, o.Volume, o.Updated, o.Cancelled)
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

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*Order)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

type Order struct {
	ID              int
	Symbol          string
	Side            string
	Price           float64
	Volume          int
	Inserted        time.Time
	Updated         bool
	Cancelled       bool
	IsUpdated       bool
	SeqNum          int
	Processed       bool
	VolumeDecreased bool
	Matched         bool
	MatchedVolume   int
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
	BuyOrders  *MaxHeap // changed to MaxHeap
	SellOrders *MinHeap // changed to MinHeap
	OrderIndex map[int]int
	Orders     map[int]*Order
	Trades     []string
}

type OrderBooks map[string]*OrderBook

func NewOrderBooks() OrderBooks {
	return make(OrderBooks)
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		BuyOrders:  &MaxHeap{},
		SellOrders: &MinHeap{},
		OrderIndex: make(map[int]int),
		Orders:     make(map[int]*Order),
		Trades:     make([]string, 0),
	}
}

var seqNumCounter int

func (ob *OrderBook) Insert(order *Order) {
	fmt.Printf("Inserting order: %+v\n", order)
	order.SeqNum = seqNumCounter
	seqNumCounter++
	order.Updated = false
	// Set the Inserted field to the current time
	order.Inserted = time.Now()

	if order.Side == "BUY" {
		heap.Push(ob.BuyOrders, order)
	} else if order.Side == "SELL" {
		heap.Push(ob.SellOrders, order)
	}

	ob.Orders[order.ID] = order
	ob.matchOrders(order.ID, order.Side)

}

func (ob *OrderBook) Update(orderID int, newPrice float64, newVolume int) {
	fmt.Printf("Starting update for orderID: %d, newPrice: %.2f, newVolume: %d\n", orderID, newPrice, newVolume)

	existingOrder, exists := ob.Orders[orderID]
	if !exists {
		fmt.Println("Order not found.")
		return
	}

	if existingOrder.Cancelled {
		fmt.Println("Order already cancelled.")
		return
	}

	fmt.Printf("Found existing order: %+v\n", existingOrder)

	if newVolume == 0 {
		fmt.Println("Order updated to zero volume, treating as cancellation.")
		ob.removeOrderFromHeap(existingOrder)
		existingOrder.Cancelled = true
		return

	}
	if newVolume > existingOrder.Volume {
		fmt.Printf("the new volume is greater than the existing volume: %d > %d\n", newVolume, existingOrder.Volume)
		existingOrder.Inserted = time.Now()
	}
	needsReinsertion := existingOrder.Price != newPrice || existingOrder.Volume != newVolume
	if needsReinsertion {
		fmt.Println("Removing order from heap for reinsertion.")
		ob.removeOrderFromHeap(existingOrder)
		existingOrder.Price = newPrice
		existingOrder.Volume = newVolume

		fmt.Printf("Updated order for reinsertion: %+v\n", existingOrder)
		ob.insertOrderIntoHeap(existingOrder)
		ob.Orders[orderID].IsUpdated = false
	} else {
		existingOrder.Volume = newVolume
	}

	ob.Orders[orderID] = existingOrder
	fmt.Printf("Order after update: %+v\n", existingOrder)
	ob.matchOrders(orderID, existingOrder.Side)
	fmt.Println("Finished update process.")
}

func (ob *OrderBook) matchOrders(initiatingOrderID int, initiatingOrderSide string) {
	if ob.SellOrders.Len() > 0 && ob.BuyOrders.Len() > 0 {
		fmt.Printf("Top Buy Order: %+v\n", (*ob.BuyOrders)[0])
		fmt.Printf("Top Sell Order: %+v\n", (*ob.SellOrders)[0])
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
		fmt.Println("Potential matching candidates:")
		fmt.Println("Buy order candidates:")
		for _, o := range *ob.BuyOrders {
			if o.Price >= sellOrder.Price {
				fmt.Printf("Buy Order ID: %d, Price: %.2f, Volume: %d\n", o.ID, o.Price, o.Volume)
			}
		}
		fmt.Println("Sell order candidates:")
		for _, o := range *ob.SellOrders {
			if o.Price <= buyOrder.Price {
				fmt.Printf("Sell Order ID: %d, Price: %.2f, Volume: %d\n", o.ID, o.Price, o.Volume)
			}
		}

		if buyOrder.IsUpdated || sellOrder.IsUpdated {
			fmt.Println("Buy or sell order has been updated, skipping matching.")
			fmt.Printf("I really should have skipped you: buyOrder.IsUpdated = %v, sellOrder.IsUpdated = %v\n", buyOrder, sellOrder)
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

func (ob *OrderBook) Cancel(orderID int) {
	fmt.Printf("Attempting to cancel order with ID: %d\n", orderID)
	order, exists := ob.Orders[orderID]
	if !exists {
		fmt.Println("Order not found. Unable to cancel.")
	} else {
		fmt.Println("Order found and cancelled successfully.")
		order.Cancelled = true
		order.SeqNum = seqNumCounter
		seqNumCounter++

		if order.Side == "BUY" {
			for i := 0; i < ob.BuyOrders.Len(); i++ {
				if (*ob.BuyOrders)[i].ID == order.ID {
					fmt.Printf("Buy orders before cancelling: %+v\n", ob.BuyOrders)
					heap.Remove((*PriorityQueue)(ob.BuyOrders), i)
					fmt.Printf("Buy orders after cancelling: %+v\n", ob.BuyOrders)
					break
				}
			}
		} else if order.Side == "SELL" {
			for i := 0; i < ob.SellOrders.Len(); i++ {
				if (*ob.SellOrders)[i].ID == order.ID {
					fmt.Printf("Sell orders before cancelling: %+v\n", ob.SellOrders)
					heap.Remove(ob.SellOrders, i)
					fmt.Printf("Sell orders after cancelling: %+v\n", ob.SellOrders)
					break
				}
			}
		}

		// Check for matches after the new order is cancelled
		// ob.matchOrders()
	}
}

func (obs OrderBooks) Insert(order *Order) {
	ob, exists := obs[order.Symbol]
	if !exists {
		ob = NewOrderBook()
		obs[order.Symbol] = ob
	}
	ob.Insert(order)
}

func (obs OrderBooks) Update(order *Order) {
	fmt.Printf("Attempting to update order with ID: %d, Symbol: %s, New Price: %.2f, New Volume: %d\n", order.ID, order.Symbol, order.Price, order.Volume)

	ob, exists := obs[order.Symbol]
	if !exists {
		fmt.Printf("OrderBook for symbol %s not found\n", order.Symbol)
		return
	}

	fmt.Printf("Found OrderBook for symbol %s. Proceeding with update.\n", order.Symbol)
	ob.Update(order.ID, order.Price, order.Volume)
	fmt.Println("Update call completed for OrderBook.")
}

func (obs OrderBooks) Cancel(orderID int, symbol string) {
	ob, exists := obs[symbol]
	if !exists {
		fmt.Printf("OrderBook for symbol %s not found\n", symbol)
		return
	}
	ob.Cancel(orderID)
}

func runMatchingEngine(operations []string) []string {
	fmt.Printf("Operations AREEEEEEEEE======: %+v\n", operations)
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
			obs.Insert(order)
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
				fmt.Printf("Order with ID %d not found\n", orderID)
				continue
			}
			order := &Order{
				ID:      orderID,
				Symbol:  symbol,
				Side:    side,
				Price:   price,
				Volume:  volume,
				Updated: true,
			}
			fmt.Printf("Before update: symbol = %s, obs = %+v\n", symbol, obs)
			obs.Update(order)
			fmt.Printf("After update: symbol = %s, obs = %+v\n", symbol, obs)
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
				fmt.Printf("OrderBook for symbol %s not found\n", symbol)
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

func formatFloat(f float64) string {
	if f == float64(int(f)) {
		return fmt.Sprintf("%.0f", f)
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func (ob *OrderBook) PrintState() {
	fmt.Println("BuyOrders: ")
	for _, order := range *ob.BuyOrders {
		fmt.Printf("ID=%d, Price=%f, Volume=%d\n", order.ID, order.Price, order.Volume)
	}
	fmt.Println("SellOrders: ")
	for _, order := range *ob.SellOrders {
		fmt.Printf("ID=%d, Price=%f, Volume=%d\n", order.ID, order.Price, order.Volume)
	}
}

// Helper function to aggregate orders by price, summing up volumes
func aggregateOrders(orders *MinHeap) map[float64]int {
	aggregated := make(map[float64]int)
	for _, order := range *orders {
		if !order.Cancelled {
			aggregated[order.Price] += order.Volume
		}
	}
	return aggregated
}
