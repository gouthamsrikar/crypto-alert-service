package controllers

import (
	"goalert/models"
	"goalert/repositories"
	"goalert/services"
	"net/http"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateOrder(c *gin.Context) {
	var order models.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order.Status = "pending"

	order.ID = uuid.New().String()[:8]

	if order.Type == "price" {
		order.MA = 1
	}

	go services.StoreFcmId(order.ID, order.FcmID)

	if err := services.AddOrderAndSubscribe(order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

func GetAllOrderStatus(c *gin.Context) {
	orders, err := repositories.GetAllOrders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	services.PrintOrder()

	c.JSON(http.StatusOK, orders)
}

func CancelOrder(c *gin.Context) {
	var req models.CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order, err := repositories.GetOrderById(req.OrderId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "order not found"})
		return
	}

	if order.Status == "Completed" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "order already executed"})
		return
	}

	err = services.CancelOrder(req.OrderId, order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel order"})
		return
	}

	color.Red("ok")

	c.JSON(http.StatusOK, "ok")

}
