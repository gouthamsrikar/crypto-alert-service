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
	Upward     SymbolOrderMap `json:"upward"`
	Downward   SymbolOrderMap `json:"downward"`
	BounceUp   SymbolOrderMap `json:"BounceUp"`
	BounceDown SymbolOrderMap `json:"BounceDown"`
}

type MAOrders map[int]RepositoryOrders

type OrderMap struct {
	sync.Mutex
	Data map[string]MAOrders
}

var orders = OrderMap{Data: make(map[string]MAOrders)}

func AddOrder(symbol string, referencePrice float64, orderPrice float64, orderID string, fcmId string, isUp bool, ma int) {
	orders.Lock()
	defer orders.Unlock()
	fmt.Println("Current Price:", referencePrice)

	// Initialize maps if they don't exist
	if _, exists := orders.Data[symbol]; !exists {
		orders.Data[symbol] = make(MAOrders)
	}

	if _, exists := orders.Data[symbol][ma]; !exists {
		orders.Data[symbol][ma] = RepositoryOrders{
			Upward:     make(SymbolOrderMap),
			Downward:   make(SymbolOrderMap),
			BounceDown: make(SymbolOrderMap),
			BounceUp:   make(SymbolOrderMap),
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
		if _, exists := orders.Data[symbol][ma].Upward[orderPrice]; !exists {
			orders.Data[symbol][ma].Upward[orderPrice] = []string{}
		}
		if isUp {
			orders.Data[symbol][ma].Upward[orderPrice] = append(orders.Data[symbol][ma].Upward[orderPrice], orderID)
		} else {
			if _, exists := orders.Data[symbol][ma].BounceDown[orderPrice]; !exists {
				orders.Data[symbol][ma].BounceDown[orderPrice] = []string{}
			}
			orders.Data[symbol][ma].BounceDown[orderPrice] = append(orders.Data[symbol][ma].BounceDown[orderPrice], orderID)
		}

	} else {
		if _, exists := orders.Data[symbol][ma].Downward[orderPrice]; !exists {
			orders.Data[symbol][ma].Downward[orderPrice] = []string{}
		}
		if isUp {
			if _, exists := orders.Data[symbol][ma].BounceUp[orderPrice]; !exists {
				orders.Data[symbol][ma].BounceUp[orderPrice] = []string{}
			}
			orders.Data[symbol][ma].BounceUp[orderPrice] = append(orders.Data[symbol][ma].BounceUp[orderPrice], orderID)
		} else {
			orders.Data[symbol][ma].Downward[orderPrice] = append(orders.Data[symbol][ma].Downward[orderPrice], orderID)
		}

	}

	go SendFCMNotification(fcmId, fmt.Sprintf("Order placed at price $%f", orderPrice), "")
}

func CheckAndExecuteOrder(symbol string, price float64, ma int) {
	orders.Lock()
	defer orders.Unlock()

	if _, exists := orders.Data[symbol]; !exists {
		fmt.Printf("Symbol %s does not exist.\n", symbol)
		go UnsubscribeCoin(symbol)
		return
	}

	// Check if the MA value exists
	if _, exists := orders.Data[symbol][ma]; !exists {
		fmt.Printf("MA %d for symbol %s does not exist.\n", ma, symbol)
		return
	}

	var executedOrderIds []string

	// removing the prices from the upward map that are below the input price
	for p := range orders.Data[symbol][ma].Upward {
		if p < price {
			executedOrderIds = append(executedOrderIds, orders.Data[symbol][ma].Upward[p]...)
			delete(orders.Data[symbol][ma].Upward, p)
		}
	}

	// removing the prices from the downward map that are greater than the input price
	for p := range orders.Data[symbol][ma].Downward {
		if p > price {
			executedOrderIds = append(executedOrderIds, orders.Data[symbol][ma].Downward[p]...)
			delete(orders.Data[symbol][ma].Downward, p)
		}
	}

	for p := range orders.Data[symbol][ma].BounceDown {
		if p <= price {
			bounced := orders.Data[symbol][ma].BounceDown[p]
			delete(orders.Data[symbol][ma].BounceDown, p)
			if _, exists := orders.Data[symbol][ma].Downward[p]; !exists {
				orders.Data[symbol][ma].Downward[p] = []string{}
			}
			orders.Data[symbol][ma].Downward[p] = append(orders.Data[symbol][ma].Downward[p], bounced...)
		}
	}

	for p := range orders.Data[symbol][ma].BounceUp {
		if p >= price {
			bounced := orders.Data[symbol][ma].BounceUp[p]
			delete(orders.Data[symbol][ma].BounceUp, p)
			if _, exists := orders.Data[symbol][ma].Upward[p]; !exists {
				orders.Data[symbol][ma].Upward[p] = []string{}
			}
			orders.Data[symbol][ma].Upward[p] = append(orders.Data[symbol][ma].Upward[p], bounced...)
		}
	}

	go repositories.UpdateOrderStatusForIDs(executedOrderIds, "completed")

	for p := range executedOrderIds {
		go NotifyWithOrderId(executedOrderIds[p], fmt.Sprintf("Order executed for orderId: %s", executedOrderIds[p]), "")
	}

	// unsubscribing the socket when all existing order are executed
	if len(orders.Data[symbol][ma].Upward) == 0 && len(orders.Data[symbol][ma].Downward) == 0 && len(orders.Data[symbol][ma].BounceDown) == 0 && len(orders.Data[symbol][ma].BounceUp) == 0 {
		delete(orders.Data[symbol], ma)
		go RemoveMovingAverageWindow(symbol, ma)
		if len(orders.Data[symbol]) == 0 {
			delete(orders.Data, symbol)
			go UnsubscribeCoin(symbol)
		}
	}
}

func CancelOrder(orderId string, order models.Order, ma int) error {
	color.Green(order.ID, order.Coin, order.Price, order.Status)

	orders.Lock()
	defer orders.Unlock()

	if maOrders, exists := orders.Data[order.Coin]; exists {

		if repoOrders, exists := maOrders[ma]; exists {

			var wg sync.WaitGroup

			if upwardList, exists := repoOrders.Upward[order.Price]; exists {
				wg.Add(1)
				go func() {
					defer wg.Done()
					repositories.UpdateOrderStatus(orderId, "canceled")
				}()
				repoOrders.Upward[order.Price] = removeStringFromSlice(upwardList, orderId)
				if len(repoOrders.Upward[order.Price]) == 0 {
					delete(repoOrders.Upward, order.Price)
				}
			}

			if downwardList, exists := repoOrders.Downward[order.Price]; exists {
				wg.Add(1)
				go func() {
					defer wg.Done()
					repositories.UpdateOrderStatus(orderId, "canceled")
				}()
				repoOrders.Downward[order.Price] = removeStringFromSlice(downwardList, orderId)
				if len(repoOrders.Downward[order.Price]) == 0 {
					delete(repoOrders.Downward, order.Price)
				}
			}

			wg.Wait()

			orders.Data[order.Coin][ma] = repoOrders
		}
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
