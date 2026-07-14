package longterm

import (
	"fmt"
	"time"
)

const educationDateLayout = "20060102"

func parseEducationDate(value string) (time.Time, error) {
	t, err := time.Parse(educationDateLayout, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("educationDate(%s) 파싱 실패: %w", value, err)
	}
	return t, nil
}

func formatEducationDate(t time.Time) string {
	return t.Format(educationDateLayout)
}

func educationDateMinusOneMonth(value string) (string, error) {
	t, err := parseEducationDate(value)
	if err != nil {
		return "", err
	}
	return formatEducationDate(t.AddDate(0, -1, 0)), nil
}

func educationDatePlusOneMonth(value string) (string, error) {
	t, err := parseEducationDate(value)
	if err != nil {
		return "", err
	}
	return formatEducationDate(t.AddDate(0, 1, 0)), nil
}
