# YADU - Yet Another Dotfiles Utility

YADU is a simple, powerful dotfiles management tool that helps you organize and
apply your configuration files in a modular way.

Example dot files repo: [kindrowboat/dots](https://github.com/kindrowboat/dots)

## Features

- Modular "units" approach to dotfiles management
- Dependency resolution between configuration units
- Easy creation of new configuration units
- Simple command-line interface
- Customizable templates for new units

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/kindrowboat/yadu.git
cd yadu

# Build and install
make build
make install
```

### Prerequisites

- Go 1.18 or later

## Usage

### Initializing a New Configuration

```bash
yadu init <directory>
```

Creates a new dotfiles repository in the specified directory.

#### or specifying a specific context for your dotfiles

```bash
yadu context <context-path>
```

### Creating a New Unit

```bash
yadu create <unit-name> -d "Description of the unit"
```

This will create a new unit template in the units directory of the context. Use
the `-e` flag to open the new unit in your editor.


### Listing Available Units

```bash
yadu list
```

This will show all available configuration units with their descriptions.

### Applying a Unit

```bash
yadu apply <unit-name>
```

This command will apply the specified unit and all its dependencies in the correct order.

## Unit Structure

Units are simple bash scripts located in the units directory. Each unit follows
this structure:

```bash
#!/bin/bash

description() {
    echo "Description of what this unit does"
}

# List dependencies (space separated)
dependencies() {
    echo "dependency1 dependency2"
}

# Main function containing the unit's logic
main() {
    # Your configuration commands go here
    echo "Applying configuration..."
}

# Execute main function when script is run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main
fi
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under [MDGPL](./LICENSE).

## Acknowledgments

- Inspired by the various dotfiles management tools in the community, especially
  [construct](https://github.com/kindrowboat/construct).
- Built with [Cobra](https://github.com/spf13/cobra) command library for Go