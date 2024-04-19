package utils

import (
	"context"

	"cloud.google.com/go/firestore"
)

// Firebase context and client used by Firestore functions throughout the program.
var ctx context.Context
var client *firestore.Client

// sets the firestore client
func SetFirestoreClient(c context.Context, cli *firestore.Client) {
	ctx = c
	client = cli
}

// function to get a document based on its id field
func getDocumentByID(ctx context.Context, collection string, dashboardID string) (*firestore.DocumentSnapshot, error) {
	// Query documents where the 'id' field matches the provided dashboardID
	query := client.Collection(collection).Where("id", "==", dashboardID).Limit(1)
	iter := query.Documents(ctx)

	// Retrieve reference to document
	doc, err := iter.Next()
	if err != nil {
		return nil, err
	}

	return doc, nil
}
