package main

import (
	"goalert/services"
)

func main() {
	services.AddPrice("AAPL", 150.00)
	services.AddPrice("AAPL", 155.00)
	services.AddPrice("AAPL", 160.00)
	services.AddMovingAverageWindow("AAPL", 2)
	services.AddMovingAverageWindow("AAPL", 3)
	services.PrintStockData("AAPL")

	services.RemoveMovingAverageWindow("AAPL", 2)
	services.PrintStockData("AAPL")

	services.AddPrice("AAPL", 150.00)
	services.AddPrice("AAPL", 155.00)
	services.AddPrice("AAPL", 160.00)
	services.PrintStockData("AAPL") // Should print "No data found for stock: AAPL"

	services.RemoveMovingAverageWindow("AAPL", 3)
	services.PrintStockData("AAPL") // Should print "No data found for stock: AAPL"
}
