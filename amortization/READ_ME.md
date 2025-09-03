# Amortization Package Documentation

## Overview

The `amortization` package provides comprehensive mortgage loan amortization calculations and related financial computations. It's designed to generate detailed payment schedules for various types of loans with support for CPR/SMM prepayments, delinquency tracking, and roll rate matrix transitions. All data structures are fully JSON-serializable for API integration.

## Table of Contents

- [Types](#types)
- [Interfaces](#interfaces)
- [Functions](#functions)
- [Usage Examples](#usage-examples)
- [Algorithm Details](#algorithm-details)
- [JSON Serialization](#json-serialization)

## Types

### LoanInfo

Represents basic loan information used for amortization calculations with full JSON support.

```go
type LoanInfo struct {
    ID         string            `json:"id"`           // Unique identifier for the loan
    Wam        int64             `json:"wam"`          // Weighted Average Maturity in months
    Wac        float64           `json:"wac"`          // Weighted Average Coupon rate per annum (e.g., 6.75)
    Face       float64           `json:"face"`         // Mortgage notional/principal amount
    PrepayCPR  float64           `json:"prepay_cpr"`   // Conditional Prepayment Rate in decimals
    SMMArr     []float64         `json:"smm_arr"`      // Single Monthly Mortality array for calculations
    AmortTable AmortizationTable `json:"amort_table"`  // Associated amortization table
    StaticDQ   bool              `json:"static_dq"`    // If true, uses roll rate matrix for amortization
}
```

**Fields:**
- `ID`: Unique loan identifier for tracking purposes
- `Wam`: Loan term in months (e.g., 360 for 30 years)
- `Wac`: Annual interest rate as percentage points (e.g., 4.5 for 4.5%)
- `Face`: Initial loan principal amount in dollars
- `PrepayCPR`: Conditional Prepayment Rate as decimal (e.g., 0.06 for 6% CPR)
- `SMMArr`: Single Monthly Mortality rates calculated from CPR for each period
- `AmortTable`: Computed amortization schedule
- `StaticDQ`: Flag for delinquency calculation method using roll rate matrices

### AmortizationTable

Contains the complete loan payment schedule with all components, fully JSON-serializable.

```go
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
```

**Fields:**
- `BegBal`: Starting balance at the beginning of each payment period
- `Interest`: Interest portion of each payment
- `Principal`: Principal portion of each payment  
- `SchedBal`: Scheduled remaining balance after regular payment
- `PrepayAmountArr`: SMM-based prepayment amounts for each period
- `EndBal`: Final balance after all payments and prepayments
- `Period`: Sequential period numbers starting from 1
- `DelinqArrays`: Delinquency tracking data structure

### DelinqArrays

Tracks loan performance across different delinquency buckets with JSON serialization support.

```go
type DelinqArrays struct {
    PerfArr    []float64 `json:"perf_arr"`    // Current/performing loans
    DQ30Arr    []float64 `json:"dq30_arr"`    // 30-day delinquent loans
    DQ60Arr    []float64 `json:"dq60_arr"`    // 60-day delinquent loans
    DQ90Arr    []float64 `json:"dq90_arr"`    // 90-day delinquent loans
    DQ120Arr   []float64 `json:"dq120_arr"`   // 120-day delinquent loans
    DQ150Arr   []float64 `json:"dq150_arr"`   // 150-day delinquent loans
    DQ180Arr   []float64 `json:"dq180_arr"`   // 180-day delinquent loans
    DefaultArr []float64 `json:"default_arr"` // Defaulted loans
}
```

### RollRateTransitionMatrix

Defines transition probabilities between different delinquency states for advanced modeling.

```go
type RollRateTransitionMatrix struct {
    PerformingTransition []float64 `json:"performing_transition"`
    DQ30Transition       []float64 `json:"dq30_transition"`
    DQ60Transition       []float64 `json:"dq60_transition"`
    DQ90Transition       []float64 `json:"dq90_transition"`
    DQ120Transition      []float64 `json:"dq120_transition"`
    DQ150Transition      []float64 `json:"dq150_transition"`
    DQ180Transition      []float64 `json:"dq180_transition"`
    DefaultTransition    []float64 `json:"default_transition"`
}
```

**Usage:**
- Each transition array should sum to 1.0 (100% probability)
- Array length should equal the number of delinquency states (8)
- Used when `LoanInfo.StaticDQ` is true for statistical modeling

## Interfaces

### MortgagePool

Defines behavior for types that can generate amortization schedules.

```go
type MortgagePool interface {
    GenerateAmortTable() AmortizationTable
    ConvertCPRToSMM() []float64
    TrueUpBalances()
}
```

**Methods:**
- `GenerateAmortTable()`: Creates a complete amortization schedule
- `ConvertCPRToSMM()`: Converts CPR to SMM array for monthly calculations
- `TrueUpBalances()`: Adjusts final period balances for mathematical accuracy

## Functions

### GetAmortizationTable

Calculates and returns a complete amortization table for a given loan with CPR/SMM prepayment support.

```go
func GetAmortizationTable(l *LoanInfo) AmortizationTable
```

**Parameters:**
- `l`: Pointer to `LoanInfo` containing loan parameters

**Returns:**
- `AmortizationTable`: Complete payment schedule with period-by-period breakdown

**Algorithm:**
1. Initializes arrays for all payment components
2. Converts CPR to SMM array using `ConvertCPRToSMM()`
3. For each period (WAM down to 1):
   - Calculates beginning balance
   - Computes interest payment using monthly rate
   - Calculates principal payment using financial library
   - Determines scheduled balance after payment
   - Applies SMM-based prepayments to scheduled balance
   - Updates remaining balance
4. Calls `TrueUpBalances()` to ensure mathematical consistency
5. Rounds all monetary values to 2 decimal places

### ConvertCPRToSMM

Converts Conditional Prepayment Rate to Single Monthly Mortality array.

```go
func ConvertCPRToSMM(l *LoanInfo)
```

**Parameters:**
- `l`: Pointer to `LoanInfo` with CPR to convert

**Formula:**
```
SMM = 1 - (1 - CPR)^(1/12)
```

**Behavior:**
- If `PrepayCPR` > 0, calculates SMM and fills `SMMArr` for all periods
- If `PrepayCPR` = 0, initializes `SMMArr` with zeros
- Only converts if `SMMArr` is nil (doesn't overwrite existing SMM data)

### TrueUpBalances (Method)

Adjusts the final period's balances to ensure mathematical consistency.

```go
func (a *AmortizationTable) TrueUpBalances()
```

**Purpose:**
- Corrects rounding discrepancies in final payment period
- Adjusts final principal payment if needed to balance equations
- Ensures final ending balance is zero for fully amortized loans
- Uses tolerance check (0.01) to avoid unnecessary adjustments

## Usage Examples

### Basic Loan Calculation with CPR

```go
// Create loan information with CPR
loanInfo := &LoanInfo{
    ID:        "LOAN001",
    Wam:       360,      // 30 years
    Wac:       4.5,      // 4.5% annual rate
    Face:      250000.0, // $250,000 loan
    PrepayCPR: 0.06,     // 6% CPR
}

// Calculate amortization table
table := GetAmortizationTable(loanInfo)

// Access payment information
fmt.Printf("First payment - Interest: $%.2f, Principal: $%.2f, Prepay: $%.2f\n", 
           table.Interest[0], table.Principal[0], table.PrepayAmountArr[0])
fmt.Printf("Final balance: $%.2f\n", table.EndBal[len(table.EndBal)-1])
```

### Working with Custom SMM Array

```go
// Loan with custom SMM rates (varying prepayment speeds)
loanInfo := &LoanInfo{
    ID:     "LOAN002", 
    Wam:    120,       // 10 years
    Wac:    6.0,       // 6% annual rate
    Face:   100000.0,  // $100,000 loan
    SMMArr: []float64{0.01, 0.015, 0.02}, // Custom SMM for first 3 periods
}

// Extend SMM array to full term if needed
if len(loanInfo.SMMArr) < int(loanInfo.Wam) {
    // Fill remaining periods with last SMM value
    lastSMM := loanInfo.SMMArr[len(loanInfo.SMMArr)-1]
    for i := len(loanInfo.SMMArr); i < int(loanInfo.Wam); i++ {
        loanInfo.SMMArr = append(loanInfo.SMMArr, lastSMM)
    }
}

// Generate amortization schedule
table := GetAmortizationTable(loanInfo)
```

### JSON Serialization Example

```go
// Create and calculate loan
loanInfo := &LoanInfo{
    ID:        "LOAN003",
    Wam:       240,
    Wac:       5.25,
    Face:      300000.0,
    PrepayCPR: 0.08,
}

table := GetAmortizationTable(loanInfo)

// Serialize to JSON
jsonData, err := json.Marshal(table)
if err != nil {
    log.Fatal(err)
}

// Save to file or send via API
err = ioutil.WriteFile("amortization.json", jsonData, 0644)
if err != nil {
    log.Fatal(err)
}

// Deserialize from JSON
var loadedTable AmortizationTable
err = json.Unmarshal(jsonData, &loadedTable)
if err != nil {
    log.Fatal(err)
}
```

### Processing Multiple Loans with API Integration

```go
loans := []*LoanInfo{
    {ID: "A001", Wam: 360, Wac: 4.25, Face: 400000, PrepayCPR: 0.05},
    {ID: "A002", Wam: 180, Wac: 5.75, Face: 200000, PrepayCPR: 0.08},
    {ID: "A003", Wam: 240, Wac: 4.95, Face: 350000, PrepayCPR: 0.06},
}

results := make(map[string]AmortizationTable)
for _, loan := range loans {
    table := GetAmortizationTable(loan)
    results[loan.ID] = table
    
    fmt.Printf("Loan %s: Monthly P&I ~$%.2f, Prepay ~$%.2f\n", 
               loan.ID, 
               table.Interest[0] + table.Principal[0],
               table.PrepayAmountArr[0])
}

// Serialize all results
allResults, _ := json.Marshal(results)
fmt.Printf("JSON size: %d bytes\n", len(allResults))
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

### CPR to SMM Conversion
```
SMM = 1 - (1 - CPR)^(1/12)
```

### Prepayment Application
```
Prepayment Amount = SMM × Scheduled Balance
```

### Balance Updates
```
New Balance = Beginning Balance - Principal Payment - Prepayment Amount
```

### Rounding
All monetary values are rounded to 2 decimal places using:
```go
math.Round(value * 100) / 100
```

## JSON Serialization

All data structures in the package are fully JSON-serializable:

- **Field Names**: All struct fields use `json` tags with snake_case naming
- **Exported Fields**: All fields are capitalized for external access
- **Nested Structures**: Complex types like `DelinqArrays` are properly nested
- **Array Handling**: All float64 and int arrays serialize correctly
- **API Ready**: Perfect for REST API responses and database storage

### JSON Structure Example

```json
{
  "beg_bal": [100000.0, 99500.23, 99000.11],
  "interest": [375.0, 372.81, 370.50],
  "principal": [124.77, 126.96, 129.19],
  "sched_bal": [99875.23, 99373.27, 98870.92],
  "prepay_amount_arr": [499.38, 496.87, 494.35],
  "end_bal": [99375.85, 98876.40, 98376.57],
  "period": [1, 2, 3],
  "delinq_arrays": {
    "perf_arr": [],
    "dq30_arr": [],
    "dq60_arr": [],
    "dq90_arr": [],
    "dq120_arr": [],
    "dq150_arr": [],
    "dq180_arr": [],
    "default_arr": []
  }
}
```

## Dependencies

- `github.com/razorpay/go-financial`: Financial calculations (PMT, PPmt functions)
- `github.com/shopspring/decimal`: Precise decimal arithmetic
- Standard library `math`: Mathematical operations and rounding
- Standard library `encoding/json`: JSON serialization support

## Notes

- All interest rates should be provided as percentage points (e.g., 4.5 for 4.5%)
- Loan terms (WAM) are specified in months
- CPR rates are expressed as decimals (e.g., 0.06 for 6% CPR)
- SMM is automatically calculated from CPR using standard mortgage industry formula
- The package handles edge cases like negative balances and ensures mathematical consistency
- All monetary calculations are rounded to 2 decimal places for precision
- Delinquency arrays are initialized but require separate roll rate matrix implementation
- JSON serialization is fully supported for API integration and data storage