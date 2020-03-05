package commands

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/openaustralia/yinyo/pkg/keyvaluestore"
)

func (app *AppImplementation) newCreatedKey(runID string) Key {
	return app.newKey(runID, "created")
}

func (app *AppImplementation) newCallbackKey(runID string) Key {
	return app.newKey(runID, "url")
}

func (app *AppImplementation) newExitDataKey(runID string, key string) Key {
	return app.newKey(runID, "exit_data/"+key)
}

func (app *AppImplementation) newExitDataFinishedKey(runID string) Key {
	return app.newExitDataKey(runID, "finished")
}

func (app *AppImplementation) newExitDataBuildKey(runID string) Key {
	return app.newExitDataKey(runID, "build")
}

func (app *AppImplementation) newExitDataRunKey(runID string) Key {
	return app.newExitDataKey(runID, "run")
}

func (app *AppImplementation) newExitDataAPIKey(runID string, key string) Key {
	return app.newExitDataKey(runID, "api/"+key)
}

func (app *AppImplementation) newExitDataAPINetworkInKey(runID string) Key {
	return app.newExitDataAPIKey(runID, "network_in")
}

func (app *AppImplementation) newExitDataAPINetworkOutKey(runID string) Key {
	return app.newExitDataAPIKey(runID, "network_out")
}

type Key struct {
	key    string
	client keyvaluestore.KeyValueStore
}

func (app *AppImplementation) newKey(runID string, key string) Key {
	return Key{key: runID + "/" + key, client: app.KeyValueStore}
}

func (key Key) set(value string) error {
	return key.client.Set(key.key, value)
}

func (key Key) get() (string, error) {
	value, err := key.client.Get(key.key)
	if errors.Is(err, keyvaluestore.ErrKeyNotExist) {
		return value, fmt.Errorf("%w", ErrNotFound)
	}
	return value, err
}

func (key Key) getAsInt() (int64, error) {
	v, err := key.get()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(v, 10, 64)
}

func (key Key) delete() error {
	return key.client.Delete(key.key)
}

func (key Key) increment(value int64) (int64, error) {
	return key.client.Increment(key.key, value)
}
