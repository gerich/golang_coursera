package main

import (
	"net/http"
	"net/http/httptest"
)

func HandleSearch(w http.ResponseWriter, r *http.Request) {

}

func SearchServer() {
	ts := httptest.NewServer(http.HandlerFunc(HandleSearch))
}
