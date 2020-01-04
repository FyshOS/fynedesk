package pacman

import "strings"

type dummyProvider struct {
}

func (d *dummyProvider) name() string {
	return "dummy"
}

func (d *dummyProvider) init() error {
	return nil
}

func (d *dummyProvider) updates() ([]*Update, error) {
	return []*Update{
		{Package{Name: "Package 1", Version: "1.0", Description: "More text"}},
	}, nil
}

func (d *dummyProvider) updateAll() error {
	return nil
}

func (d *dummyProvider) search(term string) ([]*Package, error) {
	if !strings.Contains("Package 1", term) {
		return []*Package{}, nil
	}

	return []*Package{
		{Name: "Package 1", Version: "1.0", Description: "More text"},
	}, nil
}
