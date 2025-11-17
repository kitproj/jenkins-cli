package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/bndr/gojenkins"
	"github.com/kitproj/jenkins-cli/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// runMCPServer starts the MCP server that communicates over stdio using the mcp-go library
func runMCPServer(ctx context.Context) error {
	// Load URL and username from config file
	url, username, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("Jenkins URL must be configured (use 'jenkins configure <url>' or set JENKINS_URL env var)")
	}

	// Load token from keyring
	token, err := config.LoadToken(url)
	if err != nil {
		return fmt.Errorf("Jenkins token must be set (use 'jenkins configure <url>' or set JENKINS_TOKEN env var)")
	}

	if url == "" {
		return fmt.Errorf("Jenkins URL must be configured (use 'jenkins configure <url>')")
	}
	if token == "" {
		return fmt.Errorf("Jenkins token must be set (use 'jenkins configure <url>')")
	}

	// Use username from config, or default to "admin"
	if username == "" {
		username = "admin"
	}

	// Create Jenkins client with the full URL
	jenkinsClient, err := gojenkins.CreateJenkins(nil, url, username, token).Init(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Jenkins client: %w", err)
	}

	// Create a new MCP server
	s := server.NewMCPServer(
		"jenkins-cli-mcp-server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Add list-jobs tool
	listJobsTool := mcp.NewTool("list_jobs",
		mcp.WithDescription("List all Jenkins jobs with their status and URL"),
	)
	s.AddTool(listJobsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return listJobsHandler(ctx, jenkinsClient, request)
	})

	// Add get-job tool
	getJobTool := mcp.NewTool("get_job",
		mcp.WithDescription("Get details of a specific Jenkins job including status, description, and build history"),
		mcp.WithString("job_name",
			mcp.Required(),
			mcp.Description("Jenkins job name"),
		),
	)
	s.AddTool(getJobTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return getJobHandler(ctx, jenkinsClient, request)
	})

	// Add get-build tool
	getBuildTool := mcp.NewTool("get_build",
		mcp.WithDescription("Get details of a specific build including status, duration, and timestamp"),
		mcp.WithString("job_name",
			mcp.Required(),
			mcp.Description("Jenkins job name"),
		),
		mcp.WithString("build_number",
			mcp.Required(),
			mcp.Description("Build number (e.g., '42')"),
		),
	)
	s.AddTool(getBuildTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return getBuildHandler(ctx, jenkinsClient, request)
	})

	// Add get-build-log tool
	getBuildLogTool := mcp.NewTool("get_build_log",
		mcp.WithDescription("Get the console output of a specific build"),
		mcp.WithString("job_name",
			mcp.Required(),
			mcp.Description("Jenkins job name"),
		),
		mcp.WithString("build_number",
			mcp.Required(),
			mcp.Description("Build number (e.g., '42')"),
		),
	)
	s.AddTool(getBuildLogTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return getBuildLogHandler(ctx, jenkinsClient, request)
	})

	// Start the stdio server
	return server.ServeStdio(s)
}

func listJobsHandler(ctx context.Context, client *gojenkins.Jenkins, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jobs, err := client.GetAllJobNames(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list jobs: %v", err)), nil
	}

	// Filter out disabled jobs
	enabledJobs := []gojenkins.InnerJob{}
	for _, job := range jobs {
		if !strings.HasPrefix(job.Color, "disabled") {
			enabledJobs = append(enabledJobs, job)
		}
	}

	if len(enabledJobs) == 0 {
		return mcp.NewToolResultText("No jobs found"), nil
	}

	result := fmt.Sprintf("Found %d job(s):\n\n", len(enabledJobs))
	for _, job := range enabledJobs {
		status := getStatusFromColor(job.Color)
		result += fmt.Sprintf("%-40s %-15s %s\n", job.Name, status, job.Url)
	}

	return mcp.NewToolResultText(result), nil
}

func getJobHandler(ctx context.Context, client *gojenkins.Jenkins, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jobName, err := request.RequireString("job_name")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Missing or invalid 'job_name' argument: %v", err)), nil
	}

	job, err := client.GetJob(ctx, jobName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get job: %v", err)), nil
	}

	result := fmt.Sprintf("Job Name: %s\nURL: %s",
		job.GetName(),
		job.Raw.URL,
	)

	// Only show status if it's not empty
	status := getStatusFromColor(job.Raw.Color)
	if status != "" {
		result += fmt.Sprintf("\nStatus: %s", status)
	}

	if job.GetDescription() != "" {
		result += fmt.Sprintf("\nDescription: %s", job.GetDescription())
	}

	lastBuild, err := job.GetLastBuild(ctx)
	if err == nil && lastBuild != nil {
		result += fmt.Sprintf("\nLast Build: #%d - %s", lastBuild.GetBuildNumber(), lastBuild.GetResult())
	}

	lastSuccess, err := job.GetLastSuccessfulBuild(ctx)
	if err == nil && lastSuccess != nil {
		result += fmt.Sprintf("\nLast Success: #%d", lastSuccess.GetBuildNumber())
	}

	lastFailed, err := job.GetLastFailedBuild(ctx)
	if err == nil && lastFailed != nil {
		result += fmt.Sprintf("\nLast Failed: #%d", lastFailed.GetBuildNumber())
	}

	// Add inner jobs if they exist (for folders and multi-branch pipelines)
	innerJobs := job.GetInnerJobsMetadata()
	// Filter out disabled inner jobs
	enabledInnerJobs := []gojenkins.InnerJob{}
	for _, innerJob := range innerJobs {
		if !strings.HasPrefix(innerJob.Color, "disabled") {
			enabledInnerJobs = append(enabledInnerJobs, innerJob)
		}
	}
	if len(enabledInnerJobs) > 0 {
		result += fmt.Sprintf("\n\nInner Jobs (%d):", len(enabledInnerJobs))
		for _, innerJob := range enabledInnerJobs {
			status := getStatusFromColor(innerJob.Color)
			result += fmt.Sprintf("\n  %-38s %-15s %s", innerJob.Name, status, innerJob.Url)
		}
	}

	return mcp.NewToolResultText(result), nil
}

func getBuildHandler(ctx context.Context, client *gojenkins.Jenkins, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jobName, err := request.RequireString("job_name")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Missing or invalid 'job_name' argument: %v", err)), nil
	}

	buildNumberStr, err := request.RequireString("build_number")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Missing or invalid 'build_number' argument: %v", err)), nil
	}

	buildNumber, err := strconv.ParseInt(buildNumberStr, 10, 64)
	if err != nil || buildNumber <= 0 {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid build number: %s", buildNumberStr)), nil
	}

	job, err := client.GetJob(ctx, jobName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get job: %v", err)), nil
	}

	build, err := job.GetBuild(ctx, buildNumber)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get build: %v", err)), nil
	}

	status := build.GetResult()
	if build.IsRunning(ctx) {
		status = "BUILDING"
	}

	result := fmt.Sprintf("Build Number: %d\nURL: %s\nStatus: %s",
		build.GetBuildNumber(),
		build.GetUrl(),
		status,
	)

	if build.Raw.Description != "" {
		result += fmt.Sprintf("\nDescription: %s", build.Raw.Description)
	}

	buildTime := build.GetTimestamp()
	if !buildTime.IsZero() {
		result += fmt.Sprintf("\nStarted: %s", buildTime.Format("2006-01-02 15:04:05"))
	}

	duration := build.GetDuration()
	if duration > 0 {
		result += fmt.Sprintf("\nDuration: %s", formatDuration(duration))
	}

	return mcp.NewToolResultText(result), nil
}

func getBuildLogHandler(ctx context.Context, client *gojenkins.Jenkins, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jobName, err := request.RequireString("job_name")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Missing or invalid 'job_name' argument: %v", err)), nil
	}

	buildNumberStr, err := request.RequireString("build_number")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Missing or invalid 'build_number' argument: %v", err)), nil
	}

	buildNumber, err := strconv.ParseInt(buildNumberStr, 10, 64)
	if err != nil || buildNumber <= 0 {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid build number: %s", buildNumberStr)), nil
	}

	job, err := client.GetJob(ctx, jobName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get job: %v", err)), nil
	}

	build, err := job.GetBuild(ctx, buildNumber)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get build: %v", err)), nil
	}

	log := build.GetConsoleOutput(ctx)
	return mcp.NewToolResultText(log), nil
}
