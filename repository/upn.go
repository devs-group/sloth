package repository

import (
	"path"
	"path/filepath"

	"github.com/devs-group/sloth/config"
)

type UPN string

func (u UPN) GetProjectPath() string {
	return path.Join(filepath.Clean(config.ProjectsDir), string(u))
}
