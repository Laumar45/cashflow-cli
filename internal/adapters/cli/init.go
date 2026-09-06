package cli

import (
	"fmt"
	"os/exec"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [git-repo-url]",
	Short: "Inicializa el directorio de almacenamiento y repositorio Git",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, repo, err := getService()
		if err != nil {
			return err
		}

		if err := repo.Init(); err != nil {
			return err
		}

		green := color.New(color.FgGreen).SprintFunc()
		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "%s Almacenamiento inicializado en: %s\n", green("✔"), repo.EntriesDir())

		// Optionally initialize local git repository if git CLI is available
		gitInitCmd := exec.Command("git", "init")
		gitInitCmd.Dir = repo.BaseDir()
		if err := gitInitCmd.Run(); err == nil {
			fmt.Fprintf(out, "%s Repositorio Git inicializado en: %s\n", green("✔"), repo.BaseDir())
		}

		// If remote git repo URL was passed, configure remote origin
		if len(args) == 1 && args[0] != "" {
			remoteURL := args[0]
			gitRemoteCmd := exec.Command("git", "remote", "add", "origin", remoteURL)
			gitRemoteCmd.Dir = repo.BaseDir()
			if err := gitRemoteCmd.Run(); err == nil {
				fmt.Fprintf(out, "%s Repositorio remoto configurado: %s\n", green("✔"), remoteURL)
			}
		}

		return nil
	},
}
