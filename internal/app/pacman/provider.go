package pacman

import (
	"errors"
	"os/exec"
)

type provider interface {
	name() string
	init() error
	updates() ([]*Update, error)
	updateAll() error
	search(string) ([]*Package, error)
}

func providerForSystem() (provider, error) {
	if _, err := exec.LookPath("pacman"); err != nil {
		return nil, errors.New("No update provider was found\nfor the current system")
	}

	return &pacman{}, nil
}
