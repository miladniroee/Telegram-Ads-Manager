package dates

import "testing"

func TestJalaliToGregorian(t *testing.T) {
	result, err := JalaliToGregorian("1403/01/01")
	if err != nil {
		t.Fatal(err)
	}

	expected := "2024-03-20"

	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestGregorianToJalali(t *testing.T) {
	result, err := GregorianToJalali("2024-03-20")
	if err != nil {
		t.Fatal(err)
	}

	expected := "1403/01/01"

	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestInvalidJalaliDate(t *testing.T) {
	_, err := JalaliToGregorian("1403/13/01")

	if err == nil {
		t.Error("expected an error for an invalid Jalali date")
	}
}

func TestInvalidGregorianDate(t *testing.T) {
	_, err := GregorianToJalali("2024-02-30")

	if err == nil {
		t.Error("expected an error for an invalid Gregorian date")
	}
}
