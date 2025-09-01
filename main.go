package main

import (
	"net/http"
	"os"

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
	var newCF mortgagePool

	if err := c.BindJSON(&newCF); err != nil {
		return
	}

	mortgages = append(mortgages, newCF)
	c.IndentedJSON(http.StatusCreated, newCF)

}

func log() *gin.Engine {
	config, _ := config.ReadConfig()

	LOG_PATH := config["LOG_PATH"]
	log_path, _ := LOG_PATH.(string)
	LOG_FILE := config["LOG_FILE"]
	log_file, _ := LOG_FILE.(string)

	f, _ := os.Create(log_path + log_file)

	gin.DefaultWriter = f
	gin.DefaultErrorWriter = f

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	router := gin.Default()

	return router
}

func main() {

	router := log()
	router.GET("/loans", getLoans)
	router.POST("/loans", requestCashflow)

	router.Run("localhost:8080")
}
