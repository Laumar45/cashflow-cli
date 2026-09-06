package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"cashflow/internal/adapters/vcs"
	"cashflow/internal/domain"
)

var (
	syncRemoteFlag string
	syncBranchFlag string
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sincroniza transacciones locales con el repositorio remoto de Git",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, repo, err := getService()
		if err != nil {
			return err
		}

		gitSvc := vcs.NewGitService(repo.BaseDir(), syncRemoteFlag, syncBranchFlag)
		gitSvc.SetStepLogger(cmd.OutOrStdout())

		if err := gitSvc.Sync(cmd.Context()); err != nil {
			if errors.Is(err, domain.ErrGitNetworkFailure) {
				yellow := color.New(color.FgYellow).SprintFunc()
				fmt.Fprintf(cmd.OutOrStdout(), "\n%s Sin conexión al repositorio remoto. Tus transacciones locales están seguras.\n",
					yellow("Advertencia:"))
				os.Exit(2)
			}

			if errors.Is(err, domain.ErrGitAuthFailure) {
				red := color.New(color.FgRed).SprintFunc()
				fmt.Fprintf(os.Stderr, "\n%s Falla de autenticación en Git. Verifica tus credenciales SSH/HTTPS.\n",
					red("Error:"))
				os.Exit(2)
			}

			if errors.Is(err, domain.ErrNoRemoteConfigured) {
				return fmt.Errorf("no hay repositorio remoto configurado. Ejecuta 'cash init <git-url>' primero")
			}

			return err
		}

		green := color.New(color.FgGreen).SprintFunc()
		fmt.Fprintf(cmd.OutOrStdout(), "\n%s Sincronización con Git completada con éxito.\n", green("✔"))
		return nil
	},
}

func init() {
	syncCmd.Flags().StringVar(&syncRemoteFlag, "remote", "origin", "Nombre del repositorio remoto de Git")
	syncCmd.Flags().StringVar(&syncBranchFlag, "branch", "", "Rama a sincronizar (por defecto la rama actual)")
	RootCmd.AddCommand(syncCmd)
}
