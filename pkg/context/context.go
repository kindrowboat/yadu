package context

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Context struct {
	directory string
	units     map[string]*unit
}

// Environment represents a named collection of units to be applied together
type Environment struct {
	Name  string   `yaml:"name"`
	Units []string `yaml:"units"`
}

// GetDirectory returns the context directory
func (c Context) GetDirectory() string {
	return c.directory
}

// LoadEnvironments loads environment configurations from the YAML file
func (c Context) LoadEnvironments() ([]Environment, error) {
	environmentsFile := filepath.Join(c.directory, "environments.yaml")

	data, err := os.ReadFile(environmentsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read environments file: %w", err)
	}

	var environments []Environment
	err = yaml.Unmarshal(data, &environments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse environments file: %w", err)
	}

	return environments, nil
}

// ApplyEnvironment applies all units in the specified environment
func (c *Context) ApplyEnvironment(environmentName string) error {
	environments, err := c.LoadEnvironments()
	if err != nil {
		return fmt.Errorf("failed to load environments: %w", err)
	}

	var targetEnv *Environment
	for _, env := range environments {
		if env.Name == environmentName {
			targetEnv = &env
			break
		}
	}

	if targetEnv == nil {
		return fmt.Errorf("environment '%s' not found", environmentName)
	}

	// Apply all units in the environment
	fmt.Printf("Applying environment: \x1b[1;94m%s\x1b[0m\n", environmentName)
	for _, unitName := range targetEnv.Units {
		fmt.Printf("Applying unit from environment: %s\n", unitName)
		if err := c.RunUnit(unitName, true); err != nil {
			return fmt.Errorf("failed to apply unit '%s': %w", unitName, err)
		}
	}

	return nil
}

func HydrateContext(directory string) (*Context, error) {
	c := &Context{directory: directory, units: make(map[string]*unit)}
	entries, err := os.ReadDir(c.getUnitsDir())
	if err != nil {
		return nil, fmt.Errorf("failed to read units directory: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			unitName := entry.Name()
			_, err := c.LoadUnit(unitName)
			if err != nil {
				return nil, fmt.Errorf("failed to add unit %s: %w", unitName, err)
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

func (c *Context) LoadUnit(name string) (*unit, error) {
	// if the unit already exists, return it
	if unit, ok := c.units[name]; ok {
		return unit, nil
	}
	// create a new unit
	unit := &unit{name: name, file: c.GetUnitFileName(name), hasRun: false}

	get_deps := exec.Command("/bin/bash", "-c", fmt.Sprintf("source %s && dependencies", unit.file))
	var stdout, stderr strings.Builder
	get_deps.Stdout = &stdout
	get_deps.Stderr = &stderr
	err := get_deps.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies for %s: stderr: %s: %w",
			unit.name, stderr.String(), err)
	}

	// Process dependencies from stdout
	for _, dep := range strings.Fields(stdout.String()) {
		depUnit, err := c.LoadUnit(dep)
		if err != nil {
			return nil, fmt.Errorf("failed to add dependency %s for unit %s: %w", dep, name, err)
		}
		unit.dependencies = append(unit.dependencies, depUnit)
	}

	// add the unit to the context
	c.units[name] = unit
	return unit, nil
}

func (c *Context) RunUnit(name string, run_deps bool) error {
	unit, ok := c.units[name]
	if !ok {
		return fmt.Errorf("unit %s not found", name)
	}

	// If unit has already been run, skip it
	if unit.hasRun {
		return nil
	}

	if run_deps {
		// Run dependencies first
		for _, dep := range unit.dependencies {
			if err := c.RunUnit(dep.name, true); err != nil {
				return err
			}
		}
	}

	fmt.Printf("\x1b[1;97m===== \x1b[1;94m%s\x1b[1;97m =====\x1b[0m\n", unit.name)
	cmd := exec.Command("/bin/bash", unit.file)
	cmd.Dir = c.directory // Set working directory to context directory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execution of '%s' failed: %w", unit.name, err)
	}

	unit.hasRun = true
	return nil
}
