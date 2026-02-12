package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/jiangshenghai57/andy-warhol/amortization"
	"github.com/jiangshenghai57/andy-warhol/config"
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

	var loans []amortization.LoanInfo

	// Parse JSON
	if err := c.BindJSON(&loans); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Received %d loans for processing", len(loans))

	// Validate and calculate
	results := make([]gin.H, len(loans))
	for i, loan := range loans {
		// Validate loan
		if err := loan.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Loan %d validation failed: %s", i, err.Error()),
			})
			return
		}

		// Calculate amortization table
		amortTable := loan.GetAmortizationTable()

		// Store result
		results[i] = gin.H{
			"loan_id":  loan.ID,
			"cashflow": amortTable,
		}
	}

	// Thread-safe append to mortgages
	mu.Lock()
	mortgages = append(mortgages, loans...)
	mu.Unlock()

	// Return results
	c.JSON(http.StatusOK, gin.H{
		"count":   len(loans),
		"results": results,
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
