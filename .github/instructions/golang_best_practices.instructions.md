# Go Best Practices Guide

A comprehensive guide to writing clean, efficient, and maintainable Go code for the Andy Warhol mortgage cashflow engine.

## Table of Contents

- [Code Organization](#code-organization)
- [Naming Conventions](#naming-conventions)
- [Error Handling](#error-handling)
- [Concurrency](#concurrency)
- [Performance](#performance)
- [Testing](#testing)
- [Logging](#logging)
- [API Design](#api-design)
- [Security](#security)
- [Documentation](#documentation)

---

## Code Organization

### Project Structure

Follow the standard Go project layout:

```
andy-warhol/
├── cmd/                    # Main applications
│   └── server/
│       └── main.go
├── internal/               # Private application code
│   ├── amortization/
│   │   └── amortization.go
│   └── config/
│       └── config.go
├── pkg/                    # Public library code
│   └── financial/
│       └── calculations.go
├── api/                    # API definitions (OpenAPI/Swagger)
├── configs/                # Configuration files
├── scripts/                # Build and CI scripts
├── test/                   # Additional test data
├── docs/                   # Documentation
├── go.mod
├── go.sum
└── README.md
```

### Package Design

```go
// ✅ Good: Small, focused packages
package amortization

// LoanInfo handles loan-related operations
type LoanInfo struct {
    ID   string
    Face float64
    Wac  float64
    Wam  int64
}

// ❌ Bad: Large, unfocused packages with unrelated types
package utils // Avoid generic "utils" packages

type LoanInfo struct{}
type User struct{}
type Logger struct{}
```

### Import Organization

```go
// ✅ Good: Grouped and sorted imports
import (
    // Standard library
    "context"
    "fmt"
    "log"
    "sync"

    // Third-party packages
    "github.com/gin-gonic/gin"
    "github.com/shopspring/decimal"

    // Internal packages
    "github.com/jiangshenghai57/andy-warhol/internal/amortization"
    "github.com/jiangshenghai57/andy-warhol/internal/config"
)
```

---

## Naming Conventions

### Variables and Functions

```go
// ✅ Good: Clear, descriptive names
func calculateMonthlyPayment(principal, monthlyRate float64, numPayments int) float64 {
    factor := math.Pow(1+monthlyRate, float64(numPayments))
    return principal * (monthlyRate * factor) / (factor - 1)
}

// ❌ Bad: Cryptic abbreviations
func calcPmt(p, r float64, n int) float64 {
    f := math.Pow(1+r, float64(n))
    return p * (r * f) / (f - 1)
}
```

### Acronyms

```go
// ✅ Good: Consistent acronym casing
type LoanID string
var httpClient *http.Client
func parseJSON(data []byte) error

// ❌ Bad: Inconsistent casing
type LoanId string
var HTTPclient *http.Client
func parseJson(data []byte) error
```

### Interfaces

```go
// ✅ Good: Use -er suffix for single-method interfaces
type Validator interface {
    Validate() error
}

type CashflowCalculator interface {
    CalculateCashflow() (AmortizationTable, error)
}

// ✅ Good: Accept interfaces, return structs
func ProcessLoan(calc CashflowCalculator) (*AmortizationTable, error) {
    return calc.CalculateCashflow()
}
```

### Constants

```go
// ✅ Good: Use camelCase for unexported, PascalCase for exported
const (
    maxLoanTerm      = 480  // unexported
    MaxConcurrentJobs = 100 // exported
    DefaultWAC       = 4.5
)

// ✅ Good: Use iota for enumerations
type DelinquencyState int

const (
    Performing DelinquencyState = iota
    DQ30
    DQ60
    DQ90
    DQ120
    DQ150
    DQ180
    Default
)
```

---

## Error Handling

### Return Errors, Don't Panic

```go
// ✅ Good: Return errors for expected failure cases
func (l *LoanInfo) Validate() error {
    if l.ID == "" {
        return fmt.Errorf("loan ID cannot be empty")
    }
    if l.Wam <= 0 || l.Wam > 480 {
        return fmt.Errorf("WAM must be between 1 and 480, got %d", l.Wam)
    }
    if l.Face <= 0 {
        return fmt.Errorf("face value must be positive, got %.2f", l.Face)
    }
    return nil
}

// ❌ Bad: Panicking on validation errors
func (l *LoanInfo) Validate() {
    if l.ID == "" {
        panic("loan ID cannot be empty") // Don't do this
    }
}
```

### Wrap Errors with Context

```go
// ✅ Good: Add context when wrapping errors
func (l *LoanInfo) GetAmortizationTable() (AmortizationTable, error) {
    if err := l.Validate(); err != nil {
        return AmortizationTable{}, fmt.Errorf("loan %s validation failed: %w", l.ID, err)
    }
    
    table, err := calculateAmortization(l)
    if err != nil {
        return AmortizationTable{}, fmt.Errorf("calculating amortization for loan %s: %w", l.ID, err)
    }
    
    return table, nil
}

// Check wrapped errors
if errors.Is(err, ErrInvalidLoan) {
    // Handle invalid loan error
}
```

### Custom Error Types

```go
// ✅ Good: Define custom error types for specific cases
type ValidationError struct {
    Field   string
    Message string
    Value   interface{}
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error on %s: %s (got %v)", e.Field, e.Message, e.Value)
}

func (l *LoanInfo) Validate() error {
    if l.Wam <= 0 {
        return &ValidationError{
            Field:   "wam",
            Message: "must be positive",
            Value:   l.Wam,
        }
    }
    return nil
}

// Type assertion for specific handling
var validErr *ValidationError
if errors.As(err, &validErr) {
    log.Printf("Invalid field: %s", validErr.Field)
}
```

### Handle Errors Once

```go
// ✅ Good: Handle error once at the appropriate level
func requestCashflow(c *gin.Context) {
    loans, err := parseLoans(c)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return // Handle once and return
    }
    
    results, err := processLoans(loans)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, results)
}

// ❌ Bad: Logging and returning the same error
func processLoan(loan LoanInfo) (AmortizationTable, error) {
    table, err := loan.GetAmortizationTable()
    if err != nil {
        log.Printf("Error: %v", err) // Logged here
        return AmortizationTable{}, err // And returned - now it might be logged again
    }
    return table, nil
}
```

---

## Concurrency

### Use sync.WaitGroup for Goroutine Coordination

```go
// ✅ Good: Properly coordinate goroutines
func processLoans(loans []LoanInfo) []AmortizationTable {
    var wg sync.WaitGroup
    results := make([]AmortizationTable, len(loans))
    
    for i, loan := range loans {
        wg.Add(1)
        go func(index int, l LoanInfo) {
            defer wg.Done()
            results[index] = l.GetAmortizationTable()
        }(i, loan) // Pass by value to avoid closure issues
    }
    
    wg.Wait()
    return results
}

// ❌ Bad: Race condition with loop variable
for i, loan := range loans {
    wg.Add(1)
    go func() {
        defer wg.Done()
        results[i] = loan.GetAmortizationTable() // Race condition!
    }()
}
```

### Worker Pool Pattern

```go
// ✅ Good: Limit concurrent goroutines with worker pool
type WorkerPool struct {
    workers chan struct{}
}

func NewWorkerPool(size int) *WorkerPool {
    return &WorkerPool{
        workers: make(chan struct{}, size),
    }
}

func (p *WorkerPool) Acquire() {
    p.workers <- struct{}{}
}

func (p *WorkerPool) Release() {
    <-p.workers
}

// Usage
var pool = NewWorkerPool(100)

func processLoans(loans []LoanInfo) []AmortizationTable {
    var wg sync.WaitGroup
    results := make([]AmortizationTable, len(loans))
    
    for i, loan := range loans {
        wg.Add(1)
        go func(index int, l LoanInfo) {
            pool.Acquire()
            defer func() {
                pool.Release()
                wg.Done()
            }()
            
            results[index] = l.GetAmortizationTable()
        }(i, loan)
    }
    
    wg.Wait()
    return results
}
```

### Use Context for Cancellation

```go
// ✅ Good: Support context cancellation
func processLoansWithContext(ctx context.Context, loans []LoanInfo) ([]AmortizationTable, error) {
    results := make([]AmortizationTable, len(loans))
    errChan := make(chan error, 1)
    
    var wg sync.WaitGroup
    
    for i, loan := range loans {
        wg.Add(1)
        go func(index int, l LoanInfo) {
            defer wg.Done()
            
            select {
            case <-ctx.Done():
                return // Context cancelled
            default:
                results[index] = l.GetAmortizationTable()
            }
        }(i, loan)
    }
    
    // Wait in goroutine to allow early return on context cancellation
    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(done)
    }()
    
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    case <-done:
        return results, nil
    }
}
```

### Protect Shared State with Mutexes

```go
// ✅ Good: Use sync.RWMutex for read-heavy workloads
type LoanStore struct {
    mu    sync.RWMutex
    loans []LoanInfo
}

func (s *LoanStore) Add(loan LoanInfo) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.loans = append(s.loans, loan)
}

func (s *LoanStore) GetAll() []LoanInfo {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    // Return a copy to prevent external modification
    result := make([]LoanInfo, len(s.loans))
    copy(result, s.loans)
    return result
}

func (s *LoanStore) Count() int {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return len(s.loans)
}
```

### Channel Best Practices

```go
// ✅ Good: Clear channel ownership and closing
func generateLoans(count int) <-chan LoanInfo {
    out := make(chan LoanInfo)
    
    go func() {
        defer close(out) // Producer closes the channel
        for i := 0; i < count; i++ {
            out <- LoanInfo{
                ID:   fmt.Sprintf("LOAN%04d", i),
                Face: 100000.0 + float64(i*1000),
            }
        }
    }()
    
    return out
}

// Consumer reads until channel is closed
func processLoanChannel(loans <-chan LoanInfo) []AmortizationTable {
    var results []AmortizationTable
    for loan := range loans {
        results = append(results, loan.GetAmortizationTable())
    }
    return results
}
```

---

## Performance

### Pre-allocate Slices

```go
// ✅ Good: Pre-allocate with known capacity
func (l *LoanInfo) GetAmortizationTable() AmortizationTable {
    numPeriods := int(l.Wam)
    
    periods := make([]int, numPeriods)
    begBal := make([]float64, numPeriods)
    interest := make([]float64, numPeriods)
    principal := make([]float64, numPeriods)
    
    for j := 0; j < numPeriods; j++ {
        periods[j] = j + 1
        begBal[j] = calculateBegBal(j)
        interest[j] = calculateInterest(j)
        principal[j] = calculatePrincipal(j)
    }
    
    return AmortizationTable{
        Period:    periods,
        BegBal:    begBal,
        Interest:  interest,
        Principal: principal,
    }
}

// ❌ Bad: Dynamic growth with append
func (l *LoanInfo) GetAmortizationTableSlow() AmortizationTable {
    var periods []int
    var begBal []float64
    
    for j := 0; j < int(l.Wam); j++ {
        periods = append(periods, j+1)      // Multiple reallocations
        begBal = append(begBal, calculateBegBal(j))
    }
    
    return AmortizationTable{Period: periods, BegBal: begBal}
}
```

### Avoid Expensive Operations in Loops

```go
// ✅ Good: Pre-calculate outside the loop
func (l *LoanInfo) GetAmortizationTable() AmortizationTable {
    numPeriods := int(l.Wam)
    
    // Pre-calculate constants
    monthlyRate := l.Wac / 12.0 / 100.0
    monthlyPayment := calculateMonthlyPayment(l.Face, monthlyRate, float64(numPeriods))
    
    results := make([]float64, numPeriods)
    balance := l.Face
    
    for j := 0; j < numPeriods; j++ {
        interest := balance * monthlyRate
        principal := monthlyPayment - interest
        balance -= principal
        results[j] = balance
    }
    
    return AmortizationTable{EndBal: results}
}

// ❌ Bad: Recalculating constants in loop
func (l *LoanInfo) GetAmortizationTableSlow() AmortizationTable {
    numPeriods := int(l.Wam)
    results := make([]float64, numPeriods)
    balance := l.Face
    
    for j := 0; j < numPeriods; j++ {
        monthlyRate := l.Wac / 12.0 / 100.0 // Recalculated every iteration
        monthlyPayment := calculateMonthlyPayment(l.Face, monthlyRate, float64(numPeriods))
        
        interest := balance * monthlyRate
        principal := monthlyPayment - interest
        balance -= principal
        results[j] = balance
    }
    
    return AmortizationTable{EndBal: results}
}
```

### Use sync.Pool for Frequent Allocations

```go
// ✅ Good: Reuse temporary slices with sync.Pool
var floatSlicePool = sync.Pool{
    New: func() interface{} {
        slice := make([]float64, 0, 480) // Max loan term
        return &slice
    },
}

func (l *LoanInfo) GetAmortizationTable() AmortizationTable {
    // Get slice from pool
    tempSlice := floatSlicePool.Get().(*[]float64)
    defer func() {
        *tempSlice = (*tempSlice)[:0] // Reset slice
        floatSlicePool.Put(tempSlice)
    }()
    
    // Use the slice for temporary calculations
    for j := 0; j < int(l.Wam); j++ {
        *tempSlice = append(*tempSlice, calculateValue(j))
    }
    
    // Copy to result
    result := make([]float64, len(*tempSlice))
    copy(result, *tempSlice)
    
    return AmortizationTable{EndBal: result}
}
```

### Avoid String Concatenation in Loops

```go
// ✅ Good: Use strings.Builder
func formatLoanReport(loans []LoanInfo) string {
    var sb strings.Builder
    sb.Grow(len(loans) * 100) // Pre-allocate estimated size
    
    for _, loan := range loans {
        fmt.Fprintf(&sb, "Loan %s: $%.2f at %.2f%%\n", 
            loan.ID, loan.Face, loan.Wac)
    }
    
    return sb.String()
}

// ❌ Bad: String concatenation creates many allocations
func formatLoanReportSlow(loans []LoanInfo) string {
    result := ""
    for _, loan := range loans {
        result += fmt.Sprintf("Loan %s: $%.2f at %.2f%%\n", 
            loan.ID, loan.Face, loan.Wac)
    }
    return result
}
```

### Benchmark Your Code

```go
// amortization_test.go
func BenchmarkGetAmortizationTable(b *testing.B) {
    loan := LoanInfo{
        ID:        "BENCH001",
        Wam:       360,
        Wac:       4.5,
        Face:      250000,
        PrepayCPR: 0.06,
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = loan.GetAmortizationTable()
    }
}

func BenchmarkProcessLoans(b *testing.B) {
    loans := generateTestLoans(1000)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = processLoans(loans)
    }
}

// Run with: go test -bench=. -benchmem
```

---

## Testing

### Table-Driven Tests

```go
func TestLoanValidation(t *testing.T) {
    tests := []struct {
        name    string
        loan    LoanInfo
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid loan",
            loan:    LoanInfo{ID: "L001", Wam: 360, Wac: 4.5, Face: 250000},
            wantErr: false,
        },
        {
            name:    "empty ID",
            loan:    LoanInfo{ID: "", Wam: 360, Wac: 4.5, Face: 250000},
            wantErr: true,
            errMsg:  "loan ID cannot be empty",
        },
        {
            name:    "invalid WAM",
            loan:    LoanInfo{ID: "L001", Wam: 0, Wac: 4.5, Face: 250000},
            wantErr: true,
            errMsg:  "WAM must be between 1 and 480",
        },
        {
            name:    "negative face value",
            loan:    LoanInfo{ID: "L001", Wam: 360, Wac: 4.5, Face: -100},
            wantErr: true,
            errMsg:  "face value must be positive",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.loan.Validate()
            
            if tt.wantErr {
                if err == nil {
                    t.Errorf("expected error containing %q, got nil", tt.errMsg)
                } else if !strings.Contains(err.Error(), tt.errMsg) {
                    t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
                }
            } else if err != nil {
                t.Errorf("unexpected error: %v", err)
            }
        })
    }
}
```

### Test Fixtures

```go
// testdata/loan_fixtures.go
func newTestLoan() LoanInfo {
    return LoanInfo{
        ID:        "TEST001",
        Wam:       360,
        Wac:       4.5,
        Face:      250000,
        PrepayCPR: 0.06,
    }
}

func newTestLoanWithOptions(opts ...func(*LoanInfo)) LoanInfo {
    loan := newTestLoan()
    for _, opt := range opts {
        opt(&loan)
    }
    return loan
}

func withWAM(wam int64) func(*LoanInfo) {
    return func(l *LoanInfo) {
        l.Wam = wam
    }
}

func withFace(face float64) func(*LoanInfo) {
    return func(l *LoanInfo) {
        l.Face = face
    }
}

// Usage
func TestAmortization(t *testing.T) {
    loan := newTestLoanWithOptions(
        withWAM(120),
        withFace(100000),
    )
    
    table := loan.GetAmortizationTable()
    
    if len(table.Period) != 120 {
        t.Errorf("expected 120 periods, got %d", len(table.Period))
    }
}
```

### HTTP Handler Tests

```go
func TestRequestCashflow(t *testing.T) {
    // Setup
    router := gin.New()
    router.POST("/loans", requestCashflow)
    
    tests := []struct {
        name       string
        body       string
        wantStatus int
        wantCount  int
    }{
        {
            name: "single valid loan",
            body: `[{
                "id": "L001",
                "wam": 360,
                "wac": 4.5,
                "face": 250000,
                "prepay_cpr": 0.06
            }]`,
            wantStatus: http.StatusOK,
            wantCount:  1,
        },
        {
            name:       "invalid JSON",
            body:       `{invalid}`,
            wantStatus: http.StatusBadRequest,
        },
        {
            name: "invalid loan data",
            body: `[{"id": "", "wam": 0}]`,
            wantStatus: http.StatusBadRequest,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("POST", "/loans", strings.NewReader(tt.body))
            req.Header.Set("Content-Type", "application/json")
            
            w := httptest.NewRecorder()
            router.ServeHTTP(w, req)
            
            if w.Code != tt.wantStatus {
                t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
            }
            
            if tt.wantStatus == http.StatusOK && tt.wantCount > 0 {
                var response map[string]interface{}
                json.Unmarshal(w.Body.Bytes(), &response)
                
                if int(response["count"].(float64)) != tt.wantCount {
                    t.Errorf("expected count %d, got %v", tt.wantCount, response["count"])
                }
            }
        })
    }
}
```

### Test Coverage

```bash
# Run tests with coverage
go test -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out -o coverage.html

# Check coverage percentage
go tool cover -func=coverage.out
```

---

## Logging

### Structured Logging

```go
// ✅ Good: Use structured logging
import (
    "log/slog"
    "os"
)

var logger *slog.Logger

func init() {
    logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
}

func processLoan(loan LoanInfo) {
    logger.Info("processing loan",
        slog.String("loan_id", loan.ID),
        slog.Float64("face", loan.Face),
        slog.Float64("wac", loan.Wac),
        slog.Int64("wam", loan.Wam),
    )
    
    startTime := time.Now()
    table := loan.GetAmortizationTable()
    duration := time.Since(startTime)
    
    logger.Info("loan processed",
        slog.String("loan_id", loan.ID),
        slog.Duration("processing_time", duration),
        slog.Int("periods", len(table.Period)),
    )
}
```

### Log Levels

```go
// ✅ Good: Use appropriate log levels
logger.Debug("detailed debugging info", slog.Any("data", debugData))
logger.Info("normal operations", slog.String("action", "loan_processed"))
logger.Warn("potential issues", slog.String("reason", "high_dq_rate"))
logger.Error("errors that need attention", slog.Any("error", err))
```

### Dual Output (File and Stdout)

```go
func setupLogging() *slog.Logger {
    // Create log directory
    logDir := "logs"
    os.MkdirAll(logDir, 0755)
    
    // Open log file
    logFileName := filepath.Join(logDir, fmt.Sprintf("andy-warhol-%s.log",
        time.Now().Format("2006-01-02")))
    
    logFile, err := os.OpenFile(logFileName,
        os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        log.Fatalf("Failed to open log file: %v", err)
    }
    
    // Multi-writer for both file and stdout
    multiWriter := io.MultiWriter(logFile, os.Stdout)
    
    // Create structured logger
    handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
        Level: slog.LevelInfo,
        AddSource: true,
    })
    
    return slog.New(handler)
}
```

---

## API Design

### RESTful Conventions

```go
// ✅ Good: Clear, RESTful endpoints
router.GET("/loans", getLoans)           // List all loans
router.GET("/loans/:id", getLoan)        // Get single loan
router.POST("/loans", createLoans)       // Create loans
router.PUT("/loans/:id", updateLoan)     // Update loan
router.DELETE("/loans/:id", deleteLoan)  // Delete loan
router.GET("/health", healthCheck)       // Health check
```

### Request Validation

```go
// ✅ Good: Validate requests early
func createLoans(c *gin.Context) {
    var loans []LoanInfo
    
    // Bind and validate JSON
    if err := c.ShouldBindJSON(&loans); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Invalid JSON format",
            "details": err.Error(),
        })
        return
    }
    
    // Validate each loan
    for i, loan := range loans {
        if err := loan.Validate(); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error":   fmt.Sprintf("Validation failed for loan %d", i),
                "loan_id": loan.ID,
                "details": err.Error(),
            })
            return
        }
    }
    
    // Process valid loans
    results := processLoans(loans)
    
    c.JSON(http.StatusOK, gin.H{
        "count":   len(results),
        "results": results,
    })
}
```

### Consistent Response Format

```go
// ✅ Good: Consistent response structure
type APIResponse struct {
    Success   bool        `json:"success"`
    Data      interface{} `json:"data,omitempty"`
    Error     string      `json:"error,omitempty"`
    Timestamp time.Time   `json:"timestamp"`
}

func respondSuccess(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, APIResponse{
        Success:   true,
        Data:      data,
        Timestamp: time.Now(),
    })
}

func respondError(c *gin.Context, status int, message string) {
    c.JSON(status, APIResponse{
        Success:   false,
        Error:     message,
        Timestamp: time.Now(),
    })
}
```

### Middleware

```go
// Request ID middleware
func RequestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)
        c.Next()
    }
}

// Logging middleware
func LoggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        
        c.Next()
        
        logger.Info("request completed",
            slog.String("method", c.Request.Method),
            slog.String("path", path),
            slog.Int("status", c.Writer.Status()),
            slog.Duration("latency", time.Since(start)),
            slog.String("request_id", c.GetString("request_id")),
        )
    }
}
```

---

## Security

### Input Validation

```go
// ✅ Good: Validate and sanitize all inputs
func (l *LoanInfo) Validate() error {
    // Check for empty strings
    l.ID = strings.TrimSpace(l.ID)
    if l.ID == "" {
        return errors.New("loan ID cannot be empty")
    }
    
    // Check for valid ID format
    if !isValidLoanID(l.ID) {
        return fmt.Errorf("invalid loan ID format: %s", l.ID)
    }
    
    // Check numeric bounds
    if l.Wam <= 0 || l.Wam > 480 {
        return fmt.Errorf("WAM must be between 1 and 480, got %d", l.Wam)
    }
    
    if l.Wac < 0 || l.Wac > 30 {
        return fmt.Errorf("WAC must be between 0 and 30, got %.2f", l.Wac)
    }
    
    if l.Face <= 0 || l.Face > 100000000 {
        return fmt.Errorf("face value out of acceptable range: %.2f", l.Face)
    }
    
    if l.PrepayCPR < 0 || l.PrepayCPR >= 1 {
        return fmt.Errorf("CPR must be between 0 and 1, got %.4f", l.PrepayCPR)
    }
    
    return nil
}

func isValidLoanID(id string) bool {
    // Only allow alphanumeric characters, dashes, and underscores
    match, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, id)
    return match && len(id) <= 50
}
```

### Rate Limiting

```go
import "golang.org/x/time/rate"

// Rate limiter middleware
func RateLimitMiddleware(requestsPerSecond int) gin.HandlerFunc {
    limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), requestsPerSecond*2)
    
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "rate limit exceeded",
            })
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### Graceful Shutdown

```go
func main() {
    router := gin.Default()
    setupRoutes(router)
    
    srv := &http.Server{
        Addr:    ":8080",
        Handler: router,
    }
    
    // Start server in goroutine
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server failed: %v", err)
        }
    }()
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down server...")
    
    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }
    
    log.Println("Server exited")
}
```

---

## Documentation

### Package Documentation

```go
// Package amortization provides mortgage loan amortization calculations.
//
// This package implements standard mortgage amortization schedules with
// support for prepayment modeling using CPR (Conditional Prepayment Rate)
// and delinquency transition analysis using Markov chain models.
//
// Example usage:
//
//     loan := amortization.LoanInfo{
//         ID:        "LOAN001",
//         Wam:       360,
//         Wac:       4.5,
//         Face:      250000,
//         PrepayCPR: 0.06,
//     }
//
//     table := loan.GetAmortizationTable()
//     fmt.Printf("Total interest: $%.2f\n", sum(table.Interest))
package amortization
```

### Function Documentation

```go
// GetAmortizationTable calculates and returns the complete amortization
// schedule for a loan.
//
// The function performs the following steps:
//   1. Validates loan parameters
//   2. Converts CPR to SMM for monthly prepayment calculations
//   3. Calculates monthly payment using standard amortization formula
//   4. Generates period-by-period breakdown of principal, interest, and balances
//   5. Applies delinquency transition matrices if specified
//
// Returns an AmortizationTable containing all cashflow arrays.
//
// Example:
//
//     loan := LoanInfo{ID: "L001", Wam: 360, Wac: 4.5, Face: 250000}
//     table := loan.GetAmortizationTable()
//
//     for i := 0; i < len(table.Period); i++ {
//         fmt.Printf("Period %d: Balance $%.2f\n", table.Period[i], table.EndBal[i])
//     }
func (l *LoanInfo) GetAmortizationTable() AmortizationTable {
    // Implementation
}
```

### Generate Documentation

```bash
# View documentation locally
go doc -all ./amortization

# Generate HTML documentation
godoc -http=:6060

# Then visit http://localhost:6060/pkg/github.com/jiangshenghai57/andy-warhol/
```

---

## Quick Reference

### Common Commands

```bash
# Format code
go fmt ./...

# Lint code
golangci-lint run

# Run tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...

# Run benchmarks
go test -bench=. -benchmem ./...

# Build
go build -o andy-warhol .

# Build for production
CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s' -o andy-warhol .

# Generate mocks
go generate ./...
```

### Code Review Checklist

- [ ] All exported functions have documentation
- [ ] Errors are wrapped with context
- [ ] No panics in library code
- [ ] Slices are pre-allocated when size is known
- [ ] Goroutines are properly coordinated with WaitGroups
- [ ] Shared state is protected with mutexes
- [ ] Tests cover edge cases and error paths
- [ ] No hardcoded credentials or secrets
- [ ] Inputs are validated before processing

---

**Remember: Clear code is better than clever code. When in doubt, keep it simple.**