# bricks

Command line downloader for vadapav.mov, it replicates the directory structure on website.

## Installation
### Option 1: Download Pre-built Binary
1. Go to the [Releases](https://github.com/your/repo/releases) page of this repository.
2. Download the appropriate binary for your platform (e.g., `bricks-linux` for Linux, `bricks-windows.exe` for Windows).
3. Rename the downloaded binary to `bricks`.

### Option 2: Build from Source
1. `go build -o bricks cmd/bricks/main.go`

## Usage
The basic usage of `bricks` is as follows:
```shell
./bricks [options] dirId
```
- `[options]`:
    - `-path` (default: "."): Specifies the download path on your local machine.
    - `-server` (default: "https://vadapav.mov"): Specifies the base server URL from which to download files and directories.
    - `-n` (default: 3): Specifies the number of parallel file downloads. (Note: Limited to a maximum of 10 parallel downloads to prevent DDOS.)

- `dirId`: The id of that you want to download from the server.

### Example
`./bricks -path /path/to/downloads -server https://mirror.vadapav.mov -n 4 28dc7aeb-902b-4824-8be2-fa1e4f20383c`
