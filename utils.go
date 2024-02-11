package main

import (
	"log"
	"math"
	"sort"
	"strconv"
)

type OrderBook struct {
	orders  []*Order
	sellers []*Order
	buyers  []*Order
}

type Order struct {
	OrderID int
	Symbol  string
	Side    string
	Price   float64
	Volume  int
}

type FinishedOrder struct {
	Symbol       string
	Price        float64
	Volume       int
	TakerOrderID int
	MakerOrderID int
}

func (ob *OrderBook) executeOperation(order Order) map[string][]FinishedOrder {

	var finishedOrders = make(map[string][]FinishedOrder)

	if order.Side == "BUY" {
		ob.buyers = append(ob.buyers, &order)
	} else {
		ob.sellers = append(ob.sellers, &order)
	}
	ob.sortBuyers()

	log.Printf("the seller item is: %+v", ob.sellers)

	// we iterate through the sellers and match them against the buyers
	i := 0
	for i < len(ob.sellers) {
		j := 0
		for j < len(ob.buyers) {
			buy := ob.buyers[j]
			log.Printf("the sell order is: %+v - and the price is: %v - the volume is: %v", ob.sellers[i].OrderID, ob.sellers[i].Price, ob.sellers[i].Volume)
			if buy.Price >= ob.sellers[i].Price { // there's a match here
				availableVolume := min(buy.Volume, ob.sellers[i].Volume)

				ob.buyers[j].Volume -= availableVolume
				ob.sellers[i].Volume -= availableVolume

				log.Printf("an operation was made successfully: buyerId :%d, sellerId: %d, volumne: %f, price: %f", buy.OrderID, ob.sellers[i].OrderID, math.Abs(float64(buy.Volume-ob.sellers[i].Volume)), buy.Price)
				finishedOrders[ob.sellers[i].Symbol] = append(finishedOrders[ob.sellers[i].Symbol], FinishedOrder{
					Symbol:       ob.sellers[i].Symbol,
					Price:        ob.sellers[i].Price,
					Volume:       availableVolume,
					TakerOrderID: buy.OrderID,
					MakerOrderID: ob.sellers[i].OrderID,
				})

			}
			// If the buyer's volume is now zero, remove them from the slice
			if ob.buyers[j].Volume == 0 {
				ob.buyers = append(ob.buyers[:j], ob.buyers[j+1:]...)
			} else {
				j++
			}
		}

		// Move on to the next seller
		i++

		// If there are no more buyers, break the loop
		if len(ob.buyers) == 0 {
			break
		}
	}
	// Add non-zero volume orders to ob.orders
	ob.orders = appendNonZeroVolume(ob.sellers)
	ob.orders = append(ob.orders, appendNonZeroVolume(ob.buyers)...)
	return finishedOrders
}

func (ob *OrderBook) sortBuyers() {
	sort.Slice(ob.buyers, func(i, j int) bool {
		if ob.buyers[i].Price == ob.buyers[j].Price {
			return ob.buyers[i].OrderID < ob.buyers[j].OrderID
		}
		return ob.buyers[i].Price > ob.buyers[j].Price
	})
}

// appendNonZeroVolume appends orders with non-zero volume to the result slice. This overrides the ob.orders slice which is extremely not the
// expected behavior. This function should return a new slice with the non-zero volume orders.
func appendNonZeroVolume(orders []*Order) []*Order {
	var result []*Order
	for _, item := range orders {
		if item.Volume > 0 {
			result = append(result, item)
		}
	}
	return result
}

func returnFormatted(orders []Order) []string {
	var result []string
	// reverse iterate through the orders
	for i := len(orders) - 1; i >= 0; i-- {

		order := orders[i]
		// <symbol>,<price>,<volume>,<taker_order_id>,<maker_order_id>
		result = append(result, order.Symbol+","+formatFloat(order.Price)+","+formatInt(order.Volume)+","+formatInt(order.OrderID), "==="+order.Symbol+"===", order.Side+","+formatFloat(order.Price)+","+formatInt(order.Volume))

	}
	return result

}

func (ob *OrderBook) updateOrder(operationData []string) {

}

func (ob *OrderBook) cancelOrder(operationData []string) {

}

func convertToFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func convertToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 4, 64)
}

func formatInt(i int) string {
	return strconv.Itoa(i)
}
