package commands

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/openaustralia/morph-ng/pkg/jobdispatcher"
	"github.com/openaustralia/morph-ng/pkg/keyvaluestore"
	"github.com/openaustralia/morph-ng/pkg/stream"
)

func TestStoragePath(t *testing.T) {
	assert.Equal(t, storagePath("abc", "app.tgz"), "abc/app.tgz")
	assert.Equal(t, storagePath("def", "output"), "def/output")
}

func TestStartRun(t *testing.T) {
	job := new(jobdispatcher.MockClient)
	keyValueStore := new(keyvaluestore.MockClient)

	// Expect that the job will get dispatched
	job.On(
		"StartJob",
		"run-name",
		"openaustralia/clay-scraper:v1",
		[]string{"/bin/run.sh", "run-name", "output.txt"},
		map[string]string{"FOO": "bar", "CLAY_INTERNAL_RUN_TOKEN": "supersecret"},
	).Return(nil)
	// Expect that we'll need the secret token
	job.On("GetToken", "run-name").Return("supersecret", nil)
	// Expect that we save the callback url in the key value store
	keyValueStore.On("Set", "url:run-name", "http://foo.com").Return(nil)

	app := App{JobDispatcher: job, KeyValueStore: keyValueStore}
	// TODO: Pass an options struct instead (we get named parameters effectively then)
	err := app.StartRun(
		"run-name",                      // Run name
		"output.txt",                    // Output filename
		map[string]string{"FOO": "bar"}, // Environment variables
		"http://foo.com",                // Callback URL
	)
	assert.Nil(t, err)

	job.AssertExpectations(t)
	keyValueStore.AssertExpectations(t)
}

// Make sure that setting a reserved environment variable is not allowed
func TestStartRunWithReservedEnv(t *testing.T) {
	app := App{}
	err := app.StartRun(
		"run-name",   // Run name
		"output.txt", // Output filename
		map[string]string{"CLAY_INTERNAL_FOO": "bar"}, // Environment variables
		"", // Callback URL
	)
	assert.EqualError(t, err, "Can't override environment variables starting with CLAY_INTERNAL_")
}

type MockRoundTripper struct {
	mock.Mock
}

func (m *MockRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	args := m.Called(r)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestCreateEvent(t *testing.T) {
	stream := new(stream.MockClient)
	keyValueStore := new(keyvaluestore.MockClient)

	stream.On("Add", "run-name", "{\"some\": \"json\"}").Return(nil)
	keyValueStore.On("Get", "url:run-name").Return("http://foo.com/bar", nil)

	// Mock out the http RoundTripper so that no actual http request is made
	httpClient := defaultHTTP()
	roundTripper := new(MockRoundTripper)
	roundTripper.On("RoundTrip", mock.MatchedBy(func(r *http.Request) bool {
		return r.URL.String() == "http://foo.com/bar"
	})).Return(
		&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader("")),
		},
		nil,
	)
	httpClient.Transport = roundTripper

	app := App{Stream: stream, KeyValueStore: keyValueStore, HTTP: httpClient}
	err := app.CreateEvent("run-name", "{\"some\": \"json\"}")
	assert.Nil(t, err)

	stream.AssertExpectations(t)
	keyValueStore.AssertExpectations(t)
	roundTripper.AssertExpectations(t)
}

func TestCreateEventNoCallbackURL(t *testing.T) {
	stream := new(stream.MockClient)
	keyValueStore := new(keyvaluestore.MockClient)

	stream.On("Add", "run-name", "{\"some\": \"json\"}").Return(nil)
	keyValueStore.On("Get", "url:run-name").Return("", nil)

	// Mock out the http RoundTripper so that no actual http request is made
	httpClient := defaultHTTP()
	roundTripper := new(MockRoundTripper)
	httpClient.Transport = roundTripper

	app := App{Stream: stream, KeyValueStore: keyValueStore, HTTP: httpClient}
	err := app.CreateEvent("run-name", "{\"some\": \"json\"}")
	assert.Nil(t, err)

	stream.AssertExpectations(t)
	keyValueStore.AssertExpectations(t)
	roundTripper.AssertNotCalled(t, "RoundTrip")
}
