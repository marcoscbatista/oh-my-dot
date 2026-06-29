package dotfiles

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type DotFileStore struct {
	path string
}

func NewDotFileStore(path string) (*DotFileStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(home, path)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(path, []byte("[]"), 0644); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &DotFileStore{path: path}, nil
}

func (d *DotFileStore) Load() ([]DotFile, error) {
	data, err := os.ReadFile(d.path)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return []DotFile{}, nil
	}

	var dotfiles []DotFile

	if err := json.Unmarshal(data, &dotfiles); err != nil {
		return nil, fmt.Errorf("could not parse dotfiles store: %w", err)
	}

	if dotfiles == nil {
		return []DotFile{}, nil
	}

	return dotfiles, nil
}

func (d *DotFileStore) Add(dotfile DotFile) error {
	dotfiles, err := d.Load()
	if err != nil {
		return err
	}

	for _, existing := range dotfiles {
		if existing.Name == dotfile.Name {
			return fmt.Errorf("dotfile %q already exists", dotfile.Name)
		}
	}

	dotfiles = append(dotfiles, dotfile)

	return d.SaveAll(dotfiles)
}

func (d *DotFileStore) SaveAll(dotfiles []DotFile) error {
	data, err := json.MarshalIndent(dotfiles, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(d.path, data, 0644); err != nil {
		return err
	}

	return nil
}
