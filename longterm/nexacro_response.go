package longterm

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var xmlnsStripper = regexp.MustCompile(`\s+xmlns="[^"]*"`)

type EnrollLoadJSONResponse struct {
	ErrorCode    string                         `json:"errorCode"`
	ErrorMessage string                         `json:"errorMessage"`
	Datasets     map[string][]map[string]string `json:"datasets"`
}

type EnrollLoadEduJSONResponse struct {
	ErrorCode    string              `json:"errorCode"`
	ErrorMessage string              `json:"errorMessage"`
	Rows         []map[string]string `json:"rows"`
}

const enrollLoadEduDatasetID = "ds_result"

type nexacroXMLCol struct {
	ID    string `xml:"id,attr"`
	Value string `xml:",chardata"`
}

type nexacroXMLRow struct {
	Cols []nexacroXMLCol `xml:"Col"`
}

type nexacroXMLRows struct {
	Row []nexacroXMLRow `xml:"Row"`
}

type nexacroXMLDataset struct {
	ID   string          `xml:"id,attr"`
	Rows nexacroXMLRows  `xml:"Rows"`
}

type nexacroXMLParameter struct {
	ID    string `xml:"id,attr"`
	Value string `xml:",chardata"`
}

type nexacroXMLRoot struct {
	Parameters struct {
		Parameter []nexacroXMLParameter `xml:"Parameter"`
	} `xml:"Parameters"`
	Datasets []nexacroXMLDataset `xml:"Dataset"`
}

func parseNexacroResponseXML(body []byte) (*EnrollLoadJSONResponse, error) {
	xmlText := xmlnsStripper.ReplaceAllString(string(body), "")

	var root nexacroXMLRoot
	if err := xml.Unmarshal([]byte(xmlText), &root); err != nil {
		return nil, fmt.Errorf("XML 파싱 실패: %w", err)
	}

	errorCode, errorMessage := extractNexacroError(root.Parameters.Parameter)
	datasets := extractNexacroDatasets(root.Datasets)
	dedupeDatasetsByFNMAndBDAY(datasets)

	return &EnrollLoadJSONResponse{
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
		Datasets:     datasets,
	}, nil
}

func parseEnrollLoadEduResponseXML(body []byte) (*EnrollLoadEduJSONResponse, error) {
	xmlText := xmlnsStripper.ReplaceAllString(string(body), "")

	var root nexacroXMLRoot
	if err := xml.Unmarshal([]byte(xmlText), &root); err != nil {
		return nil, fmt.Errorf("XML 파싱 실패: %w", err)
	}

	errorCode, errorMessage := extractNexacroError(root.Parameters.Parameter)
	datasets := extractNexacroDatasets(root.Datasets)

	rows := datasets[enrollLoadEduDatasetID]
	if rows == nil {
		rows = []map[string]string{}
	}

	return &EnrollLoadEduJSONResponse{
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
		Rows:         rows,
	}, nil
}

func extractNexacroError(params []nexacroXMLParameter) (string, string) {
	var errorCode, errorMessage string

	for _, param := range params {
		switch normalizeNexacroKey(param.ID) {
		case "errorcode":
			errorCode = strings.TrimSpace(param.Value)
		case "errormsg", "errormessage":
			errorMessage = strings.TrimSpace(param.Value)
		}
	}

	return errorCode, errorMessage
}

func normalizeNexacroKey(key string) string {
	key = strings.TrimSpace(key)
	key = strings.ReplaceAll(key, "_", "")
	key = strings.ReplaceAll(key, "-", "")
	return strings.ToLower(key)
}

func extractNexacroDatasets(datasets []nexacroXMLDataset) map[string][]map[string]string {
	result := make(map[string][]map[string]string)

	for _, dataset := range datasets {
		datasetID := strings.TrimSpace(dataset.ID)
		if datasetID == "" {
			continue
		}

		rows := make([]map[string]string, 0, len(dataset.Rows.Row))
		for _, row := range dataset.Rows.Row {
			item := make(map[string]string, len(row.Cols))
			for _, col := range row.Cols {
				colID := strings.TrimSpace(col.ID)
				if colID == "" {
					continue
				}
				item[colID] = strings.TrimSpace(col.Value)
			}
			rows = append(rows, item)
		}

		result[datasetID] = rows
	}

	return result
}

func dedupeDatasetsByFNMAndBDAY(datasets map[string][]map[string]string) {
	for datasetID, rows := range datasets {
		if !datasetHasFNMColumn(rows) {
			continue
		}
		datasets[datasetID] = dedupeRowsByFNMAndBDAY(rows)
	}
}

func datasetHasFNMColumn(rows []map[string]string) bool {
	for _, row := range rows {
		if _, ok := row["FNM"]; ok {
			return true
		}
	}
	return false
}

func dedupeRowsByFNMAndBDAY(rows []map[string]string) []map[string]string {
	seen := make(map[string]struct{}, len(rows))
	result := make([]map[string]string, 0, len(rows))

	for _, row := range rows {
		key := row["FNM"] + "\x00" + row["BDAY"]
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, row)
	}

	return result
}

func filterEnrollLoadDatasetsByEducationYear(datasets map[string][]map[string]string, educationYear string) {
	year, ok := parseYearYYYY(educationYear)
	if !ok {
		return
	}

	for datasetID, rows := range datasets {
		if !datasetHasFNMColumn(rows) {
			continue
		}
		datasets[datasetID] = filterRowsByEducationYearCriteria(rows, year)
	}
}

func filterRowsByEducationYearCriteria(rows []map[string]string, educationYear int) []map[string]string {
	result := make([]map[string]string, 0, len(rows))
	for _, row := range rows {
		if enrollLoadRowMatchesEducationYearCriteria(row, educationYear) {
			result = append(result, row)
		}
	}
	return result
}

func enrollLoadRowMatchesEducationYearCriteria(row map[string]string, educationYear int) bool {
	bdayYear, ok := yearFromYYYYMMDD(row["BDAY"])
	if !ok {
		return false
	}
	if (educationYear%2) != (bdayYear%2) {
		return false
	}

	if !entcoDateOnOrBeforeEducationYearOct31(row["ENTCO_DT"], educationYear) {
		return false
	}

	if fstGrantYear, ok := yearFromYYYYMMDD(row["FST_GRANT_DT"]); ok {
		if fstGrantYear == educationYear || fstGrantYear == educationYear-1 {
			return false
		}
	}

	return true
}

func entcoDateOnOrBeforeEducationYearOct31(entcoDT string, educationYear int) bool {
	entcoDT = strings.TrimSpace(entcoDT)
	if len(entcoDT) < 8 {
		return false
	}
	datePart := entcoDT[:8]
	if _, err := strconv.Atoi(datePart); err != nil {
		return false
	}

	cutoff := fmt.Sprintf("%d1031", educationYear)
	return datePart <= cutoff
}

func parseYearYYYY(value string) (int, bool) {
	value = strings.TrimSpace(value)
	if len(value) != 4 {
		return 0, false
	}
	year, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}
	return year, true
}

func yearFromYYYYMMDD(value string) (int, bool) {
	value = strings.TrimSpace(value)
	if len(value) < 4 {
		return 0, false
	}
	return parseYearYYYY(value[:4])
}
