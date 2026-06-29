package dotfiles

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type DotFilesHandler struct {
	Service     *DotFilesService
	DotfilesDir string
	ConfigPath  string
}

func (d *DotFilesHandler) download(dot DotFile, verbose bool) error {
	cmd := exec.Command("git", "clone", dot.RemoteAddress, dot.LocalPath)
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

	}

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (d *DotFilesHandler) GetAll() ([]DotFile, error) {
	dots, err := d.Service.GetAll()
	if err != nil {
		return nil, fmt.Errorf("could not get dotfiles: %w", err)
	}

	return dots, nil
}

func (d *DotFilesHandler) Create(name string, remoteAddr string, verbose bool) error {
	name = strings.TrimSpace(name)
	remoteAddr = strings.TrimSpace(remoteAddr)

	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if remoteAddr == "" {
		return fmt.Errorf("remote address cannot be empty")
	}

	pathPackage := filepath.Join(d.DotfilesDir, name)

	dot, err := NewDotFile(name, remoteAddr, pathPackage)
	if err != nil {
		return fmt.Errorf("could not create dotfile: %w", err)
	}

	if err := d.download(*dot, verbose); err != nil {
		return fmt.Errorf("could not download dotfiles: %w", err)
	}

	if err := d.Service.Create(dot); err != nil {
		_ = os.RemoveAll(dot.LocalPath)

		return fmt.Errorf("could not save dotfile: %w", err)
	}

	return nil
}

func (d *DotFilesHandler) Switch(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if err := d.Service.Switch(name, d.ConfigPath, d.DotfilesDir); err != nil {
		return fmt.Errorf("could not switch dotfiles: %w", err)
	}

	return nil
}
