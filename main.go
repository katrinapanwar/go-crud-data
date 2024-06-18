package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-contrib/cors"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// Data represents a data entry in the database
type Data struct {
	Date  int    `json:"date"`
	Day   string `json:"day"`
	Tasks string `json:"tasks"`
}

var db *sql.DB

func main() {
	var err error
	// Connect to the MySQL database
	db, err = sql.Open("mysql", "root:admin@tcp(localhost:3306)/testdb")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Println("error verifying connection with db.Ping")
		panic(err.Error())
	}

	router := gin.Default()

	// Set up CORS middleware options
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:4200"}, // Change this to your front-end origin
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}))

	router.GET("/data", getData)
	router.GET("/data/:date", getDataByDate)
	router.POST("/data", createData)
	router.PUT("/data/:date", updateData)
	router.DELETE("/data/:date", deleteData)

	// Start the server
	if err := router.Run("localhost:8080"); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}

// getData handles the retrieval of all data entries
func getData(c *gin.Context) {
	fmt.Println("Hello from getData function")

	stmt := "SELECT * FROM data"
	rows, err := db.Query(stmt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var dataEntries []Data
	for rows.Next() {
		var data Data
		if err := rows.Scan(&data.Date, &data.Day, &data.Tasks); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		dataEntries = append(dataEntries, data)
	}
	c.JSON(http.StatusOK, dataEntries)
}

// getDataByDate handles the retrieval of a single data entry by its date
func getDataByDate(c *gin.Context) {
	date := c.Param("date")
	stmt := "SELECT * FROM data WHERE date = ?"
	row := db.QueryRow(stmt, date)

	var data Data
	if err := row.Scan(&data.Date, &data.Day, &data.Tasks); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"message": "Data not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, data)
}

// createData handles the creation of a new data entry
func createData(c *gin.Context) {
	fmt.Println("Hello from createData function")
	var newData Data
	if err := c.BindJSON(&newData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to bind JSON: " + err.Error()})
		return
	}

	stmt := "INSERT INTO data (date, day, tasks) VALUES (?, ?, ?)"
	res, err := db.Exec(stmt, newData.Date, newData.Day, newData.Tasks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	newData.Date = int(id)
	c.JSON(http.StatusCreated, newData)
}

// updateData handles the updating of an existing data entry by its date
func updateData(c *gin.Context) {
	fmt.Println("Hello from updateData function")
	date := c.Param("date")
	var updatedData Data
	if err := c.BindJSON(&updatedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to bind JSON: " + err.Error()})
		return
	}

	stmt := "UPDATE data SET day = ?, tasks = ? WHERE date = ?"
	res, err := db.Exec(stmt, updatedData.Day, updatedData.Tasks, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Data not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data successfully updated"})
}

// deleteData handles the deletion of a data entry by its date
func deleteData(c *gin.Context) {
	date := c.Param("date")

	stmt := "DELETE FROM data WHERE date = ?"
	res, err := db.Exec(stmt, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Data not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data successfully deleted"})
}
