package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// Album represents data about a record Album

var db *sql.DB

func getAlbums(c *gin.Context) {
	// An albums slice to hold data from returned rows.
	var albums []Album

	albums, err := getAlbumsA(db, 0, 10)

	if err != nil {
		log.Fatal(err)
		c.JSON(http.StatusBadGateway, gin.H{"message": "Error while retrieving data"})
	}
	c.JSON(http.StatusOK, albums)
}

func getAlbumsByArtist(c *gin.Context) {
	artistName := c.Param("name")

	albums, err := getAlbumsByArtistA(db, artistName, 10)

	if err != nil {
		log.Fatal(err)
		c.JSON(http.StatusBadGateway, gin.H{"message": "Error while retrieving data"})
	}

	if len(albums) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "album not found"})
		return
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
	err := newAlbum.createAlbum(db)

	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"message": "Error while creating data"})
		return
	}

	c.JSON(http.StatusCreated, newAlbum)
}

func updateAlbum(c *gin.Context) {
	id := c.Param("id")

	album := Album{ID: id}
	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&album); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Cannot parse the req body"})
		return
	}

	err := album.updateAlbum(db)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"message": "Error while updating data"})
		return
	}

	c.JSON(http.StatusOK, album)
}

// getAlbumByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func getAlbumByID(c *gin.Context) {
	id := c.Param("id")

	// An album to hold data from the returned row.
	alb := Album{ID: id}

	err := alb.getAlbum(db)
	if err != nil {
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

	alb := Album{ID: id}

	err := alb.deleteAlbum(db)
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
	router.PATCH("/albums/:id", updateAlbum)
	router.GET("/albums/:id", getAlbumByID)
	router.GET("/albums/artist/:name", getAlbumsByArtist)
	router.DELETE("/albums/:id", deleteAlbum)
	router.Run("localhost:8080")
}
