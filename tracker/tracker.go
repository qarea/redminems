package tracker

import (
	"context"

	"../entities"
)

func NewClient() *Client {
	return &Client{}
}

type Client struct{}

func (c *Client) Project(context.Context, entities.Tracker, entities.ProjectID) (*entities.Project, error) {
	panic("not implemented")
}

func (c *Client) Projects(context.Context, entities.Tracker, entities.Pagination) ([]entities.Project, int64, error) {
	panic("not implemented")
}

func (c *Client) ProjectIssues(context.Context, entities.Tracker, entities.ProjectID, entities.Pagination) ([]entities.Issue, int64, error) {
	panic("not implemented")
}

func (c *Client) UserInfo(context.Context, entities.Tracker) (*entities.User, error) {
	panic("not implemented")
}

func (c *Client) Issue(context.Context, entities.Tracker, entities.ProjectID, entities.IssueID) (*entities.Issue, error) {
	panic("not implemented")
}

func (c *Client) IssueByURL(context.Context, entities.Tracker, entities.IssueURL) (*entities.Issue, error) {
	panic("not implemented")
}

func (c *Client) CreateIssue(context.Context, entities.Tracker, entities.NewIssue, entities.ProjectID) (*entities.Issue, error) {
	panic("not implemented")
}

func (c *Client) UpdateIssueProgress(context.Context, entities.Tracker, entities.ProjectID, entities.IssueID, entities.Progress) error {
	panic("not implemented")
}

func (c *Client) TotalReports(ctx context.Context, t entities.Tracker, date int64) (int64, error) {
	panic("not implemented")
}

func (c *Client) CreateReport(context.Context, entities.Tracker, entities.ProjectID, entities.Report) error {
	panic("not implemented")
}
