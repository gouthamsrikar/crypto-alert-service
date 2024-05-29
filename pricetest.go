package main

import (
	"encoding/json"
	"goalert/services"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/gorilla/websocket"
)

type Message struct {
	EventType string  `json:"e"`
	EventTime int64   `json:"E"`
	Symbol    string  `json:"s"`
	Price     string  `json:"c"`
	C         float64 `json:"C"`
}

func UnmarshalWelcome(data []byte) (Message, error) {
	var r Message
	err := json.Unmarshal(data, &r)
	return r, err
}

func main() {
	endpoint := "wss://stream.binance.com:9443/ws"

	symbols := []string{"btcusdt"}

	services.AddMovingAverageWindow("btcusdt", 5)
	services.AddMovingAverageWindow("btcusdt", 10)
	services.AddMovingAverageWindow("btcusdt", 15)

	c, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	for _, symbol := range symbols {
		subscribeMsg := []byte(`{"method":"SUBSCRIBE","params":["` + symbol + `@ticker"],"id":1}`)
		err := c.WriteMessage(websocket.TextMessage, subscribeMsg)
		if err != nil {
			log.Fatal("subscribe:", err)
		}
	}

	done := make(chan struct{})
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			prettyPrint(message)
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			}
			return
		}
	}
}

func prettyPrint(message []byte) {
	var m Message
	err := json.Unmarshal(message, &m)
	if err != nil {
		log.Println("unmarshal:", err)
		return
	}
	if m.Price != "" {
		floatPrice, _ := strconv.ParseFloat(m.Price, 64)
		 services.AddPrice("btcusdt", floatPrice)
		 services.PrintStockData("btcusdt")
	}

	log.Printf("Symbol: %s, Price: %s", m.Symbol, m.Price)
}
