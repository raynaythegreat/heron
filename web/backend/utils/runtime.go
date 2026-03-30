package utils

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/raynaythegreat/heron/pkg/config"
)

// GetPicoclawHome returns the heron home directory.
// Priority: $HERON_HOME > ~/.heron
func GetPicoclawHome() string {
	if home := os.Getenv(config.EnvHome); home != "" {
		return home
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".heron")
}

// GetHeronHome returns the heron home directory (alias for GetPicoclawHome).
func GetHeronHome() string {
	return GetPicoclawHome()
}

// GetDefaultConfigPath returns the default path to the heron config file.
func GetDefaultConfigPath() string {
	if configPath := os.Getenv(config.EnvConfig); configPath != "" {
		return configPath
	}
	return filepath.Join(GetPicoclawHome(), "config.json")
}

// FindPicoclawBinary locates the heron executable.
// Search order:
//  1. HERON_BINARY environment variable (explicit override)
//  2. Same directory as the current executable
//  3. Falls back to "heron" and relies on $PATH
func FindPicoclawBinary() string {
	binaryName := "heron"
	if runtime.GOOS == "windows" {
		binaryName = "heron.exe"
	}

	if p := os.Getenv(config.EnvBinary); p != "" {
		if info, _ := os.Stat(p); info != nil && !info.IsDir() {
			return p
		}
	}

	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), binaryName)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}

	return "heron"
}

// GetLocalIP returns the local IP address of the machine.
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return ""
}

// OpenBrowser automatically opens the given URL in the default browser.
func OpenBrowser(url string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}
