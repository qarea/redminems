package redmine

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/qarea/redminems/entities"
)

var testCredentials = entities.Credentials{
	Login:    "login1",
	Password: "password2",
}

func TestRedmineRequestErr(t *testing.T) {
	type test struct {
		tracker            entities.Tracker
		statusCode         int
		validateStatusFunc func(int) error
		method             string
		body               []byte
		resultExpected     interface{}
		errorExpected      bool
		err                error
	}
	var tests = map[string]test{
		"Redmine unprocessable entry response test": {
			tracker:    testTracker(),
			statusCode: 422,
			method:     "GET",
			body: []byte(`{
				"errors": [
				"err1",
				"err2"
				]
			}`),
			errorExpected: true,
			err:           entities.NewTrackerValidationErr("err1. err2"),
		},
		"Invalid credentials": {
			tracker:       testTracker(),
			errorExpected: true,
			err:           entities.ErrCredentials,
			statusCode:    http.StatusUnauthorized,
			method:        "GET",
		},
		"Not found": {
			tracker:       testTracker(),
			errorExpected: true,
			err:           errNotFound,
			statusCode:    http.StatusNotFound,
			method:        "GET",
		},
		"Invalid JSON": {
			tracker:            testTracker(),
			errorExpected:      true,
			statusCode:         http.StatusOK,
			resultExpected:     map[string]interface{}{},
			body:               []byte("invalid json"),
			method:             "GET",
			validateStatusFunc: validateStatusOK,
		},
		"We do not expect result": {
			tracker:       testTracker(),
			errorExpected: false,
			statusCode:    http.StatusOK,
			body:          []byte("invalid json"),
			method:        "GET",
		},
		"Internal service error": {
			tracker:       testTracker(),
			errorExpected: true,
			err:           entities.ErrRemoteServer,
			statusCode:    http.StatusInternalServerError,
			method:        "POST",
		},
		"Validate function test": {
			tracker:            testTracker(),
			errorExpected:      true,
			statusCode:         http.StatusCreated,
			method:             "POST",
			validateStatusFunc: validateStatusOK,
		},
		"Invalid tracker type": {
			tracker:       entities.Tracker{Type: "JIRA"},
			errorExpected: true,
			err:           entities.ErrTrackerType,
			method:        "GET",
		},
		"Forbidden": {
			tracker:       testTracker(),
			errorExpected: true,
			statusCode:    http.StatusForbidden,
			err:           entities.ErrForbidden,
			method:        "GET",
		},
	}

	for label, test := range tests {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != test.method {
				t.Errorf("Test %s, invalid HTTP method %s", label, r.Method)
			}
			login, pass, ok := r.BasicAuth()
			if !ok {
				t.Errorf("Basic auth header is missing")
			}
			if login != testCredentials.Login && pass != testCredentials.Password {
				t.Errorf("Test %s. Login or/and password invalid, expected: (%s, %s), actual: (%s, %s)", label, testCredentials.Login, testCredentials.Password, login, pass)
			}
			w.WriteHeader(test.statusCode)
			if test.body != nil {
				w.Write(test.body)
			}
		}))

		tr := test.tracker
		tr.URL = ts.URL
		err := redmineRequest(requestOpts{
			httpClient:         http.DefaultClient,
			ctx:                context.Background(),
			result:             test.resultExpected,
			tracker:            tr,
			method:             test.method,
			body:               test.body,
			validateStatusFunc: test.validateStatusFunc,
		})

		if test.errorExpected && err == nil {
			t.Errorf("Error expected for test %s", label)
		}
		if !test.errorExpected && err != nil {
			t.Errorf("Unexpected error %+v, label %s", err, label)
		}

		if test.err != nil && !reflect.DeepEqual(test.err, errors.Cause(err)) {
			t.Errorf("Another error was expected for test %s, actual error: %v", label, test.err)
		}

		ts.Close()
	}
}

func TestRedmineRequestTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer ts.Close()
	tr := testTracker()
	tr.URL = ts.URL

	err := redmineRequest(requestOpts{
		httpClient: &http.Client{
			Timeout: 50 * time.Millisecond,
		},
		ctx:     context.Background(),
		tracker: tr,
		method:  get,
	})
	if err == nil {
		t.Errorf("Error expected")
	}
	if errors.Cause(err) != entities.ErrTimeout {
		t.Errorf("Timeout err expected actual: %+v", err)
	}
}

func TestRedmineRequestContextTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		http.Error(w, "err", 500)
	}))
	defer ts.Close()
	tr := testTracker()
	tr.URL = ts.URL
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := redmineRequest(requestOpts{
		httpClient: http.DefaultClient,
		ctx:        ctx,
		tracker:    tr,
		method:     get,
	})
	if err == nil {
		t.Error("Should be err")
	}
	if errors.Cause(err) != context.DeadlineExceeded {
		t.Errorf("Timeout err expected actual: %+v %T", err, err)
	}
}

func TestRemoveLastSlash(t *testing.T) {
	type test struct {
		given    string
		expected string
	}
	tests := map[string]test{
		"Empty string": {
			given:    "",
			expected: "",
		},
		"URL with last slash": {
			given:    "qarea.com/",
			expected: "qarea.com",
		},
		"Slash only string": {
			given:    "/",
			expected: "",
		},
		"URL without last slash": {
			given:    "qarea.com",
			expected: "qarea.com",
		},
	}
	for label, test := range tests {
		if !reflect.DeepEqual(test.expected, removeLastSlash(test.given)) {
			t.Errorf("Test %s failed", label)
		}
	}
}

func testTracker() entities.Tracker {
	return entities.Tracker{
		Type:        redmineType,
		Credentials: testCredentials,
	}
}

func TestIDFromURL(t *testing.T) {
	type test struct {
		url         entities.IssueURL
		expectedID  entities.IssueID
		errExpected bool
	}
	tests := []test{
		{
			url:        "/issues/1",
			expectedID: 1,
		},
		{
			url:        "/issues/1/",
			expectedID: 1,
		},
		{
			url:        "/issues/123",
			expectedID: 123,
		},
		{
			url:         "/issues/",
			errExpected: true,
		},
		{
			url:         "/123",
			errExpected: true,
		},
		{
			url:         "/issues/sadasd",
			errExpected: true,
		},
	}

	for i, test := range tests {
		id, err := issueIDFromURL(test.url)
		if test.errExpected && err == nil {
			t.Errorf("Test %d. Errow was expected", i)
		}
		if test.expectedID != id {
			t.Errorf("Test %d. Unexpected ID expected %v | actual %v", i, test.expectedID, id)
		}
	}
}

func TestSecondsToHours(t *testing.T) {
	h := secondsToHours(1800)
	if h != 0.5 {
		t.Error("Invalid hour value", h)
	}
}
func TestAddActivities(t *testing.T) {
	activities := timeEntryActivitiesRoot{
		TimeEntryActivities: []idName{
			{1, "dev"},
			{2, "test"},
		},
	}

	projects := []entities.Project{
		{ID: 0},
		{ID: 1},
		{ID: 2},
	}

	addActivities(projects, activities)
	for _, p := range projects {
		if len(p.ActivityTypes) != len(activities.TimeEntryActivities) {
			t.Errorf("Invalid activity amount per project")
		}
		for i, ac := range p.ActivityTypes {
			if ac.ID != activities.TimeEntryActivities[i].ID ||
				ac.Name != activities.TimeEntryActivities[i].Name {
				t.Errorf("Invalid activity conversion")
			}
		}
	}
}

func TestAddLinks(t *testing.T) {
	projects := []entities.Project{
		{ID: 0},
		{ID: 1},
		{ID: 2},
	}
	trackerURL := "http://tracker.com"
	tr := entities.Tracker{URL: trackerURL}
	addLinks(projects, tr)
	for _, p := range projects {
		expectedURL := fmt.Sprintf("%s/projects/%d", trackerURL, p.ID)
		if p.Link != expectedURL {
			t.Errorf("Invalid Project URL formed %v", p.Link)
		}
	}
}
