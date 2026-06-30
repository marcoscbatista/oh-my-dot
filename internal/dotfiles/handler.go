package dotfiles

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
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
	var err error

	name, err = normalizePackageName(name)
	if err != nil {
		return err
	}

	remoteAddr = strings.TrimSpace(remoteAddr)

	if remoteAddr == "" {
		return fmt.Errorf("remote address cannot be empty")
	}

	pathPackage, err := resolvePackagePath(d.DotfilesDir, name)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(d.DotfilesDir, 0755); err != nil {
		return fmt.Errorf("could not prepare dotfiles directory: %w", err)
	}

	if err := d.Service.ValidateCreate(name, remoteAddr, pathPackage); err != nil {
		return err
	}

	if _, err := os.Stat(pathPackage); err == nil {
		return fmt.Errorf("package destination already exists: %s", pathPackage)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("could not check package destination: %w", err)
	}

	if err := d.download(remoteAddr, pathPackage, verbose); err != nil {
		return fmt.Errorf("could not download dotfiles: %w", err)
	}

	if err := d.Service.Create(name, remoteAddr, pathPackage); err != nil {
		_ = os.RemoveAll(pathPackage)
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
