package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

const RootDirId = "11111111-1111-1111-1111-111111111111"
const DefaultBaseURL = "https://vadapav.mov"

type Resp struct {
	Data File `json:"data"`
}

type File struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	IsDir  bool    `json:"dir"`
	Parent string  `json:"parent"`
	Files  []*File `json:"files,omitempty"`
}

type App struct {
	baseURL string
}

func New(baseURL string) *App {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &App{baseURL: baseURL}
}

func (app *App) Run(dirId, path string, n int) error {
	// Create a channel with a size of n for parallel downloads
	ch := make(chan struct{}, n)
	defer close(ch)
	var wg sync.WaitGroup

	err := app.downloadDir(dirId, path, ch, &wg)
	if err != nil {
		return err
	}

	// Wait for all downloads to finish
	wg.Wait()

	return nil
}

func (app *App) doReq(method, endpoint string, startPos int64) (*http.Response, error) {
	var req *http.Request
	var err error

	if startPos > 0 {
		req, err = http.NewRequest(method, app.baseURL+endpoint, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", startPos))
	} else {
		req, err = http.NewRequest("GET", app.baseURL+endpoint, nil)
		if err != nil {
			return nil, err
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		handleRateLimit(resp)
		return app.doReq(method, endpoint, startPos)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
	return resp, nil
}

func (app *App) downloadDir(id, parentPath string, ch chan struct{}, wg *sync.WaitGroup) error {
	// Get directory from API
	dir, err := app.fetchDir(id)
	if err != nil {
		return err
	}
	// DDos
	if dir.Parent == RootDirId {
		log.Fatalf("you are not allowed to download %s ", dir.Name)
	}
	// Create directory
	dirPath := filepath.Join(parentPath, dir.Name)
	log.Printf("Creating directory %s", dirPath)
	err = os.Mkdir(dirPath, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	// Loop over subdirectories
	for _, file := range dir.Files {
		if file.IsDir {
			if err = app.downloadDir(file.ID, dirPath, ch, wg); err != nil {
				return err
			}
		} else {
			wg.Add(1)
			go func(file *File) {
				defer wg.Done()
				// Limit parallel downloads using the channel
				ch <- struct{}{}
				defer func() {
					<-ch
				}()

				if err = app.downloadFile(file, dirPath); err != nil {
					log.Printf("Failed to download %s: %v", file.Name, err)
				}
			}(file)
		}
	}
	return nil
}

func (app *App) fetchDir(id string) (*File, error) {
	resp, err := app.doReq(http.MethodGet, "/api/d/"+id, 0)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var r Resp
	if err = json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func (app *App) downloadFile(file *File, dirPath string) error {
	path := filepath.Join(dirPath, file.Name)

	// Head request to get the content length
	headResp, err := app.doReq(http.MethodHead, "/f/"+file.ID, 0)
	if err != nil {
		return err
	}
	defer headResp.Body.Close()

	contentLength := headResp.ContentLength
	var startPos int64 = 0

	// Check if the file exists
	fileInfo, err := os.Stat(path)
	if err == nil {
		// File exists, check its size
		if fileInfo.Size() == contentLength {
			// File is already fully downloaded
			log.Printf("skipping file %s, file already exists with same size", file.Name)
			return nil
		} else if fileInfo.Size() < contentLength {
			// Partial file exists, resume download
			log.Printf("partial file %s found, resuming download", file.Name)
			startPos = fileInfo.Size()
		}
		// else: File is larger than expected, will be overwritten
	}

	// Get request with range header if resuming
	resp, err := app.doReq(http.MethodGet, "/f/"+file.ID, startPos)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Open file in append mode if resuming, otherwise create/overwrite
	fileFlag := os.O_CREATE | os.O_WRONLY
	if startPos > 0 {
		fileFlag |= os.O_APPEND
	} else {
		fileFlag |= os.O_TRUNC
	}

	f, err := os.OpenFile(path, fileFlag, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Create a progress bar
	bar := progressbar.DefaultBytes(contentLength-startPos, "Downloading "+file.Name)

	// Copy the response body to the file
	_, err = io.Copy(io.MultiWriter(f, bar), resp.Body)
	return err
}

func handleRateLimit(resp *http.Response) {
	retryAfterStr := resp.Header.Get("Retry-After")
	retryAfter, err := strconv.Atoi(retryAfterStr)
	if err != nil {
		log.Fatalf("failed to parse rate limit header %s", retryAfterStr)
	}
	time.Sleep(time.Duration(retryAfter+1) * time.Second)
}
