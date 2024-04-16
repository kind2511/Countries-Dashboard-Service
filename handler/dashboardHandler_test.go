package handler

import (
	"assignment2/utils"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDashboardsHandler(t *testing.T) {

	// Initialize handler instance
	handler := DashboardHandler()

	// set up structure to be used for testing
	server := httptest.NewServer(http.HandlerFunc(handler))
	// Close the server when test finishes
	defer server.Close()

	// URL where instance is running
	fmt.Println("URL: ", server.URL)

	// Create client instance
	client := http.Client{}

	// test Get request
	res, err := client.Get(server.URL + utils.DASHBOARD_PATH)
	if err != nil {
		t.Fatal("Get request to URL failed:", err.Error())
	}

	// test http status
	if res.Status != "200 OK" {
		t.Fatal("Get request has wrong status code", err.Error())
	}

	// test unsupported Head request
	res2, err2 := client.Head(server.URL + utils.DASHBOARD_PATH)
	if err2 != nil {
		t.Fatal("Head request to URL failed:", err.Error())
	}

	// test http status
	if res2.Status != "405 Method Not Allowed" {
		t.Fatal("un supported method has wrong status code", err.Error())
	}

	//////////
	req, err := http.NewRequest("POST", "", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler2 := http.HandlerFunc(DashboardHandler())

	handler2.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusMethodNotAllowed)
	}

	expected := "Method POST not supported."
	if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			strings.TrimSpace(rr.Body.String()), strings.TrimSpace(expected))
	}

}
