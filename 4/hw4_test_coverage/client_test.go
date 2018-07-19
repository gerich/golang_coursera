package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"flag"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var (
	ts    *httptest.Server
	token string
)

func TestMain(m *testing.M) {
	ts = httptest.NewServer(http.HandlerFunc(SearchServer))
	flag.Parse()
	status := m.Run()
	os.Exit(status)
}

type userXml struct {
	Id        int    `xml:"id"`
	FirstName string `xml:"first_name" json:"-"`
	LastName  string `xml:"last_name" json:"-"`
	Name      string
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

func (u *userXml) PrepareName() {
	u.Name = u.FirstName
	if u.FirstName != "" {
		u.Name += " " + u.LastName
	}
}

func usersInJSON(w io.Writer) {
	var user userXml
	file, err := ioutil.ReadFile("dataset.xml")
	if err != nil {
		panic(err)
	}
	input := bytes.NewReader(file)
	decoder := xml.NewDecoder(input)
	w.Write([]byte("["))
	isFirst := true
	for {
		tok, tokenErr := decoder.Token()
		if tokenErr != nil && tokenErr != io.EOF {
			panic(tokenErr)
		} else if tokenErr == io.EOF {
			break
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			if tok.Name.Local == "row" {
				if isFirst {
					isFirst = false
				} else {
					w.Write([]byte(","))
				}
				if err := decoder.DecodeElement(&user, &tok); err != nil {
					panic(err)
				}
				user.PrepareName()
				b, err := json.Marshal(&user)
				if err != nil {
					panic(err)
				}
				w.Write(b)
			}
		}
	}
	w.Write([]byte("]"))
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("AccessToken") != token {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}
	usersInJSON(w)
}

func TestCorrectAccessToken(t *testing.T) {
	var req SearchRequest
	token = "foo"
	client := SearchClient{
		URL:         ts.URL,
		AccessToken: "foo",
	}
	t.Run("Success Auth", func(t *testing.T) {
		_, err := client.FindUsers(req)
		if err != nil {
			if err.Error() != "Bad AccessToken" {
				t.Errorf("Wrong error type: %v", err)
			}
			t.Errorf("Unsuccess auth")
		}
	})

	t.Run("Unsuccess Auth", func(t *testing.T) {
		client.AccessToken = "bar"
		_, err := client.FindUsers(req)
		if err == nil {
			t.Errorf("Success auth")
		}
		if err.Error() != "Bad AccessToken" {
			t.Errorf("Wrong error type: %v", err)
		}
	})
	token = ""
}

func TestFindUsers(t *testing.T) {
	fixtures := []SearchRequest{
		SearchRequest{},
	}
}

func TestTestServer(t *testing.T) {
	var users []User
	var responseUsers []User
	buf := new(bytes.Buffer)
	usersInJSON(buf)
	json.Unmarshal(buf.Bytes(), &users)
	defer ts.Close()
	resp, err := http.Get(ts.URL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(data, &responseUsers)
	for index, user := range responseUsers {
		if user != users[index] {
			t.Errorf("user in response \"%v\" != user \"%v\"", user, users[index])
		}
	}
}
