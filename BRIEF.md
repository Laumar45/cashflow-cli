# Design Brief: CashFlow CLI — v1.1.0 Evolution (TUI Dashboard)

## 1. Header Block
- **Version:** 1.1.0
- **Date:** 2026-09-07
- **Status:** Ready for Implementation
- **Stack Summary:** Go 1.22+, Cobra CLI, Bubble Tea, Lip Gloss, Bubbles, Immutable JSON Event Files, Git VCS Adapter
- **Build Mode:** Evolution v1.1.0 (Hexagonal Architecture Extension)
- **Supersedes:** Design Brief v1.0.0 (2026-09-04)

---

## 2. Summary & Guiding Principle
**CashFlow CLI** es una herramienta de finanzas personales ultrarrápida y *offline-first* para desarrolladores y usuarios de terminal. En su versión 1.0 consolidó el registro atómico de ingresos y gastos en archivos inmutables (`~/.cashflow/entries/<ulid>.json`) y la sincronización distribuida con Git sin conflictos.

La **versión 1.1** incorpora un dashboard interactivo de terminal de **solo lectura** ejecutado mediante el subcomando `cash tui`. Permite auditar el estado financiero mensual, navegar entre meses y examinar transacciones en una interfaz responsiva y accesible, preservando intacta la captura rápida en línea de comandos.

**Guiding Principle:**  
> **"Captura rápida en terminal en milisegundos; auditoría visual interactiva de solo lectura con `cash tui`. Cero fricción, cero costo y cero llamadas de red espurias."**

---

## 3. Delta Matrix (v1.0 → v1.1)

| Section | Change Type | v1.0 Baseline | v1.1 Evolution | Technical Rationale |
|---|:---:|---|---|---|
| **§4 Stack** | Modified | Cobra + Color + Stdlib | Se agregan `bubbletea`, `lipgloss` y `bubbles` | Ecosistema canónico en Go para interfaces de terminal declarativas basadas en arquitectura Elm. |
| **§6 Surface** | Modified | 7 comandos de línea de texto | Se agrega subcomando `cash tui` (aliases: `dashboard`, `ui`) | Punto de entrada aislado que preserva la ayuda nativa de Cobra en el comando raíz `cash`. |
| **§8 Interactions** | Modified | Mapeo interactivo texto CLI | Se añade matriz de atajos de teclado (`h/l`, `j/k`, `t`, `q`) | Navegación de teclado intuitiva sin dependencia de eventos del ratón. |
| **§9 Out of Scope** | Modified | Excluía TUI por completo | Permite TUI de solo lectura; excluye formularios internos, edición y sparklines | Protege la velocidad de captura en CLI y evita sobrecargar la interfaz con gráficos de baja legibilidad. |
| **§10 Decisions** | Modified | DEC-001 a DEC-006 | Se incorporan DEC-TUI-01 a DEC-TUI-08 | Formaliza glifos accesibles `▲`/`▼`, breakpoint de 80 columnas y lectura de estado de Git 100% local. |
| **§11 Structure** | Modified | Adapters storage, cli, vcs | Se añade `internal/adapters/tui/` | Adaptador primario nuevo; desacoplado de la lógica de negocio gracias a los puertos existentes. |
| **§12 Roadmap** | Modified | Fases 1 a 4 (Completadas) | Se agrega **Fase 5: TUI Dashboard Adapter** | Extensión modular con entregables y criterios de cierre "Done when" verificables. |
| **§13 Criteria** | Modified | AC-01 a AC-04 | Se agregan AC-05 a AC-08 (TUI) | Criterios verificables de renderizado, responsividad, accesibilidad y navegación. |

---

## 4. Stack & Constraints

### 4.1. Stack Table
| Component | Technology | Version | Architectural Justification |
|---|---|---|---|
| **Core Runtime** | Go | 1.22+ | Binario estático compilado, arranque sub-10ms, consumo ínfimo de memoria. |
| **CLI Framework** | `spf13/cobra` | v1.8+ | Despacho de subcomandos y autocompletado estándar POSIX. |
| **TUI Framework** | `charmbracelet/bubbletea` | v1.2+ | Framework declarativo basado en The Elm Architecture (`Model`, `Update`, `View`) para manejo determinista de eventos de terminal. |
| **TUI Styling** | `charmbracelet/lipgloss` | v1.0+ | Definición de estilos modulares tipo CSS (bordes, márgenes, padding, colores y layout de cajas). |
| **TUI Components** | `charmbracelet/bubbles` | v0.20+ | Componente de viewport para scroll vertical fluido de tablas de transacciones. |
| **Storage & VCS** | Immutable JSON + Git CLI | RFC 8259 / Git 2.x+ | Persistencia atómica de eventos en `entries/<ulid>.json` y sync distribuido. |

### 4.2. Not in Stack
- Librerías GUI de escritorio (Wails, Fyne) o servidores Web (React, Vue) para v1.1.
- Motores de gráficos ASCII / Sparklines complejos basados en caracteres de bloque (` ▂▃▄▅▆▇█`).
- Captura de eventos del ratón (Mouse tracking) dentro de la TUI.

### 4.3. Agent Constraints (Implementation Rules)
- If it's not in this brief, it does not exist — do not add features, files, or dependencies "because it makes sense."
- If something is ambiguous, do not guess — ask, or mark it as an open question / assumption per this brief's policies.
- Do not add dependencies outside the Stack table without flagging it first.
- Do not create files outside Project Structure without flagging it first.
- Do not modify domain entities (`internal/domain/`) or ports (`internal/ports/`) unless explicitly specified in this evolution brief.

---

## 5. Visual Identity & Tokens (TUI)

### 5.1. Color & Style Tokens
| Token | Lip Gloss Def | Uso Semántico |
|---|---|---|
| `Primary` | `lipgloss.Color("#7D56F4")` | Títulos, bordes activos y acentos del header |
| `Success` | `lipgloss.Color("#04B575")` | Ingresos, balance neto positivo y glifo `▲` |
| `Danger` | `lipgloss.Color("#FF4444")` | Gastos, balance neto negativo y glifo `▼` |
| `Muted` | `lipgloss.Color("#626262")` | Bordes secundarios, separadores y footer |
| `NeutralText` | `lipgloss.Color("#DDDDDD")` | Texto principal de datos y filas de tabla |
| `Highlight` | `lipgloss.Color("#E5C07B")` | Categorías y glifos de selección |

### 5.2. Regla de Accesibilidad Universal (DEC-TUI-05)
Ningún estado financiero depende exclusivamente del color:
- **Ingreso:** Verde + Glifo `▲` (ej. `▲ +USD 3000.00`).
- **Gasto:** Rojo + Glifo `▼` (ej. `▼ -USD 15.50`).
- **Balance Neto:** Verde + `▲` si `>= 0`; Rojo + `▼` si `< 0`.

---

## 6. API & CLI Surface & Contracts

### 6.1. Subcomando `cash tui`
- **Invocación:** `cash tui [flags]`
- **Aliases:** `dashboard`, `ui`
- **Flags:**
  - `--dir PATH`: Directorio de almacenamiento personalizado (heredado de bandera global).
  - `--month YYYY-MM`: Iniciar la TUI posicionado en un mes específico (por defecto: mes actual).

### 6.2. Wireframes de Layout y Breakpoints

#### Layout Estándar (Ancho de Terminal ≥ 80 columnas)
```
+-----------------------------------------------------------------------+
|  $$$    C A S H F L O W   T U I                                       |
|  $ $    <  SEPTIEMBRE 2026  >                                         |
+-----------------------------------------------------------------------+
|  +---------------+  +---------------+  +---------------+  +---------+ |
|  | INGRESOS   ▲  |  | GASTOS     ▼  |  | BALANCE    ▲  |  | TOP CAT.| |
|  | +USD 3000.00  |  | -USD 600.00   |  | +USD 2400.00  |  | sueldo  | |
|  +---------------+  +---------------+  +---------------+  +---------+ |
+-----------------------------------------------------------------------+
|  FECHA       | TIPO       | CATEGORÍA      | MONTO      | DESCRIPCIÓN |
|  ------------+------------+----------------+------------+-----------  |
|  2026-09-06  | ▲ INCOME   | sueldo         | +$3,000.00 | Quincena    |
|  2026-09-05  | ▼ EXPENSE  | almuerzo       | -$15.50    | Menu del dia|
|  2026-09-04  | ▼ EXPENSE  | transporte     | -$4.50     | Metro       |
+-----------------------------------------------------------------------+
|  Fila 3/47  •  Última sync: hace 2h  •  Cambios sin sync: 0           |
|  [←/→ h/l] Mes  •  [↑/↓ j/k] Navegar  •  [t] Hoy  •  [q] Salir        |
+-----------------------------------------------------------------------+
```

#### Layout Compacto (Ancho de Terminal < 80 columnas — Termux / Móvil)
```
+---------------------------------+
| $$$ CASHFLOW TUI                |
| < SEPTIEMBRE 2026 >             |
+---------------------------------+
| INGRESOS ▲     +USD 3,000.00    |
| GASTOS   ▼       -USD 600.00    |
| BALANCE  ▲     +USD 2,400.00    |
| TOP CAT.          sueldo        |
+---------------------------------+
| FECHA      TIPO     MONTO       |
| ---------  --------  ---------  |
| 09-06      ▲ INCOME  +3,000.00  |
| 09-05      ▼ EXPENSE   -15.50   |
+---------------------------------+
| Fila 2/47  •  ↻ 2h              |
| [h/l] Mes [j/k] Nav [q] Salir   |
+---------------------------------+
```

---

## 7. Data Model & Behavior

**Code Detail Level:** Contracts Only (Interfaces y structs de TUI).

### 7.1. Reutilización de Puertos Existentes
La TUI no altera el dominio. Consume exclusivamente `ports.TransactionUseCases`:
- `GetMonthlySummary(ctx, year, month)`: Provee ingresos, gastos, balance neto y `ExpenseByCategory` (utilizado para derivar Top Categoría).
- `ListTransactions(ctx, filter)`: Provee las transacciones del mes ordenadas cronológicamente para la tabla.

### 7.2. Contrato del Modelo de Presentación TUI

```go
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"cashflow/internal/domain"
	"cashflow/internal/ports"
	"time"
)

// SyncStatusInfo provee métricas de solo lectura local del repositorio Git.
type SyncStatusInfo struct {
	LastSyncRelative string // ej: "hace 2h"
	UnsyncedCount    int    // cantidad de archivos locales modificados sin commit/sync
}

// Model define el estado determinista de Bubble Tea para el dashboard.
type Model struct {
	usecases      ports.TransactionUseCases
	currentMonth  time.Time
	summary       *domain.MonthlySummary
	transactions  []domain.Transaction
	cursorIndex   int
	width         int
	height        int
	isCompact     bool
	syncInfo      SyncStatusInfo
	err           error
}

func NewModel(svc ports.TransactionUseCases, initialMonth time.Time, syncInfo SyncStatusInfo) Model
func (m Model) Init() tea.Cmd
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m Model) View() string
```

---

## 8. Interactions & Feedback

### 8.1. Tabla de Atajos de Teclado
| Tecla | Acción | Comportamiento en TUI |
|---|---|---|
| `←` / `h` | Mes Anterior | Retrocede un mes en `currentMonth`, recalcula métricas y reinicia cursor en fila 0. |
| `→` / `l` | Mes Siguiente | Avanza un mes en `currentMonth`, recalcula métricas y reinicia cursor en fila 0. |
| `↑` / `k` | Fila Arriba | Desplaza cursor una fila arriba en la tabla de transacciones. |
| `↓` / `j` | Fila Abajo | Desplaza cursor una fila abajo en la tabla de transacciones. |
| `t` | Ir a Hoy | Resetea la fecha al mes y año actual (`time.Now()`). |
| `q` / `Esc` / `Ctrl+C` | Salir | Emite `tea.Quit`, restaura terminal y finaliza proceso con exit code 0. |

### 8.2. Estados Vacíos y de Error
- **Mes sin transacciones:** Muestra mensaje centrado en tabla: `"No hay transacciones registradas para este mes."` con tarjetas en `$0.00`.
- **Terminal en dimensiones extremas (ancho < 30 o alto < 10):** Muestra vista mínima: `"Terminal demasiado pequeña para renderizar CashFlow TUI."`

---

## 9. Out of Scope (v1.1)

1. **Formularios de captura/edición dentro de la TUI:** El registro permanece como responsabilidad de `cash in` y `cash out` en terminal.
2. **Gráficos Sparkline o caracteres de bloque:** No se incorporan gráficos de barras dentro de la terminal en v1.1.
3. **Eventos de mouse:** Interacción 100% controlada por teclado.
4. **Operaciones de red automáticas en segundo plano:** El indicador de sync en footer lee únicamente el estado local sin disparar `git fetch`, `pull` ni `push`.
5. **Filtrado dinámico interactivo en TUI:** Búsquedas por texto o categoría dentro de la TUI quedan reservadas para v1.2.

---

## 10. Closed Decisions Registry

### Inherited Decisions (Active from v1.0)
- **DEC-001 (Active):** Arquitectura Hexagonal (Ports & Adapters).
- **DEC-002 (Active):** Go 1.22+ como lenguaje único.
- **DEC-003 (Active):** Storage en archivos de eventos inmutables (`entries/<ulid>.json`).
- **DEC-004 (Active):** ULID como identificador de 26 caracteres.
- **DEC-005 (Active):** Aritmética de dinero en centavos enteros (`int64`).
- **DEC-006 (Active):** Código en inglés, interfaz de usuario en español.

### New Decisions (v1.1 Evolution)
1. **DEC-TUI-01:** Subcomando `cash tui` con aliases `dashboard` y `ui` (sin sobrecargar `cash` base).
2. **DEC-TUI-02:** Modo estricto de **solo lectura**. Cero formularios emergentes para preservar la pureza y rapidez de captura del CLI.
3. **DEC-TUI-03:** Navegación mensual dinámica reactiva con flechas y teclas Vim (`h`/`l`).
4. **DEC-TUI-04:** Stack Bubble Tea + Lip Gloss como estándar declarativo en Go.
5. **DEC-TUI-05:** Accesibilidad daltónica obligatoria mediante glifos redundantes `▲`/`▼` junto a todo color financiero.
6. **DEC-TUI-06:** Breakpoint responsivo a 80 columnas alternando entre layout estándar y compacto (Termux/móvil).
7. **DEC-TUI-07:** Estado de sync en footer de **solo lectura local** (conteo de cambios locales y timestamp relativo) sin operaciones de red.
8. **DEC-TUI-08:** Tarjeta de "Top Categoría" reutiliza los agregados ya provistos por `GetMonthlySummary` sin alterar el dominio.

### Supuestos a confirmar
- Logo ASCII: Si `dollar-sign.md` está vacío, se utiliza el logo tipográfico estilizado por defecto.
- Formato relativo humano para última sincronización (`hace 2h`, `ayer`, `hace 5m`).

---

## 11. Project Structure & Naming Dictionary

### 11.1. Directory Tree (Evolution)
```
cashflow-cli/
├── cmd/
│   └── cash/
│       └── main.go
├── internal/
│   ├── domain/                  # Intacto (v1.0)
│   ├── ports/                   # Intacto (v1.0)
│   ├── usecases/                # Intacto (v1.0)
│   └── adapters/
│       ├── storage/             # Intacto (v1.0)
│       ├── vcs/                 # Intacto (v1.0)
│       ├── cli/                 # Subcomandos existentes + registro de 'cash tui'
│       │   └── tui_cmd.go       # [NUEVO] Subcomando 'cash tui'
│       └── tui/                 # [NUEVO] Adaptador de interfaz de terminal
│           ├── model.go         # Modelo de Bubble Tea, Init, Update y View
│           ├── styles.go        # Definición de tokens Lip Gloss y layouts
│           ├── header.go        # Renderizado de logo ASCII y selector de mes
│           ├── cards.go         # Renderizado de tarjetas estándar y compactas
│           ├── table.go         # Renderizado y scroll de tabla de transacciones
│           ├── sync_status.go   # Extractor de métricas locales de Git para footer
│           └── tui_test.go      # Pruebas unitarias de renderizado y transiciones de estado
├── go.mod
├── go.sum
├── dollar-sign.md
├── BRIEF.md                     # Este documento actualizado a v1.1.0
└── IMPLEMENTATION_LOG.md
```

### 11.2. Naming Dictionary
| Concepto | Término Canónico | Utilizado en | No confundir con |
|---|---|---|---|
| Adaptador de interfaz TUI | `tui.Model` | `internal/adapters/tui/` | `cli.tuiCmd` (comando de arranque) |
| Métricas de sincronización | `SyncStatusInfo` | `internal/adapters/tui/` | `ports.SyncService` (ejecutor de sync) |
| Vista compacta | `isCompact` (< 80 col) | `tui.styles` | `viewport` (componente de scroll) |

---

## 12. Implementation Roadmap

### Phase 5 — TUI Dashboard Adapter (Bubble Tea & Lip Gloss)
**Deliverables:**
- `internal/adapters/tui/*.go` (model, styles, header, cards, table, sync_status)
- `internal/adapters/cli/tui_cmd.go`
- `internal/adapters/tui/tui_test.go`

**Done when ALL of:**
- [ ] `cash tui` renderiza el logo, selector de mes, tarjetas de métricas con `▲`/`▼` y la tabla de transacciones.
- [ ] Al presionar `←`/`→` (o `h`/`l`), el modelo actualiza el mes y recalcula las métricas en pantalla al instante.
- [ ] Al redimensionar la ventana por debajo de 80 columnas (`tea.WindowSizeMsg`), el layout colapsa a modo compacto apilado sin desbordamiento.
- [ ] Al presionar `q`, la aplicación finaliza limpiamente con exit code 0.
- [ ] La suite completa `go test ./...` pasa con 100% de éxito.

---

## 13. Acceptance Criteria & Test Plan (Evolution)

1. **AC-05 (Renderizado de Dashboard):** Al ejecutar `cash tui`, se inicializa la vista a pantalla completa con las tarjetas de Ingresos, Gastos y Balance Neto correspondientes al mes actual y la tabla de movimientos ordenada cronológicamente.
2. **AC-06 (Accesibilidad Simbólica):** Todos los valores de ingreso llevan el símbolo `▲` y color verde; todos los gastos llevan `▼` y color rojo; el balance neto cambia dinámicamente entre `▲` verde y `▼` rojo según su valor.
3. **AC-07 (Responsividad <80 Columnas):** Con un mensaje de ventana con `width < 80`, la TUI oculta la columna Descripción, apila verticalmente las tarjetas de métricas y muestra el indicador compacto de sincronización sin deformar los bordes.
4. **AC-08 (Navegación Temporal Fluida):** Enviar evento de tecla `←` cambia el mes visualizado a `time.Month - 1` y actualiza las filas de la tabla con los datos de dicho mes en <15ms.

---

## 14. Open Questions
*Ninguna pregunta abierta bloqueante.* El comportamiento de las dimensiones reducidas, los atajos de teclado y la reutilización de puertos existentes quedaron formalizados.

---

## 15. Glossary
- **The Elm Architecture (TEA):** Patrón de diseño para interfaces de usuario reactivas estructuradas en tres componentes puros: Modelo (Estado), Actualización (Update/Mensajes) y Vista (View declarativa).
- **Lip Gloss:** Librería declarativa para estilizado de componentes de terminal mediante composición de bordes, alineación, márgenes y colores ANSI de 24 bits.
- **Bubble Tea:** Framework para aplicaciones interactivas de terminal basado en TEA.
