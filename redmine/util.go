package redmine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/qarea/redminems/entities"
)

const (
	redmineType = "REDMINE"
	slash       = '/'
	dateFormat  = "2006-01-02"
	get         = "GET"
	post        = "POST"
	put         = "PUT"
)

var (
	validateStatusOK      = validateStatus(http.StatusOK)
	validateStatusCreated = validateStatus(http.StatusCreated)

	errNotFound = errors.New("not found")
)

type requestOpts struct {
	httpClient         *http.Client
	ctx                context.Context
	tracker            entities.Tracker
	resource           string
	method             string
	body               interface{}
	result             interface{}
	validateStatusFunc func(int) error
}

func redmineRequest(opts requestOpts) error {
	resp, err := authRequest(opts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return errNotFound
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return entities.ErrCredentials
	}
	if resp.StatusCode == http.StatusForbidden {
		return entities.ErrForbidden
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed read body")
	}
	if resp.StatusCode == 422 {
		return errors.Wrapf(toExternalServiceErr(b), "invalid object passed to redmine, response body: %s", string(b))
	}
	if resp.StatusCode == http.StatusInternalServerError {
		return errors.Wrapf(entities.ErrRemoteServer, "redmine internal error, body: %s", string(b))
	}
	if opts.validateStatusFunc != nil {
		if err := opts.validateStatusFunc(resp.StatusCode); err != nil {
			return errors.Wrapf(err, "validation function failed: body: %s", string(b))
		}
	}
	if opts.result == nil {
		return nil
	}
	if err := json.Unmarshal(b, opts.result); err != nil {
		return errors.Wrapf(err, "status code: %d, failed to unmarshal: %s", resp.StatusCode, string(b))
	}
	return nil
}

func validateStatus(expected int) func(s int) error {
	return func(s int) error {
		if s != expected {
			return errors.Errorf("expected status code: %d, actual: %d", expected, s)
		}
		return nil
	}
}

func authRequest(opts requestOpts) (*http.Response, error) {
	if opts.tracker.Type != redmineType {
		return nil, errors.Wrapf(entities.ErrTrackerType, "invalid type: %s", opts.tracker.Type)
	}
	var bodyBytes []byte
	if opts.body != nil {
		var err error
		bodyBytes, err = json.Marshal(opts.body)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to marshal body %+v", opts.body)
		}
	}
	url := fullURL(opts.tracker, opts.resource)
	req, err := http.NewRequest(opts.method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create http request for url %s", url)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(opts.tracker.Credentials.Login, opts.tracker.Credentials.Password)
	resp, err := opts.httpClient.Do(req.WithContext(opts.ctx))
	if opts.ctx.Err() != nil {
		return nil, opts.ctx.Err()
	}
	if err, ok := err.(net.Error); ok && err.Timeout() {
		return nil, errors.Wrapf(entities.ErrTimeout, "http request timeout on URL %s", url)
	}
	if err != nil {
		return nil, errors.Wrapf(entities.ErrTrackerURL, "http request failed err %v", err)
	}
	return resp, nil
}

func secondsToDate(sec int64) string {
	if sec == 0 {
		return ""
	}
	return time.Unix(sec, 0).UTC().Format(dateFormat)
}

func dateToSeconds(date string) int64 {
	if date == "" {
		return 0
	}
	t, err := time.Parse(dateFormat, date)
	if err != nil {
		log.ERR("Failed to parse date %s, expected format %s", date, dateFormat)
		return 0
	}
	return t.Unix()
}

func toExternalServiceErr(b []byte) error {
	var redmineError errorsResult
	if err := json.Unmarshal(b, &redmineError); err != nil {
		return errors.Wrapf(err, "failed to unmarshal error response body %s", string(b))
	}
	return entities.NewTrackerValidationErr(strings.Join(redmineError.Errors, ". "))
}

func toIssues(ir issuesRoot, tr entities.Tracker) []entities.Issue {
	issues := make([]entities.Issue, len(ir.Issues))
	for i, issue := range ir.Issues {
		issues[i] = toIssue(issue, tr)
	}
	return issues
}

func reportToTimeEntry(rep entities.Report) *timeEntryRoot {
	return &timeEntryRoot{timeEntry{
		ActivityID: rep.ActivityID,
		IssueID:    int64(rep.IssueID),
		Hours:      secondsToHours(rep.Duration),
		Comments:   rep.Comments,
		SpentOn:    secondsToDate(rep.Started),
	}}
}

func toIssue(i issue, tr entities.Tracker) entities.Issue {
	var t entities.TypeID
	var pid entities.ProjectID
	if i.Tracker != nil {
		t = entities.TypeID{
			ID:   i.Tracker.ID,
			Name: i.Tracker.Name,
		}
	}
	if i.Project != nil {
		pid = entities.ProjectID(i.Project.ID)
	}
	issueID := entities.IssueID(i.ID)
	return entities.Issue{
		ID:          issueID,
		Title:       i.Subject,
		Type:        t,
		Description: i.Description,
		Estimate:    hoursToSeconds(i.EstimatedHours),
		DueDate:     dateToSeconds(i.DueDate),
		ProjectID:   pid,
		Done:        entities.Progress(i.DoneRatio),
		Spent:       hoursToSeconds(i.SpentHours),
		URL:         fullURL(tr, issueByID(issueID)),
	}
}

func toIssueRoot(i entities.Issue) *issueRoot {
	return &issueRoot{
		Issue: issue{
			DoneRatio:      int(i.Done),
			ID:             int64(i.ID),
			Subject:        i.Title,
			DueDate:        secondsToDate(i.DueDate),
			EstimatedHours: secondsToHours(i.Estimate),
			TrackerID:      i.Type.ID,
		},
	}
}

func secondsToHours(sec int64) float64 {
	return float64(sec) / 3600
}

func hoursToSeconds(h float64) int64 {
	return int64(h * 3600)
}

func toProjects(pr projectsRoot) []entities.Project {
	ps := make([]entities.Project, len(pr.Projects))
	for i, p := range pr.Projects {
		ps[i] = toProject(p)
	}
	return ps
}

func toProject(p project) entities.Project {
	return entities.Project{
		ID:          entities.ProjectID(p.ID),
		Title:       p.Name,
		Description: p.Description,
		IssueTypes:  idNamesToTypeID(p.Trackers),
	}

}

func addActivities(ps []entities.Project, activityTypes timeEntryActivitiesRoot) {
	for i := range ps {
		ps[i].ActivityTypes = idNamesToTypeID(activityTypes.TimeEntryActivities)
	}
}

func addLinks(ps []entities.Project, tr entities.Tracker) {
	for i := range ps {
		ps[i].Link = fullURL(tr, projectByIDLink(ps[i].ID))
	}
}

func idNamesToTypeID(ids []idName) []entities.TypeID {
	var types []entities.TypeID
	for _, id := range ids {
		types = append(types, entities.TypeID{
			ID:   id.ID,
			Name: id.Name,
		})
	}
	return types
}

func fullURL(tr entities.Tracker, resource string) string {
	return removeLastSlash(tr.URL) + resource
}

func removeLastSlash(url string) string {
	l := len(url)
	if l == 0 {
		return url
	}
	lastChar := l - 1
	if url[lastChar] == slash {
		return url[:lastChar]
	}
	return url
}

var issueIDRegexp = regexp.MustCompile(`/issues/([0-9]+)`)

func issueIDFromURL(url entities.IssueURL) (entities.IssueID, error) {
	strID := issueIDRegexp.FindStringSubmatch(string(url))
	if strID == nil {
		return 0, errors.Wrapf(entities.ErrIssueURL, "invalid issue URL %s", url)
	}
	id, err := strconv.ParseInt(strID[1], 10, 0)
	if err != nil {
		return 0, errors.Wrapf(entities.ErrIssueURL, "failed to parse id from URL %s", url)
	}
	return entities.IssueID(id), nil
}

func copyHeadersOnRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return fmt.Errorf("too many redirects")
	}
	if len(via) == 0 {
		return nil
	}
	for attr, val := range via[0].Header {
		if _, ok := req.Header[attr]; !ok {
			req.Header[attr] = val
		}
	}
	return nil
}
