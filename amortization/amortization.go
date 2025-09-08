// Package amortization provides mortgage loan amortization calculations
// and related financial computations.
package amortization

import (
	"fmt"
	"math"

	"log"
)

// MortgagePool defines the behavior for generating amortization tables.
// Types implementing this interface can generate their own amortization schedules.
type MortgagePool interface {
	// GenerateAmortTable creates an amortization table for the mortgage pool.
	GenerateAmortTable() AmortizationTable
	ConvertCPRToSMM() []float64
	TrueUpBalances()
}

// LoanInfo represents basic loan information used for amortization calculations.
// This structure can be extended in the future to include additional factors
// such as interest rate adjustments, inflation factors, escrow balances, etc.
type LoanInfo struct {
	ID        string    `json:"id"`                // Unique identifier for the loan
	Wam       int64     `json:"wam"`               // Weighted Average Maturity in months
	Wac       float64   `json:"wac"`               // Weighted Average Coupon rate per annum in percentage points (e.g., 6.75)
	Face      float64   `json:"face"`              // Mortgage notional/principal amount
	PrepayCPR float64   `json:"prepay_cpr"`        // prepay CPR in decimals, could be SMM
	SMMArr    []float64 `json:"smm_arr,omitempty"` // SMM array for prepayment calculations
	StaticDQ  bool      `json:"static_dq"`         // If true amortization uses a roll rate matrix
	// AmortTable AmortizationTable `json:"amort_table,omitempty"` // Associated amortization table
	// Define the structure for the roll rate matrix
	// [0.92, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01]
	// should sum up to 1.0, and each element represents the transition probability
	// from one delinquency status to another.
	// For example, if the performing transformation is [0.92, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01],
	// it means there is a 92% chance that a performing loan will remain performing,
	// and a 1% chance it will transition to each of the delinquent statuses.
	// DQ30Transition represents the transition probabilities for loans that are 30 days delinquent.
	// [0.90, 0.03, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01]
	// 90% change to performing, 3% stay at 30-day delinquent, and so on.
	// Length of the array should be equal to the number of delinquency statuses and
	// RollRateMatrix struct length
	PerformingTransition []float64 `json:"performing_transition,omitempty"`
	DQ30Transition       []float64 `json:"dq30_transition,omitempty"`
	DQ60Transition       []float64 `json:"dq60_transition,omitempty"`
	DQ90Transition       []float64 `json:"dq90_transition,omitempty"`
	DQ120Transition      []float64 `json:"dq120_transition,omitempty"`
	DQ150Transition      []float64 `json:"dq150_transition,omitempty"`
	DQ180Transition      []float64 `json:"dq180_transition,omitempty"`
	DefaultTransition    []float64 `json:"default_transition,omitempty"`
}

// DelinqArrays contains delinquency performance arrays for different time periods.
// These arrays track loan performance across various delinquency buckets.
type DelinqArrays struct {
	PerfArr    []float64 `json:"perf_arr"`    // Current/performing loans
	DQ30Arr    []float64 `json:"dq30_arr"`    // 30-day delinquent loans
	DQ60Arr    []float64 `json:"dq60_arr"`    // 60-day delinquent loans
	DQ90Arr    []float64 `json:"dq90_arr"`    // 90-day delinquent loans
	DQ120Arr   []float64 `json:"dq120_arr"`   // 120-day delinquent loans
	DQ150Arr   []float64 `json:"dq150_arr"`   // 150-day delinquent loans (fixed typo)
	DQ180Arr   []float64 `json:"dq180_arr"`   // 180-day delinquent loans
	DefaultArr []float64 `json:"default_arr"` // Defaulted loans
}

// AmortizationTable represents a complete loan amortization schedule.
// It contains all payment components and balances for each period of the loan.
type AmortizationTable struct {
	BegBal          []float64    `json:"beg_bal"`           // Beginning balance for each period
	Interest        []float64    `json:"interest"`          // Interest payment for each period
	Principal       []float64    `json:"principal"`         // Principal payment for each period
	SchedBal        []float64    `json:"sched_bal"`         // Scheduled balance after payment
	PrepayAmountArr []float64    `json:"prepay_amount_arr"` // Prepayment amount for each period
	EndBal          []float64    `json:"end_bal"`           // Ending balance for each period
	Period          []int        `json:"period"`            // Period numbers (1, 2, 3, ...)
	DelinqArrays    DelinqArrays `json:"delinq_arrays"`     // Delinquency performance arrays
}

// ConvertCPRToSMM converts CPR to SMM array for prepayment calculations
func (l *LoanInfo) ConvertCPRToSMM() {
	if l.PrepayCPR != 0.0 {
		log.Println("Converting CPR to SMM array for loan:", l.ID)
		// Correct SMM formula: SMM = 1 - (1 - CPR)^(1/12)
		smm := 1 - math.Pow(1-l.PrepayCPR, 1.0/12.0)

		// Create SMM array with same value for all periods
		l.SMMArr = make([]float64, l.Wam)
		for i := range l.SMMArr {
			l.SMMArr[i] = smm
		}
	} else if l.SMMArr == nil {
		// Initialize with zeros if no prepayment
		l.SMMArr = make([]float64, l.Wam)
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
func (l *LoanInfo) GetAmortizationTable() AmortizationTable {
	// 🟢 PRE-ALLOCATE: Avoid dynamic slice growth
	numPeriods := int(l.Wam)
	periods := make([]int, numPeriods)
	begBal := make([]float64, numPeriods)
	schedBal := make([]float64, numPeriods)
	endBal := make([]float64, numPeriods)
	prepayAmountArr := make([]float64, numPeriods)
	interest := make([]float64, numPeriods)
	principal := make([]float64, numPeriods)

	// 🟢 PRE-CALCULATE: Move expensive calculations outside loop
	monthlyRate := l.Wac / 12.0 / 100.0

	// 🟢 PRE-CALCULATE: SMM conversion once
	l.ConvertCPRToSMM()

	// 🟢 OPTIMIZED: Use simple payment calculation instead of PPmt
	monthlyPayment := calculateMonthlyPayment(l.Face, monthlyRate, float64(l.Wam))

	tmp_face := l.Face

	// 🟢 OPTIMIZED: Single loop with pre-allocated slices
	for j := 0; j < numPeriods; j++ {
		i := l.Wam - int64(j) // Remaining periods

		periods[j] = j + 1
		begBal[j] = roundToCent(tmp_face)

		// 🟢 FAST: Simple multiplication instead of expensive PPmt
		interestPayment := tmp_face * monthlyRate
		interest[j] = roundToCent(interestPayment)

		// Calculate principal using standard formula
		var principalPayment float64
		if i == 1 {
			// Final payment: all remaining balance
			principalPayment = tmp_face
		} else {
			principalPayment = monthlyPayment - interestPayment
		}
		principal[j] = roundToCent(principalPayment)

		currentSchedBal := tmp_face - principalPayment
		schedBal[j] = roundToCent(currentSchedBal)

		// Calculate prepayment
		prepayAmount := l.SMMArr[j] * currentSchedBal
		prepayAmountArr[j] = roundToCent(prepayAmount)

		// Update remaining balance
		tmp_face = currentSchedBal - prepayAmount
		if tmp_face < 0.0 {
			tmp_face = 0.0
		}

		endBal[j] = roundToCent(tmp_face)
	}

	amortTable := AmortizationTable{
		Period:          periods,
		BegBal:          begBal,
		SchedBal:        schedBal,
		PrepayAmountArr: prepayAmountArr,
		Interest:        interest,
		Principal:       principal,
		EndBal:          endBal,
		DelinqArrays:    DelinqArrays{},
	}

	return amortTable
}

// 🟢 FAST: Inline rounding function
func roundToCent(value float64) float64 {
	return math.Round(value*100) / 100
}

// 🟢 FAST: Standard monthly payment calculation
func calculateMonthlyPayment(principal, monthlyRate float64, numPayments float64) float64 {
	if monthlyRate == 0 {
		return principal / numPayments
	}

	factor := math.Pow(1+monthlyRate, numPayments)
	return principal * (monthlyRate * factor) / (factor - 1)
}

// TrueUpBalances adjusts the final period's balances to ensure mathematical consistency
func (a *AmortizationTable) TrueUpBalances() {
	if len(a.Principal) == 0 {
		return
	}

	lastIndex := len(a.Principal) - 1
	// Get the last period's values
	lastBegBal := a.BegBal[lastIndex]
	lastPrincipal := a.Principal[lastIndex]
	lastPrepay := a.PrepayAmountArr[lastIndex]
	lastEndBal := a.EndBal[lastIndex]

	leftOver := lastBegBal - lastPrincipal - lastPrepay

	if math.Abs(leftOver-lastEndBal) < 0.01 {
		return // Already balanced within rounding tolerance
	}

	// Adjust the final principal payment to balance
	if leftOver != lastEndBal {
		adjustment := leftOver - lastEndBal
		a.Principal[lastIndex] = lastPrincipal + adjustment
		a.EndBal[lastIndex] = 0.0 // Final balance should be zero
	}
}

// Add validation function
func (l *LoanInfo) Validate() error {
	if l.ID == "" {
		return fmt.Errorf("loan ID cannot be empty")
	}
	if l.Wam <= 0 || l.Wam > 480 { // Max 40 years
		return fmt.Errorf("WAM must be between 1 and 480 months, got %d", l.Wam)
	}
	if l.Wac < 0 || l.Wac > 30 { // Reasonable rate limits
		return fmt.Errorf("WAC must be between 0 and 30 percent, got %f", l.Wac)
	}
	if l.Face <= 0 {
		return fmt.Errorf("face value must be positive, got %f", l.Face)
	}
	if l.PrepayCPR < 0 || l.PrepayCPR >= 1 {
		return fmt.Errorf("CPR must be between 0 and 1, got %f", l.PrepayCPR)
	}
	return nil
}
