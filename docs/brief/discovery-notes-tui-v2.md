# Discovery Notes: CashFlow TUI Dashboard Extension

**Date:** 2026-09-07
**Status:** Complete (v2 — Revisión post-feedback de diseño)
**Project Target:** `cashflow-cli` (v1.1 Evolution)
**Target Subcommand:** `cash tui` (Aliases: `dashboard`, `ui`)
**Supersedes:** `discovery-notes-tui.md` v1 (2026-09-06)

---

## 0. Changelog v1 → v2
1. **Accesibilidad de color:** indicadores simbólicos (`▲`/`▼`) añadidos junto al color en tarjetas y tabla, para no depender exclusivamente de rojo/verde.
2. **Responsividad:** definido un breakpoint de ancho de terminal con layout colapsado para Termux/pantallas angostas.
3. **Indicador de overflow en tabla:** contador de posición (`fila actual/total`) cuando la lista excede el viewport.
4. **Estado de carga:** mensaje transitorio al recalcular tras cambio de mes.
5. **Indicador de sincronización:** línea informativa de solo lectura sobre el estado de `cash sync` en el footer.
6. **Tarjeta de Top Categoría:** cuarta tarjeta que reutiliza el cálculo ya expuesto por `cash summary`.
7. Confirmado explícitamente: se mantiene fuera de alcance cualquier gráfico de barras/sparkline con caracteres de bloque, sin excepciones, por decisión previa registrada en Out of Scope §8.2.

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
5. **[NUEVO] Accesible por diseño:** Ningún estado (ingreso/gasto, balance positivo/negativo) se comunica únicamente por color. Todo estado con semántica financiera lleva un símbolo redundante.
6. **[NUEVO] Adaptable al ancho del terminal:** El layout debe degradar de forma predecible en anchos reducidos (ver §4.3), en lugar de truncar o romper visualmente.

---

## 3. Stack & Architecture
- **TUI Framework:** `github.com/charmbracelet/bubbletea` v1.x (Arquitectura Elm: Model-Update-View determinista).
- **Styling & Layout:** `github.com/charmbracelet/lipgloss` v1.x (Estilos CSS-like para bordes, colores, padding y cajas de métricas).
- **Componentes:** `github.com/charmbracelet/bubbles` (tabla interactiva con scroll vertical, uso del componente `viewport` para overflow).
- **Integración:** Consume directamente `ports.TransactionUseCases` (`GetMonthlySummary` y `ListTransactions`). No se agregan nuevos puertos: el indicador de sync (§4.4) y el top de categorías (§4.2) reutilizan datos ya expuestos por casos de uso existentes (`GetMonthlySummary` y `ports.SyncService` en modo lectura de estado, sin invocar `Sync()`).

---

## 4. Screen Layout & Visual Wireframe

### 4.1. Layout estándar (ancho ≥ 80 columnas)

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
|  Fila 3/47  •  Última sync: hace 2h  •  cambios locales sin sync: 3   |
|  [←/→ h/l] Mes  •  [↑/↓ j/k] Navegar  •  [t] Hoy  •  [q] Salir        |
+-----------------------------------------------------------------------+
```

### 4.2. Layout compacto (ancho < 80 columnas — Termux / móvil en retrato)

Las tarjetas colapsan de fila horizontal a bloque vertical apilado, y la tabla oculta la columna "Descripción" (última prioridad visual). El indicador de sync se reduce a un solo glifo (`↻ 2h`) para no competir por espacio con los atajos de teclado.

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

### Elementos de la Interfaz:
1. **Header:**
   - Logo ASCII (importado desde `dollar-sign.md` o fallback estilizado).
   - Selector de mes actual con indicadores de navegación `< Mes Año >`.
2. **Tarjetas de Métricas (3 columnas, 4 en layout estándar):**
   - Tarjeta Verde + `▲`: Ingresos Totales acumulados en el mes seleccionado.
   - Tarjeta Roja + `▼`: Gastos Totales acumulados.
   - Tarjeta Verde/Roja + `▲`/`▼` según signo: Balance Neto calculado.
   - **[NUEVO]** Tarjeta neutra: Categoría con mayor acumulado del mes (reutiliza el cálculo de "Top categorías" ya presente en `cash summary`).
3. **Tabla de Transacciones:**
   - Columnas: Fecha, Tipo (badge coloreado + símbolo direccional), Categoría, Monto y Descripción (esta última se oculta primero en layout compacto).
   - Navegación con cursor (`↑`/`↓` o `j`/`k`), usando `bubbles/viewport` para scroll cuando el número de filas excede el alto disponible.
   - **[NUEVO]** Indicador de posición `Fila X/N` cuando la lista no cabe completa en pantalla, para que el usuario sepa que hay más contenido.
   - Estado vacío si el mes no tiene movimientos: *"No hay transacciones registradas para este mes."*
   - **[NUEVO]** Estado transitorio de recálculo al cambiar de mes: *"Cargando..."* se muestra solo si la relectura de disco supera un umbral perceptible (evita parpadeo en el caso común de <10ms).
4. **Footer:**
   - Barra de ayuda con atajos de teclado disponibles.
   - **[NUEVO]** Línea informativa de solo lectura: última sincronización relativa (`Última sync: hace 2h`) y conteo de transacciones locales pendientes de `cash sync`. Esta línea **no dispara ninguna acción de red**; es puramente informativa, leída del estado local del `GitService` (por ejemplo, comparando `HEAD` local vs. `origin/HEAD` en caché, sin hacer fetch).

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

*(Sin cambios respecto a v1: no se agregan atajos de filtrado, búsqueda ni edición — se evaluó y se descartó para mantener el alcance de solo lectura sin ambigüedad.)*

---

## 6. Closed Decisions Made
1. **DEC-TUI-01:** Creación del subcomando `cash tui` (en lugar de sobrecargar el comando base `cash`).
2. **DEC-TUI-02:** Modo de **solo lectura**. La creación de datos se mantiene en `cash in`/`cash out` para preservar la regla de sub-10ms en terminal.
3. **DEC-TUI-03:** Navegación mensual dinámica mediante recálculo reactivo en memoria con el repositorio local.
4. **DEC-TUI-04:** Stack Bubble Tea + Lip Gloss como estándar de facto en Go.
5. **[NUEVO] DEC-TUI-05:** Todo indicador de estado financiero (ingreso/gasto/balance) lleva símbolo (`▲`/`▼`) además de color, por accesibilidad para usuarios con daltonismo.
6. **[NUEVO] DEC-TUI-06:** Se define un breakpoint de ancho de terminal (80 columnas) que alterna entre layout estándar (tarjetas en fila) y layout compacto (tarjetas apiladas, columna Descripción oculta). El umbral se recalcula en cada `WindowSizeMsg` de Bubble Tea.
7. **[NUEVO] DEC-TUI-07:** El indicador de sincronización en el footer es de **solo lectura de estado local** (sin fetch de red) para no introducir latencia ni tráfico no solicitado por el usuario; refleja el último estado conocido tras la última corrida de `cash sync`.
8. **[NUEVO] DEC-TUI-08:** La tarjeta de "Top Categoría" reutiliza el cálculo ya expuesto por `GetMonthlySummary` — no se crea lógica de dominio nueva ni un puerto adicional.

---

## 7. Needs Assumption / Gaps to Confirm
1. **Logo ASCII:** `dollar-sign.md` se encuentra vacío actualmente. Se utilizará un logo tipográfico estilizado por defecto a menos que el usuario pegue su arte ASCII en dicho archivo.
2. **[NUEVO] Formato de "última sync":** se asume formato relativo humano (`hace 2h`, `hace 3d`) en vez de timestamp absoluto, salvo que el usuario indique preferencia contraria.
3. **[NUEVO] Umbral de "carga perceptible":** se asume 150ms como umbral para mostrar el estado transitorio "Cargando...", ajustable si el equipo de implementación mide latencias reales distintas en Termux vs. desktop.

---

## 8. Out of Scope (v1.1 TUI)
1. Formularios de captura de transacciones dentro de la TUI (se mantiene en CLI).
2. Gráficos de barras o torta ASCII, incluyendo sparklines con caracteres de bloque — reservados para Fase Web o v1.2. **Confirmado explícitamente en esta revisión: no se reconsidera para v1.1 pese a mejorar el contexto visual.**
3. Soporte para clics del mouse (interacción 100% por teclado).
4. Edición de registros dentro de la tabla.
5. **[NUEVO]** Filtrado o búsqueda interactiva por categoría dentro de la TUI — se evaluó como posible mejora de UX, pero se descarta de v1.1 para no expandir el alcance de "solo lectura pasiva" sin una decisión explícita del usuario; queda como candidato para v1.2.
6. **[NUEVO]** Cualquier acción que dispare `git fetch`/`pull`/`push` desde dentro de la TUI — el indicador de sync es únicamente de lectura de estado local.
