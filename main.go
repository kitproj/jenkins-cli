package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kitproj/jenkins-cli/internal/config"
	"golang.org/x/term"
)

var (
	host   string
	token  string
	user   string
	client *http.Client
)

// Job represents a Jenkins job
type Job struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Color string `json:"color"`
}

// JobsResponse represents the response from Jenkins jobs API
type JobsResponse struct {
	Jobs []Job `json:"jobs"`
}

// Build represents a Jenkins build
type Build struct {
	Number    int    `json:"number"`
	URL       string `json:"url"`
	Result    string `json:"result"`
	Timestamp int64  `json:"timestamp"`
	Duration  int64  `json:"duration"`
	Building  bool   `json:"building"`
}

// JobDetail represents detailed job information
type JobDetail struct {
	Name            string  `json:"name"`
	URL             string  `json:"url"`
	Description     string  `json:"description"`
	Color           string  `json:"color"`
	LastBuild       *Build  `json:"lastBuild"`
	LastSuccessBuild *Build `json:"lastSuccessfulBuild"`
	LastFailedBuild  *Build `json:"lastFailedBuild"`
}

// BuildDetail represents detailed build information
type BuildDetail struct {
	Number      int    `json:"number"`
	URL         string `json:"url"`
	Result      string `json:"result"`
	Timestamp   int64  `json:"timestamp"`
	Duration    int64  `json:"duration"`
	Building    bool   `json:"building"`
	Description string `json:"description"`
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage:\n")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "  jenkins configure <host> [username] - Configure Jenkins host and API token (reads token from stdin)")
		fmt.Fprintln(w, "  jenkins list-jobs - List all Jenkins jobs")
		fmt.Fprintln(w, "  jenkins get-job <job-name> - Get details of a specific job")
		fmt.Fprintln(w, "  jenkins build-job <job-name> - Trigger a build for a job")
		fmt.Fprintln(w, "  jenkins get-build <job-name> <build-number> - Get details of a specific build")
		fmt.Fprintln(w, "  jenkins get-build-log <job-name> <build-number> - Get the console output of a build")
		fmt.Fprintln(w, "  jenkins get-last-build <job-name> - Get details of the last build")
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
			return fmt.Errorf("usage: jenkins configure <host> [username]")
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
	default:
		return fmt.Errorf("unknown sub-command: %s", command)
	}
}

func executeCommand(ctx context.Context, fn func(context.Context) error) error {
	// Load host and username from config file, or fall back to env var
	if host == "" {
		var err error
		var configUsername string
		host, configUsername, err = config.LoadConfig()
		if err != nil {
			// Fall back to environment variable
			host = os.Getenv("JENKINS_HOST")
		} else if user == "" && configUsername != "" {
			// Use username from config if not already set
			user = configUsername
		}
	}

	// Load token from keyring, or fall back to env var
	if token == "" {
		token = os.Getenv("JENKINS_TOKEN")
	}
	if token == "" {
		var err error
		token, err = config.LoadToken(host)
		if err != nil {
			return fmt.Errorf("token not found, please run 'jenkins configure <host>' first")
		}
	}

	// Load user from environment variable if not set
	if user == "" {
		user = os.Getenv("JENKINS_USER")
		if user == "" {
			user = "admin" // Default username
		}
	}

	if host == "" {
		return fmt.Errorf("host is required")
	}
	if token == "" {
		return fmt.Errorf("token is required")
	}

	// Create HTTP client with timeout and TLS configuration
	client = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		},
	}

	return fn(ctx)
}

// configure reads the token from stdin and saves it to the keyring
func configure(host, username string) error {
	if host == "" {
		return fmt.Errorf("host is required")
	}

	// Reject http:// or https:// prefix - host should be hostname only
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		return fmt.Errorf("host should not include http:// or https:// prefix, provide hostname only (e.g., jenkins.example.com)")
	}

	if username == "" {
		username = "admin"
	}

	fmt.Fprintf(os.Stderr, "To create an API token in Jenkins:\n")
	fmt.Fprintf(os.Stderr, "1. Go to: https://%s/user/%s/configure\n", host, username)
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

	// Save host and username to config file
	if err := config.SaveConfig(host, username); err != nil {
		return err
	}

	// Save token to keyring
	if err := config.SaveToken(host, token); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Configuration saved successfully for host: %s\n", host)
	fmt.Fprintf(os.Stderr, "Username will default to '%s' (override with JENKINS_USER env var)\n", username)
	return nil
}

// makeRequest makes an authenticated HTTP request to Jenkins
func makeRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	// Host should not have protocol prefix - always use https://
	jenkinsURL := "https://" + host

	reqURL := fmt.Sprintf("%s%s", jenkinsURL, path)
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication
	if user != "" && token != "" {
		req.SetBasicAuth(user, token)
	}

	return client.Do(req)
}

// listJobs lists all Jenkins jobs
func listJobs(ctx context.Context) error {
	resp, err := makeRequest(ctx, "GET", "/api/json?tree=jobs[name,url,color]", nil)
	if err != nil {
		return fmt.Errorf("failed to list jobs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to list jobs: %s - %s", resp.Status, string(body))
	}

	var jobsResp JobsResponse
	if err := json.NewDecoder(resp.Body).Decode(&jobsResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if len(jobsResp.Jobs) == 0 {
		fmt.Println("No jobs found")
		return nil
	}

	fmt.Printf("Found %d job(s):\n\n", len(jobsResp.Jobs))
	for _, job := range jobsResp.Jobs {
		status := getStatusFromColor(job.Color)
		fmt.Printf("%-40s %-15s %s\n", job.Name, status, job.URL)
	}

	return nil
}

// getJob gets details of a specific job
func getJob(ctx context.Context, jobName string) error {
	encodedJobName := url.PathEscape(jobName)
	path := fmt.Sprintf("/job/%s/api/json", encodedJobName)

	resp, err := makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to get job: %s - %s", resp.Status, string(body))
	}

	var job JobDetail
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	printField("Job Name", job.Name)
	printField("URL", job.URL)
	printField("Status", getStatusFromColor(job.Color))
	if job.Description != "" {
		printField("Description", job.Description)
	}
	if job.LastBuild != nil {
		printField("Last Build", fmt.Sprintf("#%d - %s", job.LastBuild.Number, job.LastBuild.Result))
	}
	if job.LastSuccessBuild != nil {
		printField("Last Success", fmt.Sprintf("#%d", job.LastSuccessBuild.Number))
	}
	if job.LastFailedBuild != nil {
		printField("Last Failed", fmt.Sprintf("#%d", job.LastFailedBuild.Number))
	}

	return nil
}

// buildJob triggers a build for a job
func buildJob(ctx context.Context, jobName string) error {
	encodedJobName := url.PathEscape(jobName)
	path := fmt.Sprintf("/job/%s/build", encodedJobName)

	resp, err := makeRequest(ctx, "POST", path, nil)
	if err != nil {
		return fmt.Errorf("failed to trigger build: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to trigger build: %s - %s", resp.Status, string(body))
	}

	fmt.Printf("Successfully triggered build for job: %s\n", jobName)
	return nil
}

// getBuild gets details of a specific build
func getBuild(ctx context.Context, jobName, buildNumber string) error {
	encodedJobName := url.PathEscape(jobName)
	path := fmt.Sprintf("/job/%s/%s/api/json", encodedJobName, buildNumber)

	resp, err := makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return fmt.Errorf("failed to get build: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to get build: %s - %s", resp.Status, string(body))
	}

	var build BuildDetail
	if err := json.NewDecoder(resp.Body).Decode(&build); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	printField("Build Number", build.Number)
	printField("URL", build.URL)
	status := build.Result
	if build.Building {
		status = "BUILDING"
	}
	printField("Status", status)
	if build.Description != "" {
		printField("Description", build.Description)
	}
	if build.Timestamp > 0 {
		buildTime := time.Unix(build.Timestamp/1000, 0)
		printField("Started", buildTime.Format(time.RFC1123))
	}
	if build.Duration > 0 {
		duration := time.Duration(build.Duration) * time.Millisecond
		printField("Duration", duration.String())
	}

	return nil
}

// getBuildLog gets the console output of a build
func getBuildLog(ctx context.Context, jobName, buildNumber string) error {
	encodedJobName := url.PathEscape(jobName)
	path := fmt.Sprintf("/job/%s/%s/consoleText", encodedJobName, buildNumber)

	resp, err := makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return fmt.Errorf("failed to get build log: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to get build log: %s - %s", resp.Status, string(body))
	}

	// Stream the log output to stdout
	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read build log: %w", err)
	}

	return nil
}

// getLastBuild gets details of the last build of a job
func getLastBuild(ctx context.Context, jobName string) error {
	encodedJobName := url.PathEscape(jobName)
	path := fmt.Sprintf("/job/%s/lastBuild/api/json", encodedJobName)

	resp, err := makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return fmt.Errorf("failed to get last build: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to get last build: %s - %s", resp.Status, string(body))
	}

	var build BuildDetail
	if err := json.NewDecoder(resp.Body).Decode(&build); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	printField("Build Number", build.Number)
	printField("URL", build.URL)
	status := build.Result
	if build.Building {
		status = "BUILDING"
	}
	printField("Status", status)
	if build.Description != "" {
		printField("Description", build.Description)
	}
	if build.Timestamp > 0 {
		buildTime := time.Unix(build.Timestamp/1000, 0)
		printField("Started", buildTime.Format(time.RFC1123))
	}
	if build.Duration > 0 {
		duration := time.Duration(build.Duration) * time.Millisecond
		printField("Duration", duration.String())
	}

	return nil
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
	default:
		return strings.ToUpper(color)
	}
}
