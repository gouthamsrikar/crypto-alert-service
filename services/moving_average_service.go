package services

import (
	"fmt"
	"sync"

	"github.com/fatih/color"
)

type StockData struct {
	sync.Mutex
	Prices         []float64
	MovingAverages map[int]float64
}

type StockMap struct {
	sync.Mutex
	Data map[string]*StockData
}

var stockMap = StockMap{Data: make(map[string]*StockData)}

func AddPrice(stock string, price float64) {
	//stockMap.Lock()
	//defer stockMap.Unlock()

	stockData, exists := stockMap.Data[stock]
	if !exists {
		stockData = &StockData{
			Prices:         []float64{},
			MovingAverages: make(map[int]float64),
		}
		stockMap.Data[stock] = stockData
	}

	//stockData.Lock()
	//defer stockData.Unlock()
	stockData.Prices = append(stockData.Prices, price)

	for windowSize := range stockData.MovingAverages {
		updateMovingAverage(stockData, windowSize)
	}

	maxWindowSize := getMaxWindowSize(stockData.MovingAverages)
	if len(stockData.Prices) > maxWindowSize {
		stockData.Prices = stockData.Prices[len(stockData.Prices)-maxWindowSize:]
	}
}

func AddMovingAverageWindow(stock string, windowSize int) {
	//stockMap.Lock()
	//defer stockMap.Unlock()

	stockData, exists := stockMap.Data[stock]
	if !exists {
		stockData = &StockData{
			Prices:         []float64{},
			MovingAverages: make(map[int]float64),
		}
		stockMap.Data[stock] = stockData
	}

	//stockData.Lock()
	//defer stockData.Unlock()
	color.Red("updaet window")

	if _, exists := stockData.MovingAverages[windowSize]; !exists {

		updateMovingAverage(stockData, windowSize)
	}
}

func RemoveMovingAverageWindow(stock string, windowSize int) {
	//stockMap.Lock()
	//defer stockMap.Unlock()

	stockData, exists := stockMap.Data[stock]
	if !exists {
		return
	}

	//stockData.Lock()
	//defer stockData.Unlock()
	delete(stockData.MovingAverages, windowSize)

	if len(stockData.MovingAverages) == 0 {
		delete(stockMap.Data, stock)
	}
}

func GetStockData(stock string) (prices []float64, movingAverages map[int]float64) {
	//stockMap.Lock()
	//defer stockMap.Unlock()

	stockData, exists := stockMap.Data[stock]
	if !exists {
		return nil, nil
	}

	//stockData.Lock()
	//defer stockData.Unlock()

	prices = append([]float64(nil), stockData.Prices...)
	movingAverages = make(map[int]float64)
	for windowSize, avg := range stockData.MovingAverages {
		movingAverages[windowSize] = avg
	}

	return prices, movingAverages
}

func updateMovingAverage(stockData *StockData, windowSize int) {
	if len(stockData.Prices) < windowSize {
		stockData.MovingAverages[windowSize] = 0
		return
	}

	sum := 0.0
	for i := len(stockData.Prices) - windowSize; i < len(stockData.Prices); i++ {
		sum += stockData.Prices[i]
	}
	stockData.MovingAverages[windowSize] = sum / float64(windowSize)
}

func getMaxWindowSize(movingAverages map[int]float64) int {
	max := 0
	for windowSize := range movingAverages {
		if windowSize > max {
			max = windowSize
		}
	}
	return max
}

func (s *StockData) String() string {
	s.Lock()
	defer s.Unlock()

	result := "Prices: ["
	for i, price := range s.Prices {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%.2f", price)
	}
	result += "]\nMoving Averages:\n"
	for window, avg := range s.MovingAverages {
		result += fmt.Sprintf("  Window %d: %.2f\n", window, avg)
	}
	return result
}

func PrintStockData(stock string) {
	//stockMap.Lock()
	//defer stockMap.Unlock()

	// stockData, exists := stockMap.Data[stock]
	// if !exists {
	// 	fmt.Println("No data found for stock:", stock)
	// 	return
	// }

	// fmt.Printf("Stock: %s\n%s", stock, stockData.String())
	fmt.Println(stockMap)
}
