package cmd

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/kballard/go-shellquote"
	"github.com/openaustralia/morph-ng/pkg/clayclient"
	"github.com/shirou/gopsutil/net"
	"github.com/spf13/cobra"
)

func streamLogs(run clayclient.Run, stage string, streamName string, stream io.ReadCloser, c chan error) {
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		run.CreateEvent(clayclient.LogEvent{Stage: stage, Stream: streamName, Text: scanner.Text()})
	}
	c <- scanner.Err()
}

// env is an array of strings to set environment variables to in the form "VARIABLE=value", ...
func runExternalCommand(run clayclient.Run, stage string, commandString string, env []string) (clayclient.ExitDataStage, error) {
	var exitData clayclient.ExitDataStage

	// Splits string up into pieces using shell rules
	commandParts, err := shellquote.Split(commandString)
	if err != nil {
		return exitData, err
	}
	command := exec.Command(commandParts[0], commandParts[1:]...)
	// Add the environment variables to the pre-existing environment
	// TODO: Do we want to zero out the environment?
	command.Env = append(os.Environ(), env...)
	stdout, err := command.StdoutPipe()
	if err != nil {
		return exitData, err
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		return exitData, err
	}
	// Capture the time and the network counters
	start := time.Now()
	statsStart, err := net.IOCounters(false)
	if err != nil {
		return exitData, err
	}
	// Since we're asking for the aggregates we should only ever receive one answer
	if len(statsStart) != 1 {
		return exitData, errors.New("Only expected one stat")
	}

	if err := command.Start(); err != nil {
		return exitData, err
	}

	c := make(chan error)
	go streamLogs(run, stage, "stdout", stdout, c)
	go streamLogs(run, stage, "stderr", stderr, c)
	err = <-c
	if err != nil {
		return exitData, err
	}
	err = <-c
	if err != nil {
		return exitData, err
	}

	err = command.Wait()
	if err != nil && command.ProcessState.ExitCode() == 0 {
		return exitData, err
	}
	statsEnd, err := net.IOCounters(false)
	if err != nil {
		return exitData, err
	}
	// Since we're asking for the aggregates we should only ever receive one answer
	if len(statsEnd) != 1 {
		return exitData, errors.New("Only expected one stat")
	}

	exitData.Usage.NetworkIn = statsEnd[0].BytesRecv - statsStart[0].BytesRecv
	exitData.Usage.NetworkOut = statsEnd[0].BytesSent - statsStart[0].BytesSent
	exitData.Usage.WallTime = time.Now().Sub(start).Seconds()
	exitData.Usage.CPUTime = command.ProcessState.UserTime().Seconds() +
		command.ProcessState.SystemTime().Seconds()
	// This bit will only return something when run on Linux I think
	rusage, ok := command.ProcessState.SysUsage().(*syscall.Rusage)
	if ok {
		exitData.Usage.MaxRSS = rusage.Maxrss
	}
	exitData.ExitCode = command.ProcessState.ExitCode()

	return exitData, nil
}

var appPath, importPath, cachePath, runOutput string

func init() {
	wrapperCmd.Flags().StringVar(&appPath, "app", "/app", "herokuish app path")
	wrapperCmd.Flags().StringVar(&importPath, "import", "/tmp/app", "herokuish import path")
	wrapperCmd.Flags().StringVar(&cachePath, "cache", "/tmp/cache", "herokuish cache path")
	wrapperCmd.Flags().StringVar(&runOutput, "output", "", "relative path to output file")
	rootCmd.AddCommand(wrapperCmd)
}

var wrapperCmd = &cobra.Command{
	Use:   "wrapper run_name run_token",
	Short: "Manages the building and running of a scraper",
	Long:  "Manages the building and running of a scraper inside a container. Used internally by the system.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runName := args[0]
		runToken := args[1]

		// We allow some settings to be overridden for the purposes of testing.
		// We don't allow users to change any environment variables that start
		// with CLAY_INTERNAL_SERVER_. So they can't change any of these

		// TODO: Convert these special environment variables to command line options
		serverURL, ok := os.LookupEnv("CLAY_INTERNAL_SERVER_URL")
		if !ok {
			serverURL = "http://clay-server.clay-system:8080"
		}

		buildCommand, ok := os.LookupEnv("CLAY_INTERNAL_BUILD_COMMAND")
		if !ok {
			buildCommand = "/bin/herokuish buildpack build"
		}

		runCommand, ok := os.LookupEnv("CLAY_INTERNAL_RUN_COMMAND")
		if !ok {
			runCommand = "/bin/herokuish procfile start scraper"
		}

		client := clayclient.New(serverURL)
		run := clayclient.Run{Name: runName, Token: runToken, Client: client}
		err := run.CreateEvent(clayclient.StartEvent{Stage: "build"})
		if err != nil {
			log.Fatal(err)
		}

		// Create and populate herokuish import path and cache path
		err = os.MkdirAll(importPath, 0755)
		if err != nil {
			log.Fatal(err)
		}
		err = os.MkdirAll(cachePath, 0755)
		if err != nil {
			log.Fatal(err)
		}
		err = run.GetAppToDirectory(importPath)
		if err != nil {
			log.Fatal(err)
		}
		d1 := []byte("scraper: /bin/start.sh")
		err = ioutil.WriteFile(filepath.Join(importPath, "Procfile"), d1, 0644)
		if err != nil {
			log.Fatal(err)
		}
		// If the cache doesn't exit this will not error
		err = run.GetCacheToDirectory(cachePath)
		if err != nil {
			log.Fatal(err)
		}

		env := []string{
			"APP_PATH=" + appPath,
			"CACHE_PATH=" + cachePath,
			"IMPORT_PATH=" + importPath,
		}

		// Initially do a very naive way of calling the command just to get things going
		var exitData clayclient.ExitData

		exitDataStage, err := runExternalCommand(run, "build", buildCommand, env)
		if err != nil {
			log.Fatal(err)
		}
		exitData.Build = &exitDataStage

		// Send the build finished event immediately when the build command has finished
		// Effectively the cache uploading happens between the build and run stages
		err = run.CreateEvent(clayclient.FinishEvent{Stage: "build"})
		if err != nil {
			log.Fatal(err)
		}

		err = run.PutCacheFromDirectory(cachePath)
		if err != nil {
			log.Fatal(err)
		}

		// Only do the main run if the build was succesful
		if exitData.Build.ExitCode == 0 {
			err = run.CreateEvent(clayclient.StartEvent{Stage: "run"})
			if err != nil {
				log.Fatal(err)
			}

			exitDataStage, err := runExternalCommand(run, "run", runCommand, env)
			if err != nil {
				log.Fatal(err)
			}
			exitData.Run = &exitDataStage

			if err := run.PutExitData(exitData); err != nil {
				log.Fatal(err)
			}

			if runOutput != "" {
				err = run.PutOutputFromFile(filepath.Join(appPath, runOutput))
				if err != nil {
					log.Fatal(err)
				}
			}

			err = run.CreateEvent(clayclient.FinishEvent{Stage: "run"})
			if err != nil {
				log.Fatal(err)
			}
		} else {
			// TODO: Only upload the exit data for the build
			if err := run.PutExitData(exitData); err != nil {
				log.Fatal(err)
			}
		}

		err = run.CreateLastEvent()
		if err != nil {
			log.Fatal(err)
		}
	},
}