package longterm

import "testing"

func TestParseNexacroResponseXML(t *testing.T) {
	xmlBody := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Root xmlns="http://www.nexacroplatform.com/platform/dataset">
	<Parameters>
		<Parameter id="ErrorCode">0</Parameter>
		<Parameter id="ErrorMsg">SUCCESS</Parameter>
	</Parameters>
	<Dataset id="ds_list">
		<ColumnInfo>
			<Column id="FNM" type="STRING" size="256" />
		</ColumnInfo>
		<Rows>
			<Row>
				<Col id="FNM">홍길동</Col>
				<Col id="CHE_HIPIN" />
			</Row>
			<Row>
				<Col id="FNM">김철수</Col>
				<Col id="CHE_HIPIN">110000053541561</Col>
			</Row>
		</Rows>
	</Dataset>
</Root>`)

	parsed, err := parseNexacroResponseXML(xmlBody)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if parsed.ErrorCode != "0" {
		t.Fatalf("errorCode = %q", parsed.ErrorCode)
	}
	if parsed.ErrorMessage != "SUCCESS" {
		t.Fatalf("errorMessage = %q", parsed.ErrorMessage)
	}

	rows := parsed.Datasets["ds_list"]
	if len(rows) != 2 {
		t.Fatalf("rows len = %d", len(rows))
	}
	if rows[0]["FNM"] != "홍길동" {
		t.Fatalf("first row FNM = %q", rows[0]["FNM"])
	}
	if rows[0]["CHE_HIPIN"] != "" {
		t.Fatalf("first row CHE_HIPIN = %q", rows[0]["CHE_HIPIN"])
	}
}

func TestParseEnrollLoadEduResponseXML(t *testing.T) {
	xmlBody := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Root xmlns="http://www.nexacroplatform.com/platform/dataset">
	<Parameters>
		<Parameter id="ErrorCode" type="int">0</Parameter>
		<Parameter id="ErrorMsg" type="string">Success</Parameter>
	</Parameters>
	<Dataset id="ds_result">
		<Rows>
			<Row>
				<Col id="FNM">박정자</Col>
				<Col id="EDU_CPET_YN">Y</Col>
				<Col id="CHE_HIPIN">110000053541561</Col>
			</Row>
		</Rows>
	</Dataset>
</Root>`)

	parsed, err := parseEnrollLoadEduResponseXML(xmlBody)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if parsed.ErrorCode != "0" {
		t.Fatalf("errorCode = %q", parsed.ErrorCode)
	}
	if parsed.ErrorMessage != "Success" {
		t.Fatalf("errorMessage = %q", parsed.ErrorMessage)
	}
	if len(parsed.Rows) != 1 {
		t.Fatalf("rows len = %d", len(parsed.Rows))
	}
	if parsed.Rows[0]["FNM"] != "박정자" {
		t.Fatalf("FNM = %q", parsed.Rows[0]["FNM"])
	}
}

func TestDedupeRowsByFNMAndBDAY(t *testing.T) {
	xmlBody := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Root xmlns="http://www.nexacroplatform.com/platform/dataset">
	<Parameters>
		<Parameter id="ErrorCode">0</Parameter>
		<Parameter id="ErrorMsg">SUCCESS</Parameter>
	</Parameters>
	<Dataset id="ds_rftrObjtrList">
		<Rows>
			<Row>
				<Col id="FNM">홍길동</Col>
				<Col id="BDAY">19980305</Col>
				<Col id="CHE_HIPIN">111</Col>
				<Col id="ENTCO_DT">20250301</Col>
			</Row>
			<Row>
				<Col id="FNM">홍길동</Col>
				<Col id="BDAY">19980305</Col>
				<Col id="CHE_HIPIN">222</Col>
				<Col id="ENTCO_DT">20250301</Col>
			</Row>
			<Row>
				<Col id="FNM">김철수</Col>
				<Col id="BDAY">19980305</Col>
				<Col id="ENTCO_DT">20250301</Col>
			</Row>
			<Row>
				<Col id="FNM">홍길동</Col>
				<Col id="BDAY">19990101</Col>
				<Col id="ENTCO_DT">20250301</Col>
			</Row>
		</Rows>
	</Dataset>
</Root>`)

	parsed, err := parseNexacroResponseXML(xmlBody)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	rows := parsed.Datasets["ds_rftrObjtrList"]
	if len(rows) != 3 {
		t.Fatalf("rows len = %d, want 3", len(rows))
	}
	if rows[0]["CHE_HIPIN"] != "111" {
		t.Fatalf("first duplicate row kept CHE_HIPIN = %q", rows[0]["CHE_HIPIN"])
	}
	if rows[1]["FNM"] != "김철수" {
		t.Fatalf("second row FNM = %q", rows[1]["FNM"])
	}
	if rows[2]["BDAY"] != "19990101" {
		t.Fatalf("third row BDAY = %q", rows[2]["BDAY"])
	}
}

func TestFilterRowsByEducationYearCriteria(t *testing.T) {
	rows := []map[string]string{
		{"FNM": "짝수-짝수-10월", "BDAY": "19980305", "ENTCO_DT": "20251001", "FST_GRANT_DT": "20130826"},
		{"FNM": "짝수-홀수", "BDAY": "19990101", "ENTCO_DT": "20250301", "FST_GRANT_DT": "20130826"},
		{"FNM": "11월입사", "BDAY": "19980305", "ENTCO_DT": "20251101", "FST_GRANT_DT": "20130826"},
		{"FNM": "12월입사", "BDAY": "19980305", "ENTCO_DT": "20251201", "FST_GRANT_DT": "20130826"},
		{"FNM": "조회연도10월31일", "BDAY": "19980305", "ENTCO_DT": "20261031", "FST_GRANT_DT": "20130826"},
		{"FNM": "조회연도11월1일", "BDAY": "19980305", "ENTCO_DT": "20261101", "FST_GRANT_DT": "20130826"},
		{"FNM": "조회연도이후", "BDAY": "19980305", "ENTCO_DT": "20270301", "FST_GRANT_DT": "20130826"},
		{"FNM": "이전연도11월", "BDAY": "19980305", "ENTCO_DT": "20251101", "FST_GRANT_DT": "20130826"},
		{"FNM": "ENTCO없음", "BDAY": "19980305", "FST_GRANT_DT": "20130826"},
		{"FNM": "최초자격-올해", "BDAY": "19980305", "ENTCO_DT": "20250301", "FST_GRANT_DT": "20260101"},
		{"FNM": "최초자격-작년", "BDAY": "19980305", "ENTCO_DT": "20250301", "FST_GRANT_DT": "20250101"},
	}

	filtered := filterRowsByEducationYearCriteria(rows, 2026)
	if len(filtered) != 5 {
		t.Fatalf("filtered len = %d, want 5", len(filtered))
	}

	wantNames := []string{
		"짝수-짝수-10월",
		"11월입사",
		"12월입사",
		"조회연도10월31일",
		"이전연도11월",
	}
	for i, name := range wantNames {
		if filtered[i]["FNM"] != name {
			t.Fatalf("filtered[%d] FNM = %q, want %q", i, filtered[i]["FNM"], name)
		}
	}
}

func TestEntcoDateOnOrBeforeEducationYearOct31(t *testing.T) {
	cases := []struct {
		entcoDT string
		year    int
		want    bool
	}{
		{"20251001", 2026, true},
		{"20251101", 2026, true},
		{"20261031", 2026, true},
		{"20261101", 2026, false},
		{"20270301", 2026, false},
	}

	for _, tc := range cases {
		got := entcoDateOnOrBeforeEducationYearOct31(tc.entcoDT, tc.year)
		if got != tc.want {
			t.Fatalf("ENTCO_DT %s year %d = %v, want %v", tc.entcoDT, tc.year, got, tc.want)
		}
	}
}
