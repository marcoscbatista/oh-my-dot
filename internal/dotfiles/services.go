package dotfiles

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type DotFilesService struct {
	Store *DotFileStore
}

func (d *DotFilesService) GetAll() ([]DotFile, error) {
	dots, err := d.Store.Load()
	if err != nil {
		return nil, err
	}

	return dots, nil
}

func (d *DotFilesService) Create(dotFile *DotFile) error {
	if dotFile == nil {
		return fmt.Errorf("dotfile cannot be nil")
	}

	return d.Store.Add(*dotFile)
}

func (d *DotFilesService) Switch(name string, configPath string, backupDir string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("dotfile name cannot be empty")
	}

	data, err := d.Store.Load()
	if err != nil {
		return err
	}

	var selected *DotFile

	for i := range data {
		if data[i].Name == name {
			selected = &data[i]
			break
		}
	}

	if selected == nil {
		return fmt.Errorf("could not find dotfiles %q", name)
	}

	targetPath, err := filepath.Abs(selected.LocalPath)
	if err != nil {
		return fmt.Errorf("could not resolve dotfile path: %w", err)
	}

	targetInfo, err := os.Stat(targetPath)
	if err != nil {
		return fmt.Errorf("dotfile package does not exist: %w", err)
	}

	if !targetInfo.IsDir() {
		return fmt.Errorf("dotfile package is not a directory: %s", targetPath)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("could not create config parent dir: %w", err)
	}

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("could not create backup dir: %w", err)
	}

	var restore func()

	info, err := os.Lstat(configPath)

	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			currentTarget, err := os.Readlink(configPath)
			if err != nil {
				return fmt.Errorf("could not read current symlink: %w", err)
			}

			currentTargetAbs := currentTarget
			if !filepath.IsAbs(currentTargetAbs) {
				currentTargetAbs = filepath.Join(filepath.Dir(configPath), currentTargetAbs)
			}

			currentTargetAbs, _ = filepath.Abs(currentTargetAbs)

			if currentTargetAbs == targetPath {
				return nil
			}

			if err := os.Remove(configPath); err != nil {
				return fmt.Errorf("could not remove current symlink: %w", err)
			}

			restore = func() {
				_ = os.Symlink(currentTarget, configPath)
			}
		} else {
			backupPath, err := uniqueBackupPath(filepath.Join(backupDir, "config-bkp"))
			if err != nil {
				return err
			}

			if err := os.Rename(configPath, backupPath); err != nil {
				return fmt.Errorf("could not backup current config: %w", err)
			}

			restore = func() {
				_ = os.Rename(backupPath, configPath)
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("could not check current config: %w", err)
	}

	if err := os.Symlink(targetPath, configPath); err != nil {
		if restore != nil {
			restore()
		}

		return fmt.Errorf("could not create symlink: %w", err)
	}

	return nil
}

func (d *DotFilesService) CanReplaceConfig(configPath string, dotfilesDir string) (bool, error) {
	info, err := os.Lstat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}

		return false, err
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return false, nil
	}

	target, err := os.Readlink(configPath)
	if err != nil {
		return false, err
	}

	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(configPath), target)
	}

	absTarget, err := filepath.Abs(target)
	if err != nil {
		return false, err
	}

	absDotfilesDir, err := filepath.Abs(dotfilesDir)
	if err != nil {
		return false, err
	}

	rel, err := filepath.Rel(absDotfilesDir, absTarget)
	if err != nil {
		return false, err
	}

	return rel != ".." && !strings.HasPrefix(rel, "../"), nil
}

func uniqueBackupPath(base string) (string, error) {
	for i := 0; ; i++ {
		candidate := base

		if i > 0 {
			candidate = fmt.Sprintf("%s-%d", base, i)
		}

		_, err := os.Lstat(candidate)

		if errors.Is(err, os.ErrNotExist) {
			return candidate, nil
		}

		if err != nil {
			return "", fmt.Errorf("could not check backup path %q: %w", candidate, err)
		}
	}
}
