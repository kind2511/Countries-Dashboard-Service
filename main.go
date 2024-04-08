package main

import (
	"assignment2/handler"
	"assignment2/utils"
	"context"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

var ctx context.Context
var client *firestore.Client

func main() {

	ctx = context.Background()

	sa := option.WithCredentialsFile("./firebase-json.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Println(err)
		return
	}

	client, err = app.Firestore(ctx)
	if err != nil {
		log.Println(err)
		return
	}

	defer func() {
		errClose := client.Close()
		if errClose != nil {
			log.Fatal("Closing of the Firebase client failed. Error:", errClose)
		}
	}()

	handler.SetFirestoreClient(ctx, client)

	port := os.Getenv("PORT")

	if port == "" {
		log.Println("$PORT has not been set. Default: 8080")
		port = "8080"

	}

	http.HandleFunc(utils.DEFAULT_PATH, handler.DefaultHandler)
	http.HandleFunc(utils.REGISTRATION_PATH, handler.RegistrationHandler)

	http.HandleFunc(utils.DASHBOARD_PATH, handler.DashboardHandler)
	http.HandleFunc(utils.STATUS_PATH, handler.StatusHandler)

	//Starts and listens to the port
	log.Println("Starting server on port " + port + " ...")
	log.Fatal(http.ListenAndServe(":"+port, nil))

}
