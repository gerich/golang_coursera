package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

func HandleSearch(w http.ResponseWriter, r *http.Request) {
	var users []User
	file, _ := ioutil.ReadFile("dataset.xml")

	xml.Unmarshal(file, &users)

	fmt.Println(users)
}

func SearchServer() {
	ts := httptest.NewServer(http.HandlerFunc(HandleSearch))

}
