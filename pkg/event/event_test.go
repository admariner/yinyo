package event

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// test marshalling and unmarshalling of event
func testMarshal(t *testing.T, event Event, jsonString string) {
	b, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, jsonString, string(b))
	var actual Event
	err = json.Unmarshal(b, &actual)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, event, actual)
}

func TestMarshalStartEvent(t *testing.T) {
	time := time.Date(2000, time.January, 2, 3, 45, 0, 0, time.UTC)
	testMarshal(t,
		NewStartEvent("", time, "build"),
		`{"time":"2000-01-02T03:45:00Z","type":"start","data":{"stage":"build"}}`,
	)
}

func TestMarshalFinishEvent(t *testing.T) {
	time := time.Date(2000, time.January, 2, 3, 45, 0, 0, time.UTC)
	testMarshal(t,
		NewFinishEvent("", time, "build"),
		`{"time":"2000-01-02T03:45:00Z","type":"finish","data":{"stage":"build"}}`,
	)
}

func TestMarshalLogEvent(t *testing.T) {
	time := time.Date(2000, time.January, 2, 3, 45, 0, 0, time.UTC)
	testMarshal(t,
		NewLogEvent("", time, "build", "stdout", "Hello"),
		`{"time":"2000-01-02T03:45:00Z","type":"log","data":{"stage":"build","stream":"stdout","text":"Hello"}}`,
	)
}

func TestMarshalLastEvent(t *testing.T) {
	time := time.Date(2000, time.January, 2, 3, 45, 0, 0, time.UTC)
	testMarshal(t,
		NewLastEvent("", time),
		`{"time":"2000-01-02T03:45:00Z","type":"last","data":{}}`,
	)
}

func TestNewLogEvent(t *testing.T) {
	now := time.Now()
	assert.Equal(t,
		Event{ID: "123", Time: now, Type: "log", Data: LogData{Stage: "build", Stream: "stdout", Text: "hello"}},
		NewLogEvent("123", now, "build", "stdout", "hello"),
	)
}