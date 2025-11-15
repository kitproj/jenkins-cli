package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bndr/gojenkins"
	"github.com/dustin/go-humanize"
	"github.com/kitproj/jenkins-cli/internal/config"
	"golang.org/x/term"
)

var (
	url     string
	token   string
	user    string
	jenkins *gojenkins.Jenkins
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage:\n")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "  jenkins configure <url> [username] - Configure Jenkins URL and API token (reads token from stdin)")
		fmt.Fprintln(w, "  jenkins list-jobs - List all Jenkins jobs")
		fmt.Fprintln(w, "  jenkins get-job <job-name> - Get details of a specific job")
		fmt.Fprintln(w, "  jenkins build-job <job-name> - Trigger a build for a job")
		fmt.Fprintln(w, "  jenkins get-build <job-name> <build-number> - Get details of a specific build")
		fmt.Fprintln(w, "  jenkins get-build-log <job-name> <build-number> - Get the console output of a build")
		fmt.Fprintln(w, "  jenkins get-last-build <job-name> - Get details of the last build")
		fmt.Fprintln(w, "  jenkins mcp-server - Start MCP server (Model Context Protocol)")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if err := run(ctx, flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: jenkins <command> [args...]")
	}

	// First argument is the command
	command := args[0]

	switch command {
	case "configure":
		if len(args) < 2 {
			return fmt.Errorf("usage: jenkins configure <url> [username]")
		}
		username := ""
		if len(args) >= 3 {
			username = args[2]
		}
		return configure(args[1], username)
	case "list-jobs":
		return executeCommand(ctx, listJobs)
	case "get-job":
		if len(args) < 2 {
			return fmt.Errorf("usage: jenkins get-job <job-name>")
		}
		jobName := args[1]
		return executeCommand(ctx, func(ctx context.Context) error {
			return getJob(ctx, jobName)
		})
	case "build-job":
		if len(args) < 2 {
			return fmt.Errorf("usage: jenkins build-job <job-name>")
		}
		jobName := args[1]
		return executeCommand(ctx, func(ctx context.Context) error {
			return buildJob(ctx, jobName)
		})
	case "get-build":
		if len(args) < 3 {
			return fmt.Errorf("usage: jenkins get-build <job-name> <build-number>")
		}
		jobName := args[1]
		buildNumber := args[2]
		return executeCommand(ctx, func(ctx context.Context) error {
			return getBuild(ctx, jobName, buildNumber)
		})
	case "get-build-log":
		if len(args) < 3 {
			return fmt.Errorf("usage: jenkins get-build-log <job-name> <build-number>")
		}
		jobName := args[1]
		buildNumber := args[2]
		return executeCommand(ctx, func(ctx context.Context) error {
			return getBuildLog(ctx, jobName, buildNumber)
		})
	case "get-last-build":
		if len(args) < 2 {
			return fmt.Errorf("usage: jenkins get-last-build <job-name>")
		}
		jobName := args[1]
		return executeCommand(ctx, func(ctx context.Context) error {
			return getLastBuild(ctx, jobName)
		})
	case "mcp-server":
		return runMCPServer(ctx)
	default:
		return fmt.Errorf("unknown sub-command: %s", command)
	}
}

func executeCommand(ctx context.Context, fn func(context.Context) error) error {
	// Load URL and username from config file, or fall back to env var
	if url == "" {
		var err error
		var configUsername string
		url, configUsername, err = config.LoadConfig()
		if err != nil {
			// Fall back to environment variable
			url = os.Getenv("JENKINS_URL")
		} else if user == "" && configUsername != "" {
			// Use username from config if not already set
			user = configUsername
		}
	}

	// Normalize URL from environment variable if needed
	if url != "" {
		url = config.NormalizeURL(url)
	}

	// Load token from keyring, or fall back to env var
	if token == "" {
		token = os.Getenv("JENKINS_TOKEN")
	}
	if token == "" {
		var err error
		token, err = config.LoadToken(url)
		if err != nil {
			return fmt.Errorf("token not found, please run 'jenkins configure <url>' first")
		}
	}

	// Load user from environment variable if not set
	if user == "" {
		user = os.Getenv("JENKINS_USER")
		if user == "" {
			user = "admin" // Default username
		}
	}

	if url == "" {
		return fmt.Errorf("Jenkins URL is required")
	}
	if token == "" {
		return fmt.Errorf("token is required")
	}

	// Create Jenkins client with the full URL
	var err error
	jenkins, err = gojenkins.CreateJenkins(nil, url, user, token).Init(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Jenkins client: %w", err)
	}

	return fn(ctx)
}

// configure reads the token from stdin and saves it to the keyring
func configure(jenkinsURL, username string) error {
	if jenkinsURL == "" {
		return fmt.Errorf("Jenkins URL is required")
	}

	// Normalize the URL (ensure protocol and remove trailing slashes)
	normalizedURL := config.NormalizeURL(jenkinsURL)

	if username == "" {
		username = "admin"
	}

	// Display the URL to the user
	fmt.Fprintf(os.Stderr, "To create an API token in Jenkins:\n")
	fmt.Fprintf(os.Stderr, "1. Go to: %s/user/%s/configure\n", normalizedURL, username)
	fmt.Fprintf(os.Stderr, "2. Click 'Add new Token' under API Token section\n")
	fmt.Fprintf(os.Stderr, "3. Copy the generated token\n")
	fmt.Fprintf(os.Stderr, "\nThe token will be stored securely in your system's keyring.\n")
	fmt.Fprintf(os.Stderr, "\nEnter Jenkins API token: ")

	// Read password with hidden input
	tokenBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(os.Stderr) // Print newline after hidden input
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}

	token := string(tokenBytes)
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	// Save URL and username to config file (SaveConfig will normalize the URL)
	if err := config.SaveConfig(normalizedURL, username); err != nil {
		return err
	}

	// Save token to keyring using normalized URL
	if err := config.SaveToken(normalizedURL, token); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Configuration saved successfully for URL: %s (username: %s, override with JENKINS_USER env var)\n", normalizedURL, username)
	return nil
}

// listJobs lists all Jenkins jobs
func listJobs(ctx context.Context) error {
	jobs, err := jenkins.GetAllJobs(ctx)
	if err != nil {
		return fmt.Errorf("failed to list jobs: %w", err)
	}

	if len(jobs) == 0 {
		fmt.Println("No jobs found")
		return nil
	}

	fmt.Printf("Found %d job(s):\n\n", len(jobs))
	for _, job := range jobs {
		status := getStatusFromColor(job.Raw.Color)
		fmt.Printf("%-40s %-15s %s\n", job.GetName(), status, job.Raw.URL)
	}

	return nil
}

// getJob gets details of a specific job
func getJob(ctx context.Context, jobName string) error {
	job, err := jenkins.GetJob(ctx, jobName)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	printField("Job Name", job.GetName())
	printField("URL", job.Raw.URL)
	printField("Status", getStatusFromColor(job.Raw.Color))
	if job.GetDescription() != "" {
		printField("Description", job.GetDescription())
	}
	
	lastBuild, err := job.GetLastBuild(ctx)
	if err == nil && lastBuild != nil {
		printField("Last Build", fmt.Sprintf("#%d - %s", lastBuild.GetBuildNumber(), lastBuild.GetResult()))
	}
	
	lastSuccess, err := job.GetLastSuccessfulBuild(ctx)
	if err == nil && lastSuccess != nil {
		printField("Last Success", fmt.Sprintf("#%d", lastSuccess.GetBuildNumber()))
	}
	
	lastFailed, err := job.GetLastFailedBuild(ctx)
	if err == nil && lastFailed != nil {
		printField("Last Failed", fmt.Sprintf("#%d", lastFailed.GetBuildNumber()))
	}

	return nil
}

// buildJob triggers a build for a job
func buildJob(ctx context.Context, jobName string) error {
	job, err := jenkins.GetJob(ctx, jobName)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	_, err = job.InvokeSimple(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to trigger build: %w", err)
	}

	fmt.Printf("Successfully triggered build for job: %s\n", jobName)
	return nil
}

// getBuild gets details of a specific build
func getBuild(ctx context.Context, jobName, buildNumber string) error {
	job, err := jenkins.GetJob(ctx, jobName)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	buildNum, err := parseBuildNumber(buildNumber)
	if err != nil {
		return err
	}

	build, err := job.GetBuild(ctx, buildNum)
	if err != nil {
		return fmt.Errorf("failed to get build: %w", err)
	}

	printBuildDetails(ctx, build)
	return nil
}

// getBuildLog gets the console output of a build
func getBuildLog(ctx context.Context, jobName, buildNumber string) error {
	job, err := jenkins.GetJob(ctx, jobName)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	buildNum, err := parseBuildNumber(buildNumber)
	if err != nil {
		return err
	}

	build, err := job.GetBuild(ctx, buildNum)
	if err != nil {
		return fmt.Errorf("failed to get build: %w", err)
	}

	log := build.GetConsoleOutput(ctx)
	fmt.Print(log)
	return nil
}

// getLastBuild gets details of the last build of a job
func getLastBuild(ctx context.Context, jobName string) error {
	job, err := jenkins.GetJob(ctx, jobName)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	build, err := job.GetLastBuild(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last build: %w", err)
	}

	printBuildDetails(ctx, build)
	return nil
}

// Helper functions

func parseBuildNumber(buildNumber string) (int64, error) {
	var num int64
	_, err := fmt.Sscanf(buildNumber, "%d", &num)
	if err != nil || num <= 0 {
		return 0, fmt.Errorf("invalid build number: %s", buildNumber)
	}
	return num, nil
}

func printBuildDetails(ctx context.Context, build *gojenkins.Build) {
	printField("Build Number", build.GetBuildNumber())
	printField("URL", build.GetUrl())
	status := build.GetResult()
	if build.IsRunning(ctx) {
		status = "BUILDING"
	}
	printField("Status", status)
	if build.Raw.Description != "" {
		printField("Description", build.Raw.Description)
	}
	buildTime := build.GetTimestamp()
	if !buildTime.IsZero() {
		printField("Started", humanize.Time(buildTime))
	}
	duration := build.GetDuration()
	if duration > 0 {
		printField("Duration", fmt.Sprintf("%.0fs", duration/1000))
	}
}

// printField prints a field with proper formatting
func printField(key string, value interface{}) {
	valueStr := fmt.Sprint(value)
	multiLine := strings.Contains(valueStr, "\n")
	fmt.Printf("%-20s", key+":")
	if !multiLine {
		fmt.Printf(" %s\n", valueStr)
	} else {
		fmt.Println()
		for _, line := range strings.Split(valueStr, "\n") {
			fmt.Printf("%-20s %s\n", "", line)
		}
	}
}

// getStatusFromColor converts Jenkins color to status
func getStatusFromColor(color string) string {
	switch {
	case strings.HasPrefix(color, "blue"):
		return "SUCCESS"
	case strings.HasPrefix(color, "red"):
		return "FAILURE"
	case strings.HasPrefix(color, "yellow"):
		return "UNSTABLE"
	case strings.HasPrefix(color, "grey"):
		return "PENDING"
	case strings.HasPrefix(color, "aborted"):
		return "ABORTED"
	case strings.HasPrefix(color, "notbuilt"):
		return "NOT_BUILT"
	case strings.HasPrefix(color, "disabled"):
		return "DISABLED"
	default:
		return strings.ToUpper(color)
	}
}
