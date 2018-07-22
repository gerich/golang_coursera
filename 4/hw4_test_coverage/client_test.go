package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

var (
	ts             *httptest.Server
	token          string
	allUsers       []User
	httpMiddleware func(w http.ResponseWriter)
	timeout        time.Duration
	clientErr      error
)

func TestMain(m *testing.M) {
	allUsers = getAllUsers()
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

func (u *userXml) ToUser() User {
	return User{
		Id:     u.Id,
		Name:   u.Name,
		Age:    u.Age,
		About:  u.About,
		Gender: u.Gender,
	}
}

func getAllUsers() []User {
	var allUsers []User
	var user userXml
	var mainUser User
	file, err := ioutil.ReadFile("dataset.xml")
	if err != nil {
		panic(err)
	}
	input := bytes.NewReader(file)
	decoder := xml.NewDecoder(input)
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
				}
				if err := decoder.DecodeElement(&user, &tok); err != nil {
					panic(err)
				}
				user.PrepareName()
				mainUser = user.ToUser()
				allUsers = append(allUsers, mainUser)
				if err != nil {
					panic(err)
				}
			}
		}
	}
	return allUsers
}

func usersInJSON(w http.ResponseWriter, r *http.Request) {
	var req SearchRequest
	req.OrderField = "name"

	q := r.URL.Query()

	if limit, err := strconv.Atoi(q.Get("limit")); err == nil {
		req.Limit = limit
	}

	if offset, err := strconv.Atoi(q.Get("offset")); err == nil {
		req.Offset = offset
	}

	if orderBy, err := strconv.Atoi(q.Get("order_by")); err == nil {
		req.OrderBy = orderBy
	}

	if query := q.Get("query"); query != "" {
		req.Query = query
	}

	if orderField := q.Get("order_field"); orderField != "" {
		req.OrderField = strings.ToLower(orderField)
	}

	if !(req.OrderField == "name" || req.OrderField == "age" || req.OrderField == "id") {
		errRes := SearchErrorResponse{Error: "ErrorBadOrderField"}
		b, err := json.Marshal(errRes)
		if err != nil {
			panic(err)
		}
		http.Error(w, string(b), http.StatusBadRequest)
		return
	}

	b, err := json.Marshal(getUsersByRequest(req))
	if err != nil {
		panic(err)
	}
	w.Write(b)

}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	if httpMiddleware != nil {
		httpMiddleware(w)
		return
	}
	if timeout != 0 {
		time.Sleep(timeout)
	}
	if r.Header.Get("AccessToken") != token {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}
	usersInJSON(w, r)
}

func TestTimeout(t *testing.T) {
	timeout = 2 * time.Second
	client := SearchClient{URL: ts.URL}
	req := SearchRequest{}
	_, err := client.FindUsers(req)
	if !strings.Contains(err.Error(), "timeout for") {
		t.Errorf("Test Timeout Falied")
	}
	timeout = 0
}

func TestUnknownError(t *testing.T) {
	client := SearchClient{URL: ""}
	req := SearchRequest{}
	_, err := client.FindUsers(req)
	if !strings.Contains(err.Error(), "unknown error ") {
		t.Errorf("Test Unknown Error Falied")
	}
}

func TestWorngLimit(t *testing.T) {
	client := SearchClient{URL: ts.URL}
	req := SearchRequest{Limit: -1}
	_, err := client.FindUsers(req)
	if err.Error() != "limit must be > 0" {
		t.Errorf("Test Wrong Limit Falied")
	}
}

func TestMaxLimit(t *testing.T) {
	client := SearchClient{URL: ts.URL}
	req := SearchRequest{Limit: 26}
	res, _ := client.FindUsers(req)

	if len(res.Users) > 25 {
		t.Errorf("Test max limit falied")
	}
}

func TestWrongOffset(t *testing.T) {
	client := SearchClient{URL: ts.URL}
	req := SearchRequest{Offset: -1}
	_, err := client.FindUsers(req)
	if err.Error() != "offset must be > 0" {
		t.Errorf("Test Wrong Offset Falied")
	}
}

func Test500Error(t *testing.T) {
	httpMiddleware = func(w http.ResponseWriter) {
		http.Error(w, "foo and bar", http.StatusInternalServerError)
	}
	client := SearchClient{URL: ts.URL}
	req := SearchRequest{}
	_, err := client.FindUsers(req)
	if err.Error() != "SearchServer fatal error" {
		t.Errorf("Test 500 failed")
	}
	httpMiddleware = nil
}

func TestBadRequestIncorrectJson(t *testing.T) {
	client := SearchClient{URL: ts.URL}
	req := SearchRequest{}
	httpMiddleware = func(w http.ResponseWriter) {
		http.Error(w, "test", http.StatusBadRequest)
		w.Write([]byte("foo and bar"))
	}
	_, err := client.FindUsers(req)
	if !strings.Contains(err.Error(), "cant unpack error json") {
		t.Error("Fail error incorrect json test")
	}
	httpMiddleware = nil
}

func TestBadRequestWrongOrderField(t *testing.T) {
	client := SearchClient{URL: ts.URL}
	req := SearchRequest{OrderField: "Foo"}
	_, err := client.FindUsers(req)
	if !strings.Contains(err.Error(), "OrderFeld Foo invalid") {
		t.Error("Fail bad order field test")
	}
}

func TestBadJsonBody(t *testing.T) {
	client := SearchClient{URL: ts.URL}
	req := SearchRequest{}
	httpMiddleware = func(w http.ResponseWriter) {
		w.Write([]byte("foo and bar"))
	}
	_, err := client.FindUsers(req)
	if !strings.Contains(err.Error(), "cant unpack result json") {
		t.Errorf("Fail bad json body test: %v", err)
	}
	httpMiddleware = nil
}

func TestBadRequestUnknownError(t *testing.T) {
	client := SearchClient{URL: ts.URL}
	req := SearchRequest{}
	errRes := SearchErrorResponse{}
	httpMiddleware = func(w http.ResponseWriter) {
		b, _ := json.Marshal(errRes)
		http.Error(w, string(b), http.StatusBadRequest)
	}
	_, err := client.FindUsers(req)
	if !strings.Contains(err.Error(), "unknown bad request error") {
		t.Error("Fail unknown error bad request test")
	}
	httpMiddleware = nil
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

type SortedUsersByName []User

func (a SortedUsersByName) Len() int           { return len(a) }
func (a SortedUsersByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortedUsersByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

type SortedUsersById []User

func (a SortedUsersById) Len() int           { return len(a) }
func (a SortedUsersById) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortedUsersById) Less(i, j int) bool { return a[i].Id < a[j].Id }

type SortedUsersByAge []User

func (a SortedUsersByAge) Len() int           { return len(a) }
func (a SortedUsersByAge) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortedUsersByAge) Less(i, j int) bool { return a[i].Age < a[j].Age }

func sortUsers(s sort.Interface, r SearchRequest) {
	if r.OrderBy == -1 {
		sort.Sort(s)
	}
	if r.OrderBy == 1 {
		sort.Sort(sort.Reverse(s))
	}
}

func getUsersByRequest(r SearchRequest) []User {
	// filtering
	filtered := make([]User, 0, len(allUsers))
	for _, user := range allUsers {
		hasName := strings.Contains(user.Name, r.Query)
		hasAbout := strings.Contains(user.About, r.Query)
		if hasAbout || hasName {
			filtered = append(filtered, user)
		}
	}
	// sorting
	if r.OrderBy != 0 {

		switch strings.ToLower(r.OrderField) {
		case "age":
			sorted := SortedUsersByAge(filtered)
			sortUsers(sorted, r)
			filtered = []User(sorted)
			break
		case "id":
			sorted := SortedUsersById(filtered)
			sortUsers(sorted, r)
			filtered = []User(sorted)
			break
		default:
			sorted := SortedUsersByName(filtered)
			sortUsers(sorted, r)
			filtered = []User(sorted)
			break
		}
	}
	// paging
	maxUsers := len(filtered) - 1
	if maxUsers < r.Offset {
		var empty []User
		return empty
	}

	if maxUsers > r.Offset && maxUsers < r.Offset+r.Limit {
		return filtered[r.Offset:]
	}

	return filtered[r.Offset : r.Offset+r.Limit]
}

func TestFindUsers(t *testing.T) {
	fixtures := []SearchRequest{
		SearchRequest{Limit: 2, Offset: 1, Query: "tempor"},
		SearchRequest{Limit: 2, Offset: 0, Query: "Glenn"},
		SearchRequest{Limit: 2, Offset: 1, Query: "consectetur", OrderBy: -1},
		SearchRequest{Limit: 2, Offset: 1, Query: "consectetur", OrderBy: 1},
		SearchRequest{Limit: 2, Offset: 1, Query: "consectetur", OrderBy: 1, OrderField: "Id"},
		SearchRequest{Limit: 2, Offset: 1, Query: "consectetur", OrderBy: -1, OrderField: "Id"},
		SearchRequest{Limit: 2, Offset: 1, Query: "consectetur", OrderBy: -1, OrderField: "Age"},
		SearchRequest{Limit: 2, Offset: 1, Query: "consectetur", OrderBy: 1, OrderField: "Age"},
	}
	client := SearchClient{URL: ts.URL}

	for _, req := range fixtures {
		t.Run("Find with req: "+fmt.Sprintf("%#v", req), func(t *testing.T) {
			res, err := client.FindUsers(req)
			if err != nil {
				t.Errorf("Has error: %s", err.Error())
			} else if len(res.Users) != req.Limit {
				t.Errorf("Wrong count users %d != %d", len(res.Users), req.Limit)
			} else if len(res.Users) == 0 {
				t.Errorf("No users found")
			}

			// Usless
			users := getUsersByRequest(req)
			for index, u := range users {
				if res.Users[index] != u {
					t.Errorf("Incorrect users with req: %#v", req)
				}
			}
		})
	}
}

func TestFindUsersWithNoNextPage(t *testing.T) {
	client := SearchClient{URL: ts.URL}
	maxUsers := len(allUsers)
	req := SearchRequest{Limit: 2, Offset: maxUsers - 2}
	res, _ := client.FindUsers(req)

	if res.NextPage {
		t.Errorf("Test No Next Page Failed")
	}
}
