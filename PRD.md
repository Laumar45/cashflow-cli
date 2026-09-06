# PRD: CashFlow CLI & Web Dashboard
**Versión:** 1.0
**Estado:** Planificación / MVP
**Autor:** Laumar

---

## 1. Resumen Ejecutivo y Visión del Producto
**CashFlow** es una herramienta de finanzas personales minimalista y "offline-first" diseñada para desarrolladores y usuarios avanzados. 
El producto se divide en dos fases:
1. **Fase 1 (MVP):** Un CLI (Interfaz de Línea de Comandos) ultrarrápido para registrar ingresos y gastos, sincronizado entre dispositivos (móvil/PC) de forma gratuita utilizando Git como backend.
2. **Fase 2 (Futuro):** Una Web GUI (Interfaz Gráfica Web) que consume los datos del repositorio Git para proporcionar visualizaciones, gráficos y agrupación inteligente de categorías.

**Filosofía de Diseño:** *Captura rápida en la terminal, análisis profundo en la web. Cero costos de infraestructura.*

---

## 2. Arquitectura de Datos y Sincronización

### 2.1. Almacenamiento: JSON Lines (`.jsonl`)
Para permitir la sincronización mediante Git sin conflictos de merge (merge conflicts), los datos NO se guardarán en SQLite ni en un JSON tradicional (array). Se usará **JSON Lines**, donde cada transacción es una línea independiente.

**Ruta del archivo:** `~/.cashflow/transactions.jsonl`

**Esquema de una transacción:**
```json
{
  "id": "1698412345678-a1b2c3",
  "timestamp": "2023-10-27T14:30:00Z",
  "type": "expense",
  "category": "gaseosa",
  "amount": 2.50,
  "description": "Para la reunion"
}
```
_Nota: `id` debe ser único (timestamp + string aleatorio). `category` siempre en minúsculas y sin espacios (usar guiones si es necesario, ej: `comida-rapida`)._

### 2.2. Sincronización: Git como Backend

El repositorio local de Git actuará como base de datos distribuida.

- **Repositorio Remoto:** GitHub/GitLab/Codeberg (Repositorio Privado).
- **Flujo de Sync:** `git pull --rebase` -> `git add .` -> `git commit -m "sync"` -> `git push`.
- **Manejo de Conflictos:** Al ser `.jsonl`, Git puede hacer _auto-merge_ de líneas añadidas al final del archivo. Si hay un conflicto raro, el CLI debe tener un comando para limpiar marcas de conflicto de Git automáticamente.

---

## 3. Fase 1: CLI (Minimum Viable Product)

### 3.1. Comandos y Funcionalidades

El binario/ejecutable se llamará `cash`.

| Comando           | Descripción                                                                  | Ejemplo de uso                            |
| ----------------- | ---------------------------------------------------------------------------- | ----------------------------------------- |
| `cash in`         | Registra un ingreso.                                                         | `cash in sueldo 3000 "Quincena Octubre"`  |
| `cash out`        | Registra un gasto.                                                           | `cash out pasajes 2.5`                    |
| `cash sync`       | Sincroniza con el repositorio remoto (Git).                                  | `cash sync`                               |
| `cash summary`    | Muestra resumen del mes actual y balance total.                              | `cash summary`                            |
| `cash list`       | Lista transacciones con filtros opcionales.                                  | `cash list --month 10 --category pasajes` |
| `cash categories` | Muestra todas las categorías únicas usadas.                                  | `cash categories`                         |
| `cash commands`   | Lista todos los comandos que se pueden usar y descripcion de para que sirven | `cash commands`                           |
### 3.2. Reglas de la CLI (UX en Terminal)

1. **Cero Fricción:** Los comandos `in` y `out` deben ejecutarse en menos de 3 segundos.
2. **Categorías Dinámicas (Micro-categorías):** La CLI acepta _cualquier_ string como categoría. No hay validación de categorías predefinidas.
3. **Formato de Categoría:** Se recomienda usar una sola palabra o guiones bajos/medios (ej: `alcohol`, `comida-rapida`). La CLI no debe forzar comillas.
4. **Descripción Opcional:** Si se pasan más de 3 argumentos, el resto se concatena como descripción.

### 3.3. Stack Tecnológico Sugerido (CLI)

_Opción recomendada para facilitar distribución y uso en Termux/PC:_

- **Lenguaje:** Go (Golang) o Node.js (con `pkg` o `bun` para compilar a binario) o Python (con `Typer`).
- **Librería CLI:** `Cobra` (Go), `Commander.js` (Node), o `Typer` (Python).
- **Git:** Ejecución nativa de comandos del sistema operativo (`os/exec`, `child_process`, `subprocess`).

---

## 4. Fase 2: Web GUI (Dashboard)

### 4.1. Arquitectura "Zero-Cost Backend"

- **Frontend:** React, Vue o Svelte.
- **Hosting:** Vercel, Netlify o GitHub Pages (Gratis).
- **Data Fetching:** La web leerá el archivo `transactions.jsonl` directamente desde la API de GitHub (usando un Personal Access Token o haciendo el repo público pero ofuscado, o usando GitHub Pages para servir el raw file).

### 4.2. Concepto de Macro-Categorías (Agrupación)

Mientras la CLI usa _micro-categorías_ dinámicas (`gaseosa`, `cerveza`, `uber`, `bus`), la Web GUI agrupará estos datos en **Macro-Categorías** para los gráficos.

**Ejemplo de Mapeo en la Web:**

- **Comida y Bebidas:** `gaseosa`, `cerveza`, `almuerzo`, `cafe`, `comida-rapida`
- **Transporte:** `uber`, `pasajes`, `bus`, `gasolina`
- **Ropa:** `camisa`, `zapatos`, `ropa`
- **Alcohol / Fiesta:** `discoteca`, `cerveza` (si no se agrupó en bebidas)
- **Ingresos:** `sueldo`, `freelance`, `regalo`

_Nota: En la web, el usuario podrá crear estas reglas de agrupación (Alias) mediante una interfaz visual, sin tocar código._

### 4.3. Funcionalidades de la Web

1. **Dashboard General:** Balance actual, ingresos vs gastos del mes.
2. **Gráficos:**
    - Pie Chart: Gastos por Macro-categoría.
    - Bar Chart: Gastos por día de la semana / mes.
    - Line Chart: Evolución del balance histórico.
3. **Analítica Avanzada:** "Día que más gastas", "Categoría con mayor gasto mensual".
4. **Gestor de Reglas:** Interfaz para mapear micro-categorías (CLI) a macro-categorías (Web).

---

## 5. Requerimientos No Funcionales

1. **Costo:** $0.00. No se usarán bases de datos en la nube pagas (AWS, MongoDB Atlas), ni servidores VPS. Todo se basa en Git y hosting estático gratuito.
2. **Privacidad:** Los datos financieros son sensibles. El repositorio de Git **debe ser privado**. La Web GUI debe manejar el token de acceso a la API de GitHub de forma segura (ej: variable de entorno en Vercel, no hardcodeada en el frontend).
3. **Portabilidad:** El CLI debe poder compilarse o ejecutarse fácilmente en powershell/Linux (Laptop) y Android/Termux (Celular).
4. **Tolerancia a Fallos:** Si no hay internet, el CLI debe guardar los datos localmente sin fallar. El comando `cash sync` fallará gracefully y avisará al usuario que se sincronizará la próxima vez.

---

## 6. Roadmap y Milestones

### Milestone 1: Core CLI & Local Storage 

- Inicializar proyecto y configurar estructura de carpetas.
- Implementar lectura/escritura de archivo `.jsonl`.
- Crear comandos `cash in` y `cash out`.
- Implementar comando `cash summary (lectura y suma básica).

### Milestone 2: Sincronización Git 

- Crear repositorio remoto privado.
- Implementar comando `cash sync` (llamadas a Git).
- Probar flujo multi-dispositivo (Laptop <-> Termux).
- Manejar edge-cases (ej: qué pasa si no hay internet al hacer sync).

### Milestone 3: Pulido de CLI 

- Agregar colores a la terminal (verde=ingreso, rojo=gasto).
- Implementar comando `cash list` con filtros.
- Configurar alias en `.bashrc`/`.zshrc` para velocidad extrema.

### Milestone 4: Web GUI MVP 

- Subir `.jsonl` a GitHub.
- Crear proyecto frontend (React/Vue) y configurar hosting estático.
- Implementar fetch del archivo `.jsonl` desde la API de GitHub.
- Crear sistema de mapeo de Micro a Macro categorías.
- Integrar librería de gráficos (Chart.js / Recharts) y renderizar el primer Pie Chart.

***

### ¿Cómo usar este PRD?
1. Copia todo el bloque de código y guárdalo como `PRD.md` en la raíz de tu nuevo repositorio (ej: `cashflow-cli`).
2. Úsalo como tu "brújula". Si en medio del desarrollo se te ocurre una idea nueva (ej: "quiero agregar presupuestos"), **no la programes de inmediato**. Agrégala al PRD como una "Fase 3" o "Future Work" para no desviarte.
3. Cuando empieces a programar, ve tachando los `[ ]` del Roadmap.

¿Qué te parece la estructura? 