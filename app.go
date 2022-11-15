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
	"github.com/wa-candra/webservice-go/appmode"
)

type App struct {
	DB     *sql.DB
	router *gin.Engine
}

func (app *App) Init(mode appmode.AppMode) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var dbname string
	switch mode {
	case appmode.Production:
		dbname = os.Getenv("DBNAME_PROD")
	case appmode.Development:
		dbname = os.Getenv("DBNAME_DEV")
	case appmode.Testing:
		dbname = os.Getenv("DBNAME_TEST")
	default:
		panic("Invalid App mode")
	}

	// Capture connection properties.
	cfg := mysql.Config{
		User:                 os.Getenv("DBUSER"),
		Passwd:               os.Getenv("DBPASS"),
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               dbname,
		AllowNativePasswords: true,
	}
	// Get a database handle.
	app.DB, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	err = app.DB.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("DB Connected!")

	app.initRoutes()
}

func (app *App) initRoutes() {
	app.router = gin.Default()
	app.router.GET("/albums", app.getAlbums)
	app.router.POST("/albums", app.addAlbum)
	app.router.PATCH("/albums/:id", app.updateAlbum)
	app.router.GET("/albums/:id", app.getAlbumByID)
	app.router.GET("/albums/artist/:name", app.getAlbumsByArtist)
	app.router.DELETE("/albums/:id", app.deleteAlbum)
}

func (app *App) Run(addr string) {
	//check if DB and Router already defined
	if app.DB == nil {
		log.Fatal("app.DB is not defined yet!")
	}
	if app.router == nil {
		log.Fatal("app.router is not defined yet!")
	}
	app.router.Run(addr)
}

func (app *App) getAlbums(c *gin.Context) {
	// An albums slice to hold data from returned rows.
	var albums []Album

	albums, err := getAlbumsA(app.DB, 0, 10)

	if err != nil {
		log.Fatal(err)
		c.JSON(http.StatusBadGateway, gin.H{"message": "Error while retrieving data"})
	}
	c.JSON(http.StatusOK, albums)
}

func (app *App) getAlbumsByArtist(c *gin.Context) {
	artistName := c.Param("name")

	albums, err := getAlbumsByArtistA(app.DB, artistName, 10)

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

func (app *App) addAlbum(c *gin.Context) {
	var newAlbum Album

	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&newAlbum); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Cannot parse the req body"})
		return
	}
	err := newAlbum.createAlbum(app.DB)

	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"message": "Error while creating data"})
		return
	}

	c.JSON(http.StatusCreated, newAlbum)
}

func (app *App) updateAlbum(c *gin.Context) {
	id := c.Param("id")

	album := Album{ID: id}
	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&album); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Cannot parse the req body"})
		return
	}

	err := album.updateAlbum(app.DB)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"message": "Error while updating data"})
		return
	}

	c.JSON(http.StatusOK, album)
}

// getAlbumByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func (app *App) getAlbumByID(c *gin.Context) {
	id := c.Param("id")

	// An album to hold data from the returned row.
	alb := Album{ID: id}

	err := alb.getAlbum(app.DB)
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

func (app *App) deleteAlbum(c *gin.Context) {
	id := c.Param("id")

	alb := Album{ID: id}

	err := alb.deleteAlbum(app.DB)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"message": "Error while deleting data"})
		log.Fatal(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Delete data success!"})
}
