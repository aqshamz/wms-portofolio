package master

import "testing"

func TestCatalogDecimal(t *testing.T) {
	for _, tc := range []struct {
		value           string
		positive, valid bool
	}{
		{"99999999999999.999999", true, true}, {"0.000001", true, true}, {" 1.000000 ", true, true},
		{"0", false, true}, {"0", true, false}, {"-1", false, false}, {"1e3", true, false},
		{"100000000000000", true, false}, {"1.1234567", true, false}, {"", false, false},
		{"NaN", false, false}, {"1/2", true, false},
	} {
		t.Run(tc.value, func(t *testing.T) {
			_, err := catalogDecimal(tc.value, tc.positive)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v error=%v", tc.valid, err)
			}
		})
	}
	if !decimalIsOne("1.000000") || decimalIsOne("1.000001") {
		t.Fatal("base conversion equality must be exact")
	}
}
func TestCatalogShelfLife(t *testing.T) {
	shelf, tooLong, negative := 30, 31, -1
	if catalogShelfLife(&shelf, &tooLong) == nil || catalogShelfLife(nil, &negative) == nil {
		t.Fatal("invalid shelf life accepted")
	}
	if catalogShelfLife(&shelf, &shelf) != nil || catalogShelfLife(nil, nil) != nil {
		t.Fatal("valid shelf life rejected")
	}
}
