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
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	ctx, err := context.HydrateContext(cfg.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to hydrate context: %w", err)
	}
	return ctx, nil
}

// Cobra commands
var rootCmd = &cobra.Command{
	Use:          "yadu",
	Short:        "Yet Another Dotfiles Utility",
	SilenceUsage: true,
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
		only, _ := cmd.Flags().GetBool("only")
		err = context.RunUnit(unitName, !only)
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
	Use:   "new-unit [name] \"[description]\"",
	Short: "Create a new unit",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		edit, _ := cmd.Flags().GetBool("edit")
		context, err := loadContext()
		if err != nil {
			return fmt.Errorf("couldn't load context: %w", err)
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

var initCmd = &cobra.Command{
	Use:   "init [directory]",
	Short: "Initialize a new context and set it as the current context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		directory := args[0]
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %v", err)
		}
		// create the new directory and sub dirctories
		err = os.MkdirAll(directory, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
		err = os.MkdirAll(fmt.Sprintf("%s/units", directory), 0755)
		if err != nil {
			return fmt.Errorf("failed to load config: %v", err)
		}
		cfg.SetContext(directory)
		return nil
	},
}

// Environment command with subcommands
var environmentCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environments",
	Long:  "Commands for managing and applying environments",
}

var environmentGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Display the current environment",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if cfg.Environment == "" {
			fmt.Println("No environment selected")
		} else {
			fmt.Println(cfg.Environment)
		}
		return nil
	},
}

var environmentSetCmd = &cobra.Command{
	Use:   "set [environment_name]",
	Short: "Set the current environment",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		ctx, err := loadContext()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		environments, err := ctx.LoadEnvironments()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		envNames := make([]string, len(environments))
		for i, env := range environments {
			envNames[i] = env.Name
		}

		return envNames, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		environmentName := args[0]

		// Verify environment exists
		ctx, err := loadContext()
		if err != nil {
			return fmt.Errorf("couldn't load context: %w", err)
		}

		environments, err := ctx.LoadEnvironments()
		if err != nil {
			return fmt.Errorf("failed to load environments: %w", err)
		}

		found := false
		for _, env := range environments {
			if env.Name == environmentName {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("environment '%s' not found", environmentName)
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		return cfg.SetSelectedEnvironment(environmentName)
	},
}

var environmentApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply the current environment",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get selected environment from config
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if cfg.Environment == "" {
			return fmt.Errorf("no environment selected, use 'yadu environment set [environment_name]' to select one")
		}

		// Load context and apply the environment
		ctx, err := loadContext()
		if err != nil {
			return fmt.Errorf("couldn't load context: %w", err)
		}

		return ctx.ApplyEnvironment(cfg.Environment)
	},
}

var environmentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available environments",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := loadContext()
		if err != nil {
			return fmt.Errorf("couldn't load context: %w", err)
		}

		environments, err := ctx.LoadEnvironments()
		if err != nil {
			return fmt.Errorf("failed to load environments: %w", err)
		}

		if len(environments) == 0 {
			fmt.Println("No environments found")
			return nil
		}

		// Get currently selected environment for highlighting
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		fmt.Println("Available environments:")
		for _, env := range environments {
			if env.Name == cfg.Environment {
				// Highlight the current environment
				fmt.Printf("  \x1b[1;92m*\x1b[0m \x1b[1;94m%s\x1b[0m (%d units)\n",
					env.Name, len(env.Units))
			} else {
				fmt.Printf("    \x1b[1;94m%s\x1b[0m (%d units)\n",
					env.Name, len(env.Units))
			}
		}

		return nil
	},
}

func init() {
	// Add environment subcommands
	environmentCmd.AddCommand(environmentGetCmd)
	environmentCmd.AddCommand(environmentSetCmd)
	environmentCmd.AddCommand(environmentApplyCmd)
	environmentCmd.AddCommand(environmentListCmd)
	rootCmd.AddCommand(environmentCmd)

	applyCmd.Flags().BoolP("only", "o", false, "Only apply the specified unit (not dependencies)")
	rootCmd.AddCommand(applyCmd)

	// Add commands to root
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(contextCmd)
	rootCmd.AddCommand(newUnitCmd)
	newUnitCmd.Flags().BoolP("edit", "e", false, "Open the new unit in the editor")
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(initCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
