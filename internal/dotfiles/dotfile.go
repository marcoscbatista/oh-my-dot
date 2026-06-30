package dotfiles

import (
	"fmt"
	"strings"
)

type DotFile struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	RemoteAddress string `json:"remote_address"`
	LocalPath     string `json:"local_path"`
	InUse         bool   `json:"in_use"`
}

func NewDotFile(name string, remoteAddr string, localPath string) (*DotFile, error) {
	var err error

	name, err = normalizePackageName(name)
	if err != nil {
		return nil, err
	}

	remoteAddr = strings.TrimSpace(remoteAddr)
	localPath = strings.TrimSpace(localPath)

	if remoteAddr == "" {
		return nil, fmt.Errorf("remote address cannot be empty")
	}

	if localPath == "" {
		return nil, fmt.Errorf("local path cannot be empty")
	}

	dot := DotFile{
		Name:          name,
		RemoteAddress: remoteAddr,
		LocalPath:     localPath,
		InUse:         false,
	}

	return &dot, nil
}
