package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// Album represents data about a record Album.
type Album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title" binding:"required"`
	Artist string  `json:"artist" binding:"required"`
	Price  float64 `json:"price" binding:"required"`
}

var db *sql.DB

func getAlbums(c *gin.Context) {
	// An albums slice to hold data from returned rows.
	var albums []Album

	rows, err := db.Query("SELECT * FROM album LIMIT 10")
	if err != nil {
		log.Fatal(err)
		c.JSON(http.StatusBadGateway, gin.H{"message": "Error while retrieving data"})
	}
	defer rows.Close()
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var alb Album
		if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
			log.Fatal(err)
			c.JSON(http.StatusBadGateway, gin.H{"message": "Error while retrieving data"})
		}
		albums = append(albums, alb)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
		c.JSON(http.StatusBadGateway, gin.H{"message": "Error while retrieving data"})
	}

	c.JSON(http.StatusOK, albums)
}

func addAlbum(c *gin.Context) {
	var newAlbum Album

	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&newAlbum); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Cannot parse the req body"})
		return
	}

	result, err := db.Exec("INSERT INTO album (title, artist, price) VALUES (?, ?, ?)", newAlbum.Title, newAlbum.Artist, newAlbum.Price)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"message": "Error while creating data"})
		return
	}
	id, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"message": "Error while creating data"})
		return
	}
	newAlbum.ID = strconv.FormatInt(id, 10)
	c.JSON(http.StatusCreated, newAlbum)
}

// getAlbumByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func getAlbumByID(c *gin.Context) {
	id := c.Param("id")

	// An album to hold data from the returned row.
	var alb Album
	row := db.QueryRow("SELECT * FROM album WHERE id = ?", id)
	if err := row.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"message": "album not found"})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"message": "err while retrieving data"})
		return
	}

	c.JSON(http.StatusOK, alb)
}

func deleteAlbum(c *gin.Context) {
	id := c.Param("id")

	result, err := db.Exec("DELETE FROM album WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"message": "Error while deleting data"})
		log.Fatal(err)
		return
	}
	_, err = result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"message": "Error while deleting data"})
		log.Fatal(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Delete data success!"})
}

func main() {
	errEnv := godotenv.Load(".env")
	if errEnv != nil {
		log.Fatal("Error loading .env file")
	}

	// Capture connection properties.
	cfg := mysql.Config{
		User:                 os.Getenv("DBUSER"),
		Passwd:               os.Getenv("DBPASS"),
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               os.Getenv("DBNAME"),
		AllowNativePasswords: true,
	}
	// Get a database handle.
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")

	router := gin.Default()
	router.GET("/albums", getAlbums)
	router.POST("/albums", addAlbum)
	router.GET("/albums/:id", getAlbumByID)
	router.DELETE("/albums/:id", deleteAlbum)
	router.Run("localhost:8080")
}
