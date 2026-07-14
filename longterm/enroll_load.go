package longterm

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/playwright-community/playwright-go"
)

const enrollLoadOriginURL = "https://www.longtermcare.or.kr"

const (
	enrollLoadEduBaseURL    = "https://www.longtermcare.or.kr/npbs/r/f/406/selectRftrObjtrCpetList.do"
	enrollLoadEduScreenID   = "nprf406m01"
	enrollLoadLongtermBaseURL = "https://www.longtermcare.or.kr/npbs/r/f/251/selectRftrObjtrList.do"
	enrollLoadLongtermScreenID = "nprf251m01"
)

var enrollLoadUserInfoColumns = []string{
	"YKIHO", "ADMIN_PTTN_CD_NM", "TODAY", "LTC_MGMT_NO", "BRCH_CD", "CRTF_SERIAL", "SERVER_NAME",
	"ADMIN_PTTN_CD", "USER_NM", "ADMIN_SYM", "RESULT", "NM_ID", "YOYANG_KIND_CD", "HIPIN", "PSTN_TYPE",
	"AGE", "USER_ID", "USER_TYPE_NM", "CMMB_TYPE_CD", "BDAY", "CEN_TYPE", "ADMIN_KIND_CD", "LTC_ADMIN_SYM",
	"USER_TYPE", "QLF_TYPE", "SERVER_IP", "BRCH_NM", "USER_IP", "CON_TYPE", "ADMIN_KIND_CD_NM", "USER_AUTH",
	"YOYANG_STAT_CD", "EGGR_NM", "PAYMT_VLT_CLSFC_TYPE_CD", "_ROW_STATUS", "PGM_ID", "BTN_PROC_TYPE",
	"BUSI_LOG_CONTN", "INDI_INFO_INQ_RSN_CD", "INQ_COND_CONTN", "RNE", "OBJTR_HIPIN", "INQ_DATA_CNT",
}

var enrollLoadEduCondColumns = []string{
	"EDU_YYYYQT", "FNM", "LTC_ADMIN_SYM", "EDU_FR_DT", "EDU_TO_DT", "CHE_HIPIN", "CPET_SEQ_NO", "REG_FR_DT", "REG_TO_DT",
}

var enrollLoadLongtermCondColumns = []string{
	"FNM", "CRNT_PAGE_CNT", "EDU_YYYY", "EDU_TGT_QT", "CPET_YN", "FIRST_INQ_CNT", "PAGE_INQ_CNT",
	"LTC_ADMIN_SYM", "RETR_YN", "EXEMP_INCL_YN",
}

type EnrollLoadResult struct {
	Response    *EnrollLoadJSONResponse
	EduResponse *EnrollLoadEduJSONResponse
}

func buildEnrollLoadEduCond(institutionCode, educationYear string) map[string]string {
	return map[string]string{
		"EDU_YYYYQT":    educationYear + "1",
		"LTC_ADMIN_SYM": institutionCode,
	}
}

func buildEnrollLoadLongtermCond(institutionCode, educationYear string) map[string]string {
	return map[string]string{
		"CRNT_PAGE_CNT": "1",
		"EDU_YYYY":      educationYear,
		"EDU_TGT_QT":    "1",
		"CPET_YN":       "N",
		"FIRST_INQ_CNT": "100",
		"PAGE_INQ_CNT":  "50",
		"LTC_ADMIN_SYM": institutionCode,
		"RETR_YN":       "N",
		"EXEMP_INCL_YN": "N",
	}
}

func buildEnrollLoadRequestXML(session *sessionData, screenID string, condColumns []string, cond map[string]string) string {
	userInfo := prepareUserInfoForEnrollLoad(session.UserInfo, screenID)

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString(`<Root xmlns="http://www.nexacroplatform.com/platform/dataset">`)
	writeParameterSection(&b, session.Parameters)
	writeDataset(&b, "gds_userInfo", enrollLoadUserInfoColumns, userInfo)
	writeDataset(&b, "ds_cond", condColumns, cond)
	b.WriteString(`</Root>`)
	return b.String()
}

func writeParameterSection(b *strings.Builder, params map[string]string) {
	b.WriteString(`<Parameters>`)
	writeParameter(b, "JSESSIONID", params["JSESSIONID"])
	writeParameter(b, "WMONID", params["WMONID"])
	writeParameter(b, "nxKey", params["nxKey"])
	writeParameter(b, "ahnlabFlag", params["ahnlabFlag"])
	writeParameter(b, "NetFunnel_ID", params["NetFunnel_ID"])
	b.WriteString(`</Parameters>`)
}

func writeParameter(b *strings.Builder, id, value string) {
	if strings.TrimSpace(value) == "" {
		b.WriteString(`<Parameter id="`)
		b.WriteString(xmlEscape(id))
		b.WriteString(`" />`)
		return
	}
	b.WriteString(`<Parameter id="`)
	b.WriteString(xmlEscape(id))
	b.WriteString(`">`)
	b.WriteString(xmlEscape(value))
	b.WriteString(`</Parameter>`)
}

func writeDataset(b *strings.Builder, datasetID string, columns []string, row map[string]string) {
	b.WriteString(`<Dataset id="`)
	b.WriteString(xmlEscape(datasetID))
	b.WriteString(`"><ColumnInfo>`)
	for _, column := range columns {
		b.WriteString(`<Column id="`)
		b.WriteString(xmlEscape(column))
		b.WriteString(`" type="STRING" size="256" />`)
	}
	b.WriteString(`</ColumnInfo><Rows><Row>`)
	for _, column := range columns {
		writeCol(b, column, row[column])
	}
	b.WriteString(`</Row></Rows></Dataset>`)
}

func writeCol(b *strings.Builder, id, value string) {
	if strings.TrimSpace(value) == "" {
		b.WriteString(`<Col id="`)
		b.WriteString(xmlEscape(id))
		b.WriteString(`" />`)
		return
	}
	b.WriteString(`<Col id="`)
	b.WriteString(xmlEscape(id))
	b.WriteString(`">`)
	b.WriteString(xmlEscape(value))
	b.WriteString(`</Col>`)
}

func xmlEscape(value string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
	)
	return replacer.Replace(value)
}

func enrollLoadRequestURL(baseURL, screenID, userIP string) string {
	query := url.Values{}
	query.Set("_SCRN_ID", screenID)
	query.Set("_USER_IP", userIP)
	query.Set("charset", "UTF-8")
	return baseURL + "?" + query.Encode()
}

func postEnrollLoadRequestEdu(page playwright.Page, session *sessionData, requestURL, requestBody, label string) (*EnrollLoadResult, error) {
	body, err := postEnrollLoadRequestBody(page, session, requestURL, requestBody, label)
	if err != nil {
		return nil, err
	}

	parsed, err := parseEnrollLoadEduResponseXML(body)
	if err != nil {
		return nil, fmt.Errorf("%s 응답 파싱 실패: %w", label, err)
	}

	return &EnrollLoadResult{EduResponse: parsed}, nil
}

func postEnrollLoadRequest(page playwright.Page, session *sessionData, requestURL, requestBody, label, educationYear string) (*EnrollLoadResult, error) {
	body, err := postEnrollLoadRequestBody(page, session, requestURL, requestBody, label)
	if err != nil {
		return nil, err
	}

	parsed, err := parseNexacroResponseXML(body)
	if err != nil {
		return nil, fmt.Errorf("%s 응답 파싱 실패: %w", label, err)
	}

	filterEnrollLoadDatasetsByEducationYear(parsed.Datasets, educationYear)

	return &EnrollLoadResult{Response: parsed}, nil
}

func postEnrollLoadRequestBody(page playwright.Page, session *sessionData, requestURL, requestBody, label string) ([]byte, error) {
	fmt.Printf("  %s 요청: %s\n", label, requestURL)

	resp, err := page.Context().Request().Post(requestURL, playwright.APIRequestContextPostOptions{
		Data: requestBody,
		Headers: map[string]string{
			"Accept":           "application/xml, text/xml, */*",
			"Content-Type":     "text/xml",
			"X-Requested-With": "XMLHttpRequest",
			"Origin":           enrollLoadOriginURL,
			"Referer":          session.Referer,
			"Cache-Control":    "no-cache, no-store",
			"Pragma":           "no-cache",
		},
		FailOnStatusCode: playwright.Bool(false),
	})
	if err != nil {
		return nil, fmt.Errorf("%s 요청 실패: %w", label, err)
	}
	defer resp.Dispose()

	body, err := resp.Body()
	if err != nil {
		return nil, fmt.Errorf("%s 응답 읽기 실패: %w", label, err)
	}

	fmt.Printf("  %s 응답: HTTP %d (%d bytes)\n", label, resp.Status(), len(body))

	return body, nil
}

func selectRftrObjtrCpetList(page playwright.Page, institutionCode, educationYear string) (*EnrollLoadResult, error) {
	session, err := extractSessionData(page)
	if err != nil {
		return nil, err
	}

	requestURL := enrollLoadRequestURL(enrollLoadEduBaseURL, enrollLoadEduScreenID, session.UserIP)
	requestBody := buildEnrollLoadRequestXML(
		session,
		enrollLoadEduScreenID,
		enrollLoadEduCondColumns,
		buildEnrollLoadEduCond(institutionCode, educationYear),
	)

	return postEnrollLoadRequestEdu(page, session, requestURL, requestBody, "수강내역 조회(edu)")
}

func selectRftrObjtrList(page playwright.Page, institutionCode, educationYear string) (*EnrollLoadResult, error) {
	session, err := extractSessionData(page)
	if err != nil {
		return nil, err
	}

	requestURL := enrollLoadRequestURL(enrollLoadLongtermBaseURL, enrollLoadLongtermScreenID, session.UserIP)
	requestBody := buildEnrollLoadRequestXML(
		session,
		enrollLoadLongtermScreenID,
		enrollLoadLongtermCondColumns,
		buildEnrollLoadLongtermCond(institutionCode, educationYear),
	)

	return postEnrollLoadRequest(page, session, requestURL, requestBody, "수강내역 조회(longterm)", educationYear)
}
