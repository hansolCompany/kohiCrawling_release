package api

import (
	"fmt"
)

type longtermAuthRequest struct {
	InstitutionCode string `json:"institutionCode"`
	CertName        string `json:"certName"`
	CertPassword    string `json:"certPassword"`
}

type scheduleItem struct {
	EducationDate   string `json:"educationDate"`
	StandardArea    string `json:"standardArea"`
	CourseName      string `json:"courseName"`
	CourseStartTime string `json:"courseStartTime"`
	CourseEndTime   string `json:"courseEndTime"`
	InstructorName  string `json:"instructorName"`
	Capacity        int    `json:"capacity"`
	Note            string `json:"note"`
}

type scheduleRequest struct {
	longtermAuthRequest
	Data []scheduleItem `json:"data"`
}

func validateScheduleItem(prefix string, item scheduleItem) error {
	if err := validateDateYYYYMMDD(prefix+".educationDate", item.EducationDate); err != nil {
		return err
	}
	if err := requireNonEmpty(prefix+".standardArea", item.StandardArea); err != nil {
		return err
	}
	if err := requireNonEmpty(prefix+".courseName", item.CourseName); err != nil {
		return err
	}
	if err := validateTimeHHmm(prefix+".courseStartTime", item.CourseStartTime); err != nil {
		return err
	}
	if err := validateTimeHHmm(prefix+".courseEndTime", item.CourseEndTime); err != nil {
		return err
	}
	if err := requireNonEmpty(prefix+".instructorName", item.InstructorName); err != nil {
		return err
	}
	if item.Capacity <= 0 {
		return fmt.Errorf("%s.capacity는 1 이상이어야 합니다", prefix)
	}
	return nil
}

func (r *scheduleRequest) validate() error {
	if err := validateLongtermAuth(r.InstitutionCode, r.CertName, r.CertPassword); err != nil {
		return err
	}
	if len(r.Data) == 0 {
		return fmt.Errorf("data는 1개 이상의 일정이 필요합니다")
	}
	for i, item := range r.Data {
		if err := validateScheduleItem(fmt.Sprintf("data[%d]", i), item); err != nil {
			return err
		}
	}
	return nil
}

type enrollLoadRequest struct {
	longtermAuthRequest
	EducationYear string `json:"educationYear"`
}

func (r *enrollLoadRequest) validate() error {
	if err := validateLongtermAuth(r.InstitutionCode, r.CertName, r.CertPassword); err != nil {
		return err
	}
	return validateYearYYYY("educationYear", r.EducationYear)
}

type enrollUploadRequest struct {
	longtermAuthRequest
}

func (r *enrollUploadRequest) validate() error {
	return validateLongtermAuth(r.InstitutionCode, r.CertName, r.CertPassword)
}

type annualPlanRequest struct {
	longtermAuthRequest
	EducationYear string `json:"educationYear"`
}

func (r *annualPlanRequest) validate() error {
	if err := validateLongtermAuth(r.InstitutionCode, r.CertName, r.CertPassword); err != nil {
		return err
	}
	return validateYearYYYY("educationYear", r.EducationYear)
}

type autoLearnRequest struct {
	CaregiverDbID string   `json:"caregiverDbId"`
	UserID        string   `json:"userId"`
	Password      string   `json:"password"`
	CourseName    []string `json:"courseName"`
}

func (r *autoLearnRequest) validate() error {
	if err := requireNonEmpty("caregiverDbId", r.CaregiverDbID); err != nil {
		return err
	}
	if err := requireNonEmpty("userId", r.UserID); err != nil {
		return err
	}
	if err := requireNonEmpty("password", r.Password); err != nil {
		return err
	}
	if len(r.CourseName) == 0 {
		return fmt.Errorf("courseName은 1개 이상의 강의명이 필요합니다")
	}
	for i, name := range r.CourseName {
		if err := requireNonEmpty(fmt.Sprintf("courseName[%d]", i), name); err != nil {
			return err
		}
	}
	return nil
}
