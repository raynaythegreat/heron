package run

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func getRunDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".heron", "run")
}

func getPIDFile(name string) string {
	return filepath.Join(getRunDir(), name+".pid")
}

func WritePID(name string, pid int) error {
	os.MkdirAll(getRunDir(), 0o755)
	return os.WriteFile(getPIDFile(name), []byte(strconv.Itoa(pid)), 0o644)
}

func ReadPID(name string) (int, error) {
	data, err := os.ReadFile(getPIDFile(name))
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

func RemovePID(name string) error {
	return os.Remove(getPIDFile(name))
}
