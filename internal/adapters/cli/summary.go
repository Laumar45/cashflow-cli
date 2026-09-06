package cli

import (
	"fmt"
	"sort"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	summaryMonthFlag    string
	summaryCurrencyFlag string
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Muestra el resumen mensual y balance acumulado",
	RunE: func(cmd *cobra.Command, args []string) error {
		now := time.Now()
		year := now.Year()
		month := now.Month()

		if summaryMonthFlag != "" {
			parsed, err := time.Parse("2006-01", summaryMonthFlag)
			if err != nil {
				return fmt.Errorf("formato de mes inválido, use YYYY-MM (ej: 2026-09)")
			}
			year = parsed.Year()
			month = parsed.Month()
		}

		svc, _, err := getService()
		if err != nil {
			return err
		}

		summary, err := svc.GetMonthlySummary(cmd.Context(), year, month)
		if err != nil {
			return err
		}

		green := color.New(color.FgGreen).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()
		cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()

		out := cmd.OutOrStdout()

		fmt.Fprintln(out, "==================================================")
		fmt.Fprintf(out, "        %s: %s %d\n", cyan("RESUMEN MENSUAL"), month.String(), year)
		fmt.Fprintln(out, "==================================================")
		fmt.Fprintf(out, "Ingresos Totales:   %s\n", green("+"+summary.TotalIncome.Format()))
		fmt.Fprintf(out, "Gastos Totales:     %s\n", red("-"+summary.TotalExpense.Format()))
		fmt.Fprintln(out, "--------------------------------------------------")

		balanceStr := summary.NetBalance.Format()
		if summary.NetBalance.Cents >= 0 {
			fmt.Fprintf(out, "Balance Neto:       %s\n", green("+"+balanceStr))
		} else {
			fmt.Fprintf(out, "Balance Neto:       %s\n", red(balanceStr))
		}
		fmt.Fprintf(out, "Total Movimientos:  %d\n", summary.TransactionCount)

		if len(summary.ExpenseByCategory) > 0 {
			fmt.Fprintln(out, "--------------------------------------------------")
			fmt.Fprintf(out, "%s\n", yellow("Gastos por Categoría:"))

			// Sort categories by amount descending
			type catItem struct {
				name  string
				cents int64
				count int
			}
			var items []catItem
			for name, cat := range summary.ExpenseByCategory {
				items = append(items, catItem{name: name, cents: cat.Amount.Cents, count: cat.Count})
			}
			sort.Slice(items, func(i, j int) bool {
				return items[i].cents > items[j].cents
			})

			for _, item := range items {
				fmt.Fprintf(out, "  • %-18s %10s (%d tx)\n",
					item.name,
					red(fmt.Sprintf("%.2f", float64(item.cents)/100.0)),
					item.count)
			}
		}

		fmt.Fprintln(out, "==================================================")
		return nil
	},
}

func init() {
	summaryCmd.Flags().StringVar(&summaryMonthFlag, "month", "", "Mes a consultar en formato YYYY-MM (ej: 2026-09)")
	summaryCmd.Flags().StringVar(&summaryCurrencyFlag, "currency", "USD", "Símbolo de moneda para el reporte")
}
