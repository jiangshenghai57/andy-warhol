package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"amortization"
	"config"

	"github.com/gin-gonic/gin"
)

var (
	mortgages  = []amortization.LoanInfo{}
	mu         sync.RWMutex // Protect the mortgages slice
	workerPool = make(chan struct{}, 100)
)

func getLoans(c *gin.Context) {
	mu.RLock()
	defer mu.RUnlock()
	c.IndentedJSON(http.StatusOK, mortgages)
}

func getServiceInfo(c *gin.Context) {
	info := gin.H{
		"service": "andy-warhol",
		"description": "Mortgage loan amortization calculation service",
		"version": "1.0.0",
		"endpoints": gin.H{
			"GET /info": "Get service information and capabilities",
			"GET /loans": "Retrieve list of processed loans",
			"POST /loans": "Submit loan data for amortization calculation",
		},
		"capabilities": []string{
			"Loan amortization schedule generation",
			"CPR to SMM conversion for prepayment modeling",
			"Concurrent loan processing",
			"JSON serialization for API responses",
			"Delinquency tracking support",
		},
		"loan_parameters": gin.H{
			"id": "Unique loan identifier (string)",
			"wam": "Weighted Average Maturity in months (integer)",
			"wac": "Weighted Average Coupon rate per annum as percentage (float)",
			"face": "Mortgage principal amount in dollars (float)",
			"prepay_cpr": "Conditional Prepayment Rate as decimal (float, optional)",
		},
	}
	c.IndentedJSON(http.StatusOK, info)
}

func requestCashflow(c *gin.Context) {
	log.Println("requestCashflow endpoint was hit")

	var newCFs []amortization.LoanInfo // Change to slice to accept multiple loans

	if err := c.BindJSON(&newCFs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		log.Printf("Error binding JSON: %v", err)
		return
	}

	log.Printf("Received %d loans for processing", len(newCFs))

	// Thread-safe append to mortgages
	mu.Lock()
	mortgages = append(mortgages, newCFs...)
	mu.Unlock()

	// Get current local time zone date
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		loc = time.Local
	}
	localNow := time.Now().In(loc).Format(time.RFC3339)

	// Process each loan in a separate goroutine
	for _, newCF := range newCFs {
		go func(loan amortization.LoanInfo) {
			workerPool <- struct{}{}        // Acquire worker
			defer func() { <-workerPool }() // Release worker

			log.Printf("Starting amortization calculation for loan %s", loan.ID)

			loanInfo := &amortization.LoanInfo{
				ID:   loan.ID,
				Wam:  int64(loan.Wam),
				Wac:  loan.Wac,
				Face: loan.Face,
			}

			amortTable := loanInfo.GetAmortizationTable() // Call method on LoanInfo if GenerateAmortTable is a method

			// Save to JSON file
			responseData := gin.H{
				"mortgage":    loan,
				"local_date":  localNow,
				"amort_table": amortTable,
			}

			filename := "output/cashflow_" + loan.ID + "_" + time.Now().Format("20060102_150405") + ".json"

			// Create output directory if it doesn't exist
			os.MkdirAll("output", 0755)

			file, err := os.Create(filename)
			if err != nil {
				log.Printf("Error creating file: %v", err)
				return
			}
			defer file.Close()

			encoder := json.NewEncoder(file)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(responseData); err != nil {
				log.Printf("Error writing JSON: %v", err)
			} else {
				log.Printf("Cashflow data saved to: %s", filename)
			}
			log.Printf("Completed amortization calculation for loan %s", loan.ID)
		}(newCF) // Pass loan as parameter to avoid closure issues
	}

	// Return immediate response
	c.JSON(http.StatusAccepted, gin.H{
		"message":    fmt.Sprintf("Received %d loans, amortization calculations started", len(newCFs)),
		"loan_count": len(newCFs),
		"local_date": localNow,
	})
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
	router.GET("/info", getServiceInfo)
	router.GET("/loans", getLoans)
	router.POST("/loans", requestCashflow)

	router.Run("localhost:8080")
}
