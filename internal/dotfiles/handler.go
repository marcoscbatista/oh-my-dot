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

func (d *DotFilesHandler) download(remoteAddr string, path string, verbose bool) error {
	cmd := exec.Command("git", "clone", remoteAddr, path)
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

	}

	if err := cmd.Run(); err != nil {
		_ = os.RemoveAll(path)
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

	if err := d.download(remoteAddr, pathPackage, verbose); err != nil {
		return fmt.Errorf("could not download dotfiles: %w", err)
	}

	if err := d.Service.Create(name, remoteAddr, pathPackage); err != nil {
		return fmt.Errorf("could not save dotfile: %w", err)
	}

	return nil
}

func (d *DotFilesHandler) Switch(id int, force bool) error {
	isManaged, err := d.Service.CanReplaceConfig(d.ConfigPath, d.DotfilesDir)
	if err != nil {
		return err
	}

	if !isManaged && !force {
		return fmt.Errorf(
			"%s is not managed by oh-my-dot and would be replaced.\n A backup would be created in %s.\n Run again with --force to continue",
			d.ConfigPath,
			d.DotfilesDir,
		)
	}

	if id <= 0 {
		return fmt.Errorf("id cannot be empty")
	}

	if err := d.Service.Switch(id, d.ConfigPath, d.DotfilesDir); err != nil {
		return fmt.Errorf("could not switch dotfiles: %w", err)
	}

	return nil
}
