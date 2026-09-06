package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"cashflow/internal/domain"
	"cashflow/internal/ports"
)

var categoriesTypeFlag string

var categoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "Lista todas las categorías únicas registradas",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, _, err := getService()
		if err != nil {
			return err
		}

		filter := ports.TransactionFilter{}
		if categoriesTypeFlag == "in" {
			t := domain.TypeIncome
			filter.Type = &t
		} else if categoriesTypeFlag == "out" {
			t := domain.TypeExpense
			filter.Type = &t
		}

		txs, err := svc.ListTransactions(cmd.Context(), filter)
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		if len(txs) == 0 {
			fmt.Fprintln(out, "No hay categorías registradas.")
			return nil
		}

		type catStats struct {
			count int
			cents int64
		}
		stats := make(map[string]*catStats)
		for _, tx := range txs {
			s, exists := stats[tx.Category]
			if !exists {
				s = &catStats{}
				stats[tx.Category] = s
			}
			s.count++
			s.cents += tx.Amount.Cents
		}

		uniqueCats, err := svc.ListCategories(cmd.Context())
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "CATEGORÍA\tTRANSACCIONES\tTOTAL ACUMULADO")
		fmt.Fprintln(w, "---------\t-------------\t---------------")

		for _, cat := range uniqueCats {
			st, ok := stats[cat]
			if !ok {
				continue
			}
			fmt.Fprintf(w, "%s\t%d\t$%.2f\n", cat, st.count, float64(st.cents)/100.0)
		}

		return w.Flush()
	},
}

func init() {
	categoriesCmd.Flags().StringVar(&categoriesTypeFlag, "type", "", "Filtrar por tipo: 'in' (ingresos) o 'out' (gastos)")
}
