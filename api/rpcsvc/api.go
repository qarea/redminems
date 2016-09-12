package rpcsvc

import (
	"../../entities"
	"gitlab.qarea.org/tgms/ctxtg"
)

type ProjectsReq struct {
	Context ctxtg.Context
	Tracker entities.Tracker
	entities.Pagination
}

type ProjectsResp struct {
	Projects []entities.Project
	Amount   int64
}

type ProjectDetailsReq struct {
	Context   ctxtg.Context
	Tracker   entities.Tracker
	ProjectID entities.ProjectID
}

type ProjectDetailsResp struct {
	Project entities.Project
}

type CurrentUserReq struct {
	Context ctxtg.Context
	Tracker entities.Tracker
}

type CurrentUserResp struct {
	User entities.User
}

type ProjectIssuesReq struct {
	Context   ctxtg.Context
	Tracker   entities.Tracker
	ProjectID entities.ProjectID
	entities.Pagination
}

type ProjectIssuesResp struct {
	Issues []entities.Issue
	Amount int64
}

type CreateIssueReq struct {
	Context   ctxtg.Context
	Tracker   entities.Tracker
	Issue     entities.NewIssue
	ProjectID entities.ProjectID
}

type CreateIssueResp struct {
	Issue entities.Issue
}

type GetIssueReq struct {
	Context ctxtg.Context
	Tracker entities.Tracker
	IssueID entities.IssueID
}

type GetIssueResp struct {
	Issue entities.Issue
}

type CreateReportReq struct {
	Context ctxtg.Context
	Tracker entities.Tracker
	Report  entities.Report
}

type GetIssueByURLReq struct {
	Context  ctxtg.Context
	Tracker  entities.Tracker
	IssueURL entities.IssueURL
}

type GetIssueByURLResp struct {
	Issue     entities.Issue
	ProjectID entities.ProjectID
}

type UpdateIssueProgressReq struct {
	Context  ctxtg.Context
	Tracker  entities.Tracker
	IssueID  entities.IssueID
	Progress entities.Progress
}

type GetReportsReq struct {
	Context ctxtg.Context
	Tracker entities.Tracker
	Date    int64
}

type GetReportsResp struct {
	Total int64
}
