package main

import (
	"math"

	financial "github.com/razorpay/go-financial"
	"github.com/shopspring/decimal"
)

// MortgagePool is an interface that dfeins the behavior of GenerateAmortTable.
type MortgagePool interface {
	GenerateAmortTable()
}

type LoanInfo struct {
	ID         string
	Wam        int64
	Wac        float64 // per annum
	Face       float64 // mortgage notional
	AmortTable AmortizationTable
}

type RollRateMatrix struct {
	perfArr    []float64
	dq30Arr    []float64
	dq60Arr    []float64
	dq90Arr    []float64
	dq120Arr   []float64
	dq15Arr    []float64
	dq180Arr   []float64
	defualtArr []float64
}

type PrepayArr = []float64

type AmortizationTable struct {
	ID             string
	BegBal         []float64
	Interest       []float64
	Principal      []float64
	EndBal         []float64
	Period         []int
	rollRateMatrix RollRateMatrix
}

func GetAmortizationTable(l *LoanInfo) AmortizationTable {

	// initialize
	var periods []int
	var begBal []float64
	var endBal []float64
	var interest []float64
	var principal []float64

	// initialize
	j := 0
	tmp_face := l.Face

	for i := l.Wam; i > 0; i-- {
		j += 1

		periods = append(periods, j)
		begBal = append(begBal, tmp_face)

		interest = append(interest, math.Round(tmp_face*l.Wac/12)/100)

		prinPmt, _ := financial.PPmt(
			decimal.NewFromFloat(l.Wac/12/100),
			i, l.Wam,
			decimal.NewFromFloat(-l.Face),
			decimal.NewFromFloat(0.0), 0,
		).Float64()

		principla = append(principal, math.Round(prinPmt*100)/100)

		tmp_face = tmp_face - float64(prinPmt)
	}

	var a = AmortizationTable{
		ID: l.ID,

		// interest: intPmt.Float64()
	}

	// Println(a)

	return a
}
