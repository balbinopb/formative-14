package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var db *sql.DB

type Bioskop struct {
	ID     int     `json:"id"`
	Nama   string  `json:"nama"`
	Lokasi string  `json:"lokasi"`
	Rating float64 `json:"rating"`
}

// post method
func PostBioskop(c *gin.Context) {
	var input Bioskop

	err := c.ShouldBindJSON(&input) //bind

	//
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// validasi nama dan lokasi harus diisi
	if input.Nama == "" || input.Lokasi == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama dan Lokasi wajib diisi"})
		return
	}

	query := `INSERT INTO bioskop_db  (nama, lokasi, rating) VALUES ($1, $2, $3) RETURNING id`

	err = db.QueryRow(query, input.Nama, input.Lokasi, input.Rating).Scan(&input.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menambahkan data"})
		return
	}

	c.JSON(http.StatusOK, input)
}

// get all method
func GetAllBioskop(c *gin.Context) {
	rows, err := db.Query("SELECT id, nama, lokasi, rating FROM bioskop_db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data"})
		return
	}
	defer rows.Close()

	var bioskops []Bioskop
	for rows.Next() {
		var b Bioskop
		if err := rows.Scan(&b.ID, &b.Nama, &b.Lokasi, &b.Rating); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca data"})
			return
		}
		bioskops = append(bioskops, b)
	}

	c.JSON(http.StatusOK, bioskops)
}

// get by id method
func GetBioskopByID(c *gin.Context) {
	id := c.Param("id")
	var b Bioskop

	err := db.QueryRow("SELECT id, nama, lokasi, rating FROM bioskop_db WHERE id = $1", id).Scan(&b.ID, &b.Nama, &b.Lokasi, &b.Rating)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data tidak ditemukan"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data"})
		return
	}

	c.JSON(http.StatusOK, b)
}

// update or put method
func UpdateBioskop(c *gin.Context) {
	id := c.Param("id")
	var input Bioskop
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if input.Nama == "" || input.Lokasi == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama dan Lokasi wajib diisi"})
		return
	}

	query := `UPDATE bioskop_db SET nama = $1, lokasi = $2, rating = $3 WHERE id = $4`
	res, err := db.Exec(query, input.Nama, input.Lokasi, input.Rating, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate data"})
		return
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data berhasil diupdate"})
}

// delete method
func DeleteBioskop(c *gin.Context) {
	id := c.Param("id")

	res, err := db.Exec("DELETE FROM bioskop_db WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus data"})
		return
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data berhasil dihapus"})
}

func main() {

	if _, exists := os.LookupEnv("RAILWAY_ENVIRONMENT"); !exists {
		_ = godotenv.Load()
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		os.Getenv("PGHOST"), os.Getenv("PGPORT"), os.Getenv("PGUSER"), os.Getenv("PGPASSWORD"), os.Getenv("PGDATABASE"),
	)

	var err error

	db, err = sql.Open("postgres", connStr)

	if err != nil {
		log.Printf("Gagal koneksi DB: %v", err)
		return
	}

	err = db.Ping()
	if err != nil {
		log.Printf("Gagal koneksi DB: %v", err)
		return
	}

	fmt.Println("DATABASE IS CONNECTED!!!")

	router := gin.Default()

	// create
	router.POST("/bioskop", PostBioskop)

	//read
	router.GET("/bioskop", GetAllBioskop)
	router.GET("/bioskop/:id", GetBioskopByID)

	//update
	router.PUT("/bioskop/:id", UpdateBioskop)

	//delete
	router.DELETE("/bioskop/:id", DeleteBioskop)

	port := os.Getenv("PGPORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("\nServer running di: http://localhost:%s ...\n", port)
	router.Run(":" + port)

}
