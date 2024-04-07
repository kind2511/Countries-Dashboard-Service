package main

import (
	"assignment2/handler"
	"assignment2/utils"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {

	fmt.Println("Hello World!")

	port := os.Getenv("PORT")

	if port == "" {
		log.Println("$PORT has not been set. Default: 8080")
		port = "8080"

	}

	http.HandleFunc(utils.DEFAULT_PATH, handler.DefaultHandler)
	http.HandleFunc(utils.REGISTRATION_PATH, handler.RegistrationHandler)

	http.HandleFunc(utils.DASHBOARD_PATH, handler.DashboardHandler)
	http.HandleFunc(utils.STATUS_PATH, handler.StatusHandler)
	http.HandleFunc(utils.NOTIFICATION_PATH, handler.NotificationHandler)

	// Start http Server
	log.Println("Starting server on port " + port + "...")
	log.Fatal(http.ListenAndServe(":"+port, nil))

}
