package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"cashflow/internal/adapters/storage"
	"cashflow/internal/ports"
	"cashflow/internal/usecases"
)

var (
	customDir       string
	defaultCurrency string = "USD"
	serviceInstance ports.TransactionUseCases
	repoInstance    *storage.FileRepository
)

// RootCmd represents the base command when called without any subcommands.
var RootCmd = &cobra.Command{
	Use:   "cash",
	Short: "CashFlow CLI — Fast offline-first personal finance tracker",
	Long: `CashFlow CLI is an ultra-fast, offline-first personal finance tracker.
It stores transactions as immutable event files and synchronizes via Git with zero merge conflicts.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() int {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

func init() {
	RootCmd.PersistentFlags().StringVar(&customDir, "dir", "", "Custom storage directory (defaults to ~/.cashflow)")

	// Register all child subcommands
	RootCmd.AddCommand(inCmd)
	RootCmd.AddCommand(outCmd)
	RootCmd.AddCommand(summaryCmd)
	RootCmd.AddCommand(listCmd)
	RootCmd.AddCommand(categoriesCmd)
	RootCmd.AddCommand(initCmd)
}

// getService initializes and returns the application service instance.
func getService() (ports.TransactionUseCases, *storage.FileRepository, error) {
	if serviceInstance != nil && repoInstance != nil {
		return serviceInstance, repoInstance, nil
	}

	repo, err := storage.NewFileRepository(customDir)
	if err != nil {
		return nil, nil, err
	}

	// In Phase 3, sync service is nil until Phase 4
	svc := usecases.NewTransactionService(repo, nil, defaultCurrency)
	serviceInstance = svc
	repoInstance = repo

	return serviceInstance, repoInstance, nil
}

// SetCustomService allows overriding the service instance for automated testing.
func SetCustomService(svc ports.TransactionUseCases, repo *storage.FileRepository) {
	serviceInstance = svc
	repoInstance = repo
}

// ResetCustomService restores default service wiring.
func ResetCustomService() {
	serviceInstance = nil
	repoInstance = nil
}
