package dotfiles

import (
	"fmt"
	"strings"
)

type DotFile struct {
	Name          string `json:"name"`
	RemoteAddress string `json:"remote_address"`
	LocalPath     string `json:"local_path"`
}

func NewDotFile(name string, remoteAddr string, localPath string) (*DotFile, error) {
	name = strings.TrimSpace(name)
	remoteAddr = strings.TrimSpace(remoteAddr)
	localPath = strings.TrimSpace(localPath)

	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

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
	}

	return &dot, nil
}
