package entities

// Project representation in our system
type Project struct {
	ID            ProjectID
	Title         string
	Link          string
	Description   string
	IssueTypes    []TypeID
	ActivityTypes []TypeID
}

// Tracker representation in our system
type Tracker struct {
	ID          int64
	URL         string
	Type        string
	Credentials Credentials
}

// Credentials to tracker
type Credentials struct {
	Login    string
	Password string
}

// TypeID used for IssueTypes and ActivityTypes
type TypeID struct {
	ID   int64
	Name string
}

// Issue representation in our system
type Issue struct {
	ID          IssueID
	ProjectID   ProjectID `json:"-"`
	Type        TypeID
	Title       string
	Description string
	Estimate    int64
	DueDate     int64
	Done        Progress
	Spent       int64
	URL         string
}

// NewIssue differs from the Issye in a Type field.
// Type field is int64 type. And Type field inside issue will be empty.
type NewIssue struct {
	Issue
	// Be carefule with NewIssue type
	// Type field inside Issue will be empty
	// We have only id of Type with NewIssue
	Type int64
}

// User information from tracker
type User struct {
	ID   int64
	Name string
	Mail string
}

// Report represents time report and additional information
type Report struct {
	IssueID    IssueID
	ActivityID int64
	Comments   string
	Duration   int64
	Started    int64
}

// Pagination used for pagination info in corresponding requests
type Pagination struct {
	Offset int
	Limit  int
}

// ProjectID is helper type to avoid invalid int usage
type ProjectID int64

// IssueID is helper type to avoid invalid int usage
type IssueID int64

// Progress represents progress in percents (0-100)
type Progress int

// IssueURL is helper type to avoid invalid string usage
type IssueURL string
