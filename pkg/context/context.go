package context

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type Context struct {
	directory string
	units     map[string]*unit
}

func HydrateContext(directory string) (*Context, error) {
	c := &Context{directory: directory, units: make(map[string]*unit)}
	entries, err := os.ReadDir(c.getUnitsDir())
	if err != nil {
		return nil, fmt.Errorf("failed to read units directory: %v", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			unitName := entry.Name()
			_, err := c.AddUnit(unitName)
			if err != nil {
				return nil, err
			}
		}
	}
	return c, nil
}

func (c Context) GetUnitsAndDescriptions() ([][2]string, error) {
	units := make([][2]string, 0, len(c.units))
	for name, unit := range c.units {
		desc, err := unit.GetDescription()
		if err != nil {
			return nil, err
		}
		units = append(units, [2]string{name, desc})
	}

	// Sort units alphabetically by name
	sort.Slice(units, func(i, j int) bool {
		return units[i][0] < units[j][0]
	})

	return units, nil
}

func (c Context) getUnitsDir() string {
	return filepath.Join(c.directory, "units")
}

func (c Context) GetUnitFileName(unitName string) string {
	return filepath.Join(c.getUnitsDir(), unitName)
}

type unit struct {
	name         string
	dependencies []*unit
	file         string
	hasRun       bool
}

func (u unit) GetDescription() (string, error) {
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("source %s && description", u.file))
	desc, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get description for %s: %v", u.name, err)
	}
	return string(desc), nil
}

func (c *Context) AddUnit(name string) (*unit, error) {
	// if the unit already exists, return it
	if unit, ok := c.units[name]; ok {
		return unit, nil
	}
	// create a new unit
	unit := &unit{name: name, file: c.GetUnitFileName(name), hasRun: false}
	// add dependencies
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("source %s && dependencies", unit.file))
	deps, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies for %s: %v", unit.name, err)
	}
	for _, dep := range strings.Fields(string(deps)) {
		depUnit, err := c.AddUnit(dep)
		if err != nil {
			return nil, err
		}
		unit.dependencies = append(unit.dependencies, depUnit)
	}
	// add the unit to the context
	c.units[name] = unit
	return unit, nil
}

func (u *unit) Run() error {
	if u.hasRun {
		return nil
	}
	for _, dep := range u.dependencies {
		if err := dep.Run(); err != nil {
			return err
		}
	}
	fmt.Printf("\x1b[1;97m===== \x1b[1;94m%s\x1b[1;97m =====\x1b[0m\n", u.name)
	cmd := exec.Command("/bin/bash", u.file)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execution of '%s' failed: %v", u.name, err)
	}
	u.hasRun = true
	return nil
}

func (c *Context) RunUnit(name string) error {
	unit, ok := c.units[name]
	if !ok {
		return fmt.Errorf("unit %s not found", name)
	}
	return unit.Run()
}
