# Discovery Notes: CashFlow TUI Dashboard Extension

**Date:** 2026-09-06  
**Status:** Complete  
**Project Target:** `cashflow-cli` (v1.1 Evolution)  
**Target Subcommand:** `cash tui` (Aliases: `dashboard`, `ui`)  

---

## 1. Vision & Core Value
Incorporar un dashboard interactivo de terminal de **solo lectura** ejecutado mediante el comando `cash tui`. Permite al desarrollador tener una visión panorámica instantánea de sus finanzas del mes con una interfaz estética y estilizada, sin romper el estándar POSIX del comando raíz `cash` ni la velocidad de captura de los subcomandos rápidos `cash in` y `cash out`.

**Filosofía:** *"Captura rápida por comandos CLI; análisis y auditoría visual instantánea con `cash tui`."*

---

## 2. Constraints & Non-Functional Boundaries
1. **No interferencia con CLI:** El comando raíz `cash` sigue mostrando la ayuda y lista de comandos nativa de Cobra. El dashboard vive exclusivamente en `cash tui`.
2. **Arquitectura Hexagonal Pura:** La TUI es un **adaptador de presentación primario** (`internal/adapters/tui/`). No toca ni modifica una sola línea del dominio (`internal/domain/`) ni del almacenamiento (`internal/adapters/storage/`).
3. **Solo Lectura (Read-Only):** Cero formularios emergentes o edición dentro de la TUI en esta fase; el registro de ingresos y gastos se mantiene rápido en la terminal (`cash in` / `cash out`).
4. **Resiliente y Ligera:** Salida inmediata con `q`, `Esc` o `Ctrl+C`. Compatible con terminales estándar en Windows (PowerShell/Windows Terminal), Linux y Termux en Android.

---

## 3. Stack & Architecture
- **TUI Framework:** `github.com/charmbracelet/bubbletea` v1.x (Arquitectura Elm: Model-Update-View determinista).
- **Styling & Layout:** `github.com/charmbracelet/lipgloss` v1.x (Estilos CSS-like para bordes, colores, padding y cajas de métricas).
- **Componentes:** `github.com/charmbracelet/bubbles` (tabla interactiva con scroll vertical).
- **Integración:** Consume directamente `ports.TransactionUseCases` (`GetMonthlySummary` y `ListTransactions`).

---

## 4. Screen Layout & Visual Wireframe

```
+-----------------------------------------------------------------------+
|  $$$    C A S H F L O W   T U I                                       |
|  $ $    <  SEPTIEMBRE 2026  >                                         |
+-----------------------------------------------------------------------+
|  +-------------------+  +-------------------+  +-------------------+  |
|  | INGRESOS TOTALES  |  |  GASTOS TOTALES   |  |   BALANCE NETO    |  |
|  |   +USD 3,000.00   |  |    -USD 600.00    |  |   +USD 2,400.00   |  |
|  +-------------------+  +-------------------+  +-------------------+  |
+-----------------------------------------------------------------------+
|  FECHA       | TIPO    | CATEGORÍA      | MONTO      | DESCRIPCIÓN    |
|  ------------+---------+----------------+------------+--------------  |
|  2026-09-06  | INCOME  | sueldo         | +$3,000.00 | Quincena       |
|  2026-09-05  | EXPENSE | almuerzo       | -$15.50    | Menu del dia   |
|  2026-09-04  | EXPENSE | transporte     | -$4.50     | Metro          |
+-----------------------------------------------------------------------+
|  [←/→ / h/l] Cambiar Mes  •  [↑/↓ / j/k] Navegar  •  [q] Salir        |
+-----------------------------------------------------------------------+
```

### Elementos de la Interfaz:
1. **Header:** 
   - Logo ASCII (importado desde `dollar-sign.md` o fallback estilizado).
   - Selector de mes actual con indicadores de navegación `< Mes Año >`.
2. **Tarjetas de Métricas (3 columnas):**
   - Tarjeta Verde: Ingresos Totales acumulados en el mes seleccionado.
   - Tarjeta Roja: Gastos Totales acumulados.
   - Tarjeta Verde/Roja: Balance Neto calculado.
3. **Tabla de Transacciones:**
   - Columnas: Fecha, Tipo (badge coloreado), Categoría, Monto y Descripción.
   - Navegación con cursor (`↑`/`↓` o `j`/`k`).
   - Estado vacío si el mes no tiene movimientos: *"No hay transacciones registradas para este mes."*
4. **Footer:** Barra de ayuda con atajos de teclado disponibles.

---

## 5. Interactions & Keybindings
| Tecla | Acción | Comportamiento |
|---|---|---|
| `←` / `h` | Mes Anterior | Retrocede un mes, recalcula tarjetas y recarga tabla de transacciones. |
| `→` / `l` | Mes Siguiente | Avanza un mes, recalcula tarjetas y recarga tabla. |
| `↑` / `k` | Subir cursor | Desplaza la selección de fila hacia arriba en la tabla. |
| `↓` / `j` | Bajar cursor | Desplaza la selección de fila hacia abajo en la tabla. |
| `t` | Ir a hoy | Restaura la vista al mes y año en curso. |
| `q` / `Esc` / `Ctrl+C` | Salir | Cierra la TUI limpiamente y restaura la terminal. |

---

## 6. Closed Decisions Made
1. **DEC-TUI-01:** Creación del subcomando `cash tui` (en lugar de sobrecargar el comando base `cash`).
2. **DEC-TUI-02:** Modo de **solo lectura**. La creación de datos se mantiene en `cash in`/`cash out` para preservar la regla de sub-10ms en terminal.
3. **DEC-TUI-03:** Navegación mensual dinámica mediante recálculo reactivo en memoria con el repositorio local.
4. **DEC-TUI-04:** Stack Bubble Tea + Lip Gloss como estándar de facto en Go.

---

## 7. Needs Assumption / Gaps to Confirm
1. **Logo ASCII:** `dollar-sign.md` se encuentra vacío actualmente. Se utilizará un logo tipográfico estilizado por defecto a menos que el usuario pegue su arte ASCII en dicho archivo.

---

## 8. Out of Scope (v1.1 TUI)
1. Formularios de captura de transacciones dentro de la TUI (se mantiene en CLI).
2. Gráficos de barras o torta ASCII (reservados para Fase Web o v1.2).
3. Soporte para clics del mouse (interacción 100% por teclado).
4. Edición de registros dentro de la tabla.
