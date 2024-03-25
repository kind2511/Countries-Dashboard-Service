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

// Firebase context and client used by Firestore functions throughout the program.
var ctx context.Context
var client *firestore.Client



func main() {

	// Firebase initialisation
	ctx = context.Background()

	// Loads credential file from firebase
	sa := option.WithCredentialsFile("./prog2005-assignment2-ee93a-firebase-adminsdk-9o3qm-43d9d2d766.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Println(err)
		return
	}

	// Instantiate client
	client, err = app.Firestore(ctx)

	// Check whether there is an error when connecting to Firestore
	if err != nil {
		log.Println(err)
		return
	}

	// Close down client at the end of the function
	defer func() {
		errClose := client.Close()
		if errClose != nil {
			log.Fatal("Closing of the Firebase client failed. Error:", errClose)
		}
	}()

	port := os.Getenv("PORT")

	if port == "" {
		log.Println("$PORT has not been set. Default: 8080")
		port = "8080"

	}

	http.HandleFunc(utils.DEFAULT_PATH, handler.DefaultHandler)
	http.HandleFunc(utils.REGISTRATION_PATH, handler.RegistrationHandler)

	http.HandleFunc(utils.DASHBOARD_PATH, handler.DashboardHandler)
	http.HandleFunc(utils.STATUS_PATH, handler.StatusHandler)

}
