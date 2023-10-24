package main

import (
	"database/sql"
	"flag"
	"fmt"

	// "fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// Struct to represent a record in the database
type Record struct {
	ID   int
	Name string
	Age  int
}

func main() {
	// Connect to the PostgreSQL database
	corsURL := flag.String("cors-url", "http://localhost:3000", "URL to be added to CORS AllowOrigins")
	dbUser := flag.String("db-user", "postgres", "Database user")
	dbPassword := flag.String("db-password", "", "Database password")
	dbName := flag.String("db-name", "mydatabase", "Database name")
	apiPort := flag.String("api-port", "8080", "Database name")
	fmt.Printf(*corsURL)

	// Parse the command-line arguments.
	flag.Parse()

	// Use the provided flags to construct the connection string.
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", *dbUser, *dbPassword, *dbName)

	// Connect to the PostgreSQL database.
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create a new Gin router.
	r := gin.Default()

	r.Use(cors.Default())
	r.GET("/write", func(c *gin.Context) {
		name := generateRandomName()
		age := generateRandomAge()
		_, err := db.Exec("INSERT INTO mytable (name, age) VALUES ($1, $2)", name, age)
		if err != nil {
			log.Println("Error inserting record:", err)
			c.String(http.StatusInternalServerError, "Error writing record")
			return
		}
		c.String(http.StatusOK, "Record written successfully")
	})

	// Endpoint to poll the table and get all records, recording the response time
	r.GET("/poll", func(c *gin.Context) {
		start := time.Now()
		rows, err := db.Query("SELECT * FROM mytable")
		if err != nil {
			log.Println("Error querying records:", err)
			c.String(http.StatusInternalServerError, "Error polling records")
			return
		}
		defer rows.Close()

		var records []Record
		for rows.Next() {
			var id int // Add id column
			var name string
			var age int
			err := rows.Scan(&id, &name, &age) // Include id in scanning
			if err != nil {
				log.Println("Error scanning row:", err)
				continue // Skip this row and proceed to the next one
			}
			records = append(records, Record{ID: id, Name: name, Age: age}) // Update the Record struct to include ID
		}

		if err := rows.Err(); err != nil {
			log.Println("Error iterating rows:", err)
			c.String(http.StatusInternalServerError, "Error polling records")
			return
		}

		responseTime := time.Since(start).Seconds()
		c.JSON(http.StatusOK, gin.H{
			"totalRecords": len(records),
			"responseTime": responseTime,
		})
	})

	go generateRandomRecords(db)

	// Start the server
	r.Run(fmt.Sprintf(":%s", *apiPort))
}

// Helper functions to generate random name and age
func generateRandomName() string {
	names := []string{"Alice", "Bob", "Charlie", "David", "Eva", "Frank", "Grace"}
	return names[rand.Intn(len(names))]
}

func generateRandomAge() int {
	return rand.Intn(50) + 20 // Generate age between 20 and 70
}

// Function to generate random records every 500ms
func generateRandomRecords(db *sql.DB) {
	for {
		name := generateRandomName()
		age := generateRandomAge()

		_, err := db.Exec("INSERT INTO mytable (name, age) VALUES ($1, $2)", name, age)
		if err != nil {
			log.Println("Error inserting record:", err)
		}

		time.Sleep(10 * time.Millisecond) // Wait for 500ms before generating the next record
	}
}
