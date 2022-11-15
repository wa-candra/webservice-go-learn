package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/wa-candra/webservice-go/appmode"
)

var a App

func TestMain(m *testing.M) {
	a.Init(appmode.Testing)
	ensureTableExists()
	resultCode := m.Run()
	clearTable()
	os.Exit(resultCode)
}

func TestEmptyTable(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/albums", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected empty array. Got %s", body)
	}
}

func TestGetNonExistentProduct(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/albums/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["message"] != "album not found" {
		t.Errorf("Expected the 'error' key of response to be set to 'album not found'. Got '%s'", m["message"])
	}
}

func TestCreateProduct(t *testing.T) {
	clearTable()

	var jsonStr = []byte(`{"title": "test album", "artist": "test artist", "price": 11.22}`)
	req, _ := http.NewRequest("POST", "/albums", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["title"] != "test album" {
		t.Errorf("Expected product name to be 'test album'. Got %v", m["title"])
	}

	if m["artist"] != "test artist" {
		t.Errorf("Expected product name to be 'test artist'. Got %v", m["artist"])
	}

	if m["price"] != 11.22 {
		t.Errorf("Expected product price to be '11.22'. Got %v", m["price"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["id"] != "1" {
		t.Errorf("Expected product ID to be '1'. Got '%v', type: %v", m["id"], reflect.TypeOf(m["id"]))
	}
}

func TestGetProduct(t *testing.T) {
	clearTable()
	addAlbums(1)

	req, _ := http.NewRequest("GET", "/albums/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestUpdateProduct(t *testing.T) {
	clearTable()
	addAlbums(1)

	req, _ := http.NewRequest("GET", "/albums/1", nil)
	response := executeRequest(req)
	var originalProduct map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalProduct)

	var jsonStr = []byte(`{"title":"test album - updated name", "artist": "test artist","price": 11.22}`)
	req, _ = http.NewRequest("PATCH", "/albums/1", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != originalProduct["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalProduct["id"], m["id"])
	}

	if m["title"] == originalProduct["title"] {
		t.Errorf("Expected the title to change from '%v' to '%v'. Got '%v'", originalProduct["title"], m["title"], m["title"])
	}

	if m["artist"] == originalProduct["artist"] {
		t.Errorf("Expected the artist to change from '%v' to '%v'. Got '%v'", originalProduct["artist"], m["artist"], m["artist"])
	}

	if m["price"] == originalProduct["price"] {
		t.Errorf("Expected the price to change from '%v' to '%v'. Got '%v'", originalProduct["price"], m["price"], m["price"])
	}
}

func TestDeleteProduct(t *testing.T) {
	clearTable()
	addAlbums(1)

	req, _ := http.NewRequest("GET", "/albums/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/albums/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/albums/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d \n", expected, actual)
	}
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS album
(
	id         INT AUTO_INCREMENT NOT NULL,
  title      VARCHAR(128) NOT NULL,
  artist     VARCHAR(255) NOT NULL,
  price      DECIMAL(5,2) NOT NULL,
  PRIMARY KEY (id)
)`

func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func addAlbums(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		a.DB.Exec("INSERT INTO album(title, artist, price) VALUES(?, ?, ?)", "Album "+strconv.Itoa(i), "Artist "+strconv.Itoa(i), (i+1.0)*10)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM album")
	a.DB.Exec("ALTER TABLE album AUTO_INCREMENT = 1")
}
