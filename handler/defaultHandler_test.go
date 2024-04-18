package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDefaultHandler(t *testing.T) {
	w := httptest.NewRecorder()

	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	DefaultHandler(w, r)

	response := w.Result()
	if response.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %v", response.Status)
	}

	body, _ := io.ReadAll(response.Body)
	if !strings.Contains(string(body), "This service has the following endpoints with methods:") {
		t.Errorf("expected response body to contain 'This service has the following endpoints with methods:', got %v", string(body))
	}
}
