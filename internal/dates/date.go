package dates

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	jalaali "github.com/jalaali/go-jalaali"
)

// JalaliToGregorian converts YYYY/MM/DD into YYYY-MM-DD.
func JalaliToGregorian(value string) (string, error) {
	parts := strings.Split(strings.TrimSpace(value), "/")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid Jalali date format; expected YYYY/MM/DD")
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", fmt.Errorf("invalid Jalali year")
	}

	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid Jalali month")
	}

	day, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", fmt.Errorf("invalid Jalali day")
	}

	gy, gm, gd, err := jalaali.ToGregorian(
		year,
		jalaali.Month(month),
		day,
	)
	if err != nil {
		return "", fmt.Errorf("invalid Jalali date: %w", err)
	}

	return fmt.Sprintf("%04d-%02d-%02d", gy, gm, gd), nil
}

// GregorianToJalali converts YYYY-MM-DD into YYYY/MM/DD.
func GregorianToJalali(value string) (string, error) {
	date, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil {
		return "", fmt.Errorf("invalid Gregorian date: %w", err)
	}

	jy, jm, jd, err := jalaali.ToJalaali(
		date.Year(),
		date.Month(),
		date.Day(),
	)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%04d/%02d/%02d", jy, jm, jd), nil
}
