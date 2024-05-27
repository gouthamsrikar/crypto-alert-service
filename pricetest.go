package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
)

// Message struct to represent the JSON message received from the WebSocket
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
	// Define the endpoint for the Binance WebSocket API
	endpoint := "wss://stream.binance.com:9443/ws"

	// Define the symbols you want to subscribe to
	symbols := []string{"btcusdt"} // Example symbols, you can change this to whatever you need

	// Create a WebSocket connection
	c, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	// Subscribe to the symbols
	for _, symbol := range symbols {
		subscribeMsg := []byte(`{"method":"SUBSCRIBE","params":["` + symbol + `@ticker"],"id":1}`)
		err := c.WriteMessage(websocket.TextMessage, subscribeMsg)
		if err != nil {
			log.Fatal("subscribe:", err)
		}
	}

	// Listen for messages
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

			// Cleanly close the WebSocket connection by sending a close message and then waiting (with timeout) for the server to close the connection.
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

// prettyPrint function to print the JSON message in a human-readable format
func prettyPrint(message []byte) {
	var m Message
	err := json.Unmarshal(message, &m)
	if err != nil {
		log.Println("unmarshal:", err)
		return
	}
	// log.Printf(" Price: %s", m.Price)

	log.Printf("Symbol: %s, Price: %s", m.Symbol, m.Price)
}
