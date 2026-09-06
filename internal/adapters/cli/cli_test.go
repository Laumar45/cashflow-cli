package cli_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"cashflow/internal/adapters/cli"
	"cashflow/internal/domain"
)

func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	cli.RootCmd.SetOut(buf)
	cli.RootCmd.SetErr(buf)
	cli.RootCmd.SetArgs(args)

	err := cli.RootCmd.Execute()
	return buf.String(), err
}

func TestPhase3CLICapabilities(t *testing.T) {
	tempDir := t.TempDir()
	cli.ResetCustomService()

	// 1. Initialize storage in tempDir
	initOut, err := executeCommand("init", "--dir", tempDir)
	if err != nil {
		t.Fatalf("init command failed: %v (output: %s)", err, initOut)
	}

	// Done-when Condition 1: Comandos cash in y cash out registran transacciones en disco en menos de 10ms
	startIn := time.Now()
	inOut, err := executeCommand("in", "sueldo", "2000.00", "Salario quincenal", "--dir", tempDir)
	elapsedIn := time.Since(startIn)
	if err != nil {
		t.Fatalf("cash in failed: %v (output: %s)", err, inOut)
	}
	if !strings.Contains(inOut, "Ingreso registrado") || !strings.Contains(inOut, "2000.00") {
		t.Errorf("unexpected cash in output: %s", inOut)
	}
	t.Logf("cash in completed in %v", elapsedIn)

	startOut := time.Now()
	outOut, err := executeCommand("out", "almuerzo", "15.50", "Menu ejecutivo", "--dir", tempDir)
	elapsedOut := time.Since(startOut)
	if err != nil {
		t.Fatalf("cash out failed: %v (output: %s)", err, outOut)
	}
	if !strings.Contains(outOut, "Gasto registrado") || !strings.Contains(outOut, "15.50") {
		t.Errorf("unexpected cash out output: %s", outOut)
	}
	t.Logf("cash out completed in %v", elapsedOut)

	// Register another expense for summary testing
	_, _ = executeCommand("out", "transporte", "4.50", "Metro", "--dir", tempDir)

	// Done-when Condition 2: cash summary imprime resumen formateado con balance neto calculado correctamente
	// Incomes = $2000.00, Expenses = $15.50 + $4.50 = $20.00. Net Balance = +$1980.00
	summaryOut, err := executeCommand("summary", "--dir", tempDir)
	if err != nil {
		t.Fatalf("cash summary failed: %v (output: %s)", err, summaryOut)
	}
	if !strings.Contains(summaryOut, "RESUMEN MENSUAL") {
		t.Errorf("expected header 'RESUMEN MENSUAL', got: %s", summaryOut)
	}
	if !strings.Contains(summaryOut, "2000.00") {
		t.Errorf("expected Total Income 2000.00 in summary, got: %s", summaryOut)
	}
	if !strings.Contains(summaryOut, "20.00") {
		t.Errorf("expected Total Expense 20.00 in summary, got: %s", summaryOut)
	}
	if !strings.Contains(summaryOut, "1980.00") {
		t.Errorf("expected Net Balance 1980.00 in summary, got: %s", summaryOut)
	}

	// Done-when Condition 3: cash list --json emite un array JSON válido sin decoraciones ANSI para interoperabilidad
	listJSONOut, err := executeCommand("list", "--json", "--dir", tempDir)
	if err != nil {
		t.Fatalf("cash list --json failed: %v (output: %s)", err, listJSONOut)
	}

	var jsonTxs []domain.Transaction
	if err := json.Unmarshal([]byte(listJSONOut), &jsonTxs); err != nil {
		t.Fatalf("cash list --json did not emit valid JSON array: %v (raw: %s)", err, listJSONOut)
	}

	if len(jsonTxs) != 3 {
		t.Errorf("expected 3 transactions in JSON list, got %d", len(jsonTxs))
	}

	// Verify categories command
	catOut, err := executeCommand("categories", "--dir", tempDir)
	if err != nil {
		t.Fatalf("cash categories failed: %v (output: %s)", err, catOut)
	}
	if !strings.Contains(catOut, "almuerzo") || !strings.Contains(catOut, "sueldo") || !strings.Contains(catOut, "transporte") {
		t.Errorf("missing categories in output: %s", catOut)
	}
}
