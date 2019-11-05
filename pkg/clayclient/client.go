package clayclient

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Run is what you get when you create a run and what you need to update it
type Run struct {
	Name  string `json:"run_name"`
	Token string `json:"run_token"`
	// Ignore this field when converting from/to json
	Client *Client
}

// Client is used to access the API
type Client struct {
	URL        string
	HTTPClient *http.Client
}

// New configures a new Client
func New(URL string) *Client {
	return &Client{
		URL:        URL,
		HTTPClient: http.DefaultClient,
	}
}

func checkOK(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return errors.New(resp.Status)
}

// IsNotFound checks whether a particular error message corresponds to a 404
func IsNotFound(err error) bool {
	// TODO: Don't want to depend on a hardcoded string here
	return (err.Error() == "404 Not Found")
}

func checkContentType(resp *http.Response, expected string) error {
	ct := resp.Header["Content-Type"]
	if len(ct) == 1 && ct[0] == expected {
		return nil
	}
	return errors.New("Unexpected content type")
}

// Hello does a simple ping type request to the API
func (client *Client) Hello() (string, error) {
	req, err := http.NewRequest("GET", client.URL, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	if err = checkOK(resp); err != nil {
		return "", err
	}
	if err = checkContentType(resp, "text/plain; charset=utf-8"); err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CreateRun is the first thing called. It creates a run
func (client *Client) CreateRun(namePrefix string) (Run, error) {
	run := Run{Client: client}

	uri := client.URL + "/runs"
	if namePrefix != "" {
		params := url.Values{}
		params.Add("name_prefix", namePrefix)
		uri += "?" + params.Encode()
	}
	req, err := http.NewRequest("POST", uri, nil)
	if err != nil {
		return run, err
	}

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return run, err
	}
	if err = checkOK(resp); err != nil {
		return run, err
	}
	if err = checkContentType(resp, "application/json"); err != nil {
		return run, err
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&run)
	return run, err
}

// Make an API call for a particular run. These requests are always authenticated
func (run *Run) request(method string, path string, body io.Reader) (*http.Response, error) {
	url := run.Client.URL + fmt.Sprintf("/runs/%s", run.Name) + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+run.Token)
	return run.Client.HTTPClient.Do(req)
}

// ExtractArchiveToDirectory takes a tar, gzipped archive and extracts it to a directory on the filesystem
func ExtractArchiveToDirectory(gzipTarContent io.ReadCloser, dir string) error {
	gzipReader, err := gzip.NewReader(gzipTarContent)
	if err != nil {
		return err
	}
	tarReader := tar.NewReader(gzipReader)
	for {
		file, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}
		switch file.Typeflag {
		case tar.TypeDir:
			// TODO: Extract variable
			err := os.Mkdir(filepath.Join(dir, file.Name), 0755)
			if err != nil {
				return err
			}
		case tar.TypeReg:
			f, err := os.OpenFile(
				filepath.Join(dir, file.Name),
				os.O_RDWR|os.O_CREATE|os.O_TRUNC,
				file.FileInfo().Mode(),
			)
			if err != nil {
				return err
			}
			io.Copy(f, tarReader)
			f.Close()
		case tar.TypeSymlink:
			newname := filepath.Join(dir, file.Name)
			oldname := filepath.Join(filepath.Dir(newname), file.Linkname)
			err = os.Symlink(oldname, newname)
			if err != nil {
				return err
			}
		default:
			return errors.New("Unexpected type in tar")
		}
	}
	return nil
}

// CreateArchiveFromDirectory creates an archive from a directory on the filesystem
func CreateArchiveFromDirectory(dir string) (io.Reader, error) {
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == dir {
			return nil
		}
		relativePath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		var link string
		if info.Mode()&os.ModeSymlink != 0 {
			link, err = os.Readlink(path)
			if err != nil {
				return err
			}
			if filepath.IsAbs(link) {
				// Convert the absolute link to a relative link
				absPath, err := filepath.Abs(path)
				if err != nil {
					return err
				}
				d := filepath.Dir(absPath)
				link, err = filepath.Rel(d, link)
				if err != nil {
					return err
				}
			}
		}
		header, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return err
		}
		header.Name = relativePath
		tarWriter.WriteHeader(header)

		// If it's a regular file then write the contents
		if info.Mode().IsRegular() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			io.Copy(tarWriter, f)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	// TODO: This should always get called
	tarWriter.Close()
	gzipWriter.Close()
	return &buffer, nil
}

// GetAppToDirectory downloads the scraper code into a pre-existing directory on the filesystem
func (run *Run) GetAppToDirectory(dir string) error {
	app, err := run.GetApp()
	if err != nil {
		return err
	}
	defer app.Close()
	return ExtractArchiveToDirectory(app, dir)
}

// PutAppFromDirectory uploads the scraper code from a directory on the filesystem
func (run *Run) PutAppFromDirectory(dir string) error {
	r, err := CreateArchiveFromDirectory(dir)
	if err != nil {
		return err
	}
	return run.PutApp(r)
}

// GetCacheToDirectory downloads the cache into a pre-existing directory on the filesystem
func (run *Run) GetCacheToDirectory(dir string) error {
	app, err := run.GetCache()
	if err != nil {
		// If cache doesn't exist then do nothing
		if IsNotFound(err) {
			return nil
		}
		return err
	}
	defer app.Close()
	return ExtractArchiveToDirectory(app, dir)
}

// PutCacheFromDirectory uploads the cache from a directory on the filesystem
func (run *Run) PutCacheFromDirectory(dir string) error {
	r, err := CreateArchiveFromDirectory(dir)
	if err != nil {
		return err
	}
	return run.PutCache(r)
}

// GetApp downloads the tarred & gzipped scraper code
func (run *Run) GetApp() (io.ReadCloser, error) {
	resp, err := run.request("GET", "/app", nil)
	if err != nil {
		return nil, err
	}
	if err = checkOK(resp); err != nil {
		return nil, err
	}
	if err = checkContentType(resp, "application/gzip"); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// PutApp uploads the tarred & gzipped scraper code
func (run *Run) PutApp(appData io.Reader) error {
	resp, err := run.request("PUT", "/app", appData)
	if err != nil {
		return err
	}
	return checkOK(resp)
}

// PutCache uploads the tarred & gzipped build cache
func (run *Run) PutCache(data io.Reader) error {
	resp, err := run.request("PUT", "/cache", data)
	if err != nil {
		return err
	}
	return checkOK(resp)
}

// PutOutput uploads the output of the scraper
func (run *Run) PutOutput(data io.Reader) error {
	resp, err := run.request("PUT", "/output", data)
	if err != nil {
		return err
	}
	return checkOK(resp)
}

// GetCache downloads the tarred & gzipped build cache
func (run *Run) GetCache() (io.ReadCloser, error) {
	resp, err := run.request("GET", "/cache", nil)
	if err != nil {
		return nil, err
	}
	if err = checkOK(resp); err != nil {
		return nil, err
	}
	if err = checkContentType(resp, "application/gzip"); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// StartRunOptions are options that can be used when starting a run
type StartRunOptions struct {
	Output string
}

// Start starts a run that has earlier been created
// TODO: Add setting of environment variables
func (run *Run) Start(options *StartRunOptions) error {
	b, err := json.Marshal(options)
	if err != nil {
		return err
	}
	resp, err := run.request("POST", "/start", bytes.NewReader(b))
	if err != nil {
		return err
	}
	return checkOK(resp)
}

type eventRaw struct {
	Stage  string `json:"stage"`
	Type   string `json:"type"`
	Stream string `json:"stream,omitempty"`
	Text   string `json:"text,omitempty"`
}

// Event is the interface for all event types
type Event interface {
}

// StartEvent represents the start of a build or run
type StartEvent struct {
	Stage string
}

// FinishEvent represent the completion of a build or run
type FinishEvent struct {
	Stage string
}

// LogEvent is the output of some text from the build or run of a scraper
type LogEvent struct {
	Stage  string
	Stream string
	Text   string
}

// EventIterator is a stream of events
type EventIterator struct {
	decoder *json.Decoder
}

// MarshalJSON converts a StartEvent to JSON
func (e StartEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(eventRaw{Type: "started", Stage: e.Stage})
}

// MarshalJSON converts a StartEvent to JSON
func (e FinishEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(eventRaw{Type: "finished", Stage: e.Stage})
}

// MarshalJSON converts a StartEvent to JSON
func (e LogEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(eventRaw{Type: "log", Stage: e.Stage, Stream: e.Stream, Text: e.Text})
}

// More checks whether another event is available
func (iterator *EventIterator) More() bool {
	return iterator.decoder.More()
}

// Next returns the next event
func (iterator *EventIterator) Next() (Event, error) {
	var eventRaw eventRaw
	err := iterator.decoder.Decode(&eventRaw)
	if err != nil {
		return nil, err
	}
	if eventRaw.Type == "started" {
		return StartEvent{Stage: eventRaw.Stage}, nil
	} else if eventRaw.Type == "finished" {
		return FinishEvent{Stage: eventRaw.Stage}, nil
	} else if eventRaw.Type == "log" {
		return LogEvent{Stage: eventRaw.Stage, Stream: eventRaw.Stream, Text: eventRaw.Text}, nil
	}
	return nil, errors.New("Unexpected type")
}

// GetEvents returns a stream of events from the API
func (run *Run) GetEvents() (*EventIterator, error) {
	resp, err := run.request("GET", "/events", nil)
	if err != nil {
		return nil, err
	}
	if err = checkOK(resp); err != nil {
		return nil, err
	}
	if err = checkContentType(resp, "application/ld+json"); err != nil {
		return nil, err
	}
	return &EventIterator{decoder: json.NewDecoder(resp.Body)}, nil
}

// CreateEvent sends an event
func (run *Run) CreateEvent(event Event) error {
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}
	resp, err := run.request("POST", "/events", bytes.NewReader(b))
	if err != nil {
		return err
	}
	return checkOK(resp)
}

// CreateLastEvent sends a special message to close the stream
// TODO: Figure out a better way of doing this
func (run *Run) CreateLastEvent() error {
	resp, err := run.request("POST", "/events", strings.NewReader("EOF"))
	if err != nil {
		return err
	}
	return checkOK(resp)
}

// ExitData holds information about how things ran and how much resources were used
type ExitData struct {
	Build ExitDataStage `json:"build"`
	Run   ExitDataStage `json:"run"`
}

// ExitDataStage gives the exit data for a single stage
type ExitDataStage struct {
	ExitCode int   `json:"exit_code"`
	Usage    Usage `json:"usage"`
}

// Usage gives the resource usage for a single stage
type Usage struct {
	WallTime   float64 `json:"wall_time"`   // In seconds
	CPUTime    float64 `json:"cpu_time"`    // In seconds
	MaxRSS     int64   `json:"max_rss"`     // In kilobytes
	NetworkIn  uint64  `json:"network_in"`  // In bytes
	NetworkOut uint64  `json:"network_out"` // In bytes
}

// PutExitData uploads information about how things ran and how much resources were used
func (run *Run) PutExitData(exitData ExitData) error {
	b, err := json.Marshal(exitData)
	if err != nil {
		return err
	}
	resp, err := run.request("PUT", "/exit-data", bytes.NewReader(b))
	if err != nil {
		return err
	}
	return checkOK(resp)
}

// Delete cleans up after a run is complete
func (run *Run) Delete() error {
	resp, err := run.request("DELETE", "", nil)
	if err != nil {
		return err
	}
	return checkOK(resp)
}
