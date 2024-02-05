package app

import (
	"encoding/json"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

const DefaultBaseURL = "https://vadapav.mov"

type Resp struct {
	Data File `json:"data"`
}

type File struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	IsDir bool   `json:"dir"`
	Files []File `json:"files,omitempty"`
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

func (app *App) downloadDir(id, parentPath string, ch chan struct{}, wg *sync.WaitGroup) error {
	// Get directory from API
	dir, err := app.fetchDir(id)
	if err != nil {
		return err
	}
	// Create directory
	dirPath := filepath.Join(parentPath, dir.Name)
	log.Printf("Creating directory %s", dirPath)
	err = os.Mkdir(dirPath, 0755)
	if err != nil {
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
			go func(file File) {
				defer wg.Done()
				// Limit parallel downloads using the channel
				ch <- struct{}{}
				defer func() {
					<-ch
				}()

				if err := app.downloadFile(file.ID, dirPath, file.Name); err != nil {
					log.Printf("Failed to download %s: %v", file.Name, err)
				}
			}(file)
		}
	}
	return nil
}

func (app *App) fetchDir(id string) (*File, error) {
	resp, err := http.Get(app.baseURL + "/api/d/" + id)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("something went wrong, recevied error code %d", resp.StatusCode)
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

func (app *App) downloadFile(id, dirPath, name string) error {
	path := filepath.Join(dirPath, name)

	resp, err := http.Get(app.baseURL + "/f/" + id)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, _ := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"Downloading "+name,
	)
	_, err = io.Copy(io.MultiWriter(f, bar), resp.Body)
	return err
}
