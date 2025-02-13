package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
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

var editCmd = &cobra.Command{
	Use:   "edit [unit]",
	Short: "Edit a specific unit",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		context, err := loadContext()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		units, err := context.GetUnitsAndDescriptions()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		// Extract just the unit names
		validUnits := make([]string, len(units))
		for i, unit := range units {
			validUnits[i] = unit[0]
		}

		return validUnits, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		context, err := loadContext()
		if err != nil {
			return err
		}
		unitName := args[0]
		unitFile := context.GetUnitFileName(unitName)
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}
		editCmd := exec.Command(editor, unitFile)
		editCmd.Stdin = os.Stdin
		editCmd.Stdout = os.Stdout
		editCmd.Stderr = os.Stderr
		if err := editCmd.Run(); err != nil {
			return fmt.Errorf("failed to open editor: %v", err)
		}
		return nil
	},
}

var applyCmd = &cobra.Command{
	Use:   "apply [unit]",
	Short: "Apply a specific unit",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		context, err := loadContext()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		units, err := context.GetUnitsAndDescriptions()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		// Extract just the unit names
		validUnits := make([]string, len(units))
		for i, unit := range units {
			validUnits[i] = unit[0]
		}

		return validUnits, cobra.ShellCompDirectiveNoFileComp
	},
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
		edit, _ := cmd.Flags().GetBool("edit")
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
		}

		// Open the new unit in the editor if the flag is set
		if edit {
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi"
			}
			cmd := exec.Command(editor, unitFile)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to open editor: %v", err)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(contextCmd)
	rootCmd.AddCommand(newUnitCmd)
	newUnitCmd.Flags().BoolP("edit", "e", false, "Open the new unit in the editor")
	rootCmd.AddCommand(editCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
