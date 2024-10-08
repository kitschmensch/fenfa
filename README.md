# Fenfa 分发

fēnfā: to distribute

Fenfa is a simple file-sharing program that generates temporary links for accessing files on your server. Fenfa generates links for sharing individual files or entire directories (which are zipped automatically before serving) without having to share credentials or go through an intermediary like wetransfer. This program is intended to run within a shared server environment where permissions are limited, but could also work well on a home computer if port forwarding is configured on your home network.

## Features

- **Easily Start/Stop HTTP Service**: Run Fenfa as a background service to manage the sharing of files.
- **Generate Temporary Links**: Share files or directories with time-limited access links.
- **IP Banning**: Automatically bans IPs after a configured number of failed access attempts.
- **Rate Limiting**: Limit global calls per minute.
- **Link Expiration**: Links expire after a configurable time period.
- **Directory Sharing**: Serve entire directories as a zipped archive, with options for zip depth and max zip file size.
- **Logging**: All activity is logged to a local log file.

## Installation

1. **Clone the Repository**:

   ```bash
   git clone https://github.com/kitschmensch/fenfa.git
   cd fenfa
   ```

2. **Install Dependencies**:

   To install the required dependencies specified in the go.mod file, run the following command:

   ```bash
   go mod tidy
   ```

   This command will download and install the necessary packages:

   - [github.com/joho/godotenv v1.5.1](https://github.com/joho/godotenv): Used for loading environment variables from a .env file.
   - [github.com/mattn/go-sqlite3 v1.14.23](https://github.com/mattn/go-sqlite3): SQLite3 database driver.

   go-sqlite3 requires gcc to compile. Furthermore, if you intend to cross compile you should install musl-cross (brew install FiloSottile/musl-cross/musl-cross). For example, when targeting Linux on x86_64 architecture use:

   ```bash
   CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ GOARCH=amd64 GOOS=linux CGO_ENABLED=1 go build -o fenfa -ldflags "-linkmode external -extldflags -static"
   ```

3. **Build the Project**:

   Make sure you have [Go installed](https://golang.org/doc/install), then run:

   ```bash
   go build -o fenfa
   ```

4. **Set up Environment Variables**:

   Fenfa uses environment variables for configuration. Create a `.env` file in the project directory and include the following variables:

   ```bash
    FENFA_HOST=http://localhost
    FENFA_PORT=12000
    FENFA_DEFAULT_EXPIRATION_PERIOD=86400
    FENFA_FAILED_ATTEMPT_LIMIT=10
    FENFA_MAX_ZIP_DEPTH=3
    FENFA_MAX_ZIP_SIZE=10737418240
    FENFA_ZIP_DIRECTORY=".fenfa"
   ```

5. **Add to PATH**:
   Depending on your shell, open your console config file:

   bash:

   ```bash
   nano ~/.bashrc
   ```

   zsh:

   ```bash
   nano ~/.zshrc
   ```

   Then add an export for Fenfa:

   ```bash
   export PATH="$PATH:/path/to/fenfa"
   ```

   Then restart your shell. To verify that Fenfa is in your path, you can run:

   ```bash
   which fenfa
   ```

## Usage

Fenfa supports the following commands:

- **Start the HTTP Server**:

  ```bash
  fenfa start
  ```

- **Stop the HTTP Server**:

  ```bash
  fenfa stop
  ```

- **Generate a shareable link**: This command generates a time-limited link for a specified file or directory.

  ```bash
  fenfa link /path/to/file
  $ https://localhost:1200/d9b0dc2c2a1b77aa03ee2f3e3004bca687030fba0e67d9a16ebb1fa3b78a4570
  ```

- **List active links**: List all active links stored in the system.

  ```bash
  fenfa list entries
  ```

- **List failed IP attempts**: List all active links stored in the system.

  ```bash
  fenfa list ip_attempts
  ```

- **Unban an IP**: Reset the failed attempts for a specific IP to unban it.

  ```bash
  fenfa unban [IP address]
  ```

## Configuration

Fenfa is configured using environment variables, which can be set in a `.env` file. The following settings are available:

- **`FENFA_HOST`**: A slug to configure the URL output to the terminal.
- **`FENFA_PORT`**: The port on which the service will run.
- **`FENFA_TEMPLATE_INCLUDES_PORT`**: Boolean, whether to append ":port" at the end of the URL output.
- **`FENFA_DEFAULT_EXPIRATION_PERIOD`**: The default expiration period for generated links (in seconds). For example, `86400` seconds is equal to 24 hours.
- **`FENFA_FAILED_ATTEMPT_LIMIT`**: The number of failed access attempts allowed before an IP is banned from accessing the service.
- **`FENFA_MAX_ZIP_DEPTH`**: How many subdirectories deep to consider when zipping directories
- **`FENFA_MAX_ZIP_SIZE`**: When zipping a directory, the size is estimated before zipping. If the estimated size is greater than this variable, the request will be cancelled.

## Implementation Details

- All logs are written to `fenfa.log` by default. You can change the log file location in the `main.go` file when initializing the logger.

## Planned Improvements

- Functions for tracking statistics, like the number of times a link was downloaded.
- A configurable process to automatically delete zips for expired links.
- A command to delete links, and corresponding zip files, if any.
- Better logging.
- Configure log level.
