package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"amortization"
	"config"

	"github.com/gin-gonic/gin"
)

type mortgagePool struct {
	ID       string  `json:"id"`
	WAC      float64 `json:"wac"`
	WAM      int     `json:"wam"`
	StaticDQ bool    `json:"staticdq"`
}

var mortgages = []mortgagePool{
	{ID: "1", WAC: 3.4, WAM: 240, StaticDQ: true},
	{ID: "1", WAC: 6.4, WAM: 360, StaticDQ: true},
	{ID: "1", WAC: 5.0, WAM: 120, StaticDQ: true},
}

func getLoans(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, mortgages)
}

func requestCashflow(c *gin.Context) {

	log.Println("requestCashflow endpoint was hit")

	var newCF mortgagePool

	if err := c.BindJSON(&newCF); err != nil {
		return
	}

	log.Println("New cashflow received:", newCF)

	// Convert to LoanInfo for amortization calculation
	loanInfo := &amortization.LoanInfo{
		ID:   newCF.ID,
		Wam:  int64(newCF.WAM),
		Wac:  newCF.WAC,
		Face: 100000.0, // You might want to add this field to mortgagePool
	}

	// Calculate amortization table
	amortTable := amortization.GetAmortizationTable(loanInfo)

	mortgages = append(mortgages, newCF)

	// Get current local time zone date
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		loc = time.Local
	}
	localNow := time.Now().In(loc).Format(time.RFC3339)

	// Create response data
	responseData := gin.H{
		"mortgage":    newCF,
		"local_date":  localNow,
		"amort_table": amortTable,
	}

	// Save to JSON file
	filename := "cashflow_" + newCF.ID + "_" + time.Now().Format("20060102_150405") + ".json"
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Error creating file: %v", err)
	} else {
		defer file.Close()
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(responseData); err != nil {
			log.Printf("Error writing JSON: %v", err)
		} else {
			log.Printf("Cashflow data saved to: %s", filename)
		}
	}

	c.IndentedJSON(http.StatusCreated, responseData)
}

func multiLog() *gin.Engine {
	config, _ := config.ReadConfig()

	LOG_PATH := config["LOG_PATH"]
	log_path, _ := LOG_PATH.(string)
	LOG_FILE := config["LOG_FILE"]
	log_file, _ := LOG_FILE.(string)

	f, _ := os.Create(log_path + log_file)

	mw := io.MultiWriter(f, os.Stdout)

	gin.DefaultWriter = mw
	gin.DefaultErrorWriter = mw
	log.Println(config)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	router := gin.Default()

	return router
}

func main() {

	router := multiLog()
	router.GET("/loans", getLoans)
	router.POST("/loans", requestCashflow)

	router.Run("localhost:8080")
}
