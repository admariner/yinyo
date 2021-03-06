package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/openaustralia/yinyo/pkg/apiclient"
	"github.com/openaustralia/yinyo/pkg/protocol"
	"github.com/spf13/cobra"
)

// Convert the internal text representation of a stream type ("stdout"/"stderr") to the go stream
// we can write to
func osStream(stream string) (*os.File, error) {
	switch stream {
	// TODO: Extract string constant
	case "stdout":
		return os.Stdout, nil
	case "stderr", "interr":
		return os.Stderr, nil
	default:
		return nil, fmt.Errorf("unexpected stream %v", stream)
	}
}

func display(event protocol.Event, showEventsJSON bool) error {
	if showEventsJSON {
		// Convert the event back to JSON for display
		b, err := json.Marshal(event)
		if err != nil {
			return err
		}
		fmt.Println(string(b))
	} else {
		// Only display the log events to the user
		l, ok := event.Data.(protocol.LogData)
		if ok {
			f, err := osStream(l.Stream)
			if err != nil {
				return err
			}
			fmt.Fprintln(f, l.Text)
		}
	}
	return nil
}

func apiKeysPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".yinyo"), nil
}

// Write to user's local directory as .yinyo
// Store a separate api key for each server URL
func saveAPIKeys(apiKeys map[string]string) error {
	path, err := apiKeysPath()
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	return enc.Encode(apiKeys)
}

func loadAPIKeys() (map[string]string, error) {
	apiKeys := make(map[string]string)
	path, err := apiKeysPath()
	if err != nil {
		return apiKeys, err
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return apiKeys, nil
		}
		return apiKeys, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	err = dec.Decode(&apiKeys)
	return apiKeys, err
}

func askForAndSaveAPIKey(clientServerURL string) (string, error) {
	fmt.Print("Enter your api key: ")
	var apiKey string
	fmt.Scanln(&apiKey)

	err := saveAPIKey(clientServerURL, apiKey)
	return apiKey, err
}

func saveAPIKey(clientServerURL string, apiKey string) error {
	apiKeys, err := loadAPIKeys()
	if err != nil {
		return err
	}
	apiKeys[clientServerURL] = apiKey
	return saveAPIKeys(apiKeys)
}

func getAPIKey(clientServerURL string) (string, error) {
	apiKeys, err := loadAPIKeys()
	if err != nil {
		return "", err
	}
	apiKey := apiKeys[clientServerURL]
	return apiKey, nil
}

func createRun(clientServerURL string) (apiclient.RunInterface, error) {
	var run apiclient.RunInterface
	apiKey, err := getAPIKey(clientServerURL)
	if err != nil {
		return run, err
	}
	client := apiclient.New(clientServerURL)
	for {
		// Create the run
		run, err = client.CreateRun(protocol.CreateRunOptions{APIKey: apiKey})
		if err == nil {
			return run, err
		}
		fmt.Fprintln(os.Stderr, err)

		// TODO: Display info/error message from server
		if !errors.Is(err, apiclient.ErrUnauthorized) {
			return run, err
		}
		apiKey, err = askForAndSaveAPIKey(clientServerURL)
		if err != nil {
			return run, err
		}
	}
}

// reconnectCommand returns the command to reconnect to the current run
func reconnectCommand(cmd *cobra.Command, runID string, scraperDirectory string) string {
	// We're hacking together the command-line rendering here. I hope there's a more sensible way of doing this with cobra.
	text := fmt.Sprintf("yinyo %v", scraperDirectory)
	if cmd.Flag("server").Changed {
		text += fmt.Sprintf(" --server %v", cmd.Flag("server").Value)
	}
	if cmd.Flag("output").Changed {
		text += fmt.Sprintf(" --output %v", cmd.Flag("output").Value)
	}
	if cmd.Flag("cache").Changed {
		text += " --cache"
	}
	if cmd.Flag("noprogress").Changed {
		text += " --noprogress"
	}
	text += fmt.Sprintf(" --connect %v", runID)
	return text
}

func cleanup(started bool, cmd *cobra.Command, scraperDirectory string, disableProgress bool, run apiclient.RunInterface) {
	if started {
		fmt.Println("\n\nTo reconnect to this run:")
		fmt.Println(reconnectCommand(cmd, run.GetID(), scraperDirectory))
	} else {
		if !disableProgress {
			fmt.Println("\n\n[Cleaning up]")
			fmt.Println("[Deleting run]")
		}

		if err := run.Delete(); err != nil {
			log.Fatal(err)
		}
	}
	os.Exit(1)
}

func main() {
	// Show the source of the error with the standard logger. Don't show date & time
	log.SetFlags(log.Lshortfile)

	var callbackURL, outputFile, clientServerURL, runID string
	var showEventsJSON, cache, disableProgress bool
	var environment map[string]string

	var rootCmd = &cobra.Command{
		Use:   "yinyo scraper_directory",
		Short: "Yinyo runs heaps of scrapers easily, fast and scalably",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			scraperDirectory := args[0]
			eventCallback := func(event protocol.Event) error { return display(event, showEventsJSON) }

			if runID == "" {
				run, err := createRun(clientServerURL)
				if err != nil {
					log.Fatal(err)
				}

				// Set up a handler for ctrl+c that will gracefully exit
				sigs := make(chan os.Signal, 1)

				signal.Notify(sigs, syscall.SIGINT)

				started := false
				go func() {
					<-sigs
					cleanup(started, cmd, scraperDirectory, disableProgress, run)
				}()

				runID = run.GetID()
				err = apiclient.SimpleStart(runID, scraperDirectory, clientServerURL, environment, outputFile, cache, callbackURL, !disableProgress)
				if err != nil {
					log.Fatal(err)
				}
				started = true
			}
			err := apiclient.SimpleConnect(runID, scraperDirectory, clientServerURL, outputFile, cache, eventCallback, !disableProgress)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	rootCmd.Flags().StringVar(&callbackURL, "callback", "", "Optionally provide a callback URL. For every event a POST to the URL will be made. To be able to authenticate the callback you'll need to specify a secret in the URL. Something like http://my-url-endpoint.com?key=special-secret-stuff would do the trick")
	// TODO: Check that the output file is a relative path and if not error
	rootCmd.Flags().StringVar(&outputFile, "output", "", "The output is written to the same local directory at the end. The output file path is given relative to the scraper directory")
	rootCmd.Flags().StringVar(&runID, "connect", "", "Connect to a run that has already started by giving the run ID")
	rootCmd.Flags().StringVar(&clientServerURL, "server", "https://api.yinyo.io", "Override yinyo server URL")
	rootCmd.Flags().StringToStringVar(&environment, "env", map[string]string{}, "Set one or more environment variables (e.g. --env foo=twiddle,bar=blah)")
	rootCmd.Flags().BoolVar(&showEventsJSON, "allevents", false, "Show the full events output as JSON instead of the default of just showing the log events as text")
	rootCmd.Flags().BoolVar(&cache, "cache", false, "Enable the download and upload of the build cache")
	rootCmd.Flags().BoolVar(&disableProgress, "noprogress", false, "Disable messages showing progress")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
