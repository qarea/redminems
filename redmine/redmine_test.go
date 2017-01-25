package redmine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/qarea/redminems/entities"
)

const (
	projectsFile       = "projects.json"
	projectFile        = "project.json"
	userFile           = "user.json"
	issuesFile         = "issues.json"
	issueFile          = "issue.json"
	timeentriesFile    = "timeentries.json"
	timeActivitiesFile = "timeactivities.json"
)

var (
	testCreds = entities.Credentials{
		Login:    "login1",
		Password: "password2",
	}
)

func TestProjectsReq(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != get {
			t.Errorf("invalid method %s", r.Method)
		}
		if r.URL.Path != "/projects.json" {
			t.Errorf("Invalid resource path %s", r.URL.Path)
		}
		if r.URL.Query().Get("include") != "trackers" {
			t.Errorf("Missed include query param")
		}
		w.Write(readTestFile(t, projectsFile))
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	pr, err := r.projects(context.Background(), entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	})
	if err != nil {
		t.Fatal(err)
	}
	ps := toProjects(*pr)
	if len(ps) != len(projectsJSON) {
		t.Error("Invalid project amount")
	}
	for i, p := range ps {
		expectedPr := projectsJSON[i]
		if p.ID != expectedPr.ID {
			t.Errorf("Invalid id %v != %v", p.ID, expectedPr.ID)
		}
		if p.Title != expectedPr.Title {
			t.Errorf("Invalid title %v != %v", p.Title, expectedPr.Title)
		}
		if p.Description != expectedPr.Description {
			t.Errorf("Invalid description %v != %v", p.Description, expectedPr.Description)
		}
		if !reflect.DeepEqual(p.IssueTypes, expectedPr.IssueTypes) {
			t.Errorf("Invalid issuetypes %v != %v \n %v \n %v", len(p.IssueTypes), len(expectedPr.IssueTypes), p.IssueTypes, expectedPr.IssueTypes)
		}
	}
}

func TestProjectsReqNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	_, err := r.projects(context.Background(), entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	})
	assertErr(t, err, entities.ErrTrackerURL)
}

func TestProjectsReqInternalError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	_, err := r.projects(context.Background(), entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	})
	assertErr(t, err, entities.ErrRemoteServer)
}

func TestProjectReq(t *testing.T) {
	pid := projectJSON.ID
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != get {
			t.Errorf("Invalid method %s", r.Method)
		}
		if r.URL.Path != "/projects/"+fmt.Sprintf("%d", pid)+".json" {
			t.Errorf("Invalid resource path %s", r.URL.Path)
		}
		if r.URL.Query().Get("include") != "trackers" {
			t.Error("Missed include query param")
		}
		w.Write(readTestFile(t, projectFile))
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	p, err := r.project(context.Background(), entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	}, pid)
	if err != nil {
		t.Fatal(err)
	}
	if p.ID != projectJSON.ID {
		t.Errorf("Invalid id %v != %v", p.ID, projectJSON.ID)
	}
	if p.Title != projectJSON.Title {
		t.Errorf("Invalid title %v != %v", p.Title, projectJSON.Title)
	}
	if p.Description != projectJSON.Description {
		t.Errorf("Invalid description %v != %v", p.Description, projectJSON.Description)
	}
	if !reflect.DeepEqual(p.IssueTypes, projectJSON.IssueTypes) {
		t.Errorf("Invalid issuetypes %v != %v", len(p.IssueTypes), len(projectJSON.IssueTypes))
	}
}

func TestProjectReqNotFound(t *testing.T) {
	pid := projectJSON.ID
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	_, err := r.project(context.Background(), entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	}, pid)
	assertErr(t, err, entities.ErrProjectNotFound)
}

func TestProjectReqInternal(t *testing.T) {
	pid := projectJSON.ID
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	_, err := r.project(context.Background(), entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	}, pid)
	assertErr(t, err, entities.ErrRemoteServer)
}

func TestActivitiesReq(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != get {
			t.Errorf("Invalid method %s", r.Method)
		}
		if r.URL.Path != "/enumerations/time_entry_activities.json" {
			t.Errorf("Invalid resource path %s", r.URL.Path)
		}
		w.Write(readTestFile(t, timeActivitiesFile))
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	as, err := r.activities(context.Background(), entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(activityTypesJSON, idNamesToTypeID(as.TimeEntryActivities)) {
		t.Error("Unexpected result")
	}
}

func TestActivitiesReqNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	_, err := r.activities(context.Background(), entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	})
	assertErr(t, err, entities.ErrTrackerURL)
}

func TestActivitiesReqInternalErr(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	_, err := r.activities(context.Background(), entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	})
	assertErr(t, err, entities.ErrRemoteServer)
}

func TestUserReq(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != get {
			t.Errorf("Invalid method %s", r.Method)
		}
		if r.URL.Path != currentUserResourse {
			t.Errorf("Invalid resource path %s", r.URL.Path)
		}
		w.Write(readTestFile(t, userFile))
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	u, err := r.UserInfo(context.Background(), entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(userJSON, *u) {
		t.Errorf(`Unexpected result expect: %+v \n actual: %+v`, userJSON, u)
	}
}

func TestUserReqNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	_, err := r.UserInfo(context.Background(), entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	})
	assertErr(t, err, entities.ErrTrackerURL)
}

func TestUserReqInternalErr(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	_, err := r.UserInfo(context.Background(), entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	})
	assertErr(t, err, entities.ErrRemoteServer)
}

func TestParallelIssues(t *testing.T) {
	ids := []entities.IssueID{1, 2, 3, 4, 5}
	var counter int64
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&counter, 1)
		w.Write(readTestFile(t, issueFile))
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	issues, err := r.parallelFullIssues(
		context.Background(),
		entities.Tracker{
			Credentials: testCreds,
			URL:         ts.URL,
			Type:        redmineType,
		},
		ids,
	)
	if err != nil {
		t.Fatal(err)
	}
	if counter != int64(len(ids)) {
		t.Error("Not enough requests")
	}
	if len(issues) != len(ids) {
		t.Error("Invalid amount")
	}
}

func TestParallelIssuesErr(t *testing.T) {
	ids := []entities.IssueID{1, 2, 3, 4, 5}
	var counter int64
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		new := atomic.AddInt64(&counter, 1)
		if new == 3 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Write(readTestFile(t, issueFile))
		}
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	_, err := r.parallelFullIssues(
		context.Background(),
		entities.Tracker{
			Credentials: testCreds,
			URL:         ts.URL,
			Type:        redmineType,
		},
		ids,
	)
	assertErr(t, err, entities.ErrRemoteServer)
}

func TestProjectIssuesReq(t *testing.T) {
	page := entities.Pagination{
		Offset: 20,
		Limit:  200,
	}
	pid := entities.ProjectID(1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != get {
			t.Errorf("Invalid method %s", r.Method)
		}
		if r.URL.Path != "/projects/"+strconv.Itoa(int(pid))+"/issues.json" {
			t.Errorf("Unexpected resource path %s", r.URL.Path)
		}
		if r.URL.Query().Get("assigned_to_id") != "me" {
			t.Errorf("Invalid assigned_to_id filter")
		}
		if r.URL.Query().Get("offset") != fmt.Sprintf("%d", page.Offset) {
			t.Errorf("Invalid offset")
		}
		if r.URL.Query().Get("limit") != fmt.Sprintf("%d", page.Limit) {
			t.Errorf("Invalid limit id")
		}
		w.Write(readTestFile(t, issuesFile))
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	iss, _, err := r.projectIssues(
		context.Background(),
		entities.Tracker{
			Credentials: testCreds,
			URL:         ts.URL,
			Type:        redmineType,
		},
		pid,
		page,
	)
	if err != nil {
		t.Fatal(err)
	}
	for i, issue := range iss {
		if issue.URL != fmt.Sprintf("%s%s%d", ts.URL, "/issues/", issue.ID) {
			t.Error("Invalid URL", issue.URL)
		}
		iss[i].URL = ""
	}
	if !reflect.DeepEqual(issuesJSON, iss) {
		t.Error("Unexpected result")
	}
}

func TestProjectIssuesReqNotFound(t *testing.T) {
	page := entities.Pagination{
		Offset: 20,
		Limit:  200,
	}
	pid := entities.ProjectID(1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	_, _, err := r.projectIssues(
		context.Background(),
		entities.Tracker{
			Credentials: testCreds,
			URL:         ts.URL,
			Type:        redmineType,
		},
		pid,
		page,
	)
	assertErr(t, err, entities.ErrProjectNotFound)
}

func TestProjectIssuesReqInternalErr(t *testing.T) {
	page := entities.Pagination{
		Offset: 20,
		Limit:  200,
	}
	pid := entities.ProjectID(1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	_, _, err := r.projectIssues(
		context.Background(),
		entities.Tracker{
			Credentials: testCreds,
			URL:         ts.URL,
			Type:        redmineType,
		},
		pid,
		page,
	)
	assertErr(t, err, entities.ErrRemoteServer)
}

func TestProjectIssueResourceEmptyPage(t *testing.T) {
	result := projectIssuesResource(1, entities.Pagination{})
	expected := "/projects/1/issues.json?offset=0&limit=100&assigned_to_id=me"
	if result != expected {
		t.Error("Unexpected result", result)
	}
}

func TestIssueByIDReq(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != get {
			t.Errorf("Invalid method %s", r.Method)
		}
		if r.URL.Path != "/issues/3.json" {
			t.Errorf("Unexpected resource path %s", r.URL.Path)
		}
		w.Write(readTestFile(t, issueFile))
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	issue, err := r.Issue(
		context.Background(),
		entities.Tracker{
			Credentials: testCreds,
			URL:         ts.URL,
			Type:        redmineType,
		},
		0,
		3,
	)
	if err != nil {
		t.Fatal(err)
	}
	if issue.URL != fmt.Sprintf("%s%s%d", ts.URL, "/issues/", issue.ID) {
		t.Error("Invalid URL", issue.URL)
	}
	issue.URL = ""

	if !reflect.DeepEqual(issueJSON, issue) {
		t.Error("Unexpected result")
	}

}

func TestIssueReqNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	_, err := r.issue(context.Background(), entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	}, 3)
	assertErr(t, err, entities.ErrIssueNotFound)
}

func TestIssueReqInternalErr(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	_, err := r.issue(context.Background(), entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	}, 3)
	assertErr(t, err, entities.ErrRemoteServer)
}

func TestCreateIssueReq(t *testing.T) {
	var newIssueID entities.IssueID = 555
	testIssue := entities.NewIssue{Issue: *issueJSON}
	pid := entities.ProjectID(3)
	uid := int64(3)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != post {
			t.Errorf("Invalid method %s", r.Method)
		}
		if r.URL.Path != "/projects/"+strconv.Itoa(int(pid))+"/issues.json" {
			t.Errorf("Unexpected resource path %s", r.URL.Path)
		}
		var ir issueRoot
		unmarshal(t, r.Body, &ir)
		if ir.Issue.Subject != testIssue.Title {
			t.Errorf("Invalid subject %v != %v", ir.Issue.Subject, testIssue.Title)
		}
		if ir.Issue.DueDate != "2016-06-14" {
			t.Errorf("Invalid duedate %v", ir.Issue.DueDate)
		}
		if ir.Issue.EstimatedHours != 24.0 {
			t.Errorf("Invalid estimate hours %v", ir.Issue.EstimatedHours)
		}
		if ir.Issue.AssignedToID != uid {
			t.Errorf("Invalid userID %v != %v", ir.Issue.AssignedToID, uid)
		}
		w.WriteHeader(http.StatusCreated)
		ir.Issue.ID = int64(newIssueID)
		b, _ := json.Marshal(ir)
		w.Write(b)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	issue, err := r.createIssue(
		context.Background(),
		entities.Tracker{
			Credentials: testCreds,
			URL:         ts.URL,
			Type:        redmineType,
		},
		testIssue,
		pid,
		uid,
	)
	if err != nil {
		t.Fatal(err)
	}

	if issue.ID != newIssueID {
		t.Errorf("Invalid issue ID %v", issue.ID)
	}
}

func TestCreateIssueReqNotFound(t *testing.T) {
	testIssue := entities.NewIssue{Issue: *issueJSON}
	pid := entities.ProjectID(3)
	uid := int64(3)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	_, err := r.createIssue(
		context.Background(),
		entities.Tracker{
			Credentials: testCreds,
			URL:         ts.URL,
			Type:        redmineType,
		},
		testIssue,
		pid,
		uid,
	)
	assertErr(t, err, entities.ErrProjectNotFound)
}

func TestCreateIssueReqInternalErr(t *testing.T) {
	testIssue := entities.NewIssue{Issue: *issueJSON}
	pid := entities.ProjectID(3)
	uid := int64(3)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	_, err := r.createIssue(
		context.Background(),
		entities.Tracker{
			Credentials: testCreds,
			URL:         ts.URL,
			Type:        redmineType,
		},
		testIssue,
		pid,
		uid,
	)
	assertErr(t, err, entities.ErrRemoteServer)
}

func TestCreateReportReq(t *testing.T) {
	report := entities.Report{
		Duration:   10800,
		IssueID:    1,
		Started:    1470839302,
		ActivityID: 9,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != post {
			t.Errorf("Invalid method %s", r.Method)
		}
		if r.URL.Path != "/time_entries.json" {
			t.Errorf("Unexpected resource path %s", r.URL.Path)
		}
		var timeEntryRoot timeEntryRoot
		unmarshal(t, r.Body, &timeEntryRoot)
		if timeEntryRoot.TimeEntry.IssueID != int64(report.IssueID) ||
			timeEntryRoot.TimeEntry.Hours != 3 ||
			timeEntryRoot.TimeEntry.SpentOn != "2016-08-10" {
			t.Error("Iinvalid timeEntry")
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	err := r.CreateReport(context.Background(), entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	}, 0, report)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateReportReqNotFound(t *testing.T) {
	report := entities.Report{
		Duration:   10800,
		IssueID:    1,
		Started:    1470839302,
		ActivityID: 9,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	err := r.CreateReport(context.Background(), entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	}, 0, report)
	assertErr(t, err, entities.ErrIssueNotFound)
}

func TestCreateReportReqInternalErr(t *testing.T) {
	report := entities.Report{
		Duration:   10800,
		IssueID:    1,
		Started:    1470839302,
		ActivityID: 9,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	err := r.CreateReport(context.Background(), entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	}, 0, report)
	assertErr(t, err, entities.ErrRemoteServer)
}

func TestUpdateIssueReq(t *testing.T) {
	testIssue := *issueJSON

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != put {
			t.Errorf("Invalid method %s", r.Method)
		}
		if r.URL.Path != "/issues/"+strconv.Itoa(int(testIssue.ID))+".json" {
			t.Errorf("Unexpected resource path %s", r.URL.Path)
		}
		var ir issueRoot
		unmarshal(t, r.Body, &ir)
		if ir.Issue.Subject != testIssue.Title ||
			ir.Issue.DueDate != "2016-06-14" ||
			ir.Issue.EstimatedHours != 24.0 ||
			ir.Issue.ID != int64(testIssue.ID) {
			t.Error("Invalid issue passed to the server")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	err := r.updateIssue(
		context.Background(),
		entities.Tracker{
			Credentials: testCreds,
			URL:         ts.URL,
			Type:        redmineType,
		},
		testIssue,
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateIssueReqNotFound(t *testing.T) {
	testIssue := *issueJSON

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	err := r.updateIssue(
		context.Background(),
		entities.Tracker{
			Credentials: testCreds,
			URL:         ts.URL,
			Type:        redmineType,
		},
		testIssue,
	)
	assertErr(t, err, entities.ErrIssueNotFound)
}

func TestUpdateIssueReqInternalErr(t *testing.T) {
	testIssue := entities.NewIssue{Issue: *issueJSON}
	pid := entities.ProjectID(3)
	uid := int64(3)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	_, err := r.createIssue(
		context.Background(),
		entities.Tracker{
			Credentials: testCreds,
			URL:         ts.URL,
			Type:        redmineType,
		},
		testIssue,
		pid,
		uid,
	)
	assertErr(t, err, entities.ErrRemoteServer)
}

func TestTotalReportsReq(t *testing.T) {
	date := int64(1465862400)
	dateString := "2016-06-14"
	expectedSeconds := int64(57600)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/time_entries.json" {
			t.Errorf("Unexpected resource path %s", r.URL.RawPath)
		}
		if r.Method != get {
			t.Errorf("Invalid method %s", r.Method)
		}

		date := r.URL.Query().Get("spent_on")
		if date != dateString {
			t.Errorf("Invalid date %v, expected %v", date, dateString)
		}
		w.Write(readTestFile(t, timeentriesFile))
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	totalTime, err := r.TotalReports(
		context.Background(),
		entities.Tracker{
			Credentials: testCreds,
			URL:         ts.URL,
			Type:        redmineType,
		},
		date,
	)
	if err != nil {
		t.Fatal(err)
	}
	if totalTime != expectedSeconds {
		t.Error("Unexpected result")
	}
}

func TestTotalReportsReqNotFound(t *testing.T) {
	date := int64(1465862400)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	_, err := r.TotalReports(
		context.Background(),
		entities.Tracker{
			Credentials: testCreds,
			URL:         ts.URL,
			Type:        redmineType,
		},
		date,
	)
	assertErr(t, err, entities.ErrTrackerURL)
}

func TestTotalReportsReqInternalErr(t *testing.T) {
	date := int64(1465862400)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	_, err := r.TotalReports(
		context.Background(),
		entities.Tracker{
			Credentials: testCreds,
			URL:         ts.URL,
			Type:        redmineType,
		},
		date,
	)
	assertErr(t, err, entities.ErrRemoteServer)
}

func TestUpdateIssueProgress(t *testing.T) {
	is := entities.IssueID(3)
	prog := entities.Progress(4)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != put {
			t.Errorf("Invalid method %s", r.Method)
		}
		if r.URL.Path != "/issues/"+strconv.Itoa(int(is))+".json" {
			t.Errorf("Unexpected resource path %s", r.URL.Path)
		}
		var ir issueRoot
		unmarshal(t, r.Body, &ir)
		if ir.Issue.ID != int64(is) ||
			ir.Issue.DoneRatio != int(prog) {
			t.Error("Invalid issue passed to the server")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	tr := entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	}

	r := NewClient(testTimeout())
	err := r.UpdateIssueProgress(context.Background(), tr, 0, is, prog)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateIssueProgressError(t *testing.T) {
	is := entities.IssueID(3)
	prog := entities.Progress(4)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != put {
			t.Errorf("Invalid method %s", r.Method)
		}
		if r.URL.Path != "/issues/"+strconv.Itoa(int(is))+".json" {
			t.Errorf("Unexpected resource path %s", r.URL.Path)
		}
		var ir issueRoot
		unmarshal(t, r.Body, &ir)
		if ir.Issue.ID != int64(is) ||
			ir.Issue.DoneRatio != int(prog) {
			t.Error("Invalid issue passed to the server")
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	tr := entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	}

	r := NewClient(testTimeout())
	err := r.UpdateIssueProgress(context.Background(), tr, 0, is, prog)
	if err == nil {
		t.Error("Error expected")
	}
}

func TestFullUpdateIssue(t *testing.T) {
	testIssue := *issueJSON

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == get {
			w.Write(readTestFile(t, issueFile))
		}
		if r.URL.Path != "/issues/"+strconv.Itoa(int(testIssue.ID))+".json" {
			t.Errorf("Unexpected resource path %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	_, err := r.UpdateIssue(
		context.Background(),
		entities.Tracker{
			Credentials: testCreds,
			URL:         ts.URL,
			Type:        redmineType,
		},
		0,
		testIssue,
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFullUpdateIssueError(t *testing.T) {
	testIssue := *issueJSON

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == get {
			w.Write(readTestFile(t, issueFile))
		}
		if r.URL.Path != "/issues/"+strconv.Itoa(int(testIssue.ID))+".json" {
			t.Errorf("Unexpected resource path %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	r := NewClient(testTimeout())
	_, err := r.UpdateIssue(
		context.Background(),
		entities.Tracker{
			Credentials: testCreds,
			URL:         ts.URL,
			Type:        redmineType,
		},
		0,
		testIssue,
	)
	if err == nil {
		t.Error("Error expect")
	}
}

func TestCreateIssue(t *testing.T) {
	pid := entities.ProjectID(0)
	testIssue := entities.NewIssue{Issue: *issueJSON}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		if r.URL.Path == "/users/current.json" {
			fmt.Println("users")
			w.Write(readTestFile(t, userFile))
		}
		if r.URL.Path == "/projects/0/issues.json" {
			fmt.Println("issues")
			w.WriteHeader(http.StatusCreated)
			io.Copy(w, r.Body)
		}
	}))
	defer ts.Close()

	tr := entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	}

	r := NewClient(testTimeout())
	_, err := r.CreateIssue(context.Background(), tr, testIssue, pid)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateIssueWithServerError(t *testing.T) {
	pid := entities.ProjectID(0)
	testIssue := entities.NewIssue{Issue: *issueJSON}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		io.Copy(w, r.Body)
	},
	))
	defer ts.Close()

	tr := entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	}

	r := NewClient(testTimeout())
	_, err := r.CreateIssue(context.Background(), tr, testIssue, pid)
	if err == nil {
		t.Error("Error expected")
	}
}

func TestIssueByURLInvalidURL(t *testing.T) {
	issurl := entities.IssueURL("https")

	tr := entities.Tracker{
		Credentials: testCreds,
		Type:        redmineType,
	}

	r := NewClient(testTimeout())
	_, err := r.IssueByURL(context.Background(), tr, issurl)
	if errors.Cause(err) != entities.ErrIssueURL {
		t.Fatal(err)
	}
}

func TestIssueByURL(t *testing.T) {

	issurl := entities.IssueURL("https://redmine.qarea.org/issues/80193")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Write(readTestFile(t, issueFile))
	}))
	defer ts.Close()

	tr := entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	}

	r := NewClient(testTimeout())
	_, err := r.IssueByURL(context.Background(), tr, issurl)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIssueByURLWithServerError(t *testing.T) {
	issurl := entities.IssueURL("https://redmine.qarea.org/issues/80193")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		io.Copy(w, r.Body)
	},
	))
	defer ts.Close()

	tr := entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	}

	r := NewClient(testTimeout())
	_, err := r.IssueByURL(context.Background(), tr, issurl)
	if err == nil {
		t.Error("Error expected")
	}
}

func TestProjectIssuesErr(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/projects/1/issues.json") {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer ts.Close()

	page := entities.Pagination{
		Offset: 41,
		Limit:  88,
	}

	tr := entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	}

	r := NewClient(testTimeout())
	_, _, err := r.ProjectIssues(context.Background(), tr, 1, page)
	if err == nil {
		t.Fatal(err)
	}
}

func TestProjectIssuesErrorAsync(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/projects/1/issues.json") {
			w.Write(readTestFile(t, issuesFile))
		}
		if strings.HasPrefix(r.URL.Path, "/issues") {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer ts.Close()

	page := entities.Pagination{
		Offset: 41,
		Limit:  88,
	}

	tr := entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	}

	r := NewClient(testTimeout())
	_, _, err := r.ProjectIssues(context.Background(), tr, 1, page)
	if err == nil {
		t.Fatal(err)
	}
}

func TestProjectIssues(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/projects/1/issues.json") {
			w.Write(readTestFile(t, issuesFile))
		}
		if strings.HasPrefix(r.URL.Path, "/issues") {
			w.Write(readTestFile(t, issueFile))
		}
	}))
	defer ts.Close()

	page := entities.Pagination{
		Offset: 41,
		Limit:  88,
	}

	tr := entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	}

	r := NewClient(testTimeout())
	iss, _, err := r.ProjectIssues(context.Background(), tr, 1, page)
	if len(iss) == 0 {
		t.Error("Should return issues")
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestProjects(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/projects") {
			w.Write(readTestFile(t, projectsFile))
		}
		if strings.HasPrefix(r.URL.Path, "/enumerations/time_entry_activities.json") {
			w.Write(readTestFile(t, timeentriesFile))
		}
	}))
	defer ts.Close()

	page := entities.Pagination{
		Offset: 41,
		Limit:  88,
	}

	tr := entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	}

	r := NewClient(testTimeout())
	prs, _, err := r.Projects(context.Background(), tr, page)
	if len(prs) == 0 {
		t.Error("Should return projects")
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestProjectActivityErr(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/project") {
			w.Write(readTestFile(t, projectFile))
		}
		if strings.HasPrefix(r.URL.Path, "/enumerations/time_entry_activities.json") {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer ts.Close()

	tr := entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	}

	r := NewClient(testTimeout())
	_, err := r.Project(context.Background(), tr, 1)
	if err == nil {
		t.Fatal(err)
	}
}

func TestProjectsErr(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	page := entities.Pagination{
		Offset: 41,
		Limit:  88,
	}

	tr := entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	}

	r := NewClient(testTimeout())
	_, _, err := r.Projects(context.Background(), tr, page)
	if err == nil {
		t.Fatal(err)
	}
}

func TestProject(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/project") {
			w.Write(readTestFile(t, projectFile))
		}
		if strings.HasPrefix(r.URL.Path, "/enumerations/time_entry_activities.json") {
			w.Write(readTestFile(t, timeentriesFile))
		}
	}))
	defer ts.Close()

	tr := entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	}

	r := NewClient(testTimeout())
	pr, err := r.Project(context.Background(), tr, 1)
	if pr == nil {
		t.Error("Should return project")
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestProjectsActivityErr(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/projects") {
			w.Write(readTestFile(t, projectsFile))
		}
		if strings.HasPrefix(r.URL.Path, "/enumerations/time_entry_activities.json") {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer ts.Close()

	page := entities.Pagination{
		Offset: 41,
		Limit:  88,
	}

	tr := entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	}

	r := NewClient(testTimeout())
	_, _, err := r.Projects(context.Background(), tr, page)
	if err == nil {
		t.Fatal(err)
	}
}

func TestProjectErr(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	tr := entities.Tracker{
		Credentials: testCreds,
		URL:         ts.URL,
		Type:        redmineType,
	}

	r := NewClient(testTimeout())
	_, err := r.Project(context.Background(), tr, 1)
	if err == nil {
		t.Fatal(err)
	}
}

func TestDateToSeconds(t *testing.T) {
	// Zero on empty string
	if dateToSeconds("") != 0 {
		t.Error("invalid result")
	}
	// Zero on invalid format
	if dateToSeconds("one") != 0 {
		t.Error("invalid result")
	}
}

func TestExternalSvcErr(t *testing.T) {
	// Error on invalid json
	assert.Error(t, toExternalServiceErr([]byte("123")))
}

func assertErr(t *testing.T, actual, expected error) {
	if errors.Cause(actual) != expected {
		t.Error("Unexpected error", actual)
	}
}

func testTimeout() time.Duration {
	return 10 * time.Second
}

func readTestFile(t *testing.T, path string) []byte {
	b, err := ioutil.ReadFile(filepath.Join("var", "testdata", path))
	if err != nil {
		pwd, err2 := os.Getwd()
		t.Fatal(err, pwd, err2)
	}
	return b
}

func unmarshal(t *testing.T, body io.ReadCloser, target interface{}) {
	defer body.Close()
	if err := json.NewDecoder(body).Decode(target); err != nil {
		t.Fatalf("failed to unmarshal issue")
	}
}

func idNameByID(idNames []idName, id int64) *idName {
	for _, v := range idNames {
		if v.ID == id {
			return &v
		}
	}
	return nil
}

func idNameByName(idNames []idName, name string) *idName {
	for _, v := range idNames {
		if v.Name == name {
			return &v
		}
	}
	return nil
}

func names(idNames []idName) []string {
	var sstr []string
	for _, v := range idNames {
		sstr = append(sstr, v.Name)
	}
	return sstr
}

var userJSON = entities.User{
	ID:   1131,
	Name: "Prylutskyi Anatolii",
	Mail: "prylutskyi@qarea.com",
}
var issueJSON = &entities.Issue{
	ID:          71307,
	DueDate:     1465862400,
	ProjectID:   223,
	Description: "https://docs.google.com/document",
	Type: entities.TypeID{
		ID:   19,
		Name: "Task",
	},
	Title:    "Develop Redmine Tracker Adapter MS",
	Estimate: 86400,
	Done:     10,
	Spent:    10 * 60 * 60,
}

var issuesJSON = []entities.Issue{
	{
		ID:        71307,
		DueDate:   1465862400,
		ProjectID: 223,
		Type: entities.TypeID{
			ID:   19,
			Name: "Task",
		},
		Title:       "Develop Redmine Tracker Adapter MS",
		Description: "https://docs.google.com/document/d",
		Estimate:    86400,
		Done:        20,
	},
	{
		ID:        71306,
		DueDate:   1465430400,
		ProjectID: 223,
		Type: entities.TypeID{
			ID:   19,
			Name: "Task",
		},
		Title:       "Review current architectural document and confirm that all technical side is correct and can be developed ",
		Description: "Review current architectural document and confirm that all",
		Estimate:    21600,
		Done:        55,
	},
}

var projectJSON = entities.Project{
	ID:            170,
	Title:         "Internal",
	Description:   "Project170",
	IssueTypes:    issueTypesJSON,
	ActivityTypes: []entities.TypeID{},
}

var projectsJSON = []entities.Project{
	{
		ID:            170,
		Title:         "Internal",
		Description:   "Project170",
		IssueTypes:    issueTypesJSON,
		ActivityTypes: []entities.TypeID{},
	},
	{
		ID:            577,
		Title:         "Bench",
		Description:   "Project570",
		IssueTypes:    issueTypesJSON,
		ActivityTypes: []entities.TypeID{},
	},
	{
		ID:            379,
		Title:         "Education",
		Description:   "Project370",
		IssueTypes:    issueTypesJSON,
		ActivityTypes: []entities.TypeID{},
	},
	{
		ID:            987,
		Title:         "S-match",
		Description:   "",
		IssueTypes:    issueTypesJSON,
		ActivityTypes: []entities.TypeID{},
	},
	{
		ID:            223,
		Title:         "TimeGuard",
		Description:   "TimeGuard",
		IssueTypes:    issueTypesJSON,
		ActivityTypes: []entities.TypeID{},
	},
}

var issueTypesJSON = []entities.TypeID{
	{1, "Bug"},
	{2, "Feature"},
	{3, "Support"},
	{4, "Administration"},
	{5, "Management"},
	{6, "Quality Control"},
	{7, "Planning"},
	{8, "Business analysis"},
	{9, "Technical writing"},
	{10, "Design"},
	{11, "Change Request"},
	{12, "Interview"},
	{13, "Estimation"},
	{14, "Question"},
	{15, "Vacancy"},
	{16, "Risk"},
	{17, "User Story"},
	{19, "Task"},
	{20, "Unit test"},
	{21, "Group"},
	{22, "Pseudo task"},
}

var activityTypesJSON = []entities.TypeID{
	{8, "Design"},
	{9, "Development"},
	{10, "Analysis"},
	{11, "Testing"},
	{12, "Management"},
	{13, "Administration"},
}
