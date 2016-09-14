package tracker

import (
	"context"

	"../entities"
)

//NewClient returns new instance of tracker client
func NewClient() *Client {
	return &Client{}
}

//Client to tracker
type Client struct{}

//Project return project by id or err if not foind project
func (c *Client) Project(context.Context, entities.Tracker, entities.ProjectID) (*entities.Project, error) {
	panic("not implemented")
}

//Projects return project list and total amount of projects
func (c *Client) Projects(context.Context, entities.Tracker, entities.Pagination) ([]entities.Project, int64, error) {
	panic("not implemented")
}

//ProjectIssues return issues assigned to user and total amount
func (c *Client) ProjectIssues(context.Context, entities.Tracker, entities.ProjectID, entities.Pagination) ([]entities.Issue, int64, error) {
	panic("not implemented")
}

//UserInfo returns user info from tracker t
func (c *Client) UserInfo(context.Context, entities.Tracker) (*entities.User, error) {
	panic("not implemented")
}

//Issue query issue by ID from tracker t or err if not found
func (c *Client) Issue(context.Context, entities.Tracker, entities.ProjectID, entities.IssueID) (*entities.Issue, error) {
	panic("not implemented")
}

//IssueByURL query issue by URL ar err if not found
func (c *Client) IssueByURL(context.Context, entities.Tracker, entities.IssueURL) (*entities.Issue, error) {
	panic("not implemented")
}

//CreateIssue for projectID and assign it to user
func (c *Client) CreateIssue(context.Context, entities.Tracker, entities.NewIssue, entities.ProjectID) (*entities.Issue, error) {
	panic("not implemented")
}

//UpdateIssueProgress updates issue progress by issue id
func (c *Client) UpdateIssueProgress(context.Context, entities.Tracker, entities.ProjectID, entities.IssueID, entities.Progress) error {
	panic("not implemented")
}

//TotalReports receive date as UNIX timestamp (seconds) and return total reported time at this day in seconds
func (c *Client) TotalReports(ctx context.Context, t entities.Tracker, date int64) (int64, error) {
	panic("not implemented")
}

//CreateReport for user
func (c *Client) CreateReport(context.Context, entities.Tracker, entities.ProjectID, entities.Report) error {
	panic("not implemented")
}
