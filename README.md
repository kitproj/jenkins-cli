# Jenkins CLI & MCP Server

A Golang-based CLI and MCP server for interacting with Jenkins. Inspired by the [jira-cli](https://github.com/kitproj/jira-cli), it provides a simple and efficient way for humans and automation tools to interact with Jenkins from the command line.

Being both a CLI and an MCP server means you get the best of both worlds. Humans can use the CLI commands directly, while AI agents can use the MCP server to perform Jenkins operations programmatically.

Like `jq`, it is designed to be a single, lightweight binary without the overhead of installing additional runtimes, and without the need to store your Jenkins API token in plain text files (it uses the system keyring).

## Features

- 🔐 **Secure credential storage** - Uses system keyring for API tokens
- 📦 **Single binary** - No dependencies, just download and run
- 🚀 **Simple commands** - Intuitive command structure
- 🔧 **Jenkins operations** - List jobs, trigger builds, get build status, view logs
- 🤖 **MCP Server** - Model Context Protocol server for AI agent integration

## Installation

### Supported Platforms

Binaries are available for:
- **Linux**: amd64, arm64
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)
- **Windows**: amd64

### Download and Install

Download the binary for your platform from the [release page](https://github.com/kitproj/jenkins-cli/releases).

#### Linux

**For Linux (amd64):**
```bash
sudo curl -fsL -o /usr/local/bin/jenkins https://github.com/kitproj/jenkins-cli/releases/download/v0.0.1/jenkins_v0.0.1_linux_amd64
sudo chmod +x /usr/local/bin/jenkins
```

**For Linux (arm64):**
```bash
sudo curl -fsL -o /usr/local/bin/jenkins https://github.com/kitproj/jenkins-cli/releases/download/v0.0.1/jenkins_v0.0.1_linux_arm64
sudo chmod +x /usr/local/bin/jenkins
```

#### macOS

**For macOS (Apple Silicon/arm64):**
```bash
sudo curl -fsL -o /usr/local/bin/jenkins https://github.com/kitproj/jenkins-cli/releases/download/v0.0.1/jenkins_v0.0.1_darwin_arm64
sudo chmod +x /usr/local/bin/jenkins
```

**For macOS (Intel/amd64):**
```bash
sudo curl -fsL -o /usr/local/bin/jenkins https://github.com/kitproj/jenkins-cli/releases/download/v0.0.1/jenkins_v0.0.1_darwin_amd64
sudo chmod +x /usr/local/bin/jenkins
```

#### Verify Installation

After installing, verify the installation works:
```bash
jenkins -h
```

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
   jenkins configure https://your-jenkins-host.com your-username
   # Then enter your API token when prompted
   ```
   This stores the URL and username in `~/.config/jenkins-cli/config.json` and the token securely in your system's keyring.
   
   **Note:** The URL should be a fully formed URL including the protocol (e.g., `https://jenkins.example.com` or `http://localhost:8080`). If your Jenkins instance is at a subpath, include it in the URL (e.g., `https://example.com/jenkins`).

2. **Using environment variables**:
   ```bash
   export JENKINS_URL=https://your-jenkins-host.com
   # Or with a subpath:
   # export JENKINS_URL=https://your-jenkins-host.com/jenkins
   export JENKINS_USER=your-username
   export JENKINS_TOKEN=your-api-token
   ```
   Note: The JENKINS_TOKEN environment variable is supported for backward compatibility, but using the keyring (via `jenkins configure`) is more secure on multi-user systems.

## Usage

```
Usage:
  jenkins configure <url> [username] - Configure Jenkins URL and API token (reads token from stdin)
  jenkins list-jobs - List all Jenkins jobs
  jenkins get-job <job-name> - Get details of a specific job
  jenkins build-job <job-name> - Trigger a build for a job
  jenkins get-build <job-name> <build-number> - Get details of a specific build
  jenkins get-build-log <job-name> <build-number> - Get the console output of a build
  jenkins get-last-build <job-name> - Get details of the last build
  jenkins mcp-server - Start MCP server (Model Context Protocol)
```

### Examples

**Configure Jenkins CLI:**
```bash
jenkins configure https://jenkins.example.com myusername
# Enter your API token when prompted
```

**Configure Jenkins CLI with a base path (e.g., for Jenkins at https://example.com/jenkins):**
```bash
jenkins configure https://example.com/jenkins myusername
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
# Started:             5 minutes ago
# Duration:            135s
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

## MCP Server Mode

The jenkins-cli can also run as an MCP (Model Context Protocol) server, allowing AI agents to interact with Jenkins through a standardized protocol.

### Starting the MCP Server

```bash
jenkins mcp-server
```

The MCP server communicates over standard input/output (stdio) and provides the following tools for AI agents:

- **list_jobs** - List all Jenkins jobs with their status and URL
- **get_job** - Get details of a specific Jenkins job including status, description, and build history
- **build_job** - Trigger a build for a Jenkins job
- **get_build** - Get details of a specific build including status, duration, and timestamp
- **get_build_log** - Get the console output of a specific build
- **get_last_build** - Get details of the last build of a Jenkins job

### MCP Server Configuration

The MCP server uses the same configuration as the CLI:
- Configuration file: `~/.config/jenkins-cli/config.json` (URL and username)
- Credentials stored securely in system keyring
- Environment variables `JENKINS_URL` and `JENKINS_TOKEN` are also supported

### Using with AI Agents

Configure your AI agent or MCP client to use the jenkins-cli MCP server. The server will handle all Jenkins operations through the MCP protocol, providing a secure and standardized way for AI agents to interact with Jenkins.

Example MCP client configuration:
```json
{
  "mcpServers": {
    "jenkins": {
      "command": "jenkins",
      "args": ["mcp-server"]
    }
  }
}
```

## Troubleshooting

### Common Issues

**"Jenkins URL is required" error**
- Make sure you've run `jenkins configure <url>` or set the `JENKINS_URL` environment variable
- Check that the config file exists: `cat ~/.config/jenkins-cli/config.json`

**"token not found" or authentication errors**
- Verify your API token is still valid
- Re-run the configure command to update the token: `jenkins configure https://your-jenkins-host.com your-username`
- Make sure your Jenkins user has permission to access the jobs

**Connection errors**
- Verify the Jenkins URL is accessible
- Check if your Jenkins instance requires HTTPS
- Some corporate networks may require proxy configuration

**Keyring issues on Linux**
- Some Linux systems may not have a keyring service installed
- Install `gnome-keyring` or `kwallet` for your desktop environment
- Alternatively, use environment variables: `export JENKINS_TOKEN=your-token`

### Getting Help

- Report issues: https://github.com/kitproj/jenkins-cli/issues
- Check existing issues for solutions and workarounds

## For Developers

### Releasing a New Version

When you push a new tag (e.g., `v0.0.3`), the GitHub Actions release workflow will automatically:

1. Update all version references in this README.md
2. Commit the updated README back to the main branch
3. Build binaries for all supported platforms
4. Create a GitHub release with the built binaries and checksums

To create a new release:

```bash
git tag v0.0.3
git push origin v0.0.3
```

The README will be automatically updated with the new version number in all installation instructions.

## License

See [LICENSE](LICENSE) file for details.
