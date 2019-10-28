package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/openaustralia/morph-ng/pkg/clayclient"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "clay-run run_name run_output",
	Short: "Builds and runs a scraper",
	Long:  "Builds and runs a scraper and talks back to the Clay server.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runName := args[0]
		runToken := os.Getenv("CLAY_INTERNAL_RUN_TOKEN")
		runOutput := args[1]

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

		fmt.Println("runName", runName)
		fmt.Println("runToken", runToken)
		fmt.Println("runOutput", runOutput)
		fmt.Println("serverURL", serverURL)
		fmt.Println("buildCommand", buildCommand)
		fmt.Println("runCommand", runCommand)

		client := clayclient.New(serverURL)
		run := clayclient.Run{Name: runName, Token: runToken, Client: client}
		err := run.CreateStartEvent("build")
		if err != nil {
			log.Fatal(err)
		}

		// Create and populate /tmp/app and /tmp/cache
		err = os.MkdirAll("/tmp/app", 0755)
		if err != nil {
			log.Fatal(err)
		}
		err = os.MkdirAll("/tmp/cache", 0755)
		if err != nil {
			log.Fatal(err)
		}
		err = run.GetAppToDirectory("/tmp/app")
		if err != nil {
			log.Fatal(err)
		}
		d1 := []byte("scraper: /bin/start.sh")
		err = ioutil.WriteFile("/tmp/app/Procfile", d1, 0644)
		if err != nil {
			log.Fatal(err)
		}
		// TODO: Don't fail if the cache doesn't yet exist
		err = run.GetCacheToDirectory("/tmp/cache")
		if err != nil {
			log.Fatal(err)
		}

		// TODO: ***** MUCH MORE TODO HERE ******
	},
}

// Execute makes it all happen
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}