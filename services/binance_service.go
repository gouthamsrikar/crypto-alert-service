package services

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	"goalert/models"
	"goalert/repositories"
	"goalert/utils"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type BinanceTicker struct {
	Symbol string  `json:"s"`
	Price  string  `json:"c"`
	C      float64 `json:"C"`
}

var ws *websocket.Conn
var streams map[string]bool
var mu sync.Mutex

func init() {
	streams = make(map[string]bool)
}

var Prices map[string]float64 = make(map[string]float64)

func ListenBinanceSocket(binanceSocketURL string) {
	var err error
	ws, _, err = utils.NewWebSocketClient(binanceSocketURL)
	if err != nil {
		log.Fatal("Error connecting to WebSocket:", err)
	}

	go handleMessages()

	// Initially subscribe to all active coins from db
	// subscribeToActiveCoins()
}

func handleMessages() {
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		var ticker BinanceTicker
		err = json.Unmarshal(message, &ticker)
		if err != nil {
			log.Println("json unmarshal:", err)
			continue
		}

		if ticker.Price != "" {
			price, err := strconv.ParseFloat(ticker.Price, 64)
			if err != nil {
				log.Println("price parse:", err)
				continue
			}

			var orders []models.Order
			repositories.DB.Where("coin = ? AND status = ?", strings.ToLower(ticker.Symbol), "pending").Find(&orders)

			go CheckAndRemoveOrder(strings.ToLower(ticker.Symbol), price)

		}

	}
}

func UnsubscribeCoin(symbol string) {
	streams[strings.ToLower(symbol)] = false
	utils.Unsubscribe(ws, symbol)
}

func subscribeToActiveCoins() {
	var orders []models.Order
	repositories.DB.Where("status = ?", "pending").Find(&orders)

	mu.Lock()
	defer mu.Unlock()
	for _, order := range orders {
		Prices[order.Coin], _ = getPrice(order.Coin)
	}
	for _, order := range orders {
		stream := order.Coin
		go AddOrder(order.Coin, Prices[order.Coin], order.Price, order.ID)
		if !streams[stream] {
			streams[stream] = true
			err := utils.Subscribe(ws, stream)
			if err != nil {
				log.Printf("Error subscribing to stream %s: %v", stream, err)
			} else {
				log.Printf("Subscribed to stream %s", stream)
			}
		}
	}
}

func AddOrderAndSubscribe(order models.Order) error {
	err := repositories.CreateOrder(&order)
	if err != nil {
		return err
	}

	stream := order.Coin

	mu.Lock()
	defer mu.Unlock()

	Prices[order.Coin], err = getPrice(order.Coin)
	if Prices[order.Coin] != 0 {
		color.Red("add order")
		go AddOrder(order.Coin, Prices[order.Coin], order.Price, order.ID)
	} else {
		return err
	}

	if !streams[stream] {
		streams[stream] = true

		err = utils.Subscribe(ws, stream)
		if err != nil {
			log.Printf("Error subscribing to stream %s: %v", stream, err)
			return err
		} else {
			log.Printf("Subscribed to stream %s", stream)
		}

	}
	return nil
}

type PriceResponse struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

func getPrice(symbol string) (float64, error) {
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s", strings.ToUpper(symbol))

	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("error fetching data: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response body: %v", err)
	}

	var priceResponse PriceResponse
	if err := json.Unmarshal(body, &priceResponse); err != nil {
		return 0, fmt.Errorf("error decoding JSON: %v", err)
	}

	price, err := strconv.ParseFloat(priceResponse.Price, 64)
	if err != nil {
		color.Red(err.Error())
		return price, err
	}

	color.Green("no null")
	fmt.Println(price)
	return price, nil
}
