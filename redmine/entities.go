package redmine

import "time"

type errorsResult struct {
	Errors []string `json:"errors"`
}

type projectsRoot struct {
	Projects   []project `json:"projects"`
	TotalCount int64     `json:"total_count"`
	Offset     int64     `json:"offset"`
	Limit      int64     `json:"limit"`
}

type projectRoot struct {
	Project project `json:"project"`
}

type project struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Identifier  string    `json:"identifier"`
	Description string    `json:"description"`
	Trackers    []idName  `json:"trackers"`
	CreatedOn   time.Time `json:"created_on"`
	UpdatedOn   time.Time `json:"updated_on"`
}

type userRoot struct {
	User struct {
		APIKey      string `json:"api_key"`
		CreatedOn   string `json:"created_on"`
		Firstname   string `json:"firstname"`
		ID          int64  `json:"id"`
		LastLoginOn string `json:"last_login_on"`
		Lastname    string `json:"lastname"`
		Login       string `json:"login"`
		Mail        string `json:"mail"`
	} `json:"user"`
}

type issuesRoot struct {
	Issues     []issue `json:"issues"`
	TotalCount int64   `json:"total_count"`
	Offset     int64   `json:"offset"`
	Limit      int64   `json:"limit"`
}

type issueRoot struct {
	Issue issue `json:"issue"`
}

type issue struct {
	ID             int64   `json:"id,omitempty"`
	Project        *idName `json:"project,omitempty"`
	Tracker        *idName `json:"tracker,omitempty"`
	TrackerID      int64   `json:"tracker_id,omitempty"`
	Status         *idName `json:"status,omitempty"`
	Priority       *idName `json:"priority,omitempty"`
	Author         *idName `json:"author,omitempty"`
	AssignedToID   int64   `json:"assigned_to_id,omitempty"`
	AssignedTo     *idName `json:"assigned_to,omitempty"`
	FixedVersion   *idName `json:"fixed_version,omitempty"`
	Subject        string  `json:"subject,omitempty"`
	Description    string  `json:"description,omitempty"`
	StartDate      string  `json:"start_date,omitempty"`
	DueDate        string  `json:"due_date,omitempty"`
	DoneRatio      int     `json:"done_ratio,omitempty"`
	SpentHours     float64 `json:"spent_hours,omitempty"`
	EstimatedHours float64 `json:"estimated_hours,omitempty"`
	CustomFields   []struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"custom_fields,omitempty"`
	CreatedOn time.Time `json:"created_on"`
	UpdatedOn time.Time `json:"updated_on"`
}

type timeEntryActivitiesRoot struct {
	TimeEntryActivities []idName `json:"time_entry_activities"`
}

type idName struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type timeEntriesRoot struct {
	TimeEntries []timeEntry `json:"time_entries"`
	Limit       int         `json:"limit"`
	Offset      int         `json:"offset"`
	TotalCount  int         `json:"total_count"`
}

type timeEntryRoot struct {
	TimeEntry timeEntry `json:"time_entry"`
}

type timeEntry struct {
	Activity     *idName `json:"activity,omitempty"`
	ActivityID   int64   `json:"activity_id,omitempty"`
	Comments     string  `json:"comments,omitempty"`
	CreatedOn    string  `json:"created_on,omitempty"`
	CustomFields []struct {
		ID    int    `json:"id,omitempty"`
		Name  string `json:"name,omitempty"`
		Value string `json:"value,omitempty"`
	} `json:"custom_fields,omitempty"`
	Hours     float64 `json:"hours,omitempty"`
	ID        int64   `json:"id,omitempty"`
	IssueID   int64   `json:"issue_id,omitempty"`
	Issue     *idName `json:"issue,omitempty"`
	Project   *idName `json:"project,omitempty"`
	SpentOn   string  `json:"spent_on,omitempty"`
	UpdatedOn string  `json:"updated_on,omitempty"`
	User      *idName `json:"user,omitempty"`
}
