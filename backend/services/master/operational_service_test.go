package master

import "testing"

func TestOperationalValidation(t *testing.T) {
	for _, value := range []string{"0", "0.0000", "25.1234", "100.0000"} {
		if err := validateEmptyPercent(&value); err != nil {
			t.Errorf("%s: %v", value, err)
		}
	}
	for _, value := range []string{"-1", "100.0001", "1.12345", "1e2", "", "NaN"} {
		if validateEmptyPercent(&value) == nil {
			t.Errorf("accepted %q", value)
		}
	}
	if validateEmptyPercent(nil) != nil {
		t.Fatal("null percentage rejected")
	}
	for _, date := range []string{"2026-02-30", "03-09-2026", "0000-01-01", "2026-9-3"} {
		if _, err := parseOperationalDate(date); err == nil {
			t.Errorf("accepted date %q", date)
		}
	}
	if statusFlags(true, true, false, true) == nil || statusFlags(false, false, true, true) == nil || statusFlags(true, false, false, false) == nil {
		t.Fatal("invalid status flags accepted")
	}
	if sequenceLimit(3) != 999 || sequenceLimit(18) != 999999999999999999 {
		t.Fatal("incorrect sequence bounds")
	}
}
