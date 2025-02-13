package main

import (
	_ "embed"
	"fmt"
	"os"
	"text/template"

	"github.com/kindrowboat/yadu/internal/config"
	"github.com/kindrowboat/yadu/pkg/context"
	"github.com/spf13/cobra"
)

//go:embed templates/unit.tmpl
var unitTemplate string

func loadContext() (*context.Context, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	return context.HydrateContext(cfg.Context)
}

// Cobra commands
var rootCmd = &cobra.Command{
	Use:   "yadu",
	Short: "Yet Another Dotfiles Utility",
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available units",
	RunE: func(cmd *cobra.Command, args []string) error {
		context, err := loadContext()
		if err != nil {
			return err
		}
		units, err := context.GetUnitsAndDescriptions()
		if err != nil {
			return err
		}
		for _, unit := range units {
			fmt.Printf("\x1b[1;94m%s\x1b[0m: %s", unit[0], unit[1])
		}
		return nil
	},
}

var applyCmd = &cobra.Command{
	Use:   "apply [unit]",
	Short: "Apply a specific unit",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		context, err := loadContext()
		if err != nil {
			return err
		}
		unitName := args[0]
		err = context.RunUnit(unitName)
		if err != nil {
			return err
		}
		return nil
	},
}

var contextCmd = &cobra.Command{
	Use:   "context [directory]",
	Short: "Get or set the context directory",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %v", err)
		}
		if len(args) == 0 {
			// Show current context
			fmt.Println(cfg.Context)
			return nil
		}

		// Set new context
		cfg.SetContext(args[0])
		return nil
	},
}

var newUnitCmd = &cobra.Command{
	Use:   "new-unit [name]",
	Short: "Create a new unit",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		context, err := loadContext()
		if err != nil {
			return err
		}
		unitName := args[0]
		unitDescription := args[1]

		// create the unit file using the template
		tmpl, err := template.New("unit").Parse(unitTemplate)
		if err != nil {
			return fmt.Errorf("failed to parse unit template: %v", err)
		}
		unitFile := context.GetUnitFileName(unitName)
		f, err := os.Create(unitFile)
		if err != nil {
			return fmt.Errorf("failed to create unit file: %v", err)
		}
		defer f.Close()
		err = tmpl.Execute(f,
			struct {
				Name        string
				Description string
			}{
				Name:        unitName,
				Description: unitDescription,
			},
		)
		if err != nil {
			return err
		} else {
			fmt.Println(unitName)
			return nil
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(contextCmd)
	rootCmd.AddCommand(newUnitCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
