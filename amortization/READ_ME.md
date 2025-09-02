# Amortization Package Documentation

## Overview

The `amortization` package provides comprehensive mortgage loan amortization calculations and related financial computations. It's designed to generate detailed payment schedules for various types of loans with support for prepayments and delinquency tracking.

## Table of Contents

- [Types](#types)
- [Interfaces](#interfaces)
- [Functions](#functions)
- [Usage Examples](#usage-examples)
- [Algorithm Details](#algorithm-details)

## Types

### LoanInfo

Represents basic loan information used for amortization calculations.

```go
type LoanInfo struct {
    ID         string            // Unique identifier for the loan
    Wam        int64             // Weighted Average Maturity in months
    Wac        float64           // Weighted Average Coupon rate per annum (e.g., 6.75)
    Face       float64           // Mortgage notional/principal amount
    Prepay     []float64         // Prepayment amounts by period
    AmortTable AmortizationTable // Associated amortization table
    StaticDQ   bool              // If true, uses roll rate matrix for amortization
}
```

**Fields:**
- `ID`: Unique loan identifier for tracking purposes
- `Wam`: Loan term in months (e.g., 360 for 30 years)
- `Wac`: Annual interest rate as percentage points (e.g., 4.5 for 4.5%)
- `Face`: Initial loan principal amount in dollars
- `Prepay`: Array of prepayment percentages for each period
- `AmortTable`: Computed amortization schedule
- `StaticDQ`: Flag for delinquency calculation method

### AmortizationTable

Contains the complete loan payment schedule with all components.

```go
type AmortizationTable struct {
    BegBal       []float64    // Beginning balance for each period
    Interest     []float64    // Interest payment for each period
    Principal    []float64    // Principal payment for each period
    SchedBal     []float64    // Scheduled balance after payment
    PrepayArr    []float64    // Prepayment amount for each period
    EndBal       []float64    // Ending balance for each period
    Period       []int        // Period numbers (1, 2, 3, ...)
    DelinqArrays DelinqArrays // Delinquency performance arrays
}
```

**Fields:**
- `BegBal`: Starting balance at the beginning of each payment period
- `Interest`: Interest portion of each payment
- `Principal`: Principal portion of each payment  
- `SchedBal`: Scheduled remaining balance after regular payment
- `PrepayArr`: Additional prepayment amounts for each period
- `EndBal`: Final balance after all payments and prepayments
- `Period`: Sequential period numbers starting from 1
- `DelinqArrays`: Delinquency tracking data structure

### DelinqArrays

Tracks loan performance across different delinquency buckets.

```go
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
```

## Interfaces

### MortgagePool

Defines behavior for types that can generate amortization schedules.

```go
type MortgagePool interface {
    GenerateAmortTable() AmortizationTable
    ExtendPrepayArr() []float64
    TrueUpBalances()
}
```

**Methods:**
- `GenerateAmortTable()`: Creates a complete amortization schedule
- `ExtendPrepayArr()`: Extends prepayment array to match loan term
- `TrueUpBalances()`: Adjusts final period balances for accuracy

## Functions

### GetAmortizationTable

Calculates and returns a complete amortization table for a given loan.

```go
func GetAmortizationTable(l *LoanInfo) AmortizationTable
```

**Parameters:**
- `l`: Pointer to `LoanInfo` containing loan parameters

**Returns:**
- `AmortizationTable`: Complete payment schedule with period-by-period breakdown

**Algorithm:**
1. Initializes arrays for all payment components
2. Extends prepayment array to match loan term if needed
3. For each period (WAM down to 1):
   - Calculates beginning balance
   - Computes interest payment using monthly rate
   - Calculates principal payment using financial library
   - Determines scheduled balance after payment
   - Applies prepayments based on percentage of scheduled balance
   - Updates remaining balance
4. Rounds all monetary values to 2 decimal places

### ExtendPrepayArr

Extends a single-element prepayment array to match the loan term.

```go
func ExtendPrepayArr(l *LoanInfo)
```

**Parameters:**
- `l`: Pointer to `LoanInfo` with prepayment array to extend

**Behavior:**
- If `Prepay` has only one element, fills remaining periods with zeros
- Ensures prepayment array length matches loan term (WAM)

### TrueUpBalances (Method)

Adjusts the final period's balances to ensure mathematical consistency.

```go
func (a *AmortizationTable) TrueUpBalances()
```

**Purpose:**
- Corrects rounding discrepancies in final payment period
- Ensures ending balance equals principal + prepayment for last period
- Prevents small balance remainders due to floating-point arithmetic

## Usage Examples

### Basic Loan Calculation

```go
// Create loan information
loanInfo := &LoanInfo{
    ID:   "LOAN001",
    Wam:  360,        // 30 years
    Wac:  4.5,        // 4.5% annual rate
    Face: 250000.0,   // $250,000 loan
    Prepay: []float64{0.05}, // 5% prepayment rate
}

// Calculate amortization table
table := GetAmortizationTable(loanInfo)

// Access payment information
fmt.Printf("First payment - Interest: $%.2f, Principal: $%.2f\n", 
           table.Interest[0], table.Principal[0])
fmt.Printf("Final balance: $%.2f\n", table.EndBal[len(table.EndBal)-1])
```

### Working with Prepayments

```go
// Loan with varying prepayment rates
loanInfo := &LoanInfo{
    ID:   "LOAN002", 
    Wam:  120,       // 10 years
    Wac:  6.0,       // 6% annual rate
    Face: 100000.0,  // $100,000 loan
    Prepay: []float64{0.0, 0.1, 0.05}, // 0%, 10%, 5% for first 3 periods
}

// Extend prepayment array to full term if a single lenght prepay is passed in
ExtendPrepayArr(loanInfo)

// Generate amortization schedule
table := GetAmortizationTable(loanInfo)

// True up final balances
table.TrueUpBalances()
```

### Processing Multiple Loans

```go
loans := []*LoanInfo{
    {ID: "A001", Wam: 360, Wac: 4.25, Face: 400000, Prepay: []float64{0.0}},
    {ID: "A002", Wam: 180, Wac: 5.75, Face: 200000, Prepay: []float64{0.1}},
}

for _, loan := range loans {
    table := GetAmortizationTable(loan)
    fmt.Printf("Loan %s: Monthly payment ~$%.2f\n", 
               loan.ID, table.Interest[0] + table.Principal[0])
}
```

## Algorithm Details

### Interest Calculation
```
Monthly Interest = Beginning Balance × (Annual Rate ÷ 12) ÷ 100
```

### Principal Payment
Uses the financial library's `PPmt` function with:
- Monthly interest rate as decimal
- Remaining periods 
- Total loan term
- Present value (negative loan amount)
- Future value (0 for full payoff)

### Prepayment Application
```
Prepayment Amount = Prepayment Rate × Scheduled Balance
```

### Balance Updates
```
New Balance = Beginning Balance - Principal Payment - Prepayment
```

### Rounding
All monetary values are rounded to 2 decimal places using:
```go
math.Round(value * 100) / 100
```

## Dependencies

- `github.com/razorpay/go-financial`: Financial calculations (PMT, PPmt functions)
- `github.com/shopspring/decimal`: Precise decimal arithmetic
- Standard library `math`: Mathematical operations and rounding

## Notes

- All interest rates should be provided as percentage points (e.g., 4.5 for 4.5%)
- Loan terms (WAM) are specified in months
- Prepayment rates are expressed as decimals (e.g., 0.05 for 5%)
- The package handles edge cases like negative balances and ensures mathematical consistency
- Delinquency arrays are currently defined but not fully implemented in calculations