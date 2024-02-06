package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/vadapavmov/bricks/internal/app"
)

const Version = "v1.1.0"
const MaxAllowedParallelDownloads = 10

func main() {
	// Define default values and usage messages for flags
	downloadPath := flag.String("path", ".", "Download path")
	baseURL := flag.String("server", "https://vadapav.mov", "Base server url")
	parallelDownloads := flag.Int("n", 3, "Number of parallel file downloads")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	// Show version
	if *showVersion {
		fmt.Printf("Bricks %s\n", Version)
		return
	}

	// Check if the remaining argument (dirId) is provided
	if flag.NArg() != 1 {
		fmt.Println("Usage: bricks -path /your/download/path -url https://mirror.url dirId")
		return
	}

	// Get dirId
	dirId := flag.Arg(0)

	// Build absolute path
	abspath, err := filepath.Abs(*downloadPath)
	if err != nil {
		log.Fatalf("invalid path %s", *downloadPath)
	}

	// Check if the specified download path exists
	if _, err := os.Stat(abspath); os.IsNotExist(err) {
		log.Fatalf("download path %s does not exist", abspath)
	}

	// To save site from DDOS
	if *parallelDownloads > MaxAllowedParallelDownloads {
		log.Fatalf("max parallel downloads can't be larger than %d", MaxAllowedParallelDownloads)
	}

	// Run the app
	bricks := app.New(*baseURL)
	if err = bricks.Run(dirId, abspath, *parallelDownloads); err != nil {
		log.Fatalf("failed to download %v", err)
	}
}
