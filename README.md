# bricks

Command line downloader for vadapav.mov, it replicates the directory structure on website.

## Installation
### Option 1: Download Pre-built Binary
1. Go to the [Releases](https://github.com/vadapavmov/bricks/releases) page of this repository.
2. Download the appropriate binary for your platform (e.g., `bricks-linux` for Linux, `bricks-windows.exe` for Windows).
3. Rename the downloaded binary to `bricks`.

### Option 2: Build from Source
1. `go build -o bricks cmd/bricks/main.go`

## Usage
The basic usage of `bricks` is as follows:
```shell
./bricks [options]
```
- `[options]`:
    - `-path` (default: "."): Specifies the download path on your local machine.
    - `-n` (default: 3): Specifies the number of parallel file downloads. (Note: Limited to a maximum of 5 parallel downloads to prevent DDOS.)

### Example
`./bricks --path=Downloads/`

### TODO
- Retry capability 
- Better progressbar
