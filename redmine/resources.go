package redmine

import (
	"fmt"
	"strconv"

	"github.com/qarea/redminems/entities"
)

const (
	projectsResource      = "/projects.json?include=trackers"
	currentUserResourse   = "/users/current.json"
	reportsResource       = "/time_entries.json"
	timeEntriesActivities = "/enumerations/time_entry_activities.json"
)

const (
	projectLinkTemplate         = "/projects/%d"
	projectResourceTemplate     = "/projects/%d.json?include=trackers"
	projectsIssuesTemplate      = "/projects/%d/issues.json?offset=%d&limit=%d&assigned_to_id=me"
	createIssuesTemplate        = "/projects/%d/issues.json"
	timeEntriesResourceTemplate = "/time_entries.json?user_id=me&spent_on=%s&limit=100"
)

func projectIssuesResource(id entities.ProjectID, p entities.Pagination) string {
	if p.Offset == 0 && p.Limit == 0 {
		return fmt.Sprintf(projectsIssuesTemplate, id, 0, 100)
	}
	return fmt.Sprintf(projectsIssuesTemplate, id, p.Offset, p.Limit)
}

func timeEntriesResource(date int64) string {
	return fmt.Sprintf(timeEntriesResourceTemplate, secondsToDate(date))
}

func issueByID(id entities.IssueID) string {
	return "/issues/" + strconv.FormatInt(int64(id), 10)
}

func issueByIDResource(id entities.IssueID) string {
	return issueByID(id) + ".json"
}

func projectByIDLink(id entities.ProjectID) string {
	return fmt.Sprintf(projectLinkTemplate, id)
}

func projectByIDResource(id entities.ProjectID) string {
	return fmt.Sprintf(projectResourceTemplate, id)
}

func createIssueResource(projectID entities.ProjectID) string {
	return fmt.Sprintf(createIssuesTemplate, projectID)
}
