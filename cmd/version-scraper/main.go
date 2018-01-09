package main

import (
	"github.com/heroku/herald";
	"github.com/fatih/color";
	"github.com/google/go-github/github";
	"golang.org/x/oauth2"
)
import (
	"time";
	"log";
	"fmt";
	"os";
	"context"
)

// Personal GitHub token. TODO: Create a bit account.
var GITHUB_TOKEN = os.Getenv("GITHUB_TOKEN")


// Opens an issue on GitHub for the given buildpack and new target.
// 
// Note: Uses the GITHUB_TOKEN environment variable, which is currently
//   Set to Kenneth's personal GitHub account. Need to create a bot account
//   for this service. 
func open_issue(bp herald.Buildpack, target string) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GITHUB_TOKEN},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	title := fmt.Sprintf("New release (%s) available! (Herald System)", target)
	body := fmt.Sprintf("This issue created programmatically and automatically by Heroku, on behalf of %s, the owner of the %s buildpack.", bp.Owner, bp.Name)
	
	newIssue := github.IssueRequest{
		Title: &title,
		Body: &body,
// 		Labels: ["New Build Target"],
		Assignee: &bp.Owner,
	}
	// list all repositories for the authenticated user
	bp_name := fmt.Sprintf("heroku-buildpack-%s", bp.Name)
	fmt.Println(bp_name)
	
	issue, _, err := client.Issues.Create(ctx, "heroku", bp_name, &newIssue)
	if err != nil {
		// do something
	}
// 	_ = issue
	fmt.Println(issue)

	fmt.Println(fmt.Sprintf("New issue created on %s buildpack on GitHub.", bp.Name))
}


func main() {

	// Redis stuff.
	redis := herald.NewRedis("")

	// Color Stuff.
	color.NoColor = false

	red := color.New(color.FgRed).SprintFunc()
    blue := color.New(color.FgBlue).SprintFunc()
	magenta := color.New(color.FgMagenta).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	bold := color.New(color.Bold, color.FgWhite).SprintFunc()

	for {

		// Get a list of the buildpacks (as types).
		buildpacks := herald.GetBuildpacks()

		// Iterate over them.
		for _, bp := range(buildpacks) {

			// Download and extract each Buildpack.
			log.Printf(bold("Downloading '%s'…"), red(bp.Name))
			path := bp.Download()

			log.Printf("Buildpack '%s' downloaded to '%s'!", red(bp), green(path))

			// Find all version executables for the given buildpack.
			executables := bp.FindVersionScripts()

			for _, exe := range(executables) {

				log.Printf(yellow("Executing '%s:%s' script…"), red(bp), magenta(exe))

				// TODO: Ensure chmod for the executable.
				exe.EnsureExecutable()

				// Execute the executable, print the results.
				results := exe.Execute()

				for _, result := range(results) {
					key := fmt.Sprintf("%s:%s:%s", bp, exe, result)
					value := herald.NewVersion().JSON()

					// Store the results in Redis.
					result, err := redis.Connection.Do("SETNX", key, value)

					// The insert was successful (e.g. it didn't exist already)
					if result.(int64) != int64(0) {
						// TODO: Send a notification to the buildpack owner.
						fmt.Println("Notifying", blue(bp.Owner), "about", red(key), ".")

						// Open an issue on GitHub (work in progress).
						open_issue(bp, key)
                    }
                    
					if err != nil {
						log.Fatal(err)
					}

				}

				// Log results.
				log.Printf("%s:%s results: %s", red(bp), magenta(exe), results)
			}
		}

		log.Print(bold("Sleeping for 10 minutes…"))

		// Sleep for ten minutes.
		time.Sleep(10*time.Minute)

		}

}