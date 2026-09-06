# Design Brief: CashFlow CLI

## 1. Header Block
- **Version:** 1.0.0
- **Date:** 2026-09-04
- **Status:** Ready for Implementation
- **Stack Summary:** Go 1.22+, Cobra CLI, Immutable JSON Event Files, Git VCS Adapter
- **Build Mode:** Greenfield v1.0 (Clean Hexagonal Architecture)

---

## 2. Summary & Guiding Principle
**CashFlow CLI** es una herramienta de registro y seguimiento de finanzas personales ultrarrápida, diseñada para desarrolladores y usuarios avanzados de terminal. Opera bajo un modelo *offline-first* estricto donde el almacenamiento local se estructura como un log de eventos inmutables en archivos individuales, sincronizados entre múltiples dispositivos (Laptop, PC, Termux en Android) mediante un repositorio privado de Git como backend distribuido a costo cero.

Reemplaza hojas de cálculo manuales y aplicaciones móviles de finanzas sobrecargadas con interfaces lentas y sincronización centralizada propietaria.

**Guiding Principle:**  
> **"Captura en terminal en milisegundos, almacenamiento distribuido de eventos inmutables: cero fricción, cero costo y cero conflictos en Git."**

---

## 3. Delta Matrix
*(Omitido por Skip Rule: Aplica únicamente a briefs de evolución v2+).*

---

## 4. Stack & Constraints

### 4.1. Stack Table
| Component | Technology | Version | Architectural Justification |
|---|---|---|---|
| **Core Runtime** | Go | 1.22+ | Compilación nativa a un solo binario estático, arranque en <10ms, consumo ínfimo de memoria y portabilidad directa en Windows, Linux y Termux (Android). |
| **CLI Framework** | `spf13/cobra` | v1.8+ | Estándar de la industria en Go para subcomandos (`in`, `out`, `sync`, `summary`), parsing robusto de flags, help contextual y autocompletado. |
| **ID Generator** | `oklog/ulid` | v2.1+ | Identificadores únicos universales lexicográficamente ordenables por timestamp (128 bits). Garantiza orden cronológico natural al listar archivos sin leer su contenido. |
| **Storage Engine** | Immutable JSON Files | RFC 8259 | Patrón Event Sourcing: un archivo por transacción (`~/.cashflow/entries/<ulid>.json`). Resuelve de raíz los conflictos de merge en Git. |
| **VCS Sync** | Git CLI | 2.x+ | Reutiliza los comandos nativos de Git del sistema (`os/exec`) para commit, pull y push hacia repositorio privado sin dependencias de red en el runtime. |
| **Terminal Output** | `fatih/color` + `text/tabwriter` | v1.16+ / Stdlib | Formato tabular limpio y resaltado ANSI de ingresos/gastos sin sobrecargar el binario. |

### 4.2. Not in Stack
- **SQLite / Motores SQL embebidos:** No son amigables para versionado distribuido directo en Git (archivos binarios sufren colisiones constantes).
- **Un solo archivo `.jsonl`:** Descartado como storage primario para evitar colisiones de líneas en Git rebase.
- **Librerías GUI / Web (React, Vue, Node):** Estrictamente fuera del alcance de la v1.0.
- **Punto flotante (`float32`/`float64`) para dinero:** Prohibido por imprecisión acumulativa IEEE 754.

### 4.3. Agent Constraints (Implementation Rules)
- If it's not in this brief, it does not exist — do not add features, files, or dependencies "because it makes sense."
- If something is ambiguous, do not guess — ask, or mark it as an open question / assumption per this brief's policies.
- Do not add dependencies outside the Stack table without flagging it first.
- Do not create files outside Project Structure without flagging it first.
- Do not rename established identifiers (see Naming Dictionary) without flagging it first.

---

## 5. Visual Identity
*(Omitido por Skip Rule: Proyecto backend/CLI sin interfaz gráfica).*

---

## 6. API & CLI Surface & Contracts

El binario ejecutable se compilará con el nombre `cash`.

### 6.1. Command Specification Table
| Command | Arguments | Flags | Stdout Contract | Stderr / Failure | Exit Code |
|---|---|---|---|---|---|
| `cash in` | `<category> <amount> [desc]` | `--date DD-MM-YYYY` | `✔ Ingreso registrado: +$3000.00 en 'sueldo' [ID: 01H...]` (en verde) | Error si amount <= 0 o categoría inválida | `0` éxito, `1` error |
| `cash out` | `<category> <amount> [desc]` | `--date DD-MM-YYYY` | `✔ Gasto registrado: -$15.50 en 'almuerzo' [ID: 01H...]` (en rojo) | Error si amount <= 0 o categoría inválida | `0` éxito, `1` error |
| `cash summary` | *(Ninguno)* | `--month MM-YYYY`, `--currency SYM` | Tarjeta con: Total Ingresos, Total Gastos, Balance Neto y Top categorías. | Error si formato de mes es inválido | `0` éxito, `1` error |
| `cash list` | *(Ninguno)* | `--month MM-YYYY`, `--category CAT`, `--limit N`, `--json` | Tabla ASCII con columnas: Fecha, ID, Tipo, Categoría, Monto, Descripción. Si `--json`, array JSON plano. | Error si parámetros son inválidos | `0` éxito, `1` error |
| `cash categories`| *(Ninguno)* | `--type in\|out` | Lista ordenada alfabéticamente de categorías únicas con conteo de transacciones y total acumulado. | Ninguno | `0` éxito |
| `cash sync` | *(Ninguno)* | `--remote NAME`, `--branch NAME` | Progreso paso a paso: `[1/3] Commit local... [2/3] Pull rebase... [3/3] Push... Sincronización exitosa.` | Advertencia si falla conexión; transacciones locales se preservan intactas. | `0` éxito, `2` error git |
| `cash init` | `[git-repo-url]` | *(Ninguno)* | Crea `~/.cashflow/entries` y configura el repositorio Git local/remoto. | Error si ruta no tiene permisos de escritura | `0` éxito, `1` error |

---

## 7. Data Model & Behavior

**Code Detail Level:** Contracts Only (Go interfaces, structs, pre/post-conditions).

### 7.1. Domain Entities & Value Objects

```go
package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type TransactionType string

const (
	TypeIncome  TransactionType = "income"
	TypeExpense TransactionType = "expense"
)

// Money representa una cantidad monetaria exacta en centavos para evitar floating-point drift.
type Money struct {
	Cents    int64  `json:"cents"`
	Currency string `json:"currency"` // Default: "USD"
}

func NewMoney(amount float64, currency string) (Money, error) {
	if amount <= 0 {
		return Money{}, fmt.Errorf("amount must be greater than zero")
	}
	cents := int64(amount*100 + 0.5)
	if currency == "" {
		currency = "USD"
	}
	return Money{Cents: cents, Currency: currency}, nil
}

func (m Money) Format() string {
	return fmt.Sprintf("%s %.2f", m.Currency, float64(m.Cents)/100.0)
}

// Transaction representa el evento inmutable de una operación financiera.
type Transaction struct {
	ID          string          `json:"id"`          // ULID de 26 caracteres
	Timestamp   time.Time       `json:"timestamp"`   // UTC ISO 8601
	Type        TransactionType `json:"type"`        // "income" | "expense"
	Category    string          `json:"category"`    // Slug normalizado: [a-z0-9-_]+
	Amount      Money           `json:"amount"`      // Objeto de valor
	Description string          `json:"description"` // Opcional
}

var categoryRegex = regexp.MustCompile(`^[a-z0-9-_]+$`)

func NormalizeCategory(raw string) (string, error) {
	slug := strings.ToLower(strings.TrimSpace(raw))
	slug = strings.ReplaceAll(slug, " ", "-")
	if !categoryRegex.MatchString(slug) {
		return "", fmt.Errorf("invalid category format: %s", raw)
	}
	return slug, nil
}
```

### 7.2. Hexagonal Ports (Interfaces)

```go
package ports

import (
	"context"
	"cashflow/internal/domain"
	"time"
)

type TransactionFilter struct {
	Month    *time.Time
	Category string
	Type     *domain.TransactionType
	Limit    int
}

// TransactionRepository define el puerto secundario (salida) para persistencia.
type TransactionRepository interface {
	// Save guarda una transacción como un archivo inmutable entries/<ID>.json.
	// Precondition: tx.ID es un ULID válido y no vacío. tx.Amount.Cents > 0.
	// Postcondition: Escribe el archivo en disco de forma atómica. Nunca sobrescribe un ID existente.
	// Error: ErrStorageUnavailable, ErrDuplicateTransaction.
	Save(ctx context.Context, tx domain.Transaction) error

	// FindAll recupera todas las transacciones que satisfacen el filtro.
	// Precondition: filter contiene criterios válidos o valores cero.
	// Postcondition: Retorna slice ordenado por Timestamp DESC. Retorna slice vacío si no hay coincidencias.
	// Error: ErrStorageUnavailable.
	FindAll(ctx context.Context, filter TransactionFilter) ([]domain.Transaction, error)

	// ListCategories extrae todas las categorías únicas utilizadas.
	// Postcondition: Retorna slice ordenado alfabéticamente sin duplicados.
	ListCategories(ctx context.Context) ([]string, error)
}

// SyncService define el puerto secundario para interactuar con Git.
type SyncService interface {
	// Sync ejecuta el ciclo de sincronización distribuida.
	// Precondition: Directorio de almacenamiento es un repositorio Git válido.
	// Postcondition: Realiza git add -> git commit -> git pull --rebase -> git push.
	// Error: ErrNoRemoteConfigured, ErrGitConflict, ErrNetworkUnavailable.
	Sync(ctx context.Context) error
}

// TransactionUseCases define el puerto primario (entrada) para la CLI.
type TransactionUseCases interface {
	RecordIncome(ctx context.Context, category string, amount float64, desc string, date *time.Time) (*domain.Transaction, error)
	RecordExpense(ctx context.Context, category string, amount float64, desc string, date *time.Time) (*domain.Transaction, error)
	GetMonthlySummary(ctx context.Context, year int, month time.Month) (*domain.MonthlySummary, error)
	ListTransactions(ctx context.Context, filter TransactionFilter) ([]domain.Transaction, error)
	Synchronize(ctx context.Context) error
}
```

---

## 8. Interactions & Feedback & Error Taxonomy

### 8.1. Action -> Response Mapping
| Action | Immediate Feedback | Final Result |
|---|---|---|
| `cash in sueldo 3000` | Ninguno (latencia < 5ms) | `✔ Ingreso registrado: +USD 3000.00 en 'sueldo' [ID: 01H...]` |
| `cash out café 3.5 "En la plaza"` | Normalización de slug `café` -> `cafe` | `✔ Gasto registrado: -USD 3.50 en 'cafe' [ID: 01H...]` |
| `cash sync` | Print step-by-step: `[1/3] Guardando...` | `✔ Sincronización con Git completada con éxito.` |

### 8.2. Error Taxonomy Table
| Error Type | Trigger | CLI User Message | Technical Handling | Exit Code | Retryable? |
|---|---|---|---|---|---|
| `ErrInvalidAmount` | Monto <= 0 o caracteres no numéricos | `Error: El monto debe ser un número decimal positivo (ej: 15.50).` | Abortar ejecución, no tocar disco | `1` | No |
| `ErrInvalidCategory` | Categoría contiene caracteres especiales no permitidos | `Error: Categoría inválida. Usa solo letras, números y guiones.` | Abortar ejecución | `1` | No |
| `ErrStorageUninitialized`| `~/.cashflow/entries` no existe al intentar guardar/leer | `Error: Almacenamiento no inicializado. Ejecuta 'cash init' primero.` | Guía al usuario hacia `cash init` | `1` | Sí (auto-init) |
| `ErrGitNetworkFailure` | Timeout de red o sin conexión durante `cash sync` | `Advertencia: Sin conexión al repositorio remoto. Tus datos están seguros en local.` | Exit gracefully, mantener transacciones locales | `2` | Sí |
| `ErrGitAuthFailure` | Clave SSH o credencial rechazada por GitHub/GitLab | `Error: Falla de autenticación en Git. Verifica tus credenciales SSH/HTTPS.` | Mostrar stderr de Git sin exponer tokens | `2` | No |
| `ErrCorruptFile` | Un archivo en `entries/` no es JSON válido | `Advertencia: Se omitió archivo corrupto: <path>` | Ignorar archivo, advertir en stderr, continuar parseando el resto | `0` | No |

---

## 9. Out of Scope (v1.0)

1. **Web Dashboard / GUI:** Cualquier interfaz web, gráfica o servidor HTTP queda formalmente diferido para la versión 2.0. La v1.0 es exclusivamente CLI de terminal.
2. **Conversión automática de divisas (FX Rates):** La v1.0 almacena transacciones con una moneda base configurada por el usuario (default `USD`). No hace llamadas a APIs de tasas de cambio.
3. **Interfaz TUI interactiva (Ncurses/Bubbletea):** La v1.0 opera por comandos y argumentos directos sin interfaces a pantalla completa.
4. **Modificación o eliminación destructiva in situ:** Por ser un modelo de eventos inmutables, la v1.0 no implementa `cash edit` ni `cash delete`. Las correcciones se registran mediante transacciones compensatorias.
5. **Presupuestos y metas de ahorro automáticas:** Reglas de alerta de límites presupuestarios no forman parte del MVP.

---

## 10. Closed Decisions Registry

1. **DEC-001: Arquitectura Hexagonal (Ports & Adapters)**  
   *Justificación:* Aísla las reglas de negocio de finanzas (`Transaction`, `Money`, `Summary`) del sistema de archivos y del comando Git. Permite probar el 100% de la lógica con un repositorio en memoria en microsegundos.
2. **DEC-002: Go 1.22+ como Lenguaje Único**  
   *Justificación:* Compilación en un único binario sin dependencias externas, velocidad de arranque inigualable (<10ms) esencial para UX de terminal, y soporte nativo en Termux (Android).
3. **DEC-003: Almacenamiento de Eventos Inmutables (Archivo por Transacción)**  
   *Justificación:* Soluciona la falacia del auto-merge en `.jsonl`. En Git, agregar archivos independientes con nombres únicos basados en ULID jamás produce conflictos de merge al sincronizar entre dispositivos desconectados.
4. **DEC-004: ULID como Identificador Universal**  
   *Justificación:* Combina 48 bits de timestamp y 80 bits de entropía aleatoria. Proporciona orden cronológico por defecto en el sistema de archivos sin necesidad de abrir ni indexar los JSON.
5. **DEC-005: Aritmética de Dinero en Centavos Enteros (`int64`)**  
   *Justificación:* Elimina los errores de redondeo de punto flotante binario (IEEE 754) en operaciones financieras.
6. **DEC-006: Política de Idiomas**  
   *Justificación:* Código fuente, interfaces, nombres de campos JSON, comandos y flags en **inglés**. Mensajes de cara al usuario en terminal y nombres de categorías del dominio en **español** (con soporte neutro).

### Supuestos a confirmar
- `~/.cashflow` como ruta base en todos los OS (resuelto mediante `os.UserHomeDir()`).
- Moneda por defecto `USD` si el usuario no especifica configuración.

---

## 11. Project Structure & Naming Dictionary

### 11.1. Directory Tree
```
cashflow-cli/
├── cmd/
│   └── cash/
│       └── main.go                  # Bootstrap: inyección de dependencias y ejecución de root command
├── internal/
│   ├── domain/                      # Núcleo puro (sin dependencias externas)
│   │   ├── transaction.go           # Entidad Transaction y enum TransactionType
│   │   ├── money.go                 # Value Object Money y aritmética de centavos
│   │   ├── summary.go               # Agregados y estadísticas mensuales
│   │   └── errors.go                # Errores centinela de dominio
│   ├── ports/                       # Interfaces de entrada y salida
│   │   ├── repository.go            # Puerto secundario de persistencia
│   │   ├── sync.go                  # Puerto secundario de sincronización Git
│   │   └── usecases.go              # Puerto primario de la aplicación
│   ├── usecases/                    # Orquestación de lógica de negocio
│   │   ├── record_transaction.go    # Casos de uso de ingreso y gasto
│   │   ├── get_summary.go           # Cálculo de balances y sumatorias
│   │   ├── list_transactions.go     # Filtros y ordenamiento
│   │   └── sync_data.go             # Orquestación de sincronización
│   └── adapters/                    # Adaptadores de infraestructura
│       ├── storage/
│       │   └── file_repository.go   # Implementación JSON en entries/<ulid>.json
│       ├── vcs/
│       │   └── git_service.go       # Adaptador de comandos Git nativos
│       └── cli/                     # Adaptador primario (Cobra CLI)
│           ├── root.go              # Comando base y configuración global
│           ├── in.go                # Subcomando 'cash in'
│           ├── out.go               # Subcomando 'cash out'
│           ├── summary.go           # Subcomando 'cash summary'
│           ├── list.go              # Subcomando 'cash list'
│           ├── categories.go        # Subcomando 'cash categories'
│           ├── sync.go              # Subcomando 'cash sync'
│           └── init.go              # Subcomando 'cash init'
├── go.mod
├── go.sum
└── README.md
```

### 11.2. Naming Dictionary
| Concepto | Término Canónico | Utilizado en | No confundir con |
|---|---|---|---|
| Registro inmutable individual | `Transaction` | Domain, Ports, Storage | `Entry` (nombre de la carpeta de almacenamiento) |
| Cantidad monetaria exacta | `Money` (`Cents int64`) | Domain, Value Object | `Amount` (float64 usado solo en parsing de CLI) |
| Puerto de almacenamiento | `TransactionRepository` | Ports, Adapters | `GitService` (puerto de transporte/sync) |

---

## 12. Implementation Roadmap

### Phase 1 — Domain Core & In-Memory Ports
**Deliverables:** `internal/domain/*.go`, `internal/ports/*.go`, `internal/usecases/*.go`, `internal/domain/*_test.go`  
**Done when ALL of:**
- [ ] `Money` previene montos <= 0 y convierte floats a centavos enteros con precisión exacta.
- [ ] `NormalizeCategory` sanitiza espacios y mayúsculas (`"Comida Rápida"` -> `"comida-rapida"`).
- [ ] Pruebas unitarias de casos de uso ejecutadas contra un mock en memoria pasan con 100% de éxito.

### Phase 2 — File Storage Adapter (Event Log)
**Deliverables:** `internal/adapters/storage/file_repository.go`, `file_repository_test.go`  
**Done when ALL of:**
- [ ] Cada llamada a `Save()` crea un archivo único `~/.cashflow/entries/<ulid>.json`.
- [ ] `FindAll()` lee el directorio, parsea los archivos JSON y retorna los registros ordenados por timestamp descendente.
- [ ] La presencia de un archivo con JSON inválido emite una advertencia pero no aborta el parseo de los demás archivos.

### Phase 3 — Cobra CLI Commands & Rendering
**Deliverables:** `internal/adapters/cli/*.go`, `cmd/cash/main.go`  
**Done when ALL of:**
- [ ] Comandos `cash in` y `cash out` registran transacciones en disco en menos de 10ms.
- [ ] `cash summary` imprime resumen formateado con balance neto calculado correctamente.
- [ ] `cash list --json` emite un array JSON válido sin decoraciones ANSI para interoperabilidad.

### Phase 4 — Git Sync Adapter & Verification
**Deliverables:** `internal/adapters/vcs/git_service.go`, `cash sync`, `cash init`  
**Done when ALL of:**
- [ ] `cash init` inicializa el directorio local y el repositorio Git si no existen.
- [ ] `cash sync` ejecuta `git add entries/`, `git commit`, `git pull --rebase` y `git push`.
- [ ] Si no hay conexión a internet, `cash sync` falla con código de salida `2`, informando que los datos locales están a salvo.

---

## 13. Acceptance Criteria & Test Plan

1. **AC-01 (Registro Rápido):** Al ejecutar `cash out almuerzo 12.50 "Menu ejecutivo"`, se crea un archivo en `entries/` con `amount_cents: 1250`, `type: "expense"`, y la CLI retorna exit code 0 en <100ms.
2. **AC-02 (Cálculo Financiero):** Para 2 ingresos de $1000 y 3 gastos de $200, `cash summary` reporta exactamente: Ingresos: $2000.00, Gastos: $600.00, Balance: +$1400.00.
3. **AC-03 (Sincronización Concurrente):** Si el Dispositivo A crea una transacción desconectado y el Dispositivo B crea otra desconectado, al ejecutar `cash sync` en ambos, ambos repositorios contienen ambas transacciones sin intervención manual ni conflicto de Git.
4. **AC-04 (Manejo de Errores):** Ejecutar `cash in comida -50` debe mostrar `Error: El monto debe ser un número decimal positivo` y finalizar con exit code 1 sin crear archivos.

---

## 14. Open Questions
*Ninguna pregunta abierta pendiente.* Todos los supuestos iniciales fueron resueltos en las decisiones cerradas (Go como lenguaje, archivo por transacción como modelo de almacenamiento, y exclusión estricta de la web para v1.0).

---

## 15. Glossary
- **ULID (Universally Unique Lexicographically Sortable Identifier):** Identificador de 26 caracteres compatible con UUID pero ordenable por tiempo.
- **Event Sourcing / Event Log:** Patrón de persistencia donde cada cambio de estado es un evento inmutable añadido secuencialmente en lugar de sobreescribir un registro previo.
- **Hexagonal Architecture:** Patrón arquitectónico que aísla la lógica central de negocio del mundo exterior mediante puertos (interfaces) y adaptadores (implementaciones concretas).
