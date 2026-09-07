package cli

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	autoRemoteFlag   bool
	autoRepoNameFlag string
)

var initCmd = &cobra.Command{
	Use:   "init [git-repo-url]",
	Short: "Inicializa el directorio de almacenamiento y repositorio Git",
	Long: `Inicializa el directorio de almacenamiento local (~/.cashflow) y el repositorio Git.
Con el flag --auto, detecta si ya existe o crea automáticamente un repositorio privado
('cashflow-data') en GitHub usando GitHub CLI (gh) y lo vincula como remoto origin.`,
	Example: `  # Inicializar almacenamiento local y repositorio Git:
  cash init

  # Crear o vincular automáticamente un repositorio privado en GitHub:
  cash init --auto

  # Vincular automáticamente con un nombre personalizado de repositorio:
  cash init --auto --repo mis-finanzas

  # Vincular manualmente con una URL de Git existente:
  cash init https://github.com/usuario/cashflow-data.git`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if autoRemoteFlag && len(args) > 0 {
			return fmt.Errorf("no puedes combinar --auto con una URL manual de repositorio")
		}

		_, repo, err := getService()
		if err != nil {
			return err
		}

		if err := repo.Init(); err != nil {
			return err
		}

		green := color.New(color.FgGreen).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()
		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "%s Almacenamiento inicializado en: %s\n", green("✔"), repo.EntriesDir())

		// Initialize local git repository if git CLI is available
		gitInitCmd := exec.Command("git", "init")
		gitInitCmd.Dir = repo.BaseDir()
		if err := gitInitCmd.Run(); err == nil {
			fmt.Fprintf(out, "%s Repositorio Git inicializado en: %s\n", green("✔"), repo.BaseDir())
		}

		// Handle manual URL if passed
		if len(args) == 1 && args[0] != "" {
			remoteURL := args[0]
			if err := setGitRemoteOrigin(repo.BaseDir(), remoteURL); err != nil {
				return fmt.Errorf("error al configurar repositorio remoto: %w", err)
			}
			fmt.Fprintf(out, "%s Repositorio remoto configurado: %s\n", green("✔"), remoteURL)
			return nil
		}

		// Handle --auto flag
		if autoRemoteFlag {
			// 1. Check if remote already configured in local repo
			existingRemoteCmd := exec.Command("git", "remote", "get-url", "origin")
			existingRemoteCmd.Dir = repo.BaseDir()
			if existingOut, err := existingRemoteCmd.Output(); err == nil && len(strings.TrimSpace(string(existingOut))) > 0 {
				fmt.Fprintf(out, "%s Repositorio remoto 'origin' ya configurado en local: %s\n",
					yellow("ℹ"), strings.TrimSpace(string(existingOut)))
				return nil
			}

			// 2. Verify gh CLI is installed
			if _, err := exec.LookPath("gh"); err != nil {
				return fmt.Errorf("GitHub CLI ('gh') no está instalado. Instálalo con 'pkg install gh' o vincula manualmente: cash init <git-url>")
			}

			// 3. Verify active gh session
			authStatusCmd := exec.Command("gh", "auth", "status")
			if err := authStatusCmd.Run(); err != nil {
				return fmt.Errorf("no hay una sesión activa en GitHub CLI. Ejecuta 'gh auth login' primero")
			}

			targetRepo := autoRepoNameFlag
			if targetRepo == "" {
				targetRepo = "cashflow-data"
			}

			// 4. Check if repo already exists on GitHub
			viewCmd := exec.Command("gh", "repo", "view", targetRepo, "--json", "url", "-q", ".url")
			viewOut, err := viewCmd.Output()
			var remoteURL string

			if err == nil && len(strings.TrimSpace(string(viewOut))) > 0 {
				repoURL := strings.TrimSpace(string(viewOut))
				remoteURL = repoURL + ".git"
				fmt.Fprintf(out, "%s Repositorio '%s' detectado en GitHub.\n", green("✔"), targetRepo)
			} else {
				fmt.Fprintf(out, "Creando repositorio privado '%s' en GitHub con gh...\n", targetRepo)
				createCmd := exec.Command("gh", "repo", "create", targetRepo, "--private", "--description", "CashFlow personal finance storage")
				if createOut, err := createCmd.CombinedOutput(); err != nil {
					return fmt.Errorf("error al crear el repositorio en GitHub: %s (%w)", strings.TrimSpace(string(createOut)), err)
				}

				viewCmd2 := exec.Command("gh", "repo", "view", targetRepo, "--json", "url", "-q", ".url")
				if viewOut2, err := viewCmd2.Output(); err == nil && len(strings.TrimSpace(string(viewOut2))) > 0 {
					remoteURL = strings.TrimSpace(string(viewOut2)) + ".git"
				} else {
					remoteURL = fmt.Sprintf("https://github.com/%s.git", targetRepo)
				}
				fmt.Fprintf(out, "%s Repositorio privado '%s' creado exitosamente en GitHub.\n", green("✔"), targetRepo)
			}

			// 5. Link as origin
			if err := setGitRemoteOrigin(repo.BaseDir(), remoteURL); err != nil {
				return fmt.Errorf("error al vincular remoto origin: %w", err)
			}
			fmt.Fprintf(out, "%s Repositorio remoto vinculado: %s\n", green("✔"), remoteURL)
		}

		return nil
	},
}

func setGitRemoteOrigin(dir, url string) error {
	checkCmd := exec.Command("git", "remote", "get-url", "origin")
	checkCmd.Dir = dir
	if _, err := checkCmd.Output(); err == nil {
		setCmd := exec.Command("git", "remote", "set-url", "origin", url)
		setCmd.Dir = dir
		return setCmd.Run()
	}

	addCmd := exec.Command("git", "remote", "add", "origin", url)
	addCmd.Dir = dir
	return addCmd.Run()
}

func init() {
	initCmd.Flags().BoolVar(&autoRemoteFlag, "auto", false, "Crea o vincula automáticamente un repositorio privado 'cashflow-data' en GitHub usando gh")
	initCmd.Flags().StringVar(&autoRepoNameFlag, "repo", "cashflow-data", "Nombre del repositorio en GitHub para la sincronización automática")
}
