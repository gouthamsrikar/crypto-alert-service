package repositories

// all db related actions
import (
	"goalert/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(databaseURL string) {
	var err error
	DB, err = gorm.Open(sqlite.Open(databaseURL), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	DB.AutoMigrate(&models.Order{})

	cleanDB()
}

func CreateOrder(order *models.Order) error {
	return DB.Create(order).Error
}

func GetAllOrders() ([]models.Order, error) {
	var orders []models.Order
	err := DB.Find(&orders).Error
	return orders, err
}

func UpdateOrderStatus(id string, status string) error {
	return DB.Model(&models.Order{}).Where("id = ?", id).Update("status", status).Error
}

func UpdateOrderStatusForIDs(ids []string, status string) error {
	for _, id := range ids {
		err := UpdateOrderStatus(id, status)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetOrderById(id string) (models.Order, error) {
	var order models.Order
	err := DB.First(&order, "id = ?", id).Error
	return order, err
}

func cleanDB() {
	DB.Exec("DELETE FROM orders")
}
