# Go Best Practices Guide

A concise, repository-agnostic guide to writing clean, efficient, and maintainable Go code. Focuses on principles, idiomatic patterns, and small examples you can apply to any Go project.

---

## Table of Contents

- [Project Layout](#project-layout)
- [Naming & Style](#naming--style)
- [Error Handling](#error-handling)
- [Concurrency](#concurrency)
- [Performance](#performance)
- [Testing](#testing)
- [Logging](#logging)
- [API Design](#api-design)
- [Security](#security)
- [Documentation & Tooling](#documentation--tooling)
- [Quick Commands](#quick-commands)

---

## Project Layout

Prefer small, focused packages. Example layout:

```
cmd/           # Applications (main packages)
internal/      # Private application code
pkg/           # Reusable libraries
api/           # OpenAPI / API schemas
configs/       # Configuration files
scripts/       # Build / CI scripts
testdata/      # Fixtures and sample data
```

Keep domain types, calculations, and transport (HTTP/CLI) separated.

---

## Naming & Style

- Use clear, descriptive names.
- Acronyms: use consistent casing (e.g., `HTTPClient`, `LoanID`).
- Exported identifiers use PascalCase; internal use camelCase.
- Keep functions short and focused.

Example:

```go
// Good
func calculateMonthlyPayment(principal, monthlyRate float64, numPeriods int) float64

// Bad
func calcPmt(p, r float64, n int) float64
```

---

## Error Handling

- Return errors; do not panic in library code.
- Wrap errors with context using fmt.Errorf("%w") or errors.Join.
- Prefer sentinel or typed errors for special cases.

```go
if err := validateLoan(l); err != nil {
    return fmt.Errorf("validate loan %q: %w", l.ID, err)
}
```

---

## Concurrency

- Use sync.WaitGroup to coordinate goroutines.
- Limit concurrency with worker pools or semaphores (buffered channels).
- Accept context.Context for cancellation.

```go
var wg sync.WaitGroup
sem := make(chan struct{}, maxWorkers)

for _, task := range tasks {
    wg.Add(1)
    sem <- struct{}{}
    go func(t Task) {
        defer wg.Done()
        defer func(){ <-sem }()
        doWork(ctx, t)
    }(task)
}
wg.Wait()
```

Protect shared state with mutexes or design to avoid shared mutable state.

---

## Performance

- Pre-allocate slices when size is known.
- Avoid allocations in hot loops; reuse buffers or use sync.Pool.
- Move invariant computations out of loops.

```go
buf := make([]byte, 0, 1024) // pre-allocate
for i := 0; i < n; i++ {
    // reuse buf
}
```

For small fixed-size math (e.g., 8-state vectors), use fixed-size arrays to avoid allocations.

---

## Testing

- Use table-driven tests.
- Add benchmarks for hot code with `go test -bench`.
- Use fixtures in testdata/ and avoid network calls in unit tests.

```go
func TestCalculateMonthlyPayment(t *testing.T) {
    tests := []struct{ name string; p, r float64; n int; want float64 }{
        {"zero rate", 1000, 0, 12, 83.33},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := calculateMonthlyPayment(tt.p, tt.r, tt.n)
            if math.Abs(got-tt.want) > 1e-2 { t.Fatalf("got %.2f want %.2f", got, tt.want) }
        })
    }
}
```

---

## Logging

- Use structured logging for services (allow JSON output).
- Inject logger via interfaces to ease testing.
- Avoid logging sensitive data.

```go
type Logger interface { Info(msg string, fields ...Field) }
```

---

## API Design

- Validate input early and return clear error payloads.
- Keep interfaces small and focused.
- Prefer consistent response format.

Example response wrapper:

```go
type APIResponse struct {
    Success   bool        `json:"success"`
    Data      interface{} `json:"data,omitempty"`
    Error     string      `json:"error,omitempty"`
    Timestamp time.Time   `json:"timestamp"`
}
```

---

## Security

- Validate and sanitize inputs.
- Do not commit secrets. Use environment variables or secret stores.
- Rate-limit and use context timeouts for external calls.

Quick grep for leaks:

```bash
rg -n --hidden '(apikey|api_key|secret|password|BEGIN PRIVATE KEY)' || true
```

---

## Documentation & Tooling

- Document packages and exported APIs with Go doc comments.
- Use gofmt/gofumpt and golangci-lint for consistent style.
- Generate API docs and README for high-level usage.

Helpful commands:

```bash
gofmt -w .
gofumpt -w .
golangci-lint run
go test ./... -v
```

---

## Quick Commands

- Format: gofmt or gofumpt
- Lint: golangci-lint run
- Test: go test ./...
- Bench: go test -bench=. -benchmem ./...

---

...existing code...