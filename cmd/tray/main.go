// Command tray is the long-running background companion to the per-job router.
// It sits in the Windows system tray, shows recent routing activity (by tailing
// the router's log file), and offers quick actions: open config, open logs,
// reload, quit. The actual print routing is done by cmd/router (invoked by
// Redmon per job) — this process does no printing itself.
//
// On Windows a true Windows Service runs in session 0 and cannot show a tray
// icon, so this is meant to run as a per-user startup app (the installer adds it
// to the Run key) with no visible console window. Build with:
//
//	go build -ldflags "-H windowsgui" -o spr-tray.exe ./cmd/tray
package main

import (
	"bufio"
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"fyne.io/systray"

	"github.com/yourorg/smart-print-router/config"
)

// Optional tray icon. Drop a 16x16/32x32 .ico at cmd/tray/assets/tray.ico
// before building and it will be embedded. If absent, the tray still works
// (the entry may appear without a custom icon).
//
//go:embed assets/tray.ico
var iconICO []byte

func main() {
	systray.Run(onReady, func() {})
}

func onReady() {
	if len(iconICO) > 0 {
		systray.SetIcon(iconICO)
	}
	systray.SetTitle("Smart Print Router")
	systray.SetTooltip("Smart Print Router — running")

	mStatus := systray.AddMenuItem("Idle — waiting for print jobs", "")
	mStatus.Disable()
	systray.AddSeparator()
	mOpenLog := systray.AddMenuItem("Open log", "Open router.log")
	mOpenCfg := systray.AddMenuItem("Edit config.json", "Open the config in the default editor")
	mOpenDir := systray.AddMenuItem("Open install folder", "")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Stop the tray (routing keeps working)")

	cfg, _ := config.Load("")
	logPath := cfg.LogFile
	cfgPath := config.DefaultPath()
	installDir := filepath.Dir(cfgPath)

	// Tail the log in the background and surface the latest line as status.
	go tailStatus(logPath, mStatus)

	go func() {
		for {
			select {
			case <-mOpenLog.ClickedCh:
				openPath(logPath)
			case <-mOpenCfg.ClickedCh:
				openPath(cfgPath)
			case <-mOpenDir.ClickedCh:
				openPath(installDir)
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

var statusMu sync.Mutex

// tailStatus follows the log file and pushes the most recent meaningful line
// into the status menu item. Polls so it survives log rotation / late creation.
func tailStatus(path string, item *systray.MenuItem) {
	if path == "" {
		return
	}
	var lastSize int64
	for {
		time.Sleep(1 * time.Second)
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		info, err := f.Stat()
		if err != nil {
			f.Close()
			continue
		}
		if info.Size() == lastSize {
			f.Close()
			continue
		}
		// Read from where we left off (or near the end on first pass).
		start := lastSize
		if start > info.Size() {
			start = 0 // file was truncated/rotated
		}
		f.Seek(start, 0)
		sc := bufio.NewScanner(f)
		var last string
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line != "" {
				last = line
			}
		}
		lastSize = info.Size()
		f.Close()
		if last != "" {
			statusMu.Lock()
			item.SetTitle(truncate(last, 80))
			statusMu.Unlock()
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

// openPath opens a file or folder with the OS default handler.
func openPath(path string) {
	if path == "" {
		return
	}
	switch runtime.GOOS {
	case "windows":
		exec.Command("cmd", "/c", "start", "", path).Start()
	case "darwin":
		exec.Command("open", path).Start()
	default:
		exec.Command("xdg-open", path).Start()
	}
}
