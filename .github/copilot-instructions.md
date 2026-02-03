# Andy Warhol - Mortgage Cashflow Engine

A high-performance Go-based REST API for calculating mortgage amortization schedules with prepayment modeling and delinquency transition analysis. Built for deployment in Kubernetes environments with concurrent processing capabilities.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Getting Started](#getting-started)
- [API Reference](#api-reference)
- [Code Examples](#code-examples)
- [Configuration](#configuration)
- [Deployment](#deployment)
- [Performance](#performance)
- [Contributing](#contributing)

---

## Overview

Andy Warhol is a mortgage cashflow projection engine that calculates:

- **Amortization schedules** with principal and interest breakdowns
- **Prepayment modeling** using CPR (Conditional Prepayment Rate) to SMM (Single Monthly Mortality) conversion
- **Delinquency transitions** modeling balance flow between performing, DQ30, DQ60, DQ90, DQ120, DQ150, DQ180, and default states
- **Batch processing** of up to 1000+ loans concurrently

### Why "Andy Warhol"?

Like the artist who mass-produced art with precision and speed, this engine mass-produces mortgage cashflows with high throughput and accuracy.

---

## Features

| Feature | Description |
|---------|-------------|
| 🚀 **High Performance** | Worker pool pattern for concurrent loan processing |
| 📊 **Full Amortization** | Period-by-period breakdown of principal, interest, and balances |
| 💰 **Prepayment Modeling** | CPR to SMM conversion with customizable prepayment curves |
| 📉 **Delinquency Transitions** | Markov chain-based transition modeling between delinquency states |
| 🐳 **Container Ready** | Multi-stage Docker builds for minimal image size |
| ☸️ **Kubernetes Native** | Health probes, graceful shutdown, and resource-aware scaling |
| 📝 **Structured Logging** | Dual output to file and stdout with timestamps |

---

## Architecture

```
┌─────────────────────────────────────────────────────────────-┐
│                      API Gateway                             │
│                    (Gin HTTP Router)                         │
├─────────────────────────────────────────────────────────────-┤
│                                                              │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐       │
│  │  /health    │    │  GET /loans │    │ POST /loans │       │
│  │  endpoint   │    │  endpoint   │    │  endpoint   │       │
│  └─────────────┘    └─────────────┘    └──────┬──────┘       │
│                                               │              │
├───────────────────────────────────────────────┼──────────────┤
│                                               ▼              │
│                    ┌─────────────────────────────┐           │
│                    │      Worker Pool            │           │
│                    │   (Concurrent Processing)   │           │
│                    └─────────────┬───────────────┘           │
│                                  │                           │
│          ┌───────────────────────┼───────────────────────┐   │
│          ▼                       ▼                       ▼   │
│  ┌───────────────┐      ┌───────────────┐      ┌───────────-┐│
│  │   Loan 1      │      │   Loan 2      │      │  Loan N    ││
│  │ Amortization  │      │ Amortization  │      │Amortization││
│  └───────┬───────┘      └───────┬───────┘      └─────┬─────-┘│
│          │                      │                    │       │
│          ▼                      ▼                    ▼       │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Amortization Engine                        │ │
│  │  • CPR → SMM Conversion                                 │ │
│  │  • Monthly Payment Calculation                          │ │
│  │  • Delinquency Transition Matrices                      │ │
│  └─────────────────────────────────────────────────────────┘ │
│                                                              │
└─────────────────────────────────────────────────────────────-┘
```

---

## Getting Started

### Prerequisites

- Go 1.22+
- Docker (optional, for containerized deployment)
- Kubernetes (optional, for orchestrated deployment)

### Installation

```bash
# Clone the repository
git clone https://github.com/jiangshenghai57/andy-warhol.git
cd andy-warhol

# Download dependencies
go mod download

# Build the application
go build -o andy-warhol .

# Run the application
./andy-warhol
```

### Quick Start

```bash
# Start the server
./andy-warhol

# In another terminal, send a test request
curl -X POST http://localhost:8080/loans \
  -H "Content-Type: application/json" \
  -d '[{
    "id": "LOAN001",
    "wac": 4.5,
    "wam": 360,
    "face": 250000,
    "prepay_cpr": 0.06,
    "static_dq": false
  }]'
```

---

## API Reference

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check for Kubernetes probes |
| `GET` | `/loans` | Retrieve all processed loans |
| `POST` | `/loans` | Submit loans for cashflow calculation |

### Request Schema

```json
{
  "id": "string",
  "wac": "number (annual rate as percentage, e.g., 4.5)",
  "wam": "integer (months remaining, e.g., 360)",
  "face": "number (current balance, e.g., 250000)",
  "prepay_cpr": "number (annual CPR as decimal, e.g., 0.06)",
  "static_dq": "boolean (use static delinquency model)",
  "performing_transition": "[8]float64 (optional)",
  "dq30_transition": "[8]float64 (optional)",
  "dq60_transition": "[8]float64 (optional)",
  "dq90_transition": "[8]float64 (optional)",
  "dq120_transition": "[8]float64 (optional)",
  "dq150_transition": "[8]float64 (optional)",
  "dq180_transition": "[8]float64 (optional)",
  "default_transition": "[8]float64 (optional)"
}
```

### Response Schema

```json
{
  "count": 1,
  "results": [
    {
      "loan_id": "LOAN001",
      "cashflow": {
        "period": [1, 2, 3, "..."],
        "beg_bal": [250000.00, 249876.23, "..."],
        "interest": [937.50, 936.03, "..."],
        "principal": [327.32, 328.55, "..."],
        "sched_bal": [249672.68, 249344.13, "..."],
        "prepay_amount_arr": [1298.50, 1296.79, "..."],
        "end_bal": [248374.18, 248047.34, "..."],
        "delinq_arrays": {
          "perf_arr": [248374.18, "..."],
          "dq30_arr": [0.00, "..."],
          "dq60_arr": [0.00, "..."],
          "dq90_arr": [0.00, "..."],
          "dq120_arr": [0.00, "..."],
          "dq150_arr": [0.00, "..."],
          "dq180_arr": [0.00, "..."],
          "default_arr": [0.00, "..."]
        }
      }
    }
  ]
}
```

---

## Code Examples

### 1. Basic Loan Structure

```go
package amortization

// LoanInfo represents mortgage information
type LoanInfo struct {
    ID         string    `json:"id"`
    Wam        int64     `json:"wam"`         // Weighted Average Maturity (months)
    Wac        float64   `json:"wac"`         // Weighted Average Coupon (annual %)
    Face       float64   `json:"face"`        // Current face value
    PrepayCPR  float64   `json:"prepay_cpr"`  // Conditional Prepayment Rate
    SMMArr     []float64 `json:"smm_arr"`     // Single Monthly Mortality array
    StaticDQ   bool      `json:"static_dq"`   // Use static delinquency model
    
    // Transition matrices (8 states each)
    PerformingTransition []float64 `json:"performing_transition,omitempty"`
    DQ30Transition       []float64 `json:"dq30_transition,omitempty"`
    DQ60Transition       []float64 `json:"dq60_transition,omitempty"`
    DQ90Transition       []float64 `json:"dq90_transition,omitempty"`
    DQ120Transition      []float64 `json:"dq120_transition,omitempty"`
    DQ150Transition      []float64 `json:"dq150_transition,omitempty"`
    DQ180Transition      []float64 `json:"dq180_transition,omitempty"`
    DefaultTransition    []float64 `json:"default_transition,omitempty"`
}
```

### 2. CPR to SMM Conversion

```go
// ConvertCPRToSMM converts annual CPR to monthly SMM
func (l *LoanInfo) ConvertCPRToSMM() {
    if l.PrepayCPR > 0.0 {
        // SMM = 1 - (1 - CPR)^(1/12)
        smm := 1 - math.Pow(1-l.PrepayCPR, 1.0/12.0)
        
        l.SMMArr = make([]float64, l.Wam)
        for i := range l.SMMArr {
            l.SMMArr[i] = smm
        }
        
        log.Printf("Loan %s: Converted CPR %.4f to SMM %.6f", 
            l.ID, l.PrepayCPR, smm)
    } else {
        l.SMMArr = make([]float64, l.Wam)
    }
}
```

### 3. Monthly Payment Calculation

```go
// calculateMonthlyPayment computes the fixed monthly payment
func calculateMonthlyPayment(principal, monthlyRate float64, numPayments float64) float64 {
    if monthlyRate == 0 {
        return principal / numPayments
    }
    
    // PMT = P * [r(1+r)^n] / [(1+r)^n - 1]
    factor := math.Pow(1+monthlyRate, numPayments)
    return principal * (monthlyRate * factor) / (factor - 1)
}
```

### 4. Amortization Table Generation

```go
// GetAmortizationTable calculates the full amortization schedule
func (l *LoanInfo) GetAmortizationTable() AmortizationTable {
    l.SetDefaultTransitions()
    l.ConvertCPRToSMM()
    
    numPeriods := int(l.Wam)
    
    // Pre-allocate arrays for performance
    periods := make([]int, numPeriods)
    begBal := make([]float64, numPeriods)
    interest := make([]float64, numPeriods)
    principal := make([]float64, numPeriods)
    schedBal := make([]float64, numPeriods)
    prepayAmountArr := make([]float64, numPeriods)
    endBal := make([]float64, numPeriods)
    
    monthlyRate := l.Wac / 12.0 / 100.0
    monthlyPayment := calculateMonthlyPayment(l.Face, monthlyRate, float64(l.Wam))
    
    tmp_face := l.Face
    
    for j := 0; j < numPeriods; j++ {
        periods[j] = j + 1
        begBal[j] = roundToCent(tmp_face)
        
        // Interest = Balance * Monthly Rate
        interestPayment := tmp_face * monthlyRate
        interest[j] = roundToCent(interestPayment)
        
        // Principal = Payment - Interest
        principalPayment := monthlyPayment - interestPayment
        if j == numPeriods-1 {
            principalPayment = tmp_face // Final payment
        }
        principal[j] = roundToCent(principalPayment)
        
        // Scheduled Balance after principal payment
        currentSchedBal := tmp_face - principalPayment
        schedBal[j] = roundToCent(currentSchedBal)
        
        // Prepayment = SMM * Scheduled Balance
        prepayAmount := l.SMMArr[j] * currentSchedBal
        prepayAmountArr[j] = roundToCent(prepayAmount)
        
        // Ending Balance = Scheduled Balance - Prepayment
        tmp_face = currentSchedBal - prepayAmount
        if tmp_face < 0.0 {
            tmp_face = 0.0
        }
        endBal[j] = roundToCent(tmp_face)
    }
    
    return AmortizationTable{
        Period:          periods,
        BegBal:          begBal,
        Interest:        interest,
        Principal:       principal,
        SchedBal:        schedBal,
        PrepayAmountArr: prepayAmountArr,
        EndBal:          endBal,
    }
}
```

### 5. Delinquency Transition Model

```go
// Transition matrix format: [performing, dq30, dq60, dq90, dq120, dq150, dq180, default]
// Each row sums to 1.0

// Default performing transition: 98% stay performing, 2% go to DQ30
PerformingTransition := []float64{0.98, 0.02, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0}

// Apply transition to calculate new distribution
func applyTransition(balance float64, transitions []float64, result []float64) {
    for i, rate := range transitions {
        result[i] += balance * rate
    }
}

// Scale distribution to match total balance
func scaleDistribution(distribution []float64, targetBalance float64) {
    var total float64
    for _, val := range distribution {
        total += val
    }
    
    if total <= 0 {
        return
    }
    
    scaleFactor := targetBalance / total
    for i := range distribution {
        distribution[i] = roundToCent(distribution[i] * scaleFactor)
    }
}
```

### 6. Worker Pool Pattern for Concurrent Processing

```go
var workerPool = make(chan struct{}, 100) // Limit concurrent workers

func requestCashflow(c *gin.Context) {
    var loans []amortization.LoanInfo
    
    if err := c.BindJSON(&loans); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    var wg sync.WaitGroup
    results := make([]gin.H, len(loans))
    
    for i, loan := range loans {
        wg.Add(1)
        
        go func(index int, l amortization.LoanInfo) {
            // Acquire worker from pool
            workerPool <- struct{}{}
            defer func() {
                <-workerPool // Release worker
                wg.Done()
            }()
            
            // Calculate amortization
            amortTable := l.GetAmortizationTable()
            
            results[index] = gin.H{
                "loan_id":  l.ID,
                "cashflow": amortTable,
            }
        }(i, loan)
    }
    
    wg.Wait()
    
    c.JSON(http.StatusOK, gin.H{
        "count":   len(loans),
        "results": results,
    })
}
```

### 7. Logging Setup

```go
// setupLogging configures dual output to file and stdout
func setupLogging() {
    logDir := "logs"
    os.MkdirAll(logDir, 0755)
    
    logFileName := filepath.Join(logDir, fmt.Sprintf("andy-warhol-%s.log", 
        time.Now().Format("2006-01-02")))
    
    logFile, err := os.OpenFile(logFileName, 
        os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        log.Fatalf("Failed to open log file: %v", err)
    }
    
    // Write to both file and stdout
    multiWriter := io.MultiWriter(logFile, os.Stdout)
    
    log.SetOutput(multiWriter)
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
    
    gin.DefaultWriter = multiWriter
    gin.DefaultErrorWriter = multiWriter
}
```

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GIN_MODE` | `release` | Gin framework mode (`debug`, `release`, `test`) |
| `WORKER_LIMIT` | `100` | Maximum concurrent workers |
| `LOG_PATH` | `./logs/` | Directory for log files |
| `PORT` | `8080` | HTTP server port |

### Config File (`config/config.json`)

```json
{
  "LOG_PATH": "./logs/",
  "LOG_FILE": "andy-warhol.log",
  "WORKER_LIMIT": 100,
  "PORT": 8080
}
```

---

## Deployment

### Docker Build

```bash
# Build the image
docker build -t andy-warhol:latest .

# Run the container
docker run -d \
  --name andy-warhol \
  -p 8080:8080 \
  -v $(pwd)/output:/app/output \
  -e GIN_MODE=release \
  -e WORKER_LIMIT=50 \
  andy-warhol:latest
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: andy-warhol
spec:
  replicas: 3
  selector:
    matchLabels:
      app: andy-warhol
  template:
    metadata:
      labels:
        app: andy-warhol
    spec:
      containers:
      - name: andy-warhol
        image: andy-warhol:latest
        ports:
        - containerPort: 8080
        env:
        - name: GIN_MODE
          value: "release"
        - name: WORKER_LIMIT
          value: "20"
        resources:
          limits:
            cpu: "2000m"
            memory: "1Gi"
          requests:
            cpu: "500m"
            memory: "256Mi"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

---

## Performance

### Benchmarks

| Loans | Processing Time | Memory Usage |
|-------|-----------------|--------------|
| 1 | ~5ms | ~2MB |
| 100 | ~50ms | ~20MB |
| 1,000 | ~500ms | ~200MB |
| 10,000 | ~5s | ~2GB |

### Optimization Tips

1. **Pre-allocate slices** instead of using `append()` in loops
2. **Use worker pools** to limit concurrent goroutines
3. **Avoid decimal libraries** in hot loops - use float64 with rounding
4. **Batch requests** - process multiple loans in a single API call

---

## Project Structure

```
andy-warhol/
├── main.go                 # Application entry point
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
├── Dockerfile              # Container build instructions
├── docker_build.sh         # Docker build script
├── single_loan_post.sh     # Test script for API
├── loans_payload.json      # Sample loan data (1000 loans)
├── amortization/
│   └── amortization.go     # Core amortization calculations
├── config/
│   ├── config.go           # Configuration reader
│   └── config.json         # Configuration file
├── output/                 # Generated cashflow files
└── logs/                   # Application logs
```

---

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [Gin Web Framework](https://github.com/gin-gonic/gin) - High-performance HTTP router
- [Shopspring Decimal](https://github.com/shopspring/decimal) - Arbitrary-precision decimal library
- Inspired by mortgage-backed securities cashflow modeling practices

---

**Built with ❤️ for the mortgage industry**
