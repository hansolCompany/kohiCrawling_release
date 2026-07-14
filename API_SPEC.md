# KohiCrawling API 명세서

로컬 HTTP 서버 기본 주소: `http://localhost:7061`

모든 API는 **POST** 요청만 허용하며, 요청 본문은 **JSON** (`Content-Type: application/json`) 형식을 사용합니다.

---

## 응답 방식 요약

| URL | HTTP 코드 | 설명 |
|-----|-----------|------|
| `POST /api/kohi/autoLearn` | `200 OK` | 작업 완료 후 성공/실패 JSON 반환 |
| `POST /api/longterm/schedule` | `200 OK` | 작업 완료 후 성공/실패 JSON 반환 |
| `POST /api/longterm/enrollLoad/edu` | `200 OK` | 수강내역 조회 후 JSON 반환 |
| `POST /api/longterm/enrollLoad/longterm` | `200 OK` | 수강내역 조회 후 JSON 반환 |
| `POST /api/longterm/enrollUpload` | `200 OK` | 작업 완료 후 성공/실패 JSON 반환 |

- **크롤링 API** (`autoLearn`, `schedule`, `enrollUpload`): 작업이 끝날 때까지 HTTP 연결을 유지하고, 결과를 JSON으로 반환합니다.
- **수강내역 불러오기** (`enrollLoad/*`): 로그인 → 공단 API 호출 → JSON 변환 후 반환합니다.
- **동시 실행 제한**: KOHI API는 `kohi/browser` 락, 롱텀 API는 `longterm/browser` 락을 사용합니다. 같은 그룹 내 다른 요청이 실행 중이면 `409 Conflict`를 반환합니다.
- **브라우저 세션 재사용**: 작업 완료 후 브라우저를 유지합니다. 다음 요청 시 **동일 계정**이면 재로그인 없이 기존 브라우저를 사용합니다. 계정(또는 롱텀 로그인 유형)이 다르면 브라우저를 닫고 처음부터 다시 시작합니다.

---

## 공통 규칙

### 필드 명명

- JSON 키는 **camelCase** 영문을 사용합니다.
- URL 경로도 **camelCase** 영문을 사용합니다.

### 날짜·시간 형식

| 필드 유형 | 형식 | 예시 |
|-----------|------|------|
| 교육일자 (`educationDate`) | `YYYYMMDD` | `20260707` |
| 과목시작·종료시간 (`courseStartTime`, `courseEndTime`) | `HHmm` (24시간) | `0930`, `1745` |
| 교육년도 (`educationYear`) | `YYYY` | `2026` |

### 요청 검증 오류 (400)

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json; charset=utf-8
```

```json
{
  "status": "error",
  "message": "오류 설명"
}
```

### 이미 실행 중 (409)

```http
HTTP/1.1 409 Conflict
Content-Type: application/json; charset=utf-8
```

```json
{
  "status": "running",
  "message": "해당 크롤링이 이미 실행 중입니다"
}
```

### POST 외 메서드 (405)

```json
{
  "status": "error",
  "message": "POST만 허용됩니다"
}
```

### 크롤링 작업 결과 (200)

작업 실행 중 오류(로그인 실패, 자동화 실패 등):

```http
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
```

```json
{
  "status": "error",
  "message": "실패 사유"
}
```

성공:

```json
{
  "status": "success"
}
```

---

## 롱텀 공통 필드

| 한글 | JSON 키 | 타입 | 필수 | 설명 |
|------|---------|------|------|------|
| 기관기호 | `institutionCode` | string | O | 장기요양기관기호 |
| 인증서명 | `certName` | string | O | 법인인증서 표시명 (포함 일치) |
| 인증서 비밀번호 | `certPassword` | string | O | 인증서 비밀번호 |

### 롱텀 로그인 유형

| API | 로그인 유형 |
|-----|-------------|
| `schedule`, `enrollLoad/edu`, `enrollUpload` | 보수교육기관 |
| `enrollLoad/longterm` | 장기요양기관 |

---

## 롱텀 API

### 보수교육 일정관리

| 항목 | 값 |
|------|-----|
| URL | `POST /api/longterm/schedule` |
| 응답 | `200 OK` (성공/실패 JSON) |
| 구현 상태 | 브라우저 자동화 구현됨 |

롱텀 공단 로그인(보수교육기관) 후 보수교육 일정관리 화면에서 `data` 배열의 일정을 순서대로 등록합니다.

#### 요청 본문

| 한글 | JSON 키 | 타입 | 필수 | 형식·설명 |
|------|---------|------|------|-----------|
| 기관기호 | `institutionCode` | string | O | 롱텀 공통 |
| 인증서명 | `certName` | string | O | 롱텀 공통 |
| 인증서 비밀번호 | `certPassword` | string | O | 롱텀 공통 |
| 일정 목록 | `data` | array | O | 1개 이상 |

`data` 배열 각 항목:

| 한글 | JSON 키 | 타입 | 필수 | 형식·설명 |
|------|---------|------|------|-----------|
| 교육일자 | `educationDate` | string | O | `YYYYMMDD` |
| 표준영역 | `standardArea` | string | O | 콤보박스 **완전 일치** |
| 과목명 | `courseName` | string | O | 콤보박스 **포함 일치** |
| 과목시작시간 | `courseStartTime` | string | O | `HHmm` |
| 과목종료시간 | `courseEndTime` | string | O | `HHmm` |
| 강사명 | `instructorName` | string | O | 강사 선택 모달 **완전 일치** |
| 정원 | `capacity` | number | O | 1 이상 |
| 비고 | `note` | string | X | 빈 문자열 가능 |

#### 요청 예시

```json
{
  "institutionCode": "50652000009",
  "certName": "명성요양보호사 교육원",
  "certPassword": "~MJMS954044",
  "data": [
    {
      "educationDate": "20260715",
      "standardArea": "요양",
      "courseName": "노인인권보호",
      "courseStartTime": "0930",
      "courseEndTime": "1200",
      "instructorName": "홍길동",
      "capacity": 30,
      "note": ""
    }
  ]
}
```

#### 응답

크롤링 **작업 결과 (200)** 형식을 사용합니다.

---

### 보수교육 수강내역 관리 — 수강내역 불러오기 (교육원)

| 항목 | 값 |
|------|-----|
| URL | `POST /api/longterm/enrollLoad/edu` |
| 응답 | `200 OK` (동기 JSON) |
| 구현 상태 | 구현됨 |

법인인증서 로그인(보수교육기관) 후 공단 API `selectRftrObjtrCpetList.do`를 호출하고, 응답 XML의 `ds_result` Dataset을 JSON으로 변환해 반환합니다.

- `educationYear`(`YYYY`)는 공단 조회 조건 `EDU_YYYYQT`로 `{YYYY}1` 형식(1분기)으로 변환됩니다.  
  예: `"2026"` → `"20261"`
- `institutionCode`는 `LTC_ADMIN_SYM`으로 사용됩니다.

#### 요청 본문

| 한글 | JSON 키 | 타입 | 필수 | 형식·설명 |
|------|---------|------|------|-----------|
| 기관기호 | `institutionCode` | string | O | 롱텀 공통 |
| 인증서명 | `certName` | string | O | 롱텀 공통 |
| 인증서 비밀번호 | `certPassword` | string | O | 롱텀 공통 |
| 교육년도 | `educationYear` | string | O | `YYYY` |

#### 성공 응답

```json
{
  "errorCode": "0",
  "errorMessage": "Success",
  "rows": [
    {
      "FNM": "박정자",
      "EDU_CPET_YN": "Y",
      "CHE_HIPIN": "110000053541561",
      "EDU_MTH_NM": "대면",
      "BULTN_YYYY": "2026"
    }
  ]
}
```

| JSON 키 | 설명 |
|---------|------|
| `errorCode` | 공단 XML `<Parameters>`의 `ErrorCode` |
| `errorMessage` | 공단 XML `<Parameters>`의 `ErrorMsg` |
| `rows` | `ds_result` Dataset의 Row 목록 (Col id → 값 객체) |

공단 API 호출은 성공했지만 `errorCode`가 오류값인 경우에도 위 JSON 구조로 반환됩니다.

#### 작업 실패 응답

로그인 실패, 세션 조회 실패, XML 파싱 실패 등:

```json
{
  "status": "error",
  "message": "오류 설명"
}
```

---

### 보수교육 수강내역 관리 — 수강내역 불러오기 (롱텀)

| 항목 | 값 |
|------|-----|
| URL | `POST /api/longterm/enrollLoad/longterm` |
| 응답 | `200 OK` (동기 JSON) |
| 구현 상태 | 구현됨 |

법인인증서 로그인(**장기요양기관**) 후 공단 API `selectRftrObjtrList.do`를 호출하고, 응답 XML을 JSON으로 변환해 반환합니다.

- `educationYear`(`YYYY`)는 `ds_cond.EDU_YYYY`로 사용됩니다.
- `EDU_TGT_QT = "1"`, `CPET_YN = "N"`, `RETR_YN = "N"`, `EXEMP_INCL_YN = "N"` 고정값을 사용합니다.
- `institutionCode`는 `LTC_ADMIN_SYM`으로 사용됩니다.

요청 본문은 `enrollLoad/edu`와 동일합니다.

응답 데이터는 아래 필터를 적용한 뒤 반환합니다.

- `FNM` + `BDAY` 중복 행 제거 (먼저 나온 행 유지)
- `educationYear`와 `BDAY` 연도의 짝/홀 일치
- `ENTCO_DT`가 조회 연도 10월 31일 이후인 행 제외
- `FST_GRANT_DT` 연도가 `educationYear` 또는 `educationYear - 1`인 행 제외

#### 성공 응답

```json
{
  "errorCode": "0",
  "errorMessage": "Success",
  "datasets": {
    "ds_rftrObjtrList": [
      {
        "FNM": "강군자",
        "CHE_HIPIN": "110000523586850",
        "EDU_CPET_YN": "N"
      }
    ]
  }
}
```

| JSON 키 | 설명 |
|---------|------|
| `errorCode` | 공단 XML `<Parameters>`의 `ErrorCode` |
| `errorMessage` | 공단 XML `<Parameters>`의 `ErrorMsg` |
| `datasets` | Dataset `id`별 row 배열 |

#### 작업 실패 응답

`enrollLoad/edu`와 동일합니다.

---

### 보수교육 수강내역 관리 — 수강내역 업로드

| 항목 | 값 |
|------|-----|
| URL | `POST /api/longterm/enrollUpload` |
| 응답 | `200 OK` (성공/실패 JSON) |
| 구현 상태 | 로그인 → 메뉴 이동 → 엑셀 업로드 버튼 클릭 → 바탕화면 파일 선택 (부분 구현) |

바탕화면(`Desktop`, `바탕 화면`, OneDrive 바탕화면)에서 파일명에 **보수교육 수강내역 등록**을 포함한 파일을 찾아 업로드합니다. 여러 파일이 있으면 가장 최근 수정 파일을 사용합니다.

#### 요청 본문

롱텀 공통 필드만 필요합니다 (`excelFileUrl` 없음).

#### 응답

크롤링 **작업 결과 (200)** 형식을 사용합니다.

---

## KOHI API

### 보수교육 자동 학습

| 항목 | 값 |
|------|-----|
| URL | `POST /api/kohi/autoLearn` |
| 응답 | `200 OK` (성공/실패 JSON) |
| 구현 상태 | 브라우저 자동화 구현됨 |

KOHI 로그인 후 `courseName` 배열의 각 강의를 순서대로 **수강신청 → 학습**합니다. 강의명은 제목 **포함 일치**로 검색합니다.

#### 요청 본문

| 한글 | JSON 키 | 타입 | 필수 | 설명 |
|------|---------|------|------|------|
| 요양보호사 db id | `caregiverDbId` | string | O | 요양보호사 DB 식별자 (로그 출력용) |
| 아이디 | `userId` | string | O | KOHI 로그인 아이디 |
| 비밀번호 | `password` | string | O | KOHI 로그인 비밀번호 |
| 강의명 | `courseName` | string[] | O | 수강할 강의명 (1개 이상) |

#### 응답

크롤링 **작업 결과 (200)** 형식을 사용합니다.

---

## URL 요약

| 서비스명 | Method | URL | 응답 |
|----------|--------|-----|------|
| 보수교육 일정관리 | POST | `/api/longterm/schedule` | 200 `{status}` |
| 보수교육 수강내역 불러오기 (교육원) | POST | `/api/longterm/enrollLoad/edu` | 200 `{errorCode, errorMessage, rows}` |
| 보수교육 수강내역 불러오기 (롱텀) | POST | `/api/longterm/enrollLoad/longterm` | 200 `{errorCode, errorMessage, datasets}` |
| 보수교육 수강내역 업로드 | POST | `/api/longterm/enrollUpload` | 200 `{status}` |
| 보수교육 자동 학습 | POST | `/api/kohi/autoLearn` | 200 `{status}` |

---

## 호출 예시 (curl)

```bash
# 보수교육 일정관리
curl -X POST http://localhost:7061/api/longterm/schedule \
  -H "Content-Type: application/json" \
  -d '{"institutionCode":"50652000009","certName":"명성요양보호사 교육원","certPassword":"~MJMS954044","data":[{"educationDate":"20260715","standardArea":"요양","courseName":"노인인권보호","courseStartTime":"0930","courseEndTime":"1200","instructorName":"홍길동","capacity":30,"note":""}]}'

# 보수교육 수강내역 불러오기 — 교육원
curl -X POST http://localhost:7061/api/longterm/enrollLoad/edu \
  -H "Content-Type: application/json" \
  -d '{"institutionCode":"50652000009","certName":"명성요양보호사 교육원","certPassword":"~MJMS954044","educationYear":"2026"}'

# 보수교육 수강내역 불러오기 — 롱텀
curl -X POST http://localhost:7061/api/longterm/enrollLoad/longterm \
  -H "Content-Type: application/json" \
  -d '{"institutionCode":"34615000250","certName":"송이노인복지센터","certPassword":"비밀번호","educationYear":"2026"}'

# 보수교육 수강내역 업로드
curl -X POST http://localhost:7061/api/longterm/enrollUpload \
  -H "Content-Type: application/json" \
  -d '{"institutionCode":"50652000009","certName":"명성요양보호사 교육원","certPassword":"~MJMS954044"}'

# 보수교육 자동 학습
curl -X POST http://localhost:7061/api/kohi/autoLearn \
  -H "Content-Type: application/json" \
  -d '{"caregiverDbId":"cg-1001","userId":"example_user","password":"example_password","courseName":["노인인권 및 학대신고의무자 교육","노인학대 예방 및 대응"]}'
```
