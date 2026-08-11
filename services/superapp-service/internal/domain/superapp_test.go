package domain

import "testing"

func TestSemverCompatible(t *testing.T) {
	if !SemverCompatible("2.1.0", "2.0.0") {
		t.Fatal()
	}
	if SemverCompatible("1.9.0", "2.0.0") {
		t.Fatal()
	}
}

func TestValidatePermissions(t *testing.T) {
	if err := ValidatePermissions([]string{PermOrders, PermSearch}); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePermissions([]string{"root"}); err == nil {
		t.Fatal()
	}
}

func TestUpdateRatingAvg(t *testing.T) {
	avg, n := UpdateRatingAvg(4, 1, 5)
	if n != 2 || avg != 4.5 {
		t.Fatal(avg, n)
	}
}
