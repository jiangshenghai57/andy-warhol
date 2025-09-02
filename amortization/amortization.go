// Package amortization provides mortgage loan amortization calculations
// and related financial computations.
package amortization

import (
	"math"

	financial "github.com/razorpay/go-financial"
	"github.com/shopspring/decimal"
)

// MortgagePool defines the behavior for generating amortization tables.
// Types implementing this interface can generate their own amortization schedules.
type MortgagePool interface {
	// GenerateAmortTable creates an amortization table for the mortgage pool.
	GenerateAmortTable() AmortizationTable
	ExtendPrepayArr() []float64
	TrueUpBalances()
}

// LoanInfo represents basic loan information used for amortization calculations.
// This structure can be extended in the future to include additional factors
// such as interest rate adjustments, inflation factors, escrow balances, etc.
type LoanInfo struct {
	ID         string            // Unique identifier for the loan
	Wam        int64             // Weighted Average Maturity in months
	Wac        float64           // Weighted Average Coupon rate per annum in percentage points (e.g., 6.75)
	Face       float64           // Mortgage notional/principal amount
	PrepayCPR  []float64         // prepay CPR in decimals, could be SMM
	AmortTable AmortizationTable // Associated amortization table
	StaticDQ   bool              // If true amortization uses a roll rate matrix
}

type RollRateTransitionMatrix struct {
	// Define the structure for the roll rate matrix
	// [0.92, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01]
	// should sum up to 1.0, and each element represents the transition probability
	// from one delinquency status to another.
	// For example, if the performing transformation is [0.92, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01],
	// it means there is a 92% chance that a performing loan will remain performing,
	// and a 1% chance it will transition to each of the delinquent statuses.
	// Length of the array should be equal to the number of delinquency statuses and
	// RollRateMatrix struct length
	PerformingTransition []float64
	DQ30Transition       []float64
	DQ60Transition       []float64
	DQ90Transition       []float64
	DQ120Transition      []float64
	DQ150Transition      []float64
	DQ180Transition      []float64
	DefaultTransition    []float64
}

// DelinqArrays contains delinquency performance arrays for different time periods.
// These arrays track loan performance across various delinquency buckets.
type DelinqArrays struct {
	perfArr    []float64 // Current/performing loans
	dq30Arr    []float64 // 30-day delinquent loans
	dq60Arr    []float64 // 60-day delinquent loans
	dq90Arr    []float64 // 90-day delinquent loans
	dq120Arr   []float64 // 120-day delinquent loans
	dq15Arr    []float64 // 150-day delinquent loans
	dq180Arr   []float64 // 180-day delinquent loans
	defaultArr []float64 // Defaulted loans
}

// AmortizationTable represents a complete loan amortization schedule.
// It contains all payment components and balances for each period of the loan.
type AmortizationTable struct {
	BegBal          []float64    // Beginning balance for each period
	Interest        []float64    // Interest payment for each period
	Principal       []float64    // Principal payment for each period
	SchedBal        []float64    // Scheduled balance after payment
	PrepayAmountArr []float64    // Prepayment amount for each period
	EndBal          []float64    // Ending balance for each period
	Period          []int        // Period numbers (1, 2, 3, ...)
	DelinqArrays    DelinqArrays // Delinquency performance arrays
}

// if LoanInfo.Prepay on has only one element, extend the prepay array to
// match the loan term
func ExtendPrepayArr(l *LoanInfo) {
	if len(l.Prepay) == 1 {
		smm = []float64{1 + math.Pow(1-l.Prepay[0], 1.0/12.0)}
		l.Prepay = append(l.Prepay, make([]float64, l.Wam-1)...)
	}
}

// GetAmortizationTable calculates and returns a complete amortization table
// for the given loan information.
//
// The function uses the loan's WAC (interest rate), WAM (term), and Face (principal)
// to calculate monthly payments, interest, and principal components for each period.
// All monetary values are rounded to 2 decimal places.
//
// Parameters:
//   - l: Pointer to LoanInfo containing loan parameters
//
// Returns:
//   - AmortizationTable: Complete amortization schedule with period-by-period breakdown
//
// Example:
//
//	loanInfo := &LoanInfo{
//	    ID:   "LOAN001",
//	    Wam:  360,        // 30 years
//	    Wac:  4.5,        // 4.5% annual rate
//	    Face: 250000.0,   // $250,000 loan
//	}
//	table := GetAmortizationTable(loanInfo)
func GetAmortizationTable(l *LoanInfo) AmortizationTable {
	// Initialize arrays to store amortization components
	var periods []int
	var begBal []float64
	var schedBal []float64
	var endBal []float64
	var prepayAmountArr []float64
	var interest []float64
	var principal []float64

	// Initialize counters and working variables
	j := 0
	tmp_face := l.Face

	// Extend prepayment array if necessary
	ExtendPrepayArr(l)

	// Calculate amortization for each period from WAM down to 1
	for i := l.Wam; i > 0; i-- {
		j += 1

		// Store period number
		periods = append(periods, j)

		// Calculate and store beginning balance (rounded to 2 decimal places)
		begBal = append(begBal, math.Round(tmp_face*100)/100)

		// Calculate monthly interest payment
		// Formula: Principal * (Annual Rate / 12) / 100
		interest = append(interest, math.Round(tmp_face*l.Wac/12)/100)

		// Calculate principal payment using financial library
		// PPmt calculates principal payment for a given period
		prinPmt, _ := financial.PPmt(
			decimal.NewFromFloat(l.Wac/12/100), // Monthly interest rate as decimal
			i,                                  // Current period (remaining periods)
			l.Wam,                              // Total loan term
			decimal.NewFromFloat(-l.Face),      // Present value (negative for payment calculation)
			decimal.NewFromFloat(0.0),          // Future value (loan fully paid)
			0,                                  // Payment timing (0 = end of period)
		).Float64()

		// Store principal payment (rounded to 2 decimal places)
		principal = append(principal, math.Round(prinPmt*100)/100)

		schedBal = append(schedBal, math.Round((tmp_face-prinPmt)*100)/100)

		// Store prepayment amount (rounded to 2 decimal places)
		prepayAmountArr = append(prepayAmountArr, math.Round(l.Prepay[j-1]*schedBal[j-1]*100)/100)

		// Reduce remaining balance by principal payment
		tmp_face = tmp_face - float64(prinPmt)

		// Ensure balance doesn't go negative
		if tmp_face < 0.0 {
			tmp_face = 0
		}

		// Store ending balance (rounded to 2 decimal places)
		endBal = append(endBal, math.Round(tmp_face*100)/100)
	}

	// Construct and return the complete amortization table
	amortTable := AmortizationTable{
		Period:          periods,
		BegBal:          begBal,
		SchedBal:        schedBal,
		PrepayAmountArr: prepayAmountArr,
		Interest:        interest,
		Principal:       principal,
		EndBal:          endBal,
	}

	amortTable.TrueUpBalances()

	return amortTable
}

// true up the last period's ending balance in principal amount, prepay, and endBal

func (a *AmortizationTable) TrueUpBalances() {

	lastIndex := len(a.Principal) - 1
	// Get the last period's values
	lastBegBal := a.BegBal[lastIndex]
	lastPrincipal := a.Principal[lastIndex]
	lastPrepay := a.PrepayAmountArr[lastIndex]
	lastEndBal := a.EndBal[lastIndex]

	leftOver := lastBegBal - lastPrincipal - lastPrepay

	if leftOver == lastEndBal {
		return
	}

	if leftOver < 0 {
		a.Principal[lastIndex] = lastPrincipal + leftOver
	}

	// Ensure all values are consistent
	a.Principal[lastIndex] = lastPrincipal
}
