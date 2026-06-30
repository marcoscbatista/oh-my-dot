package dotfiles

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const storeRelativePath = "oh-my-dot/db.json"

func DefaultStorePath() string {
	return storeRelativePath
}

func DefaultDotfilesDir(home string) string {
	return filepath.Join(home, "oh-my-dot")
}

func DefaultConfigPath(home string) string {
	if configHome := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); configHome != "" {
		return configHome
	}

	return filepath.Join(home, ".config")
}

func normalizePackageName(name string) (string, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return "", fmt.Errorf("name cannot be empty")
	}

	if filepath.IsAbs(name) || strings.ContainsAny(name, `/\`) {
		return "", fmt.Errorf("name must be a single directory name")
	}

	if name == "." || name == ".." {
		return "", fmt.Errorf("name %q is not allowed", name)
	}

	return name, nil
}

func resolvePackagePath(dotfilesDir string, name string) (string, error) {
	name, err := normalizePackageName(name)
	if err != nil {
		return "", err
	}

	return filepath.Join(dotfilesDir, name), nil
}
