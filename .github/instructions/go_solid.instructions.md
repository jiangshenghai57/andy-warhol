# Go + SOLID — Practical Instructions for Andy Warhol

Concise, practical guide to applying SOLID in Go with short code examples tailored to the amortization project.

---

## Single Responsibility Principle (SRP)
Each package / file / type should have one reason to change. Keep domain types, calculations, I/O, and HTTP handlers separate.

Example: separate types and calculations.

```go
// amortization/types.go
package amortization

type LoanInfo struct {
    ID        string
    Wam       int
    Wac       float64
    Face      float64
    PrepayCPR float64
    StaticDQ  bool
}

// amortization/calc.go
package amortization

func CalculateMonthlyPayment(principal float64, monthlyRate float64, numPeriods int) float64 {
    if monthlyRate == 0 {
        return principal / float64(numPeriods)
    }
    f := math.Pow(1+monthlyRate, float64(numPeriods))
    return principal * (monthlyRate*f)/(f-1)
}
```

---

## Open/Closed Principle (OCP)
Make behavior extensible without modifying existing code. Use interfaces for extension points (e.g., different delinquency models).

```go
// amortization/pool.go
package amortization

type MortgagePool interface {
    GenerateAmortTable() AmortizationTable
}

// new model can be added without editing existing code:
type MarkovPool struct{ /* ... */ }
func (m *MarkovPool) GenerateAmortTable() AmortizationTable { /* ... */ }
```

---

## Liskov Substitution Principle (LSP)
Subtypes must be usable via base interfaces without surprises. Keep contracts clear and avoid side-effects that break callers.

```go
// contract: GenerateAmortTable returns a complete table and does not mutate shared global state
func Process(pool MortgagePool) AmortizationTable {
    return pool.GenerateAmortTable()
}
```

Ensure implementations obey validation and error semantics: return deterministic, documented results.

---

## Interface Segregation Principle (ISP)
Prefer focused interfaces over large "fat" interfaces. Consumers should depend only on the methods they use.

```go
package amortization

type PrepayCalculator interface {
    ConvertCPRToSMM() []float64
}

type AmortCalculator interface {
    GenerateAmortTable() AmortizationTable
}
```

Use small interfaces in function signatures:

```go
func RunPrepay(pc PrepayCalculator) []float64 { return pc.ConvertCPRToSMM() }
```

---

## Dependency Inversion Principle (DIP)
High-level modules should depend on abstractions. Inject dependencies (loggers, storage, worker pools) via constructors or parameters.

```go
package api

type Logger interface { Info(args ...interface{}) }
type WorkerPool interface { Do(func()) }

type Service struct {
    logger Logger
    pool   WorkerPool
}

func NewService(logger Logger, pool WorkerPool) *Service {
    return &Service{logger: logger, pool: pool}
}

func (s *Service) ProcessLoan(l amortization.LoanInfo) {
    s.pool.Do(func() {
        s.logger.Info("processing", l.ID)
        table := l.GenerateAmortTable() // uses amortization implementation
        _ = table
    })
}
```

Prefer simple interfaces (one or two methods) for easier testing and swapping implementations.

---

## Practical tips & patterns

- Favor composition over inheritance: embed smaller structs for shared behavior.
- Keep methods small and pure where possible; side-effects should be explicit and documented.
- Validate inputs early; return errors not panics.
- Use fixed-size arrays for hot numeric paths (8-state vectors) to avoid allocations.
- Unit test each principle: validation tests, interface conformance tests, and benchmarks for hot code.

Example: Matrix type for transitions (fast, testable)

```go
// amortization/matrix.go
package amortization

type StateVec [8]float64
type TransMatrix [8][8]float64

func (m *TransMatrix) Mul(src StateVec) (dst StateVec) {
    for i := range dst {
        dst[i] = 0
    }
    for from := 0; from < 8; from++ {
        f := src[from]
        if f == 0 { continue }
        row := m[from]
        for to := 0; to < 8; to++ {
            dst[to] += f * row[to]
        }
    }
    return
}
```

---

## Testing & Maintenance

- Use table-driven tests per small unit (prepay conversion, payment calc, transition matrix).
- Benchmark hot functions (amortization loop) with `go test -bench`.
- Add interface mocks for unit tests; prefer constructor injection for dependencies.

---

## File layout (recommended)

```
/amortization
  ├─ types.go
  ├─ calc.go
  ├─ prepay.go
  ├─ transitions.go
  ├─ matrix.go
  └─ utils.go
/cmd/server
  └─ main.go
/internal/api
  └─ service.go
/docs
  └─ GO_SOLID.md
```

---

Adopt these patterns incrementally. Small, focused interfaces and composition make the amortization engine easier to test, extend, and