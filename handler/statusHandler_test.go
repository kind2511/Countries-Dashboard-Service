package handler

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"cloud.google.com/go/firestore"
)

func TestCheckHTTPError(t *testing.T) {
	// Redirect standard output to a buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call checkHTTPError with an error
	testErr := errors.New("test error")
	checkHTTPError(testErr)

	// Capture standard output
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stdout = oldStdout

	expected := "error occurred while making HTTP request: test error"
	actual := buf.String()

	// Trim spaces and convert both strings to lower case for equality check
	expected = strings.ToLower(strings.TrimSpace(expected))
	actual = strings.ToLower(strings.TrimSpace(actual))

	if expected != actual {
		t.Errorf("unexpected output: got %v want %v", actual, expected)
	}

}

func TestCheckHTTPErrorNoError(t *testing.T) {
	// Redirect standard output to a buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call checkHTTPError with nil
	checkHTTPError(nil)

	// Capture standard output
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stdout = oldStdout

	// Check that nothing was printed
	if buf.Len() != 0 {
		t.Errorf("unexpected output: got %v want nothing", buf.String())
	}
}

func TestGetWebhookSize(t *testing.T) {
	mockDocuments := []*firestore.DocumentSnapshot{ /* Initialize with the documents you want the mock function to return */ }

	size, err := GetWebhookSize(func() ([]*firestore.DocumentSnapshot, error) {
		return mockDocuments, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Check the size
	if size != len(mockDocuments) {
		t.Errorf("Expected size to be %v, got %v", len(mockDocuments), size)
	}

	// Test error
	size, err = GetWebhookSize(func() ([]*firestore.DocumentSnapshot, error) {
		return nil, errors.New("test error")
	})

}
