package pacman

import (
	"os/exec"
	"strings"
)

type pacman struct {
	needsDescription *Package
}

func (p *pacman) name() string {
	return "Arch Linux"
}

func (p *pacman) init() error {
	update := exec.Command("pacman", "-Syy") // TODO sudo?
	return update.Run()
}

func (p *pacman) forEachCommandOutputLine(cmd string, args []string, each func(string)) error {
	update := exec.Command(cmd, args...)
	out, err := update.CombinedOutput()

	if err != nil {
		return err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		each(line)
	}

	return nil
}

func (p *pacman) updates() ([]*Update, error) {
	var updates []*Update
	err := p.forEachCommandOutputLine("pacman", []string{"-Qu"}, func(line string) {
		words := strings.Split(line, " ")
		updates = append(updates, &Update{Package{Name: words[0]}})
	})

	if err != nil {
		return nil, err
	}
	return updates, nil
}

func (p *pacman) updateAll() error {
	update := exec.Command("pacman", "-Syu") // TODO sudo
	return update.Run()
}

func (p *pacman) search(text string) ([]*Package, error) {
	var results []*Package
	err := p.forEachCommandOutputLine("pacman", []string{"-Ss", text}, func(line string) {
		if len(line) == 0 {
			return
		}
		if line[0] == ' ' {
			if p.needsDescription != nil {
				p.needsDescription.Description = strings.TrimSpace(line)
				p.needsDescription = nil
			}
			return
		}

		words := strings.Split(line, " ")
		parts := strings.Split(words[0], "/")
		pkg := &Package{
			Name:    parts[1],
			Version: words[1],
		}
		p.needsDescription = pkg
		results = append(results, pkg)
	})

	if err != nil {
		return nil, err
	}
	return results, nil
}
