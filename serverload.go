package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Order struct {
	Coin  string  `json:"coin"`
	Price float64 `json:"price"`
	FcmId string  `json:"fcmId"`
}

func main() {
	baseURL := "http://localhost:8080/createOrder"
	basePrice := 65650.0
	coin := "btcusdt"
	fcmId := "uiiuy"

	client := &http.Client{}

	trigger:= func (i int) {price := basePrice + float64(i)*1
		order := Order{
			Coin:  coin,
			Price: price,
			FcmId: fcmId,
		}

		jsonData, err := json.Marshal(order)
		if err != nil {
			fmt.Printf("Error marshalling JSON: %v\n", err)
			
		}

		req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error making request: %v\n", err)
			
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Received non-OK response code: %d\n", resp.StatusCode)
		} else {
			fmt.Printf("Successfully sent request with price: %.1f\n", price)
		}

		resp.Body.Close()
	}

	for i := 0; i < 1000; i++ {
		 trigger(i)
	}
}
