package logger

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewLogger_Success(t *testing.T) {
	tests := []struct {
		name   string
		logDir string
	}{
		{
			name:   "simple directory",
			logDir: t.TempDir(),
		},
		{
			name:   "nested directory creation",
			logDir: filepath.Join(t.TempDir(), "logs", "nested", "deep"),
		},
		{
			name:   "current directory",
			logDir: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.logDir)

			if err != nil {
				t.Errorf("NewLogger() unexpected error: %v", err)
				return
			}

			if logger == nil {
				t.Error("NewLogger() returned nil logger")
				return
			}

			if logger.Logger == nil {
				t.Error("NewLogger() returned logger with nil *slog.Logger")
			}
		})
	}
}

func TestNewLogger_CreatesLogFile(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}

	if logger == nil {
		t.Fatal("NewLogger() returned nil logger")
	}

	// Verify log file exists with today's date
	expectedFileName := time.Now().Format("2006-01-02") + ".log"
	logFilePath := filepath.Join(tempDir, expectedFileName)

	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		t.Errorf("expected log file %s does not exist", logFilePath)
	}
}

func TestNewLogger_InvalidPermissions(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping test when running as root")
	}

	// Create directory with no write permissions
	tempDir := t.TempDir()
	noWriteDir := filepath.Join(tempDir, "no-write")
	if err := os.Mkdir(noWriteDir, 0444); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	logDir := filepath.Join(noWriteDir, "logs")
	_, err := NewLogger(logDir)

	if err == nil {
		t.Error("NewLogger() expected permission error, got nil")
	}
}

func TestLogger_InfoLogging(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}

	// Log test message matching Andy Warhol loan structure
	logger.Info("processing loan",
		slog.String("loan_id", "LOAN001"),
		slog.Float64("face", 250000),
		slog.Int("wam", 360),
		slog.Float64("wac", 4.5),
	)

	// Read and parse log file
	logFile := filepath.Join(tempDir, time.Now().Format("2006-01-02")+".log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// Verify JSON structure and expected fields
	var logEntry map[string]interface{}
	if err := json.Unmarshal(content, &logEntry); err != nil {
		t.Fatalf("log output is not valid JSON: %v", err)
	}

	// Check required fields
	expectedFields := map[string]interface{}{
		"level":   "INFO",
		"msg":     "processing loan",
		"loan_id": "LOAN001",
		"face":    float64(250000),
		"wam":     float64(360),
		"wac":     float64(4.5),
	}

	for field, expectedValue := range expectedFields {
		actualValue, exists := logEntry[field]
		if !exists {
			t.Errorf("log entry missing field: %s", field)
			continue
		}

		if actualValue != expectedValue {
			t.Errorf("field %s: expected %v, got %v", field, expectedValue, actualValue)
		}
	}

	// Verify source location is included (AddSource: true)
	if _, hasSource := logEntry["source"]; !hasSource {
		t.Error("log entry missing source location")
	}
}

func TestLogger_ErrorLogging(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}

	testErr := os.ErrNotExist

	// Log error matching Andy Warhol error patterns
	logger.Error("amortization calculation failed",
		slog.String("loan_id", "LOAN002"),
		slog.Any("error", testErr),
		slog.String("reason", "invalid WAM value"),
	)

	// Read log file
	logFile := filepath.Join(tempDir, time.Now().Format("2006-01-02")+".log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	logContent := string(content)

	// Verify error level and fields
	if !strings.Contains(logContent, `"level":"ERROR"`) {
		t.Error("log missing ERROR level")
	}
	if !strings.Contains(logContent, `"msg":"amortization calculation failed"`) {
		t.Error("log missing error message")
	}
	if !strings.Contains(logContent, `"loan_id":"LOAN002"`) {
		t.Error("log missing loan_id field")
	}
	if !strings.Contains(logContent, `"reason":"invalid WAM value"`) {
		t.Error("log missing reason field")
	}
}

func TestLogger_WarnLogging(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}

	// Log warning for edge cases in cashflow calculations
	logger.Warn("high prepayment rate detected",
		slog.String("loan_id", "LOAN003"),
		slog.Float64("prepay_cpr", 0.25),
		slog.String("recommendation", "review prepayment assumptions"),
	)

	// Read log file
	logFile := filepath.Join(tempDir, time.Now().Format("2006-01-02")+".log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	logContent := string(content)

	// Verify warning level
	if !strings.Contains(logContent, `"level":"WARN"`) {
		t.Error("log missing WARN level")
	}
	if !strings.Contains(logContent, `"prepay_cpr":0.25`) {
		t.Error("log missing prepay_cpr field")
	}
}

func TestLogger_SourceLocationIncluded(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}

	logger.Info("test with source location")

	// Read log file
	logFile := filepath.Join(tempDir, time.Now().Format("2006-01-02")+".log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	logContent := string(content)

	// Verify source location metadata
	requiredSourceFields := []string{
		`"source"`,
		"logger_test.go",
	}

	for _, field := range requiredSourceFields {
		if !strings.Contains(logContent, field) {
			t.Errorf("log content missing source field: %s\nGot: %s", field, logContent)
		}
	}
}

func TestLogger_MultipleLogLevels(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}

	// Log at different levels matching Andy Warhol workflow
	logger.Info("batch processing started",
		slog.Int("loan_count", 1000),
	)
	logger.Warn("worker pool nearing capacity",
		slog.Int("active_workers", 95),
		slog.Int("max_workers", 100),
	)
	logger.Error("batch processing failed",
		slog.String("reason", "timeout exceeded"),
	)

	// Read log file
	logFile := filepath.Join(tempDir, time.Now().Format("2006-01-02")+".log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	logContent := string(content)

	// Verify all levels are present
	levels := []string{
		`"level":"INFO"`,
		`"level":"WARN"`,
		`"level":"ERROR"`,
	}

	for _, level := range levels {
		if !strings.Contains(logContent, level) {
			t.Errorf("log content missing expected level: %s", level)
		}
	}
}

func TestLogger_AppendToExistingFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create first logger and write
	logger1, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("NewLogger() first instance failed: %v", err)
	}
	logger1.Info("first message", slog.String("batch", "1"))

	// Create second logger (should append to same file)
	logger2, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("NewLogger() second instance failed: %v", err)
	}
	logger2.Info("second message", slog.String("batch", "2"))

	// Read log file
	logFile := filepath.Join(tempDir, time.Now().Format("2006-01-02")+".log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	logContent := string(content)

	// Verify both messages are present (append mode working)
	if !strings.Contains(logContent, "first message") {
		t.Error("log file missing first message")
	}
	if !strings.Contains(logContent, "second message") {
		t.Error("log file missing second message")
	}
	if !strings.Contains(logContent, `"batch":"1"`) {
		t.Error("log file missing first batch identifier")
	}
	if !strings.Contains(logContent, `"batch":"2"`) {
		t.Error("log file missing second batch identifier")
	}
}

func TestLogger_ConcurrentWrites(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		t.Fatalf("NewLogger() failed: %v", err)
	}

	// Simulate concurrent loan processing logs (Andy Warhol worker pool pattern)
	const numWorkers = 10
	done := make(chan bool, numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			logger.Info("processing loan",
				slog.Int("worker_id", workerID),
				slog.String("loan_id", "LOAN"+string(rune(workerID+'0'))),
			)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numWorkers; i++ {
		<-done
	}

	// Read log file
	logFile := filepath.Join(tempDir, time.Now().Format("2006-01-02")+".log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	logContent := string(content)

	// Count log entries (should have at least numWorkers entries)
	lines := strings.Split(strings.TrimSpace(logContent), "\n")
	if len(lines) < numWorkers {
		t.Errorf("expected at least %d log entries, got %d", numWorkers, len(lines))
	}
}

func BenchmarkLogger_Info(b *testing.B) {
	tempDir := b.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		b.Fatalf("NewLogger() failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("processing loan",
			slog.String("loan_id", "LOAN001"),
			slog.Float64("face", 250000),
			slog.Int("wam", 360),
			slog.Float64("wac", 4.5),
		)
	}
}

func BenchmarkLogger_Error(b *testing.B) {
	tempDir := b.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		b.Fatalf("NewLogger() failed: %v", err)
	}

	testErr := os.ErrNotExist

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Error("calculation failed",
			slog.String("loan_id", "LOAN001"),
			slog.Any("error", testErr),
		)
	}
}

func BenchmarkLogger_ConcurrentWrites(b *testing.B) {
	tempDir := b.TempDir()

	logger, err := NewLogger(tempDir)
	if err != nil {
		b.Fatalf("NewLogger() failed: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("concurrent write",
				slog.String("loan_id", "LOAN001"),
				slog.Float64("face", 250000),
			)
		}
	})
}
