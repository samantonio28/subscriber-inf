package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

func MakeResponse(w http.ResponseWriter, statusCode int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(v)
}

func MonthToInt(m time.Month) int {
	switch m {
	case time.January:
		return 1
	case time.February:
		return 2
	case time.March:
		return 3
	case time.April:
		return 4
	case time.May:
		return 5
	case time.June:
		return 6
	case time.July:
		return 7
	case time.August:
		return 8
	case time.September:
		return 9
	case time.October:
		return 10
	case time.November:
		return 11
	case time.December:
		return 12
	default:
		return -1
	}
}

func ParseMonthYear(s string) (time.Time, error) {
	re := regexp.MustCompile(`^\d{2}-\d{4}$`)
	if !re.MatchString(s) {
		return time.Time{}, fmt.Errorf("not compiles XX-XXXX")
	}

	parts := regexp.MustCompile(`-`).Split(s, -1)
	first, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("not number 1")
	}

	second, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("not number 2")
	}

	t := time.Date(second, time.Month(first), 1, 0, 0, 0, 0, time.UTC)
	return t, nil
}

func DateString(t time.Time) string {
	var s string
	if t.IsZero() {
		return ""
	}
	if m := MonthToInt(t.Month()); m < 10 {
		s = "0" + strconv.Itoa(m)
	} else {
		s = strconv.Itoa(m)
	}
	s += "-" + strconv.Itoa(t.Year())
	return s
}
