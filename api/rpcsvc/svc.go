// Package rpcsvc provides handlers for JSON-RPC 2.0.
package rpcsvc

import (
	"context"
	"net/http"
	"net/rpc"

	"gitlab.qarea.org/tgms/ctxtg"

	"github.com/pkg/errors"
	"github.com/powerman/narada-go/narada"

	"../../cfg"
	"../../entities"

	"github.com/powerman/rpc-codec/jsonrpc2"
)

var log = narada.NewLog("rpcsvc: ")

func init() {
	http.Handle(cfg.HTTP.BasePath+"/rpc", jsonrpc2.HTTPHandler(nil))
	if err := rpc.Register(&API{}); err != nil {
		log.Fatal(err)
	}
}

type TrackerClient interface {
	Project(context.Context, entities.Tracker, entities.ProjectID) (*entities.Project, error)
	//Projects return project list and total amount of projects
	Projects(context.Context, entities.Tracker, entities.Pagination) ([]entities.Project, int64, error)
	//ProjectIssues return issues assigned to user and total amount
	ProjectIssues(context.Context, entities.Tracker, entities.ProjectID, entities.Pagination) ([]entities.Issue, int64, error)
	UserInfo(context.Context, entities.Tracker) (*entities.User, error)
	Issue(context.Context, entities.Tracker, entities.IssueID) (*entities.Issue, error)
	IssueByURL(context.Context, entities.Tracker, entities.IssueURL) (*entities.Issue, error)
	CreateIssue(context.Context, entities.Tracker, entities.NewIssue, entities.ProjectID) (*entities.Issue, error)
	UpdateIssueProgress(context.Context, entities.Tracker, entities.IssueID, entities.Progress) error
	//TotalReports receive date as UNIX timestamp (seconds) and return total reported time at this day in seconds
	TotalReports(ctx context.Context, t entities.Tracker, date int64) (int64, error)
	CreateReport(context.Context, entities.Tracker, entities.Report) error
}

func Init(r TrackerClient, p ctxtg.TokenParser) {
	if err := rpc.Register(newAPI(r, p)); err != nil {
		log.Fatal(err)
	}
	http.Handle(cfg.HTTP.BasePath+"/rpc", jsonrpc2.HTTPHandler(nil))
}

func newAPI(r TrackerClient, p ctxtg.TokenParser) *API {
	return &API{
		tracker:     r,
		tokenParser: p,
	}
}

type API struct {
	tracker     TrackerClient
	tokenParser ctxtg.TokenParser
}

func (*API) Version(args *struct{}, res *string) error {
	*res, _ = narada.Version()
	return nil
}

type ProjectsReq struct {
	Context ctxtg.Context
	Tracker entities.Tracker
	entities.Pagination
}

type ProjectsResp struct {
	Projects []entities.Project
	Amount   int64
}

func (r *API) GetProjects(req *ProjectsReq, resp *ProjectsResp) error {
	err := r.tokenParser.ParseCtxWithClaims(req.Context, func(ctx context.Context, c ctxtg.Claims) error {
		projects, amount, err := r.tracker.Projects(ctx, req.Tracker, req.Pagination)
		*resp = ProjectsResp{
			Projects: projects,
			Amount:   amount,
		}
		return err
	})
	return errWithLog(req.Context, "fail to GetProjects", err)
}

type ProjectDetailsReq struct {
	Context   ctxtg.Context
	Tracker   entities.Tracker
	ProjectID entities.ProjectID
}

type ProjectDetailsResp struct {
	Project entities.Project
}

func (r *API) GetProjectDetails(req *ProjectDetailsReq, resp *ProjectDetailsResp) error {
	err := r.tokenParser.ParseCtxWithClaims(req.Context, func(ctx context.Context, c ctxtg.Claims) error {
		project, err := r.tracker.Project(ctx, req.Tracker, req.ProjectID)
		if project != nil {
			*resp = ProjectDetailsResp{
				Project: *project,
			}
		}
		return err
	})
	return errWithLog(req.Context, "fail to GetProjectDetails", err)
}

type CurrentUserReq struct {
	Context ctxtg.Context
	Tracker entities.Tracker
}

type CurrentUserResp struct {
	User entities.User
}

func (r *API) GetCurrentUser(req *CurrentUserReq, resp *CurrentUserResp) error {
	err := r.tokenParser.ParseCtxWithClaims(req.Context, func(ctx context.Context, c ctxtg.Claims) error {
		u, err := r.tracker.UserInfo(ctx, req.Tracker)
		if u != nil {
			*resp = CurrentUserResp{
				User: *u,
			}
		}
		return err
	})
	return errWithLog(req.Context, "current user info err", err)
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

func (r *API) GetProjectIssues(req *ProjectIssuesReq, resp *ProjectIssuesResp) error {
	err := r.tokenParser.ParseCtxWithClaims(req.Context, func(ctx context.Context, c ctxtg.Claims) error {
		is, amount, err := r.tracker.ProjectIssues(ctx, req.Tracker, req.ProjectID, req.Pagination)
		*resp = ProjectIssuesResp{
			Issues: is,
			Amount: amount,
		}
		return err
	})
	return errWithLog(req.Context, "issues err", err)
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

func (r *API) CreateIssue(req *CreateIssueReq, resp *CreateIssueResp) error {
	err := r.tokenParser.ParseCtxWithClaims(req.Context, func(ctx context.Context, c ctxtg.Claims) error {
		issue, err := r.tracker.CreateIssue(ctx, req.Tracker, req.Issue, req.ProjectID)
		if issue != nil {
			*resp = CreateIssueResp{
				Issue: *issue,
			}
		}
		return err
	})
	return errWithLog(req.Context, "create issue err", err)
}

type GetIssueReq struct {
	Context ctxtg.Context
	Tracker entities.Tracker
	IssueID entities.IssueID
}

type GetIssueResp struct {
	Issue entities.Issue
}

func (r *API) GetIssue(req *GetIssueReq, resp *GetIssueResp) error {
	err := r.tokenParser.ParseCtxWithClaims(req.Context, func(ctx context.Context, c ctxtg.Claims) error {
		issue, err := r.tracker.Issue(ctx, req.Tracker, req.IssueID)
		if issue != nil {
			*resp = GetIssueResp{
				Issue: *issue,
			}
		}
		return err
	})
	return errWithLog(req.Context, "issue by ID err", err)
}

type UpdateIssueProgressReq struct {
	Context  ctxtg.Context
	Tracker  entities.Tracker
	IssueID  entities.IssueID
	Progress entities.Progress
}

func (r *API) UpdateIssueProgress(req *UpdateIssueProgressReq, _ *struct{}) error {
	err := r.tokenParser.ParseCtxWithClaims(req.Context, func(ctx context.Context, c ctxtg.Claims) error {
		return r.tracker.UpdateIssueProgress(ctx, req.Tracker, req.IssueID, req.Progress)
	})
	return errWithLog(req.Context, "update issue err", err)
}

type CreateReportReq struct {
	Context ctxtg.Context
	Tracker entities.Tracker
	Report  entities.Report
}

func (r *API) CreateReport(req *CreateReportReq, _ *struct{}) error {
	err := r.tokenParser.ParseCtxWithClaims(req.Context, func(ctx context.Context, c ctxtg.Claims) error {
		return r.tracker.CreateReport(ctx, req.Tracker, req.Report)
	})
	return errWithLog(req.Context, "create report err", err)
}

type GetReportsReq struct {
	Context ctxtg.Context
	Tracker entities.Tracker
	Date    int64
}

type GetReportsResp struct {
	Total int64
}

func (r *API) GetTotalReports(req *GetReportsReq, resp *GetReportsResp) error {
	err := r.tokenParser.ParseCtxWithClaims(req.Context, func(ctx context.Context, c ctxtg.Claims) error {
		time, err := r.tracker.TotalReports(ctx, req.Tracker, req.Date)
		*resp = GetReportsResp{
			Total: time,
		}
		return err
	})
	return errWithLog(req.Context, "total reports err", err)
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

func (r *API) GetIssueByURL(req *GetIssueByURLReq, resp *GetIssueByURLResp) error {
	err := r.tokenParser.ParseCtxWithClaims(req.Context, func(ctx context.Context, c ctxtg.Claims) error {
		issue, err := r.tracker.IssueByURL(ctx, req.Tracker, req.IssueURL)
		if issue == nil {
			return entities.ErrIssueNotFound
		}
		*resp = GetIssueByURLResp{
			Issue:     *issue,
			ProjectID: issue.ProjectID,
		}
		return err
	})
	return errWithLog(req.Context, "issue by URL err", err)
}

func errWithLog(ctx ctxtg.Context, prefix string, err error) error {
	if err == nil {
		return nil
	}
	log.ERR("tracking id: %s, token: %s, %s: %+v", ctx.TracingID, ctx.Token, prefix, err)
	err = errors.Cause(err)
	if err == context.DeadlineExceeded {
		return entities.ErrTimeout
	}
	return err
}
