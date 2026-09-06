package cli

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"cashflow/internal/domain"
	"cashflow/internal/ports"
)

var (
	listMonthFlag    string
	listCategoryFlag string
	listLimitFlag    int
	listJSONFlag     bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista transacciones registradas con filtros opcionales",
	RunE: func(cmd *cobra.Command, args []string) error {
		filter := ports.TransactionFilter{
			Limit: listLimitFlag,
		}

		if listMonthFlag != "" {
			parsed, err := time.Parse("2006-01", listMonthFlag)
			if err != nil {
				return fmt.Errorf("formato de mes inválido, use YYYY-MM (ej: 2026-09)")
			}
			filter.Month = &parsed
		}

		if listCategoryFlag != "" {
			normalized, err := domain.NormalizeCategory(listCategoryFlag)
			if err != nil {
				return err
			}
			filter.Category = normalized
		}

		svc, _, err := getService()
		if err != nil {
			return err
		}

		txs, err := svc.ListTransactions(cmd.Context(), filter)
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()

		// If --json flag is passed, output pure JSON array without ANSI escapes
		if listJSONFlag {
			encoder := json.NewEncoder(out)
			encoder.SetIndent("", "  ")
			return encoder.Encode(txs)
		}

		if len(txs) == 0 {
			fmt.Fprintln(out, "No se encontraron transacciones con los filtros especificados.")
			return nil
		}

		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "FECHA\tID\tTIPO\tCATEGORÍA\tMONTO\tDESCRIPCIÓN")
		fmt.Fprintln(w, "-----\t--\t----\t---------\t-----\t-----------")

		green := color.New(color.FgGreen).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()

		for _, tx := range txs {
			dateStr := tx.Timestamp.Format("2006-01-02 15:04")
			shortID := tx.ID
			if len(shortID) > 10 {
				shortID = shortID[:10] + "..."
			}

			var typeStr, amountStr string
			if tx.Type == domain.TypeIncome {
				typeStr = green("INCOME")
				amountStr = green("+" + tx.Amount.Format())
			} else {
				typeStr = red("EXPENSE")
				amountStr = red("-" + tx.Amount.Format())
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				dateStr,
				shortID,
				typeStr,
				tx.Category,
				amountStr,
				tx.Description)
		}

		return w.Flush()
	},
}

func init() {
	listCmd.Flags().StringVar(&listMonthFlag, "month", "", "Filtrar por mes en formato YYYY-MM")
	listCmd.Flags().StringVar(&listCategoryFlag, "category", "", "Filtrar por categoría")
	listCmd.Flags().IntVar(&listLimitFlag, "limit", 0, "Limitar cantidad de registros devueltos (0 = sin límite)")
	listCmd.Flags().BoolVar(&listJSONFlag, "json", false, "Emitir resultado en formato JSON puro")
}
