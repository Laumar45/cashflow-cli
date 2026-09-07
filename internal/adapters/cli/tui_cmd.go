package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"cashflow/internal/adapters/tui"
)

var tuiMonthFlag string

var tuiCmd = &cobra.Command{
	Use:     "tui",
	Aliases: []string{"dashboard", "ui"},
	Short:   "Abre el dashboard visual interactivo en la terminal",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, repo, err := getService()
		if err != nil {
			return err
		}

		initialMonth := time.Now().UTC()
		if tuiMonthFlag != "" {
			parsed, err := time.Parse("2006-01", tuiMonthFlag)
			if err != nil {
				return fmt.Errorf("formato de mes inválido, use YYYY-MM (ej: 2026-09)")
			}
			initialMonth = parsed
		}

		// Look for dollar-sign.md in working directory or repo root
		var logoContent string
		candidatePaths := []string{
			"dollar-sign.md",
			filepath.Join(repo.BaseDir(), "dollar-sign.md"),
		}
		for _, p := range candidatePaths {
			if data, err := os.ReadFile(p); err == nil && len(data) > 0 {
				logoContent = string(data)
				break
			}
		}

		syncInfo := tui.GetLocalSyncStatus(repo.BaseDir())

		model := tui.NewModel(svc, initialMonth, syncInfo, logoContent)
		p := tea.NewProgram(model, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			return fmt.Errorf("error al ejecutar dashboard TUI: %w", err)
		}

		return nil
	},
}

func init() {
	tuiCmd.Flags().StringVar(&tuiMonthFlag, "month", "", "Mes inicial a mostrar en formato YYYY-MM")
	RootCmd.AddCommand(tuiCmd)
}
