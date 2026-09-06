package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var inDateFlag string

var inCmd = &cobra.Command{
	Use:   "in <category> <amount> [description]",
	Short: "Registra un nuevo ingreso",
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
		if inDateFlag != "" {
			d, err := time.Parse("2006-01-02", inDateFlag)
			if err != nil {
				return fmt.Errorf("formato de fecha inválido, use YYYY-MM-DD (ej: 2026-09-04)")
			}
			targetDate = &d
		}

		svc, _, err := getService()
		if err != nil {
			return err
		}

		tx, err := svc.RecordIncome(cmd.Context(), category, amount, desc, targetDate)
		if err != nil {
			return err
		}

		green := color.New(color.FgGreen).SprintFunc()
		fmt.Fprintf(cmd.OutOrStdout(), "%s Ingreso registrado: +%s en '%s' [ID: %s]\n",
			green("✔"), tx.Amount.Format(), tx.Category, tx.ID)

		return nil
	},
}

func init() {
	inCmd.Flags().StringVar(&inDateFlag, "date", "", "Fecha de la transacción (formato YYYY-MM-DD)")
}
