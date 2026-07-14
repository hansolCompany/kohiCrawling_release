package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"kohiCrawling/kohi"
	"kohiCrawling/longterm"
)

const Addr = ":7061"

const (
	taskKohiBrowser     = "kohi/browser"
	taskLongtermBrowser = "longterm/browser"
)

type Server struct {
	tasks *taskRunner
}

func NewServer() *Server {
	return &Server{
		tasks: newTaskRunner(),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/kohi/autoLearn", s.handleKohiAutoLearn)
	mux.HandleFunc("/api/longterm/schedule", s.handleLongtermSchedule)
	mux.HandleFunc("/api/longterm/enrollLoad/edu", s.handleLongtermEnrollLoadEdu)
	mux.HandleFunc("/api/longterm/enrollLoad/longterm", s.handleLongtermEnrollLoadLongterm)
	mux.HandleFunc("/api/longterm/enrollUpload", s.handleLongtermEnrollUpload)
	return mux
}

func (s *Server) Run() error {
	fmt.Printf("HTTP 서버 시작: http://localhost%s\n", Addr)
	fmt.Println("  POST /api/kohi/autoLearn")
	fmt.Println("  POST /api/longterm/schedule")
	fmt.Println("  POST /api/longterm/enrollLoad/edu")
	fmt.Println("  POST /api/longterm/enrollLoad/longterm")
	fmt.Println("  POST /api/longterm/enrollUpload")
	return http.ListenAndServe(Addr, s.Handler())
}

func (s *Server) handleKohiAutoLearn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}

	var req autoLearnRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err.Error())
		return
	}
	if err := req.validate(); err != nil {
		writeError(w, err.Error())
		return
	}
	if !s.tasks.tryStart(taskKohiBrowser) {
		writeRunning(w)
		return
	}
	defer s.tasks.finish(taskKohiBrowser)

	opts := kohi.AutoLearnOptions{
		CaregiverDbID: req.CaregiverDbID,
		UserID:        req.UserID,
		Password:      req.Password,
		CourseNames:   req.CourseName,
	}

	if err := kohi.RunAutoLearn(opts); err != nil {
		fmt.Printf("kohi/autoLearn 오류: %v\n", err)
		writeTaskError(w, err.Error())
		return
	}

	writeTaskSuccess(w)
}

func (s *Server) handleLongtermSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}

	var req scheduleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err.Error())
		return
	}
	if err := req.validate(); err != nil {
		writeError(w, err.Error())
		return
	}
	if !s.tasks.tryStart(taskLongtermBrowser) {
		writeRunning(w)
		return
	}
	defer s.tasks.finish(taskLongtermBrowser)

	auth := toLongtermAuth(req.longtermAuthRequest)
	payload := toLongtermScheduleItems(req.Data)

	if err := longterm.RunSchedule(auth, payload); err != nil {
		fmt.Printf("longterm/schedule 오류: %v\n", err)
		writeTaskError(w, err.Error())
		return
	}

	writeTaskSuccess(w)
}

func (s *Server) handleLongtermEnrollLoadEdu(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}

	req, auth, ok := s.parseEnrollLoadRequest(w, r)
	if !ok {
		return
	}

	if !s.tasks.tryStart(taskLongtermBrowser) {
		writeRunning(w)
		return
	}
	defer s.tasks.finish(taskLongtermBrowser)

	result, err := longterm.FetchEnrollLoadEdu(auth, req.EducationYear)
	if err != nil {
		fmt.Printf("longterm/enrollLoad/edu 오류: %v\n", err)
		writeTaskError(w, err.Error())
		return
	}

	writeJSONData(w, http.StatusOK, result.EduResponse)
}

func (s *Server) handleLongtermEnrollLoadLongterm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}

	req, auth, ok := s.parseEnrollLoadRequest(w, r)
	if !ok {
		return
	}

	if !s.tasks.tryStart(taskLongtermBrowser) {
		writeRunning(w)
		return
	}
	defer s.tasks.finish(taskLongtermBrowser)

	result, err := longterm.FetchEnrollLoadLongterm(auth, req.EducationYear)
	if err != nil {
		fmt.Printf("longterm/enrollLoad/longterm 오류: %v\n", err)
		writeTaskError(w, err.Error())
		return
	}

	writeJSONData(w, http.StatusOK, result.Response)
}

func (s *Server) parseEnrollLoadRequest(w http.ResponseWriter, r *http.Request) (enrollLoadRequest, longterm.Auth, bool) {
	var req enrollLoadRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err.Error())
		return req, longterm.Auth{}, false
	}
	if err := req.validate(); err != nil {
		writeError(w, err.Error())
		return req, longterm.Auth{}, false
	}
	return req, toLongtermAuth(req.longtermAuthRequest), true
}

func (s *Server) handleLongtermEnrollUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}

	var req enrollUploadRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err.Error())
		return
	}
	if err := req.validate(); err != nil {
		writeError(w, err.Error())
		return
	}
	if !s.tasks.tryStart(taskLongtermBrowser) {
		writeRunning(w)
		return
	}
	defer s.tasks.finish(taskLongtermBrowser)

	auth := toLongtermAuth(req.longtermAuthRequest)

	if err := longterm.RunEnrollUpload(auth); err != nil {
		fmt.Printf("longterm/enrollUpload 오류: %v\n", err)
		writeTaskError(w, err.Error())
		return
	}

	writeTaskSuccess(w)
}

func (s *Server) handleLongtermAnnualPlan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}

	var req annualPlanRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err.Error())
		return
	}
	if err := req.validate(); err != nil {
		writeError(w, err.Error())
		return
	}
	if !s.tasks.tryStart(taskLongtermBrowser) {
		writeRunning(w)
		return
	}
	defer s.tasks.finish(taskLongtermBrowser)

	auth := toLongtermAuth(req.longtermAuthRequest)
	year := req.EducationYear

	if err := longterm.RunAnnualPlan(auth, year); err != nil {
		fmt.Printf("longterm/annualPlan 오류: %v\n", err)
		writeTaskError(w, err.Error())
		return
	}

	writeTaskSuccess(w)
}

func decodeJSON(r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return fmt.Errorf("요청 JSON 파싱 실패")
	}
	return nil
}

func toLongtermAuth(req longtermAuthRequest) longterm.Auth {
	return longterm.Auth{
		InstitutionCode: req.InstitutionCode,
		CertName:        req.CertName,
		CertPassword:    req.CertPassword,
	}
}

func toLongtermScheduleItems(items []scheduleItem) []longterm.ScheduleRequest {
	payload := make([]longterm.ScheduleRequest, len(items))
	for i, item := range items {
		payload[i] = longterm.ScheduleRequest{
			EducationDate:   item.EducationDate,
			StandardArea:    item.StandardArea,
			CourseName:      item.CourseName,
			CourseStartTime: item.CourseStartTime,
			CourseEndTime:   item.CourseEndTime,
			InstructorName:  item.InstructorName,
			Capacity:        item.Capacity,
			Note:            item.Note,
		}
	}
	return payload
}
