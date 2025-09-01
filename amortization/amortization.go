package amortization

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

type AmortizationTable struct {
	ID             string
	BegBal         []float64
	Interest       []float64
	Principal      []float64
	SchedBal       []float64
	PrepayArr      []float64
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
		begBal = append(begBal, math.Round(tmp_face*100)/100)

		interest = append(interest, math.Round(tmp_face*l.Wac/12)/100)

		prinPmt, _ := financial.PPmt(
			decimal.NewFromFloat(l.Wac/12/100),
			i, l.Wam,
			decimal.NewFromFloat(-l.Face),
			decimal.NewFromFloat(0.0), 0,
		).Float64()

		principal = append(principal, math.Round(prinPmt*100)/100)

		tmp_face = tmp_face - float64(prinPmt)

		// zero out
		if tmp_face < 0.0 {
			tmp_face = 0
		}

		endBal = append(endBal, math.Round(tmp_face*100)/100)
	}

	amortTable := AmortizationTable{
		ID:        l.ID,
		Period:    periods,
		BegBal:    begBal,
		EndBal:    endBal,
		Interest:  interest,
		Principal: principal,
	}

	return amortTable
}
