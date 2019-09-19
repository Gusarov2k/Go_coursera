package main

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"io/ioutil"
	"encoding/xml"
	"encoding/json"
	_"fmt"
	"strconv"
	"io"
	"time"
)

type xmlRow struct {
	Id int `xml:"id"`
	Guid string `xml:"guid"`
	IsActive bool `xml:"isActive"`
	Balance string `xml:"balance"`
	Picture string `xml:"picture"`
	Age int `xml:"age"`
	EyeColor string `xml:"eyeColor"`
	FirstName string `xml:"first_name"`
	LastName string `xml:"last_name"`
	Gender string `xml:"gender"`
	Company string `xml:"company"`
	Email string `xml:"email"`
	Phone string `xml:"phone"`
	Address string `xml:"address"`
	About string `xml:"about"`
}

type xmlStructure struct {
	Version string `xml:"version"`
	Row []xmlRow `xml:"row"`

}

const pageSize = 25


func SearchServerSuccess(w http.ResponseWriter, r *http.Request)  {
	dataFile, err := ioutil.ReadFile("dataset.xml")
	checkError(err)

	usersXml := &xmlStructure{}
	xml.Unmarshal(dataFile, &usersXml)

	var currentUsers []User

	for _, curnUser := range usersXml.Row {
		currentUsers = append(currentUsers, User{
			Id: curnUser.Id,
			Name: curnUser.FirstName,
			Age: curnUser.Age,
			About: curnUser.About,
			Gender: curnUser.Gender,
		})
	}

	offset, _ := strconv.Atoi(r.FormValue("offset"))
	limit, _ := strconv.Atoi(r.FormValue("limit"))

	var startRow int
	if offset > 0 {
		startRow = offset * pageSize
	}

	endRow := startRow + limit
	currentUsers = currentUsers[ startRow: endRow ]

	jsonResponse, err := json.Marshal(currentUsers)
	checkError(err)

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func SearchServerLimitFail(w http.ResponseWriter, r *http.Request)  {
	dataFile, err := ioutil.ReadFile("dataset.xml")
	checkError(err)

	usersXml := &xmlStructure{}
	xml.Unmarshal(dataFile, &usersXml)

	var currentUsers []User

	for _, curUser := range usersXml.Row {
		currentUsers = append(currentUsers, User{
			Id: curUser.Id,
			Name: curUser.FirstName,
			Age: curUser.Age,
			About: curUser.About,
			Gender: curUser.Gender,
		})
	}

	jsonResponse, err := json.Marshal(currentUsers)
	checkError(err)

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func SearchServerJsonFail(w http.ResponseWriter, r *http.Request)  {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `"err": "bad json"}`)
}

func SearchServerTimeoutError(w http.ResponseWriter, r *http.Request)  {
	time.Sleep(time.Second * 2)
	w.WriteHeader(http.StatusOK)
}

func SearchServerUnknownError(w http.ResponseWriter, r *http.Request)  {}

func SearchServerUnauthorized(w http.ResponseWriter, r *http.Request)  {
	w.WriteHeader(http.StatusUnauthorized)
}

func SearchServerInternalServerError(w http.ResponseWriter, r *http.Request)  {
	w.WriteHeader(http.StatusInternalServerError)
}

func SearchServerBadRequest(w http.ResponseWriter, r *http.Request)  {
	w.WriteHeader(http.StatusBadRequest)
}

func SearchServerBadField(w http.ResponseWriter, r *http.Request)  {
	w.WriteHeader(http.StatusBadRequest)
	jsonResponse, _ := json.Marshal(SearchErrorResponse{Error:"ErrorBadOrderField"})
	w.Write(jsonResponse)
}

func SearchServerBadError(w http.ResponseWriter, r *http.Request)  {
	w.WriteHeader(http.StatusBadRequest)
	jsonResponse, _ := json.Marshal(SearchErrorResponse{Error:"Unknown error"})
	w.Write(jsonResponse)
}


func TestErrorResponse(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(SearchServerSuccess))

	searchClient := &SearchClient{
		URL: ts.URL,
	}

	searchRequest := SearchRequest{
		Limit: 5,
		Offset: 0,
	}

	_, err := searchClient.FindUsers(searchRequest)

	if err != nil {
		t.Error("Dosn't work success request")
	}

	searchRequest.Limit = -1

	_, err = searchClient.FindUsers(searchRequest)
	if err.Error() != "limit must be > 0" {
		t.Error("limit must be > 0")
	}

	searchRequest.Limit = 1
	searchRequest.Offset = -1
	_, err = searchClient.FindUsers(searchRequest)
	if err.Error() != "offset must be > 0" {
		t.Error("offset must be > 0")
	}

	ts.Close()
}

func TestLimitFailed(t *testing.T)  {
	limit := 7
	ts := httptest.NewServer(http.HandlerFunc(SearchServerLimitFail))

	searchClient := &SearchClient{
		URL: ts.URL,
	}

	response, _ := searchClient.FindUsers(SearchRequest{Limit: limit})

	if limit == len(response.Users) {
		t.Error("Limit not true")
	}
	ts.Close()
}

func TestBadJson(t *testing.T)  {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerJsonFail))
	searchClient := &SearchClient{
		URL: ts.URL,
	}
	_, err := searchClient.FindUsers(SearchRequest{})

	if err.Error() != `cant unpack result json: invalid character ':' after top-level value` {
		t.Error("Bad json test :(")
	}
	ts.Close()
}

func TestPerelimit(t *testing.T)  {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerSuccess))
	searchClient := &SearchClient{
		URL: ts.URL,
	}

	response, _ := searchClient.FindUsers(SearchRequest{Limit:26})

	if 25 != len(response.Users) {
		t.Error("Perelimit :(")
	}
	ts.Close()
}

func TestTimeoutError(t *testing.T)  {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerTimeoutError))
	searchClient := &SearchClient{
		URL: ts.URL,
	}

	_, err := searchClient.FindUsers(SearchRequest{})

	if err == nil {
		t.Error("Timeout chck error :(")
	}

	ts.Close()
}

func TestUnknownError(t *testing.T)  {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerUnknownError))
	searchClient := &SearchClient{
		URL: "bad_link",
	}

	_, err := searchClient.FindUsers(SearchRequest{})

	if err == nil {
		t.Error("TestUnknownError :(")
	}

	ts.Close()
}

func TestStatusUnauthorized(t *testing.T)  {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerUnauthorized))
	searchClient := &SearchClient{ URL: ts.URL}
	_, err := searchClient.FindUsers(SearchRequest{})

	if err.Error() != "Bad AccessToken" {
		t.Error("Bad AccessToken is not done :(")
	}

	ts.Close()
}

func TestStatusInternalServerError(t *testing.T)  {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerInternalServerError))
	searchClient := &SearchClient{ URL: ts.URL}
	_, err := searchClient.FindUsers(SearchRequest{})

	if err.Error() != "SearchServer fatal error" {
		t.Error("SearchServer fatal error is not done :(")
	}

	ts.Close()
}

func TestBadRequest(t *testing.T)  {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerBadRequest))
	searchClient := &SearchClient{ URL: ts.URL}
	_, err := searchClient.FindUsers(SearchRequest{})

	if err.Error() != "cant unpack error json: unexpected end of JSON input" {
		t.Error("TestBadRequest is not done")
	}

	ts.Close()
}

func TestBadField(t *testing.T)  {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerBadField))
	searchClient := &SearchClient{ URL: ts.URL}
	_, err := searchClient.FindUsers(SearchRequest{})
	if err.Error() != "OrderFeld  invalid" {
		t.Error("ErrorBadOrderField is not done")
	}

	ts.Close()
}

func TestBadRequestError(t *testing.T)  {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerBadError))
	searchClient := &SearchClient{ URL: ts.URL}
	_, err := searchClient.FindUsers(SearchRequest{})
	if err == nil {
		t.Error("TestBadRequestError is not done")
	}

	ts.Close()
}


func checkError(err error)  {
	if err != nil {
		panic(err)
	}
}
