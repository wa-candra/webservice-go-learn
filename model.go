package main

import (
	"database/sql"
	"strconv"
)

type Album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title" binding:"required"`
	Artist string  `json:"artist" binding:"required"`
	Price  float64 `json:"price" binding:"required"`
}

func getAlbumsA(db *sql.DB, start, count int) ([]Album, error) {
	rows, err := db.Query("SELECT * FROM album LIMIT ? OFFSET ?", count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	albums := []Album{}

	for rows.Next() {
		var alb Album
		if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
			return nil, err
		}
		albums = append(albums, alb)
	}

	return albums, nil
}

func getAlbumsByArtistA(db *sql.DB, artistName string, count int) ([]Album, error) {
	rows, err := db.Query("SELECT * FROM album WHERE artist = ? LIMIT ?", artistName, count)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	albums := []Album{}

	for rows.Next() {
		var alb Album
		if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
			return nil, err
		}
		albums = append(albums, alb)
	}

	return albums, nil
}

func (album *Album) getAlbum(db *sql.DB) error {
	return db.QueryRow("SELECT title, artist, price FROM album WHERE id=?",
		album.ID).Scan(&album.Title, &album.Artist, &album.Price)
}

func (album *Album) createAlbum(db *sql.DB) error {
	result, err := db.Exec("INSERT INTO album (title, artist, price) VALUES (?, ?, ?)", album.Title, album.Artist, album.Price)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	album.ID = strconv.FormatInt(id, 10)
	return nil
}

func (album *Album) updateAlbum(db *sql.DB) error {
	_, err :=
		db.Exec("UPDATE album SET title=?, artist=?, price=? WHERE id=?",
			album.Title, album.Artist, album.Price, album.ID)

	return err
}

func (album *Album) deleteAlbum(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM album WHERE id=?", album.ID)

	return err
}
