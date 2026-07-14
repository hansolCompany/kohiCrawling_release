package api

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	reDateYYYYMMDD = regexp.MustCompile(`^\d{8}$`)
	reTimeHHmm     = regexp.MustCompile(`^\d{4}$`)
	reYearYYYY     = regexp.MustCompile(`^\d{4}$`)
)

func requireNonEmpty(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s가 필요합니다", field)
	}
	return nil
}

func validateDateYYYYMMDD(field, value string) error {
	if err := requireNonEmpty(field, value); err != nil {
		return err
	}
	if !reDateYYYYMMDD.MatchString(value) {
		return fmt.Errorf("%s는 YYYYMMDD 형식이어야 합니다", field)
	}
	return nil
}

func validateTimeHHmm(field, value string) error {
	if err := requireNonEmpty(field, value); err != nil {
		return err
	}
	if !reTimeHHmm.MatchString(value) {
		return fmt.Errorf("%s는 HHmm 형식이어야 합니다", field)
	}
	hour := value[:2]
	minute := value[2:]
	if hour > "23" || minute > "59" {
		return fmt.Errorf("%s는 유효한 24시간 HHmm 형식이어야 합니다", field)
	}
	return nil
}

func validateYearYYYY(field, value string) error {
	if err := requireNonEmpty(field, value); err != nil {
		return err
	}
	if !reYearYYYY.MatchString(value) {
		return fmt.Errorf("%s는 YYYY 형식이어야 합니다", field)
	}
	return nil
}

func validateURL(field, value string) error {
	if err := requireNonEmpty(field, value); err != nil {
		return err
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("%s는 유효한 URL이어야 합니다", field)
	}
	return nil
}

func validateLongtermAuth(institutionCode, certName, certPassword string) error {
	if err := requireNonEmpty("institutionCode", institutionCode); err != nil {
		return err
	}
	if err := requireNonEmpty("certName", certName); err != nil {
		return err
	}
	if err := requireNonEmpty("certPassword", certPassword); err != nil {
		return err
	}
	return nil
}
