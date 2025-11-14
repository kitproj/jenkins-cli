# Jenkins CLI

A Golang-based CLI for interacting with Jenkins. Inspired by the [jira-cli](https://github.com/kitproj/jira-cli), it provides a simple and efficient way for humans and automation tools to interact with Jenkins from the command line.

Like `jq`, it is designed to be a single, lightweight binary without the overhead of installing additional runtimes, and without the need to store your Jenkins API token in plain text files (it uses the system keyring).

## Features

- 🔐 **Secure credential storage** - Uses system keyring for API tokens
- 📦 **Single binary** - No dependencies, just download and run
- 🚀 **Simple commands** - Intuitive command structure
- 🔧 **Jenkins operations** - List jobs, trigger builds, get build status, view logs

## Installation

### Supported Platforms

Binaries will be available for:
- **Linux**: amd64, arm64
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)
- **Windows**: amd64

### Build from Source

```bash
git clone https://github.com/kitproj/jenkins-cli.git
cd jenkins-cli
go build -o jenkins .
```

Then move the binary to your PATH:

```bash
# Linux/macOS
sudo mv jenkins /usr/local/bin/

# Or just add to your local bin
mkdir -p ~/bin
mv jenkins ~/bin/
export PATH="$HOME/bin:$PATH"
```

## Configuration

### Getting a Jenkins API Token

Before configuring, you'll need to create a Jenkins API token:

1. Log in to your Jenkins instance
2. Go to your user configuration page: `https://your-jenkins-host/user/your-username/configure`
3. Under the "API Token" section, click "Add new Token"
4. Give it a name (e.g., "jenkins-cli")
5. Click "Generate"
6. Copy the generated token (you won't be able to see it again)

### Configure the CLI

The `jenkins` CLI can be configured in two ways:

1. **Using the configure command (recommended, secure)**:
   ```bash
   jenkins configure your-jenkins-host.com your-username
   # Then enter your API token when prompted
   ```
   This stores the host in `~/.config/jenkins-cli/config.json` and the token securely in your system's keyring.
   
   **Note:** Provide only the hostname without `http://` or `https://` prefix. The CLI will automatically use HTTPS.

2. **Using environment variables**:
   ```bash
   export JENKINS_HOST=your-jenkins-host.com
   export JENKINS_USER=your-username
   export JENKINS_TOKEN=your-api-token
   ```
   Note: The JENKINS_TOKEN environment variable is supported for backward compatibility, but using the keyring (via `jenkins configure`) is more secure on multi-user systems.

## Usage

```
Usage:
  jenkins configure <host> [username] - Configure Jenkins host and API token (reads token from stdin)
  jenkins list-jobs - List all Jenkins jobs
  jenkins get-job <job-name> - Get details of a specific job
  jenkins build-job <job-name> - Trigger a build for a job
  jenkins get-build <job-name> <build-number> - Get details of a specific build
  jenkins get-build-log <job-name> <build-number> - Get the console output of a build
  jenkins get-last-build <job-name> - Get details of the last build
```

### Examples

**Configure Jenkins CLI:**
```bash
jenkins configure jenkins.example.com myusername
# Enter your API token when prompted
```

**List all jobs:**
```bash
jenkins list-jobs
# Output:
# Found 5 job(s):
#
# my-application-build              SUCCESS         https://jenkins.example.com/job/my-application-build/
# integration-tests                 FAILURE         https://jenkins.example.com/job/integration-tests/
# nightly-deploy                    SUCCESS         https://jenkins.example.com/job/nightly-deploy/
```

**Get job details:**
```bash
jenkins get-job my-application-build
# Output:
# Job Name:            my-application-build
# URL:                 https://jenkins.example.com/job/my-application-build/
# Status:              SUCCESS
# Description:         Builds the main application
# Last Build:          #42 - SUCCESS
# Last Success:        #42
```

**Trigger a build:**
```bash
jenkins build-job my-application-build
# Output:
# Successfully triggered build for job: my-application-build
```

**Get build details:**
```bash
jenkins get-build my-application-build 42
# Output:
# Build Number:        42
# URL:                 https://jenkins.example.com/job/my-application-build/42/
# Status:              SUCCESS
# Started:             Mon, 14 Nov 2025 10:30:00 UTC
# Duration:            2m15s
```

**Get the last build:**
```bash
jenkins get-last-build my-application-build
# Shows details of the most recent build
```

**View build logs:**
```bash
jenkins get-build-log my-application-build 42
# Streams the console output of build #42
```

## Troubleshooting

### Common Issues

**"host is required" error**
- Make sure you've run `jenkins configure <host>` or set the `JENKINS_HOST` environment variable
- Check that the config file exists: `cat ~/.config/jenkins-cli/config.json`

**"token not found" or authentication errors**
- Verify your API token is still valid
- Re-run the configure command to update the token: `jenkins configure your-jenkins-host.com your-username`
- Make sure your Jenkins user has permission to access the jobs

**Connection errors**
- Verify the Jenkins host is accessible
- Check if your Jenkins instance requires HTTPS
- Some corporate networks may require proxy configuration

**Keyring issues on Linux**
- Some Linux systems may not have a keyring service installed
- Install `gnome-keyring` or `kwallet` for your desktop environment
- Alternatively, use environment variables: `export JENKINS_TOKEN=your-token`

### Getting Help

- Report issues: https://github.com/kitproj/jenkins-cli/issues
- Check existing issues for solutions and workarounds

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o jenkins .
```

### Cross-compilation

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o jenkins-linux-amd64 .
GOOS=linux GOARCH=arm64 go build -o jenkins-linux-arm64 .

# macOS
GOOS=darwin GOARCH=amd64 go build -o jenkins-darwin-amd64 .
GOOS=darwin GOARCH=arm64 go build -o jenkins-darwin-arm64 .

# Windows
GOOS=windows GOARCH=amd64 go build -o jenkins-windows-amd64.exe .
```

## License

See [LICENSE](LICENSE) file for details.
