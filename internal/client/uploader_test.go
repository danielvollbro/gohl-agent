package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielvollbro/gohl/internal/game"
)

func TestUploadReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var receivedReport game.GrandReport
		if err := json.NewDecoder(r.Body).Decode(&receivedReport); err != nil {
			t.Errorf("Could not decode body: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	
	defer server.Close()

	dummyReport := game.GrandReport{
		Rank:       "Test Pilot",
		TotalScore: 100,
	}

	err := UploadReport(server.URL, dummyReport)

	if err != nil {
		t.Errorf("UploadReport failed unexpectedly: %v", err)
	}
}

func TestUploadReport_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	dummyReport := game.GrandReport{}
	
	err := UploadReport(server.URL, dummyReport)

	if err == nil {
		t.Error("Expected error from UploadReport when server returns 500, but got nil")
	}
}