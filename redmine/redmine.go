// Package redmine provides connection to redmine tracker.
package redmine

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/powerman/narada-go/narada"

	"github.com/qarea/redminems/entities"
)

var log = narada.NewLog("redmine-client: ")

//NewClient returns new instance of redmine rest client
func NewClient(httpTimeout time.Duration) *RestClient {
	return &RestClient{
		httpClient: &http.Client{
			Timeout:       httpTimeout,
			CheckRedirect: copyHeadersOnRedirect,
		},
	}
}

//RestClient provide access to Redmine tracking service via REST API
type RestClient struct {
	httpClient *http.Client
}

//Project return project by id or err if not foind project
func (r *RestClient) Project(ctx context.Context, tr entities.Tracker, pid entities.ProjectID) (*entities.Project, error) {
	p, err := r.project(ctx, tr, pid)
	if err != nil {
		return nil, err
	}
	ac, err := r.activities(ctx, tr)
	if err != nil {
		return nil, err
	}
	p.ActivityTypes = idNamesToTypeID(ac.TimeEntryActivities)
	p.Link = fullURL(tr, projectByIDLink(p.ID))
	return p, nil
}

func (r *RestClient) project(ctx context.Context, t entities.Tracker, pid entities.ProjectID) (*entities.Project, error) {
	var pr projectRoot
	err := redmineRequest(requestOpts{
		httpClient:         r.httpClient,
		ctx:                ctx,
		resource:           projectByIDResource(pid),
		tracker:            t,
		result:             &pr,
		method:             get,
		validateStatusFunc: validateStatusOK,
	})
	if err == errNotFound {
		return nil, errors.Wrapf(entities.ErrProjectNotFound, "invalid project ID % for tracker ID: %d, URL: %s", pid, t.ID, t.URL)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load project by ID % from tracker ID: %d, URL: %s", pid, t.ID, t.URL)
	}
	project := toProject(pr.Project)
	return &project, nil
}

//Projects returns project list for current user from t
func (r *RestClient) Projects(ctx context.Context, t entities.Tracker, p entities.Pagination) (projects []entities.Project, totalAmount int64, err error) {
	pr, err := r.projects(ctx, t)
	if err != nil {
		return nil, 0, errors.Wrap(err, "projects request failed")
	}
	ac, err := r.activities(ctx, t)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "activities request failed")
	}
	ps := toProjects(*pr)
	addActivities(ps, *ac)
	addLinks(ps, t)
	return ps, pr.TotalCount, nil
}

//ProjectIssues returns issues for user from tracker by projectID and total amount of them
//Maximum paginatation limit is 100 items
func (r *RestClient) ProjectIssues(ctx context.Context, t entities.Tracker, projectID entities.ProjectID, p entities.Pagination) ([]entities.Issue, int64, error) {
	issues, amount, err := r.projectIssues(ctx, t, projectID, p)
	if err != nil {
		return nil, 0, err
	}
	var ids []entities.IssueID
	for _, is := range issues {
		ids = append(ids, is.ID)
	}
	fullIssues, err := r.parallelFullIssues(ctx, t, ids)
	if err != nil {
		return nil, 0, err
	}
	return fullIssues, amount, nil
}

func (r *RestClient) projectIssues(ctx context.Context, t entities.Tracker, projectID entities.ProjectID, p entities.Pagination) ([]entities.Issue, int64, error) {
	var issues issuesRoot
	err := redmineRequest(requestOpts{
		httpClient:         r.httpClient,
		ctx:                ctx,
		resource:           projectIssuesResource(projectID, p),
		tracker:            t,
		result:             &issues,
		method:             get,
		validateStatusFunc: validateStatusOK,
	})
	if err == errNotFound {
		return nil, 0, errors.Wrapf(entities.ErrProjectNotFound, "invalid project id %d from tracker ID: %d, login %s, URL: %s", projectID, t.ID, t.Credentials.Login, t.URL)
	}
	if err != nil {
		return nil, 0, errors.Wrapf(err, "failed to load issues from tracker ID: %d, login %s, URL: %s", t.ID, t.Credentials.Login, t.URL)
	}
	return toIssues(issues, t), issues.TotalCount, nil
}

func (r *RestClient) parallelFullIssues(ctx context.Context, t entities.Tracker, ids []entities.IssueID) ([]entities.Issue, error) {
	pairsC := make(chan issueErrPair)
	f := func(id entities.IssueID) {
		var p issueErrPair
		p.issue, p.err = r.issue(ctx, t, id)
		pairsC <- p
	}
	for _, id := range ids {
		go f(id)
	}
	var fullIssues []entities.Issue
	for range ids {
		pair := <-pairsC
		if pair.err != nil {
			return nil, pair.err
		}
		if pair.issue == nil {
			return nil, entities.ErrIssueNotFound
		}
		fullIssues = append(fullIssues, *pair.issue)
	}
	return fullIssues, nil
}

type issueErrPair struct {
	issue *entities.Issue
	err   error
}

//UserInfo returns user info from tracker t
func (r *RestClient) UserInfo(ctx context.Context, t entities.Tracker) (*entities.User, error) {
	var u userRoot
	err := redmineRequest(requestOpts{
		httpClient:         r.httpClient,
		resource:           currentUserResourse,
		ctx:                ctx,
		tracker:            t,
		result:             &u,
		method:             get,
		validateStatusFunc: validateStatusOK,
	})
	if err == errNotFound {
		return nil, errors.Wrapf(entities.ErrTrackerURL, "failed to load user info from tracker ID: %d,login %s, URL: %s", t.ID, t.Credentials.Login, t.URL)
	}

	if err != nil {
		return nil, errors.Wrapf(err, "failed to load user info from tracker ID: %d, URL: %s", t.ID, t.URL)
	}
	return &entities.User{
		ID:   u.User.ID,
		Name: strings.TrimSpace(fmt.Sprintf("%s %s", u.User.Lastname, u.User.Firstname)),
		Mail: u.User.Mail,
	}, nil
}

//Issue query issue by ID from tracker t
func (r *RestClient) Issue(ctx context.Context, t entities.Tracker, _ entities.ProjectID, issueID entities.IssueID) (*entities.Issue, error) {
	return r.issue(ctx, t, issueID)
}

func (r *RestClient) issue(ctx context.Context, t entities.Tracker, issueID entities.IssueID) (*entities.Issue, error) {
	var ir issueRoot
	err := redmineRequest(requestOpts{
		httpClient:         r.httpClient,
		ctx:                ctx,
		resource:           issueByIDResource(issueID),
		tracker:            t,
		result:             &ir,
		method:             get,
		validateStatusFunc: validateStatusOK,
	})
	if err == errNotFound {
		return nil, errors.Wrapf(entities.ErrIssueNotFound, "invalid issue ID % for tracker ID: %d, URL: %s", issueID, t.ID, t.URL)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load issue by ID % from tracker ID: %d, URL: %s", issueID, t.ID, t.URL)
	}
	issue := toIssue(ir.Issue, t)
	return &issue, nil
}

//IssueByURL query issue by URL
//Pattern /issues/([0-9]+
func (r *RestClient) IssueByURL(ctx context.Context, t entities.Tracker, issueURL entities.IssueURL) (*entities.Issue, error) {
	id, err := issueIDFromURL(issueURL)
	if err != nil {
		return nil, err
	}
	return r.issue(ctx, t, id)
}

//CreateIssue for projectID and assign it to user
func (r *RestClient) CreateIssue(ctx context.Context, t entities.Tracker, i entities.NewIssue, projectID entities.ProjectID) (*entities.Issue, error) {
	u, err := r.UserInfo(ctx, t)
	if err != nil {
		return nil, err
	}
	return r.createIssue(ctx, t, i, projectID, u.ID)
}

func (r *RestClient) createIssue(ctx context.Context, t entities.Tracker, i entities.NewIssue, projectID entities.ProjectID, userID int64) (*entities.Issue, error) {
	var ir issueRoot
	newIssue := i.Issue
	newIssue.Type.ID = i.Type
	issue := toIssueRoot(newIssue)
	issue.Issue.AssignedToID = userID
	err := redmineRequest(requestOpts{
		httpClient:         r.httpClient,
		ctx:                ctx,
		resource:           createIssueResource(projectID),
		tracker:            t,
		result:             &ir,
		method:             post,
		body:               issue,
		validateStatusFunc: validateStatusCreated,
	})
	if err == errNotFound {
		return nil, errors.Wrapf(entities.ErrProjectNotFound, "failed to create issue for tracker ID: %d, login %s, URL: %s", t.ID, t.Credentials.Login, t.URL)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create issue for tracker ID: %d, login %s, URL: %s", t.ID, t.Credentials.Login, t.URL)
	}
	createdIssue := toIssue(ir.Issue, t)
	return &createdIssue, nil
}

//UpdateIssueProgress updates issue progress by issue id
func (r *RestClient) UpdateIssueProgress(ctx context.Context, t entities.Tracker, pid entities.ProjectID, id entities.IssueID, pr entities.Progress) error {
	err := r.updateIssue(ctx, t, entities.Issue{
		ID:   id,
		Done: pr,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to update issue progress for tracker ID: %d, URL: %s", t.ID, t.URL)
	}
	return nil
}

//UpdateIssue - issue id are required
func (r *RestClient) UpdateIssue(ctx context.Context, t entities.Tracker, pid entities.ProjectID, i entities.Issue) (*entities.Issue, error) {
	err := r.updateIssue(ctx, t, i)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update issue for tracker ID: %d, URL: %s", t.ID, t.URL)
	}
	return r.issue(ctx, t, i.ID)
}

func (r *RestClient) updateIssue(ctx context.Context, t entities.Tracker, i entities.Issue) error {
	err := redmineRequest(requestOpts{
		httpClient:         r.httpClient,
		ctx:                ctx,
		resource:           issueByIDResource(i.ID),
		tracker:            t,
		method:             put,
		body:               toIssueRoot(i),
		validateStatusFunc: validateStatusOK,
	})
	if err == errNotFound {
		return entities.ErrIssueNotFound
	}
	return err
}

//TotalReports returns seconds amount for user 1 day for date
func (r *RestClient) TotalReports(ctx context.Context, t entities.Tracker, date int64) (int64, error) {
	var ts timeEntriesRoot
	err := redmineRequest(requestOpts{
		httpClient:         r.httpClient,
		ctx:                ctx,
		resource:           timeEntriesResource(date),
		tracker:            t,
		result:             &ts,
		method:             get,
		validateStatusFunc: validateStatusOK,
	})
	if err == errNotFound {
		return 0, errors.Wrapf(entities.ErrTrackerURL, "failed to load total reports from tracker ID %s, user login: %s, date %d", t.ID, t.Credentials.Login, date)
	}
	if err != nil {
		return 0, errors.Wrapf(err, "failed to load total reports from tracker ID %s, user login: %s, date %d", t.ID, t.Credentials.Login, date)
	}
	return sumReportHours(ts.TimeEntries), nil
}

//CreateReport for user
func (r *RestClient) CreateReport(ctx context.Context, t entities.Tracker, _ entities.ProjectID, rep entities.Report) error {
	err := redmineRequest(requestOpts{
		httpClient:         r.httpClient,
		ctx:                ctx,
		resource:           reportsResource,
		tracker:            t,
		method:             post,
		body:               reportToTimeEntry(rep),
		validateStatusFunc: validateStatusCreated,
	})
	if err == errNotFound {
		return errors.Wrapf(entities.ErrIssueNotFound, "invalid issueID %d for tracker ID: %d, URL: %s", rep.IssueID, t.ID, t.URL)
	}
	if err != nil {
		return errors.Wrapf(err, "failed to create report for tracker ID: %d, URL: %s", t.ID, t.URL)
	}
	return nil
}

func sumReportHours(ts []timeEntry) int64 {
	var totalSeconds int64
	for _, t := range ts {
		totalSeconds = totalSeconds + hoursToSeconds(t.Hours)
	}
	return totalSeconds
}

func (r *RestClient) activities(ctx context.Context, t entities.Tracker) (*timeEntryActivitiesRoot, error) {
	var a timeEntryActivitiesRoot
	err := redmineRequest(requestOpts{
		httpClient:         r.httpClient,
		ctx:                ctx,
		tracker:            t,
		resource:           timeEntriesActivities,
		result:             &a,
		method:             get,
		validateStatusFunc: validateStatusOK,
	})
	if err == errNotFound {
		return nil, errors.Wrapf(entities.ErrTrackerURL, "failed to load activities from tracker ID: %d,login %s, URL: %s", t.ID, t.Credentials.Login, t.URL)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load activities from tracker ID: %d,login %s, URL: %s", t.ID, t.Credentials.Login, t.URL)
	}
	return &a, nil
}

func (r *RestClient) projects(ctx context.Context, t entities.Tracker) (*projectsRoot, error) {
	var p projectsRoot
	err := redmineRequest(requestOpts{
		httpClient:         r.httpClient,
		ctx:                ctx,
		tracker:            t,
		resource:           projectsResource,
		result:             &p,
		method:             get,
		validateStatusFunc: validateStatusOK,
	})
	if err == errNotFound {
		return nil, errors.Wrapf(entities.ErrTrackerURL, "failed to load projects from tracker ID: %d,login %s, URL: %s", t.ID, t.Credentials.Login, t.URL)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load projects from tracker ID: %d,login %s, URL: %s", t.ID, t.Credentials.Login, t.URL)
	}
	return &p, nil
}
