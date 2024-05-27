package main

import (
	"goalert/config"
	"goalert/controllers"
	"goalert/repositories"
	"goalert/services"

	"github.com/gin-gonic/gin"
)

func main() {

	config.InitConfig()

	repositories.InitDB(config.AppConfig.DatabaseURL)

	// services.InitFCMService()

	go services.ListenBinanceSocket(config.AppConfig.BinanceSocketURL)

	r := gin.Default()

	r.POST("/createOrder", controllers.CreateOrder)
	r.GET("/getAllOrderStatus", controllers.GetAllOrderStatus)
	r.POST("/cancelOrder", controllers.CancelOrder)

	r.Run(":8080")
}
