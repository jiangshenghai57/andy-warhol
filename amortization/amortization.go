// Package amortization provides mortgage loan amortization calculations
// and related financial computations.
package amortization

import (
	"fmt"
	"math"

	financial "github.com/razorpay/go-financial"
	"github.com/shopspring/decimal"
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
	ID         string            `json:"id"`                    // Unique identifier for the loan
	Wam        int64             `json:"wam"`                   // Weighted Average Maturity in months
	Wac        float64           `json:"wac"`                   // Weighted Average Coupon rate per annum in percentage points (e.g., 6.75)
	Face       float64           `json:"face"`                  // Mortgage notional/principal amount
	PrepayCPR  float64           `json:"prepay_cpr"`            // prepay CPR in decimals, could be SMM
	SMMArr     []float64         `json:"smm_arr,omitempty"`     // SMM array for prepayment calculations
	AmortTable AmortizationTable `json:"amort_table,omitempty"` // Associated amortization table
	StaticDQ   bool              `json:"static_dq"`             // If true amortization uses a roll rate matrix
	// Define the structure for the roll rate matrix
	// [0.92, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01]
	// should sum up to 1.0, and each element represents the transition probability
	// from one delinquency status to another.
	// For example, if the performing transformation is [0.92, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01],
	// it means there is a 92% chance that a performing loan will remain performing,
	// and a 1% chance it will transition to each of the delinquent statuses.
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
	if l.PrepayCPR != 0.0 && l.SMMArr == nil {
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

	// Convert CPR to SMM array if necessary
	l.ConvertCPRToSMM()

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

		// Calculate scheduled balance after principal payment
		currentSchedBal := math.Round((tmp_face-prinPmt)*100) / 100
		schedBal = append(schedBal, currentSchedBal)

		// Calculate prepayment using SMM array
		prepayAmount := math.Round(l.SMMArr[j-1]*currentSchedBal*100) / 100
		prepayAmountArr = append(prepayAmountArr, prepayAmount)

		// Reduce remaining balance by principal payment and prepayment
		tmp_face = tmp_face - float64(prinPmt) - prepayAmount

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
		DelinqArrays:    DelinqArrays{}, // Initialize with empty arrays
	}

	amortTable.TrueUpBalances()

	return amortTable
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
