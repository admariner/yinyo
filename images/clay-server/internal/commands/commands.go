package commands

import (
	"io"

	"github.com/dchest/uniuri"
	"github.com/go-redis/redis"

	"github.com/openaustralia/morph-ng/pkg/jobdispatcher"
	"github.com/openaustralia/morph-ng/pkg/store"
)

const filenameApp = "app.tgz"
const filenameCache = "cache.tgz"
const filenameOutput = "output"
const filenameExitData = "exit-data.json"
const dockerImage = "openaustralia/clay-scraper:v1"
const runBinary = "/bin/run.sh"

type createResult struct {
	RunName  string `json:"run_name"`
	RunToken string `json:"run_token"`
}

type logMessage struct {
	// TODO: Make the stream, stage and type an enum
	Log, Stream, Stage, Type string
}

func Create(k jobdispatcher.Client, namePrefix string) (createResult, error) {
	if namePrefix == "" {
		namePrefix = "run"
	}
	// Generate random token
	runToken := uniuri.NewLen(32)
	runName, err := k.CreateJobAndToken(namePrefix, runToken)

	createResult := createResult{
		RunName:  runName,
		RunToken: runToken,
	}
	return createResult, err
}

func GetApp(storeAccess store.Client, runName string, w io.Writer) error {
	return getData(storeAccess, runName, filenameApp, w)
}

func PutApp(storeAccess store.Client, reader io.Reader, objectSize int64, runName string) error {
	return putData(storeAccess, reader, objectSize, runName, filenameApp)
}

func GetCache(storeAccess store.Client, runName string, w io.Writer) error {
	return getData(storeAccess, runName, filenameCache, w)
}

func PutCache(storeAccess store.Client, reader io.Reader, objectSize int64, runName string) error {
	return putData(storeAccess, reader, objectSize, runName, filenameCache)
}

func GetOutput(storeAccess store.Client, runName string, w io.Writer) error {
	return getData(storeAccess, runName, filenameOutput, w)
}

func PutOutput(storeAccess store.Client, reader io.Reader, objectSize int64, runName string) error {
	return putData(storeAccess, reader, objectSize, runName, filenameOutput)
}

func GetExitData(storeAccess store.Client, runName string, w io.Writer) error {
	return getData(storeAccess, runName, filenameExitData, w)
}

func PutExitData(storeAccess store.Client, reader io.Reader, objectSize int64, runName string) error {
	return putData(storeAccess, reader, objectSize, runName, filenameExitData)
}

func Start(k jobdispatcher.Client, runName string, output string, env map[string]string) error {
	return k.StartJob(runName, dockerImage, []string{runBinary, runName, output}, env)
}

func GetEvent(redisClient *redis.Client, runName string, id string) (newId string, jsonString string, finished bool, err error) {
	// For the moment get one event at a time
	// TODO: Grab more than one at a time for a little more efficiency
	result, err := redisClient.XRead(&redis.XReadArgs{
		Streams: []string{runName, id},
		Count:   1,
		Block:   0,
	}).Result()
	if err != nil {
		return
	}
	newId = result[0].Messages[0].ID
	jsonString = result[0].Messages[0].Values["json"].(string)

	if jsonString == "EOF" {
		finished = true
	}
	return
}

func CreateEvent(redisClient *redis.Client, runName string, eventJson string) error {
	// TODO: Send the event to the user with an http POST

	// Send the json to a redis stream
	return redisClient.XAdd(&redis.XAddArgs{
		// TODO: Use something like runName-events instead for the stream name
		Stream: runName,
		Values: map[string]interface{}{"json": eventJson},
	}).Err()
}

func Delete(k jobdispatcher.Client, storeAccess store.Client, redisClient *redis.Client, runName string) error {
	err := k.DeleteJobAndToken(runName)
	if err != nil {
		return err
	}

	err = deleteData(storeAccess, runName, filenameApp)
	if err != nil {
		return err
	}
	err = deleteData(storeAccess, runName, filenameOutput)
	if err != nil {
		return err
	}
	err = deleteData(storeAccess, runName, filenameExitData)
	if err != nil {
		return err
	}
	err = deleteData(storeAccess, runName, filenameCache)
	if err != nil {
		return err
	}
	return redisClient.Del(runName).Err()
}

func storagePath(runName string, fileName string) string {
	return runName + "/" + fileName
}

func getData(m store.Client, runName string, fileName string, writer io.Writer) error {
	reader, err := m.Get(storagePath(runName, fileName))
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, reader)
	return err
}

func putData(m store.Client, reader io.Reader, objectSize int64, runName string, fileName string) error {
	return m.Put(
		storagePath(runName, fileName),
		reader,
		objectSize,
	)
}

func deleteData(m store.Client, runName string, fileName string) error {
	return m.Delete(storagePath(runName, fileName))
}
