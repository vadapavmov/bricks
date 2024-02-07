package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
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

	// Get dirId
	dirId := getDir()

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

func getDir() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter URL: ")
	inputURL, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("failed to read URL: %v", err)
	}
	inputURL = strings.TrimSpace(inputURL) // Trim newline and whitespaces

	// Parse URL and extract UUID
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		log.Fatalf("failed to parse URL: %v", err)
	}
	uuidStr := parsedURL.Fragment

	// Validate UUID
	if _, err = uuid.Parse(uuidStr); err != nil {
		log.Fatalf("invalid UUID format in URL: %v", err)
	}

	return uuidStr
}
