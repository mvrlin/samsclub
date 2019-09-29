package chromium

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/mholt/archiver"

	"github.com/cavaliercoder/grab"
)

func getExecutablePath(localPath string) string {
	folder := path.Join(localPath, getPlatformFolder())

	switch runtime.GOOS {
	case "darwin":
		return path.Join(folder, "Chromium.app/Contents/MacOS/Chromium")
	case "linux":
		return path.Join(folder, "chrome")
	case "windows":
		return path.Join(folder, "chrome.exe")
	}

	return ""
}

func getPlatformFolder() string {
	switch runtime.GOOS {
	case "darwin":
		return "chrome-mac"
	case "linux":
		return "chrome-linux"
	case "windows":
		return "chrome-win"
	}

	return ""
}

func getPlatformURL() string {
	platformFolder := getPlatformFolder()

	slice := map[string]string{
		"darwin":  "https://storage.googleapis.com/chromium-browser-snapshots/Mac/%d/%s.zip",
		"linux":   "https://storage.googleapis.com/chromium-browser-snapshots/Linux_x64/%d/%s.zip",
		"windows": "https://storage.googleapis.com/chromium-browser-snapshots/Win_x64/%d/%s.zip",
	}

	return fmt.Sprintf(slice[runtime.GOOS], 698779, platformFolder)
}

// Download is downloading chromium based on platform.
func Download() (string, error) {
	downloadPath := path.Join(os.TempDir(), "chromium.zip")

	client := grab.NewClient()
	req, _ := grab.NewRequest(downloadPath, getPlatformURL())

	resp := client.Do(req)

	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

	bar := pb.New64(resp.Size)
	bar.SetTemplateString(`{{ "Downloading" }} ({{ percent . }}) {{ bar . "[" "=" "=" "." "]"}} {{ speed . }}`)
	bar.Set(pb.Bytes, true)
	bar.Start()

	go func() {
		for {
			select {
			case <-t.C:
				bar.SetCurrent(resp.BytesComplete())

			case <-resp.Done:
				bar.SetCurrent(resp.Size)
				bar.Finish()

				return
			}
		}
	}()

	if err := resp.Err(); err != nil {
		return "", err
	}

	return downloadPath, nil
}

// ExecPath returns path where chromium is at.
func ExecPath() (string, error) {
	homeDir, _ := os.UserHomeDir()
	localPath := path.Join(homeDir, ".local-chromium")

	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		downloadPath, err := Download()
		if err != nil {
			return "", err
		}

		if err = archiver.Unarchive(downloadPath, localPath); err != nil {
			return "", err
		}
	}

	return getExecutablePath(localPath), nil
}
