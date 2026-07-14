package longterm

import (
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

type sessionData struct {
	Parameters map[string]string
	UserInfo   map[string]string
	UserIP     string
	Referer    string
}

func extractSessionData(page playwright.Page) (*sessionData, error) {
	cookies, err := page.Context().Cookies()
	if err != nil {
		return nil, fmt.Errorf("세션 쿠키 조회 실패: %w", err)
	}

	params := map[string]string{
		"NetFunnel_ID": "",
	}
	for _, cookie := range cookies {
		switch cookie.Name {
		case "JSESSIONID", "WMONID", "nxKey", "ahnlabFlag", "NetFunnel_ID":
			params[cookie.Name] = cookie.Value
		}
	}
	if params["JSESSIONID"] == "" {
		return nil, fmt.Errorf("JSESSIONID 쿠키를 찾을 수 없음")
	}

	result, err := page.Evaluate(`() => {
		const emptyRow = () => ({});
		try {
			if (typeof nexacro === 'undefined' || !nexacro.getApplication) {
				return { row: null };
			}

			const app = nexacro.getApplication();
			let ds = app.gds_userInfo;
			if (!ds && app.all) {
				for (const key of Object.keys(app.all)) {
					const obj = app.all[key];
					if (obj && (key === 'gds_userInfo' || obj.name === 'gds_userInfo' || obj.id === 'gds_userInfo')) {
						ds = obj;
						break;
					}
				}
			}
			if (!ds || !ds.rowcount) {
				return { row: null };
			}

			const row = emptyRow();
			const colCount = ds.colcount || (ds.colinfos ? ds.colinfos.length : 0);
			for (let i = 0; i < colCount; i++) {
				let colId = ds.colinfos?.[i]?.id;
				if (!colId && typeof ds.getColID === 'function') {
					colId = ds.getColID(i);
				}
				if (!colId) {
					continue;
				}
				const value = ds.getColumn(0, colId);
				row[colId] = value == null ? '' : String(value);
			}
			return { row };
		} catch (error) {
			return { row: null, error: String(error) };
		}
	}`)
	if err != nil {
		return nil, fmt.Errorf("사용자 정보 조회 실패: %w", err)
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("사용자 정보 결과 파싱 실패")
	}

	userInfo, err := mapFromEvaluateRow(data["row"])
	if err != nil {
		return nil, err
	}
	if len(userInfo) == 0 {
		return nil, fmt.Errorf("nexacro gds_userInfo를 찾을 수 없음")
	}

	userIP := strings.TrimSpace(userInfo["USER_IP"])
	if userIP == "" {
		userIP = "127.0.0.1"
	}

	return &sessionData{
		Parameters: params,
		UserInfo:   userInfo,
		UserIP:     userIP,
		Referer:    page.URL(),
	}, nil
}

func mapFromEvaluateRow(value interface{}) (map[string]string, error) {
	if value == nil {
		return nil, nil
	}

	raw, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("사용자 정보 row 파싱 실패")
	}

	row := make(map[string]string, len(raw))
	for key, item := range raw {
		switch v := item.(type) {
		case string:
			row[key] = v
		case float64, float32, int, int32, int64, bool:
			row[key] = fmt.Sprint(v)
		default:
			if item == nil {
				row[key] = ""
			} else {
				row[key] = fmt.Sprint(item)
			}
		}
	}
	return row, nil
}

func prepareUserInfoForEnrollLoad(userInfo map[string]string, screenID string) map[string]string {
	row := make(map[string]string, len(userInfo))
	for key, value := range userInfo {
		row[key] = value
	}

	row["TODAY"] = time.Now().Format("20060102")
	row["PGM_ID"] = screenID
	row["BTN_PROC_TYPE"] = "R"
	row["_ROW_STATUS"] = "1"
	row["BUSI_LOG_CONTN"] = ""
	row["INDI_INFO_INQ_RSN_CD"] = ""
	row["INQ_COND_CONTN"] = ""
	row["RNE"] = ""
	row["OBJTR_HIPIN"] = ""
	row["INQ_DATA_CNT"] = ""

	return row
}
