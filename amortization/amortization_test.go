package amortization

import (
	"math"
	"testing"
)

func TestEnsureSMMArrayType(t *testing.T) {
	// Test with valid type
	prepay := &PrepayInfo{
		SMMArr: []float64{0.01, 0.02, 0.03},
	}

	result := prepay.ensureSMMArrayType()
	if len(result) != 3 {
		t.Errorf("Expected length 3, got %d", len(result))
	}

}

func TestConvertCPRToSMM(t *testing.T) {
	prepay := &PrepayInfo{
		PrepayCPR: 0.05,
	}

	prepay.ConvertCPRToSMM(360)

	// Verify SMM array length
	if len(prepay.SMMArr) != 360 {
		t.Errorf("Expected SMM array length 360, got %d", len(prepay.SMMArr))
	}

	// Calculate expected SMM: SMM = 1 - (1 - CPR)^(1/12)
	expectedSMM := 1 - math.Pow(1-0.05, 1.0/12.0)

	// Verify all values are correct
	for i, smm := range prepay.SMMArr {
		if math.Abs(smm-expectedSMM) > 0.0001 {
			t.Errorf("Period %d: Expected SMM %.6f, got %.6f", i, expectedSMM, smm)
		}
	}
}

func TestConvertCPRToSMM_ZeroCPR(t *testing.T) {
	prepay := &PrepayInfo{
		PrepayCPR: 0.0,
	}

	prepay.ConvertCPRToSMM(360)

	// Verify SMM array is initialized with zeros
	if len(prepay.SMMArr) != 360 {
		t.Errorf("Expected SMM array length 360, got %d", len(prepay.SMMArr))
	}

	for i, smm := range prepay.SMMArr {
		if smm != 0.0 {
			t.Errorf("Period %d: Expected SMM 0.0, got %.6f", i, smm)
		}
	}
}

func TestConvertCPRToSMM_VariousRates(t *testing.T) {
	testCases := []struct {
		name        string
		cpr         float64
		numMonths   int
		expectedSMM float64
	}{
		{
			name:        "6% CPR (typical residential)",
			cpr:         0.06,
			numMonths:   360,
			expectedSMM: 1 - math.Pow(1-0.06, 1.0/12.0),
		},
		{
			name:        "20% CPR (high prepayment)",
			cpr:         0.20,
			numMonths:   180,
			expectedSMM: 1 - math.Pow(1-0.20, 1.0/12.0),
		},
		{
			name:        "2% CPR (low prepayment)",
			cpr:         0.02,
			numMonths:   240,
			expectedSMM: 1 - math.Pow(1-0.02, 1.0/12.0),
		},
		{
			name:        "15% CPR (moderate prepayment)",
			cpr:         0.15,
			numMonths:   300,
			expectedSMM: 1 - math.Pow(1-0.15, 1.0/12.0),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prepay := &PrepayInfo{
				PrepayCPR: tc.cpr,
			}

			prepay.ConvertCPRToSMM(tc.numMonths)

			// Verify array length
			if len(prepay.SMMArr) != tc.numMonths {
				t.Errorf("Expected SMM array length %d, got %d", tc.numMonths, len(prepay.SMMArr))
			}

			// Verify SMM calculation
			for i, smm := range prepay.SMMArr {
				if math.Abs(smm-tc.expectedSMM) > 0.000001 {
					t.Errorf("Period %d: Expected SMM %.8f, got %.8f", i, tc.expectedSMM, smm)
				}
			}
		})
	}
}

func TestConvertCPRToSMM_ShortTerm(t *testing.T) {
	prepay := &PrepayInfo{
		PrepayCPR: 0.10,
	}

	prepay.ConvertCPRToSMM(12) // 1 year

	if len(prepay.SMMArr) != 12 {
		t.Errorf("Expected SMM array length 12, got %d", len(prepay.SMMArr))
	}

	expectedSMM := 1 - math.Pow(1-0.10, 1.0/12.0)

	for i, smm := range prepay.SMMArr {
		if math.Abs(smm-expectedSMM) > 0.0001 {
			t.Errorf("Period %d: Expected SMM %.6f, got %.6f", i, expectedSMM, smm)
		}
	}
}

func TestConvertCPRToSMM_NilSMMArr(t *testing.T) {
	prepay := &PrepayInfo{
		PrepayCPR: 0.0,
		SMMArr:    nil, // Explicitly nil
	}

	prepay.ConvertCPRToSMM(120)

	if prepay.SMMArr == nil {
		t.Error("Expected SMMArr to be initialized, got nil")
	}

	if len(prepay.SMMArr) != 120 {
		t.Errorf("Expected SMM array length 120, got %d", len(prepay.SMMArr))
	}
}

func TestConvertCPRToSMM_BoundaryValues(t *testing.T) {
	testCases := []struct {
		name      string
		cpr       float64
		numMonths int
	}{
		{
			name:      "Very low CPR",
			cpr:       0.001,
			numMonths: 360,
		},
		{
			name:      "Very high CPR",
			cpr:       0.99,
			numMonths: 360,
		},
		{
			name:      "Single month",
			cpr:       0.05,
			numMonths: 1,
		},
		{
			name:      "40 year mortgage",
			cpr:       0.06,
			numMonths: 480,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prepay := &PrepayInfo{
				PrepayCPR: tc.cpr,
			}

			prepay.ConvertCPRToSMM(tc.numMonths)

			if len(prepay.SMMArr) != tc.numMonths {
				t.Errorf("Expected SMM array length %d, got %d", tc.numMonths, len(prepay.SMMArr))
			}

			expectedSMM := 1 - math.Pow(1-tc.cpr, 1.0/12.0)

			for _, smm := range prepay.SMMArr {
				if math.Abs(smm-expectedSMM) > 0.000001 {
					t.Errorf("Expected SMM %.8f, got %.8f", expectedSMM, smm)
				}
			}
		})
	}
}

func TestConvertCPRToSMM_Formula(t *testing.T) {
	// Test the mathematical relationship: SMM = 1 - (1 - CPR)^(1/12)
	// Also verify that CPR = 1 - (1 - SMM)^12
	prepay := &PrepayInfo{
		PrepayCPR: 0.08,
	}

	prepay.ConvertCPRToSMM(12)

	smm := prepay.SMMArr[0]

	// Reverse calculate CPR from SMM
	calculatedCPR := 1 - math.Pow(1-smm, 12.0)

	if math.Abs(calculatedCPR-prepay.PrepayCPR) > 0.0001 {
		t.Errorf("CPR roundtrip failed: Original CPR %.6f, Calculated CPR %.6f",
			prepay.PrepayCPR, calculatedCPR)
	}
}
