package longterm

type Auth struct {
	InstitutionCode string
	CertName        string
	CertPassword    string
}

var DefaultAuth = Auth{
	InstitutionCode: "50652000009",
	CertName:        "명성요양보호사 교육원",
	CertPassword:    "~MJMS954044",
}

type ScheduleRequest struct {
	EducationDate   string
	StandardArea    string
	CourseName      string
	CourseStartTime string
	CourseEndTime   string
	InstructorName  string
	Capacity        int
	Note            string
}
