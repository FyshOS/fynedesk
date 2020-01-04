package pacman

// Package represents a system package that can be searched for
type Package struct {
	Name, Version, Description string
}

// Update is a type that contains update information and the package that it refers to
type Update struct {
	Package
}

// Manager is a high level API providing system independent package management.
// It provides caching and error handling over the lower level system providers.
type Manager struct {
	provider provider
	updates  []*Update
}

func (m *Manager) getUpdates() []*Update {
	if m.updates == nil && m.provider != nil {
		ret, _ := m.provider.updates()

		m.updates = ret
	}

	return m.updates
}

// HasUpdates returns true if the system package manager reports that updates are available
func (m *Manager) HasUpdates() bool {
	if m.provider == nil {
		return false
	}

	ret := m.getUpdates()
	return ret != nil && len(ret) > 0
}

// Updates returns a list of any updates that are currently available for the system
func (m *Manager) Updates() []*Update {
	if m.provider == nil {
		return []*Update{}
	}

	return m.getUpdates()
}

// UpdateAll instructs the system package manager to update out of date packages
func (m *Manager) UpdateAll() {
	if m.provider == nil {
		return
	}

	_ = m.provider.updateAll()
}

// Search returns a list of all available system packages that match the term passed
func (m *Manager) Search(term string) []*Package {
	if m.provider == nil {
		return nil
	}

	ret, _ := m.provider.search(term)
	return ret
}

// NewManager returns a new manager instance for handling package installation and updating
func NewManager() (*Manager, error) {
	provider, err := providerForSystem()
	if err != nil {
		return nil, err
	}

	return &Manager{provider: provider}, nil
}
