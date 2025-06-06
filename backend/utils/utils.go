package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/goombaio/namegenerator"

	"github.com/devs-group/sloth/backend/config"
)

func GenerateRandomName() string {
	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)
	return nameGenerator.Generate()
}

func DeleteFile(filename string, relativePath string) error {
	cfg := config.GetConfig()

	p := path.Join(filepath.Clean(cfg.ProjectsDir), relativePath, filename)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return os.Remove(p)
}

func CreateFolderIfNotExists(p string) (string, error) {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		if err := os.MkdirAll(p, os.ModePerm); err != nil {
			return "", fmt.Errorf("failed to crete folder in path %s, err: %v", p, err)
		} else {
			return p, nil
		}
	} else if err != nil {
		return "", fmt.Errorf("unable to check if folder exists in path %s, err: %v", p, err)
	} else {
		return p, nil
	}
}

func DeleteFolder(p string) error {
	err := os.RemoveAll(p)
	if err != nil {
		return err
	}
	return nil
}

func RandStringRunes(n int) (string, error) {
	var runes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, n)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(runes))))
		if err != nil {
			return "", err
		}
		b[i] = runes[n.Int64()]
	}
	return string(b), nil
}

func IsProduction() bool {
	mode, _ := os.LookupEnv("GIN_MODE")
	return mode == "release"
}

func StringAsPointer(s string) *string {
	return &s
}
