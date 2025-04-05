package commands

import (
	"fmt"
	"strings"

	"github.com/f/mcptools/pkg/alias"
	"github.com/spf13/cobra"
)

// AliasCmd creates the alias command.
func AliasCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alias",
		Short: "Manage MCP server aliases",
		Long: `Manage aliases for MCP servers.

This command allows you to register MCP server commands with a friendly name and
reuse them later.

Aliases are stored in $HOME/.mcpt/aliases.json.

Examples:
  # Add a new server alias
  mcp alias add myfs npx -y @modelcontextprotocol/server-filesystem ~/

  # List all registered server aliases
  mcp alias list

  # Remove a server alias
  mcp alias remove myfs

  # Use an alias with any MCP command
  mcp tools myfs`,
	}

	cmd.AddCommand(aliasAddCmd())
	cmd.AddCommand(aliasListCmd())
	cmd.AddCommand(aliasRemoveCmd())

	return cmd
}

func aliasAddCmd() *cobra.Command {
	addCmd := &cobra.Command{
		Use:                "add [alias] [command args...]",
		Short:              "Add a new MCP server alias",
		DisableFlagParsing: true,
		Long: `Add a new alias for an MCP server command.

The alias will be registered and can be used in place of the server command.

Example:
  mcp alias add myfs npx -y @modelcontextprotocol/server-filesystem ~/`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(thisCmd *cobra.Command, args []string) error {
			if len(args) == 1 && (args[0] == FlagHelp || args[0] == FlagHelpShort) {
				_ = thisCmd.Help()
				return nil
			}

			aliasName := args[0]
			serverCommand := strings.Join(args[1:], " ")

			aliases, err := alias.Load()
			if err != nil {
				return fmt.Errorf("error loading aliases: %w", err)
			}

			aliases[aliasName] = alias.ServerAlias{
				Command: serverCommand,
			}

			if saveErr := alias.Save(aliases); saveErr != nil {
				return fmt.Errorf("error saving aliases: %w", saveErr)
			}

			fmt.Fprintf(thisCmd.OutOrStdout(), "Alias '%s' registered for command: %s\n", aliasName, serverCommand)
			return nil
		},
	}
	return addCmd
}

func aliasListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all registered MCP server aliases",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Load existing aliases
			aliases, err := alias.Load()
			if err != nil {
				return fmt.Errorf("error loading aliases: %w", err)
			}

			if len(aliases) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No aliases registered.")
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Registered MCP server aliases:")
			for name, a := range aliases {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s: %s\n", name, a.Command)
			}

			return nil
		},
	}
}

func aliasRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove an MCP server alias",
		Long: `Remove a registered alias for an MCP server command.

Example:
  mcp alias remove myfs`,
		Args: cobra.ExactArgs(1),
		RunE: func(thisCmd *cobra.Command, args []string) error {
			if len(args) == 1 && (args[0] == FlagHelp || args[0] == FlagHelpShort) {
				_ = thisCmd.Help()
				return nil
			}

			aliasName := args[0]

			aliases, err := alias.Load()
			if err != nil {
				return fmt.Errorf("error loading aliases: %w", err)
			}

			if _, exists := aliases[aliasName]; !exists {
				return fmt.Errorf("alias '%s' does not exist", aliasName)
			}

			delete(aliases, aliasName)

			if saveErr := alias.Save(aliases); saveErr != nil {
				return fmt.Errorf("error saving aliases: %w", saveErr)
			}

			fmt.Fprintf(thisCmd.OutOrStdout(), "Alias '%s' removed.\n", aliasName)
			return nil
		},
	}
}
