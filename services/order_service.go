package services

import (
	"fmt"
	"goalert/models"
	"goalert/repositories"
	"sync"

	"github.com/fatih/color"
)

type SymbolOrderMap map[float64][]string

type RepositoryOrders struct {
	Upward   SymbolOrderMap `json:"upward"`
	Downward SymbolOrderMap `json:"downward"`
}

type OrderMap struct {
	sync.Mutex
	Data map[string]RepositoryOrders
}

var orders = OrderMap{Data: make(map[string]RepositoryOrders)}

func AddOrder(symbol string, referencePrice float64, orderPrice float64, orderID string) {
	orders.Lock()
	defer orders.Unlock()
	fmt.Println("Current Price:", referencePrice)

	// Initialize maps if they don't exist
	if _, exists := orders.Data[symbol]; !exists {
		orders.Data[symbol] = RepositoryOrders{
			Upward:   make(SymbolOrderMap),
			Downward: make(SymbolOrderMap),
		}
	}

	// check if the order is upward or downward to current reference price
	var direction string
	if orderPrice > referencePrice {
		direction = "upward"
	} else {
		direction = "downward"
	}

	// Add the order to the upward/downward map
	if direction == "upward" {
		if _, exists := orders.Data[symbol].Upward[orderPrice]; !exists {
			orders.Data[symbol].Upward[orderPrice] = []string{}
		}
		orders.Data[symbol].Upward[orderPrice] = append(orders.Data[symbol].Upward[orderPrice], orderID)

	} else {
		if _, exists := orders.Data[symbol].Downward[orderPrice]; !exists {
			orders.Data[symbol].Downward[orderPrice] = []string{}
		}
		orders.Data[symbol].Downward[orderPrice] = append(orders.Data[symbol].Downward[orderPrice], orderID)
	}
}

func CheckAndRemoveOrder(symbol string, price float64) {
	orders.Lock()
	defer orders.Unlock()

	// Check if the token symbol exists
	if _, exists := orders.Data[symbol]; !exists {
		fmt.Printf("Symbol %s does not exist.\n", symbol)
		go UnsubscribeCoin(symbol)
		return
	}

	var executedOrderIds []string

	direction := "up"

	// removing the prices from the upward map that are below the input price
	for p := range orders.Data[symbol].Upward {
		if p <= price {
			executedOrderIds = append(executedOrderIds, orders.Data[symbol].Upward[p]...)
			delete(orders.Data[symbol].Upward, p)
			direction = "up"
		}
	}

	// removing the prices from the downward map that are greater than the input price
	for p := range orders.Data[symbol].Downward {
		if p >= price {
			executedOrderIds = append(executedOrderIds, orders.Data[symbol].Downward[p]...)
			delete(orders.Data[symbol].Downward, p)
			direction = "downward"
		}
	}

	go repositories.UpdateOrderStatusForIDs(executedOrderIds, "completed", direction)

	// unsubscribing the socket when all existing order are executed
	if len(orders.Data[symbol].Upward) == 0 && len(orders.Data[symbol].Downward) == 0 {
		delete(orders.Data, symbol)
		go UnsubscribeCoin(symbol)
	}
	// updating db for executed triggers

}

func CancelOrder(orderId string, order models.Order) error {
	color.Green(order.ID, order.Coin, order.Price, order.Status)

	orders.Lock()
	defer orders.Unlock()

	// Check if the mainKey exists in the map
	if repoOrders, exists := orders.Data[order.Coin]; exists {
		color.Green("level 1")

		// WaitGroup to wait for the goroutines to complete
		var wg sync.WaitGroup

		// Remove the string from the upward map
		if upwardList, exists := repoOrders.Upward[order.Price]; exists {
			color.Green("level 2")
			wg.Add(1)
			go func() {
				defer wg.Done()
				repositories.UpdateOrderStatus(orderId, "canceled", "")
			}()
			repoOrders.Upward[order.Price] = removeStringFromSlice(upwardList, orderId)
			if len(repoOrders.Upward[order.Price]) == 0 {
				delete(repoOrders.Upward, order.Price)
			}
		}

		// Remove the string from the downward map
		if downwardList, exists := repoOrders.Downward[order.Price]; exists {
			color.Green("level 3")
			wg.Add(1)
			go func() {
				defer wg.Done()
				repositories.UpdateOrderStatus(orderId, "canceled", "")
			}()
			repoOrders.Downward[order.Price] = removeStringFromSlice(downwardList, orderId)
			if len(repoOrders.Downward[order.Price]) == 0 {
				delete(repoOrders.Downward, order.Price)
			}
		}

		// Wait for all goroutines to complete
		wg.Wait()

		// Update the orders map
		orders.Data[order.Coin] = repoOrders
	}

	return nil
}

func removeStringFromSlice(slice []string, strToRemove string) []string {
	for i, v := range slice {
		if v == strToRemove {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func PrintOrder() {
	fmt.Println(orders)
}
