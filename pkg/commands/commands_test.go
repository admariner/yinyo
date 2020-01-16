package commands

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	blobstoremocks "github.com/openaustralia/yinyo/mocks/pkg/blobstore"
	jobdispatchermocks "github.com/openaustralia/yinyo/mocks/pkg/jobdispatcher"
	keyvaluestoremocks "github.com/openaustralia/yinyo/mocks/pkg/keyvaluestore"
	streammocks "github.com/openaustralia/yinyo/mocks/pkg/stream"
	"github.com/openaustralia/yinyo/pkg/keyvaluestore"
	"github.com/openaustralia/yinyo/pkg/protocol"
)

func TestStoragePath(t *testing.T) {
	assert.Equal(t, blobStoreStoragePath("abc", "app.tgz"), "abc/app.tgz")
	assert.Equal(t, blobStoreStoragePath("def", "output"), "def/output")
}

func TestStartRun(t *testing.T) {
	job := new(jobdispatchermocks.Jobs)
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)
	blobStore := new(blobstoremocks.BlobStore)

	// Expect that the job will get dispatched
	job.On(
		"StartJob",
		"run-name",
		"openaustralia/yinyo-scraper:v1",
		[]string{"/bin/yinyo", "wrapper", "run-name", "supersecret", "--output", "output.txt", "--env", "FOO=bar"},
		int64(86400),
	).Return(nil)
	// Expect that we'll need the secret token
	keyValueStore.On("Get", "run-name/token").Return("supersecret", nil)
	// Expect that we save the callback url in the key value store
	keyValueStore.On("Set", "run-name/url", "http://foo.com").Return(nil)
	// Expect that we try to get the code just to see if it exists
	blobStore.On("Get", "run-name/app.tgz").Return(nil, nil)

	app := AppImplementation{JobDispatcher: job, KeyValueStore: keyValueStore, BlobStore: blobStore}
	// TODO: Pass an options struct instead (we get named parameters effectively then)
	err := app.StartRun(
		"run-name",                      // Run name
		"output.txt",                    // Output filename
		map[string]string{"FOO": "bar"}, // Environment variables
		"http://foo.com",                // Callback URL
		86400,                           // Max run time
	)
	assert.Nil(t, err)

	job.AssertExpectations(t)
	keyValueStore.AssertExpectations(t)
	blobStore.AssertExpectations(t)
}

type MockRoundTripper struct {
	mock.Mock
}

func (m *MockRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	args := m.Called(r)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestCreateEvent(t *testing.T) {
	stream := new(streammocks.Stream)
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)

	time := time.Now()
	stream.On("Add", "run-name", protocol.NewStartEvent("", time, "build")).Return(protocol.NewStartEvent("123", time, "build"), nil)
	keyValueStore.On("Get", "run-name/url").Return("http://foo.com/bar", nil)

	// Mock out the http RoundTripper so that no actual http request is made
	httpClient := http.DefaultClient
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

	app := AppImplementation{Stream: stream, KeyValueStore: keyValueStore, HTTP: httpClient}
	err := app.CreateEvent("run-name", protocol.NewStartEvent("", time, "build"))
	assert.Nil(t, err)

	stream.AssertExpectations(t)
	keyValueStore.AssertExpectations(t)
	roundTripper.AssertExpectations(t)
}

func TestCreateEventNoCallbackURL(t *testing.T) {
	stream := new(streammocks.Stream)
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)

	time := time.Now()
	stream.On("Add", "run-name", protocol.NewStartEvent("", time, "build")).Return(protocol.NewStartEvent("123", time, "build"), nil)
	keyValueStore.On("Get", "run-name/url").Return("", nil)

	// Mock out the http RoundTripper so that no actual http request is made
	httpClient := http.DefaultClient
	roundTripper := new(MockRoundTripper)
	httpClient.Transport = roundTripper

	app := AppImplementation{Stream: stream, KeyValueStore: keyValueStore, HTTP: httpClient}
	err := app.CreateEvent("run-name", protocol.NewStartEvent("", time, "build"))
	assert.Nil(t, err)

	stream.AssertExpectations(t)
	keyValueStore.AssertExpectations(t)
	roundTripper.AssertNotCalled(t, "RoundTrip")
}

func TestCreateEventErrorDuringCallback(t *testing.T) {
	stream := new(streammocks.Stream)
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)

	time := time.Now()
	stream.On("Add", "run-name", protocol.NewStartEvent("", time, "build")).Return(protocol.NewStartEvent("123", time, "build"), nil)
	keyValueStore.On("Get", "run-name/url").Return("http://foo.com/bar", nil)

	// Mock out the http RoundTripper so that no actual http request is made
	httpClient := http.DefaultClient
	roundTripper := new(MockRoundTripper)
	roundTripper.On("RoundTrip", mock.MatchedBy(func(r *http.Request) bool {
		return r.URL.String() == "http://foo.com/bar"
	})).Return(
		&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader("")),
		},
		errors.New("An error while doing the postback"),
	)
	httpClient.Transport = roundTripper

	app := AppImplementation{Stream: stream, KeyValueStore: keyValueStore, HTTP: httpClient}
	err := app.CreateEvent("run-name", protocol.NewStartEvent("", time, "build"))
	assert.EqualError(t, err, "Post http://foo.com/bar: An error while doing the postback")

	stream.AssertExpectations(t)
	keyValueStore.AssertExpectations(t)
	roundTripper.AssertExpectations(t)
}

func TestGetEvents(t *testing.T) {
	stream := new(streammocks.Stream)

	time := time.Now()
	stream.On("Get", "run-name", "0").Return(protocol.NewStartEvent("123", time, "build"), nil)
	stream.On("Get", "run-name", "123").Return(protocol.NewLastEvent("456", time), nil)

	app := AppImplementation{Stream: stream}

	events := app.GetEvents("run-name", "0")

	// We're expecting two events in the stream. Let's hardcode what would normally be in a loop
	assert.True(t, events.More())
	e, err := events.Next()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, protocol.NewStartEvent("123", time, "build"), e)
	assert.True(t, events.More())
	e, err = events.Next()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, protocol.NewLastEvent("456", time), e)
	assert.False(t, events.More())

	stream.AssertExpectations(t)
}

func TestDeleteRun(t *testing.T) {
	jobDispatcher := new(jobdispatchermocks.Jobs)
	blobStore := new(blobstoremocks.BlobStore)
	stream := new(streammocks.Stream)
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)

	jobDispatcher.On("DeleteJobAndToken", "run-name").Return(nil)
	blobStore.On("Delete", "run-name/app.tgz").Return(nil)
	blobStore.On("Delete", "run-name/output").Return(nil)
	blobStore.On("Delete", "run-name/cache.tgz").Return(nil)
	stream.On("Delete", "run-name").Return(nil)
	keyValueStore.On("Delete", "run-name/url").Return(nil)
	keyValueStore.On("Delete", "run-name/token").Return(nil)
	keyValueStore.On("Delete", "run-name/exit_data").Return(nil)

	app := AppImplementation{
		JobDispatcher: jobDispatcher,
		BlobStore:     blobStore,
		Stream:        stream,
		KeyValueStore: keyValueStore,
	}
	err := app.DeleteRun("run-name")
	assert.Nil(t, err)

	jobDispatcher.AssertExpectations(t)
	blobStore.AssertExpectations(t)
	stream.AssertExpectations(t)
	keyValueStore.AssertExpectations(t)
}

func TestTokenCacheNotFound(t *testing.T) {
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)
	keyValueStore.On("Get", "does-not-exit/token").Return("", keyvaluestore.ErrKeyNotExist)

	app := AppImplementation{KeyValueStore: keyValueStore}
	// This run name should not exist
	_, err := app.GetTokenCache("does-not-exit")
	assert.True(t, errors.Is(err, ErrNotFound))

	keyValueStore.AssertExpectations(t)
}

func TestPutApp(t *testing.T) {
	blobStore := new(blobstoremocks.BlobStore)
	app := AppImplementation{BlobStore: blobStore}

	blobStore.On("Put", "run-name/app.tgz", mock.Anything, mock.Anything).Return(nil)

	// Open a file which has the simplest possible archive which is empty but valid
	file, _ := os.Open("testdata/empty.tgz")
	defer file.Close()
	stat, _ := file.Stat()

	err := app.PutApp(file, stat.Size(), "run-name")
	if err != nil {
		t.Fatal(err)
	}

	blobStore.AssertExpectations(t)
}

func TestGetCache(t *testing.T) {
	blobStore := new(blobstoremocks.BlobStore)
	app := AppImplementation{BlobStore: blobStore}

	file, _ := os.Open("testdata/empty.tgz")
	defer file.Close()

	blobStore.On("Get", "run-name/cache.tgz").Return(file, nil)

	r, err := app.GetCache("run-name")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, file, r)

	blobStore.AssertExpectations(t)
}

func TestPutCache(t *testing.T) {
	blobStore := new(blobstoremocks.BlobStore)
	app := AppImplementation{BlobStore: blobStore}

	file, _ := os.Open("testdata/empty.tgz")
	stat, _ := file.Stat()

	blobStore.On("Put", "run-name/cache.tgz", mock.Anything, stat.Size()).Return(nil)

	err := app.PutCache(file, stat.Size(), "run-name")
	if err != nil {
		t.Fatal(err)
	}

	blobStore.AssertExpectations(t)
}

func TestGetOutput(t *testing.T) {
	blobStore := new(blobstoremocks.BlobStore)
	app := AppImplementation{BlobStore: blobStore}

	blobStore.On("Get", "run-name/output").Return(strings.NewReader("output"), nil)

	r, err := app.GetOutput("run-name")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := ioutil.ReadAll(r)

	assert.Equal(t, "output", string(b))
	blobStore.AssertExpectations(t)
}

func TestPutOutput(t *testing.T) {
	blobStore := new(blobstoremocks.BlobStore)
	app := AppImplementation{BlobStore: blobStore}

	reader := strings.NewReader("output")
	blobStore.On("Put", "run-name/output", reader, int64(6)).Return(nil)

	err := app.PutOutput(reader, 6, "run-name")
	if err != nil {
		t.Fatal(err)
	}
	blobStore.AssertExpectations(t)
}

func TestGetExitData(t *testing.T) {
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)
	app := AppImplementation{KeyValueStore: keyValueStore}

	keyValueStore.On("Get", "run-name/exit_data").Return(`{"build":{"exit_code":15,"usage":{"wall_time":0,"cpu_time":0,"max_rss":0,"network_in":0,"network_out":0}}}`, nil)
	e, err := app.GetExitData("run-name")
	if err != nil {
		t.Fatal(err)
	}
	expectedExitData := protocol.ExitData{Build: &protocol.ExitDataStage{ExitCode: 15}}

	assert.Equal(t, expectedExitData, e)
	keyValueStore.AssertExpectations(t)
}

func TestPutExitData(t *testing.T) {
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)
	app := AppImplementation{KeyValueStore: keyValueStore}

	keyValueStore.On("Set", "run-name/exit_data", `{"build":{"exit_code":15,"usage":{"wall_time":0,"cpu_time":0,"max_rss":0,"network_in":0,"network_out":0}}}`).Return(nil)
	exitData := protocol.ExitData{Build: &protocol.ExitDataStage{ExitCode: 15}}

	err := app.PutExitData("run-name", exitData)
	if err != nil {
		t.Fatal(err)
	}
	keyValueStore.AssertExpectations(t)
}

func TestCreateRun(t *testing.T) {
	jobDispatcher := new(jobdispatchermocks.Jobs)
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)
	app := AppImplementation{JobDispatcher: jobDispatcher, KeyValueStore: keyValueStore}

	jobDispatcher.On("CreateJobAndToken", "run", mock.Anything).Return("run-foo", nil)
	keyValueStore.On("Set", "run-foo/token", mock.Anything).Return(nil)

	run, err := app.CreateRun("")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "run-foo", run.Name)
	jobDispatcher.AssertExpectations(t)
	keyValueStore.AssertExpectations(t)
}

// If we haven't uploaded an app error when starting a run
func TestStartNoApp(t *testing.T) {
	blobstoreClient := new(blobstoremocks.BlobStore)
	app := AppImplementation{BlobStore: blobstoreClient}

	blobstoreClient.On("Get", "foo/app.tgz").Return(nil, errors.New("Doesn't exist"))
	blobstoreClient.On("IsNotExist", errors.New("Doesn't exist")).Return(true)

	err := app.StartRun("foo", "", map[string]string{}, "", 0)
	assert.True(t, errors.Is(err, ErrAppNotAvailable))

	blobstoreClient.AssertExpectations(t)
}