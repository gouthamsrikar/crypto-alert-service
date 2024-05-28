package services

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
)

var fcmClient *messaging.Client

var clientFcmIdMap = make(map[string]string)

func InitFCMService() {
	opt := option.WithCredentialsFile("firebase.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	fcmClient, err = app.Messaging(context.Background())
	if err != nil {
		log.Fatalf("error getting Messaging client: %v", err)
	}
}

func NotifyWithOrderId(orderId string, title string, body string) {
	SendFCMNotification(clientFcmIdMap[orderId], title, body)
}

func StoreFcmId(orderId string, fcmId string) {
	clientFcmIdMap[orderId] = fcmId
}

func SendFCMNotification(token, title, body string) {
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Token: token,
	}

	response, err := fcmClient.Send(context.Background(), message)
	if err != nil {
		log.Printf("Error sending FCM message: %v\n", err)
	} else {
		log.Printf("Successfully sent FCM message: %s\n", response)
	}
}
