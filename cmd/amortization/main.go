package main

import (
	"log"

	"github.com/jiangshenghai57/andy-warhol/amortization"
)

func main() {

	loanInfo := amortization.LoanInfo{
		ID:   "LOAN001",
		Wam:  360,
		Wac:  4.5,
		Face: 250000.0,
	}

	// Create a new PrepayInfo instance
	prepayInfo := &amortization.PrepayInfo{
		PrepayCPR: 0.05, // 5% CPR
	}

	// Convert CPR to SMM
	prepayInfo.ConvertCPRToSMM(int(loanInfo.Wam))

	// Log the results
	log.Println("SMM Array:", prepayInfo.SMMArr)
	log.Println("Length of SMM Array:", len(prepayInfo.SMMArr))
}
