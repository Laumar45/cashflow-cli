package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var outDateFlag string

var outCmd = &cobra.Command{
	Use:   "out <category> <amount> [description]",
	Short: "Registra un nuevo gasto",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		category := args[0]
		amountStr := args[1]

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil || amount <= 0 {
			return fmt.Errorf("el monto debe ser un número decimal positivo (ej: 15.50)")
		}

		var desc string
		if len(args) > 2 {
			desc = strings.Join(args[2:], " ")
		}

		var targetDate *time.Time
		if outDateFlag != "" {
			d, err := time.Parse("2006-01-02", outDateFlag)
			if err != nil {
				return fmt.Errorf("formato de fecha inválido, use YYYY-MM-DD (ej: 2026-09-04)")
			}
			targetDate = &d
		}

		svc, _, err := getService()
		if err != nil {
			return err
		}

		tx, err := svc.RecordExpense(cmd.Context(), category, amount, desc, targetDate)
		if err != nil {
			return err
		}

		red := color.New(color.FgRed).SprintFunc()
		fmt.Fprintf(cmd.OutOrStdout(), "%s Gasto registrado: -%s en '%s' [ID: %s]\n",
			red("✔"), tx.Amount.Format(), tx.Category, tx.ID)

		return nil
	},
}

func init() {
	outCmd.Flags().StringVar(&outDateFlag, "date", "", "Fecha de la transacción (formato YYYY-MM-DD)")
}
