package main

import (
	"assignment2/handler"
	"assignment2/utils"
	"log"
	"net/http"
	"os"

	"context"

	"cloud.google.com/go/firestore"   //Firestore-specific support
	firebase "firebase.google.com/go" // State handling across API boundaries; part of native GoLang API
	"google.golang.org/api/option"    // Generic firebase support
)

// Firebase context and client used by Firestore functions throuhout the program
var ctx context.Context
var client *firestore.Client

func main() {
	// Firebase initialisation
	ctx = context.Background()

	// Loads credential file from firebase
	sa := option.WithCredentialsFile("prog2005-assignment-2-c48bb-firebase-adminsdk-3t1ay-f029236b93.json")
	app, err := firebase.NewApp(context.Background(), nil, sa)
	if err != nil {
		if err != nil {
			log.Println(err)
			return
		}
	}

	// Initiate client
	client, err = app.Firestore(ctx)

	// Check if there is an error when connecting to Firestore
	if err != nil {
		log.Println(err)
		return
	}

	// Close down client at the end of the function
	defer func() {
		errClose := client.Close()
		if errClose != nil {
			log.Fatal("Closing of the firebase client failed. Error", errClose)
		}
	}()

	port := os.Getenv("PORT")

	if port == "" {
		log.Println("$PORT has not been set. Default: 8080")
		port = "8080"

	}

	addr := ":" + port

	http.HandleFunc(utils.DEFAULT_PATH, handler.DefaultHandler)
	http.HandleFunc(utils.REGISTRATION_PATH, handler.RegistrationHandler)

	http.HandleFunc(utils.DASHBOARD_PATH, handler.DashboardHandler)
	http.HandleFunc(utils.STATUS_PATH, handler.StatusHandler)

	// Start server
	log.Println("Firestore REST service listening on %...\n", addr)
	if errSrv := http.ListenAndServe(addr, nil); errSrv != nil {
		panic(errSrv)
	}

}
