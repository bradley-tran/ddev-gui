# Features

## Titlebar & Menubar

A custom frameless titlebar replaces the native OS chrome. It is drag-to-move and double-click-to-maximise.

**Left side - Menubar:**

| Menu | Items |
|------|-------|
| **Projects** | New Project, Refresh (`F5`), Stop All |
| **View** | Toggle Log (`Ctrl+L`), Settings |
| **Help** | About, Check Environment |

**Center** - displays `Home` on the project list, or `<name> | <type>` when viewing a project detail.

**Right side - quick-access buttons:**
- **Browser toggle** - switches project links between system browser and embedded window (persisted to config)
- **Log toggle** - show/hide the log panel
- **Window controls** - minimize, maximize/restore, close

## Project List

Displays all DDEV projects detected by `ddev list -j`.

**Views:**
- **List view** - sortable table with Name, Status, Type, URL, and per-row action buttons
- **Grid view** - card layout with project type logo, status badge, and quick-action buttons

**Per-project actions (both views):** Start / Stop, Restart, Open Folder, Open CLI

**States:**
- Loading spinner while the initial `ddev list` runs
- Empty state with hint when no projects exist
- Error state with Retry and "Check Environment" shortcuts when loading fails

**Auto-refresh:** The list polls periodically and re-renders on project changes without resetting the view.

## Project Detail

A two-panel layout (toolbar/URL bar + sidebar navigation) for drilling into a single project.

### Toolbar

- **Start / Stop / Restart** - runs the corresponding DDEV command and auto-refreshes status
- **Open CLI** - opens an external terminal in the project root
- **Init Site** - runs the site initialiser for supported project types (shown only if not yet initialised)
- **Drupal** dropdown *(Drupal projects only)*:
  - Export DB - dumps the database to a local file (WSL backend only)
  - Import DB - imports a `.sql` / `.sql.gz` / `.tgz` file, replaces the current DB
  - Masquerade - log in as any Drupal user by UID or name (opens a user-picker modal)
  - Clear Cache - runs `drush cr`
- **More** dropdown - Open Folder, Open in Editor, Re-init Site, Delete Project

The "Open in Editor" action launches the configured editor with the project root as the working directory. Supported editors include:
1. VS Code
2. PHP Storm
3. Neovim
4. Sublime Text
5. Antigravity

### URL Bar

Quick-access pills below the toolbar:
- **Back** - returns to the project list
- **Open Site** - opens the primary HTTPS URL
- **Mailpit** - opens the Mailpit mail-catcher URL (shown when available)
- **Open Site (Admin)** - runs `drush uli` and opens the one-time login link *(Drupal only)*

### Navigation Sidebar

Vertical icon tabs on the left:

| Tab | Content |
|-----|---------|
| **Overview** | Key-value metadata: name, status, type, docroot, location, router, Node.js version; plus a Services section showing each container's status, URLs, and ports |
| **Add-on** | Installed add-ons table + add-on marketplace picker |
| **Snapshots** | Database snapshot management (create, restore, delete) |
| **Files** | Integrated file explorer (see below) |
| **Logs** | Project logs
| **Terminal** | Embedded interactive terminal (see below) |

## File Explorer

A split-panel file browser inside the **Files** tab:

| Panel | Width | Purpose |
|-------|-------|---------|
| File tree | 1/4 | Browse directories and select files |
| Preview | 3/4 | Format-aware file preview |

**Capabilities:**
- Breadcrumb navigation with clickable path segments
- Lazy directory loading - subdirectories load on click
- Directories displayed above files, both sorted alphabetically
- File and folder icons for quick visual scanning
- `..` row to navigate up one level
- Selected file highlighted in the tree
- Stale file-load responses discarded on fast switching (request-ID guard)

**Performance:**
- **Dual persistent WSL shells** - `d.shell` for directory listings, `d.fileShell` for file reads (separate mutexes, no cross-blocking)
- All file content piped through `base64` for safe transport
- Directory listings cached with 2-minute TTL (`dirCache`)
- File content cached with 5-minute TTL (`fileCache`)

**Backend API:**
- `ListDir(project, relPath)` - lists files/directories using `find -printf`, returns JSON array of `FileEntry`
- `ReadFile(project, relPath)` - reads file via `head | base64`, decodes in Go, returns text string (capped at 1 MB)
- `ReadFileBase64(project, relPath)` - reads binary file via `head | base64`, returns raw base64 string (capped at 5 MB)

## File Preview

A routing component (`FilePreview`) that delegates to the correct viewer based on file type:

| File type | Viewer | Detection |
|-----------|--------|-----------|
| `.md`, `.markdown`, `.mdx` | MarkdownViewer | Extension match |
| `.png`, `.jpg`, `.gif`, `.svg`, `.webp`, etc. | ImageViewer | Extension match |
| `.js`, `.ts`, `.go`, `.php`, `.css`, `.json`, etc. | CodeViewer | Extension match (50+ extensions) |
| `.zip`, `.exe`, `.pdf`, `.mp4`, `.woff`, etc. | "Binary file" message | Extension match (no backend call) |
| Everything else | Plain text `<pre>` | Fallback |

## Image Viewer

Displays base64-encoded images with interactive controls:

- **MIME detection** from extension (png, jpg, gif, svg, webp, ico, bmp, avif)
- **Zoom controls** in the preview header bar: zoom in/out buttons (25%–400%), clickable percentage to reset
- **Drag-to-pan** when zoomed beyond 100% (grab/grabbing cursor, mouse events)
- **Checkered transparency background** for images with alpha channels

## Code Viewer

Renders source code with syntax highlighting:

- **highlight.js** with `atom-one-dark` theme for token colouring
- **Line numbers** gutter on the left
- **Language badge** in the top-right corner (auto-detected from extension)
- **50+ extensions** supported including `Dockerfile`, `Makefile`, `.gitignore` (name-based detection)

## Markdown Viewer

Renders markdown strings as styled HTML using **marked** (GFM) and **highlight.js** (fenced code blocks).

**Supported elements:**
- Headings (h1–h6), fenced code blocks with syntax highlighting, inline code
- Blockquotes, tables, links (external), images, lists, horizontal rules

## Embedded Terminal

An interactive shell panel inside the **Terminal** tab of the project detail view.

- Commands run inside the DDEV container via `DdevService.execCommand`
- **Streaming output** - lines appear in real-time via Wails `Runtime.on` events (`terminal:output:<project>`, `terminal:done:<project>`)
- **ANSI colour rendering** - `ansiToHtml()` converts escape codes to `<span style="color:…">` (One Dark palette, 16 colours)
- **Command history** - `↑`/`↓` arrow keys cycle through the last 100 commands (de-duplicated)
- **Built-in `clear` / `cls`** - resets the output buffer without sending a command to the shell
- **Auto-scroll** - output scrolls to the bottom unless the user has manually scrolled up
- **Spinner indicator** while a command is running; input is disabled until completion
- Exit code ≠ 0 is displayed as an error line

## Settings Modal

Persistent configuration stored in the app's JSON config file via `ConfigService`.

| Setting | Options | Notes |
|---------|---------|-------|
| **Language** | English, 简体中文, Tiếng Việt | See *Multilingual Support* below |
| **Open links in** | System browser / Embedded window | Also toggleable from the titlebar |
| **Preferred Editor** | VS Code / PHP Storm / Neovim / Sublime Text / Antigravity | Used by the project "Open in Editor" action |
| **Theme** | Default, Windows Acrylic, Windows Tabbed | Windows only; requires app restart |
| **Backend** | WSL, SSH *(dev mode)*, Local *(dev mode)* | Controls how DDEV commands are executed |
| **WSL Distribution** | Dropdown of detected distros | Shown when backend = WSL |
| **SSH Host / Port / User / Key** | Text fields | Shown when backend = SSH (dev mode) |

## Multilingual Support (i18n)

Zero-dependency, PO-file-based internationalisation system.

**Locales:** English (`en`), Simplified Chinese (`zh`), Vietnamese (`vi`)

**Architecture:**
- Translation strings stored in standard **gettext PO files** (`src/lib/locales/*.po`), editable with tools like Poedit or Weblate
- Lightweight **PO parser** built into `i18n.tsx` (~40 lines) - no external library
- English PO is **statically bundled** (zero network cost on first load)
- Non-English locales are **dynamically imported** via Vite `?raw` code-splitting - loaded on demand
- `{placeholder}` interpolation supported (e.g. `t('detail.delete.confirm', { name: projectName })`)
- **Automatic fallback** to English if a key is missing in the active locale

**React integration:**
- `I18nProvider` wraps the entire app and injects the current locale + message map via React Context
- `useTranslation()` hook exposes the `t()` function to any component
- Language preference is persisted to the app config (`locale` field) and restored on next launch

**Translated surfaces:** all menus, modals, form labels, table headers, tooltips, confirmations, toasts, and log messages.

## Environment Info Modal

Displays the current DDEV version detected from the active backend. If DDEV is not found:

- Shows a descriptive error (distinguishes WSL/pipe errors from missing-binary errors)
- Provides an **Install DDEV** button that streams installation progress line-by-line
- For WSL configuration errors, shows an **Open Settings** shortcut instead

Also includes a **Developer Mode** toggle that unlocks experimental features (SSH backend, Local backend).

## About Modal

Displays app version (from the Wails runtime), DDEV version, and the build stack (Wails v2, React 19, TypeScript).

## Toast Notifications

A non-blocking notification system (`Toast.tsx`, `showToast` helper in the app store):

- Stacked toasts in the bottom-right corner
- Three types: `success`, `error`, `info` (distinct colours)
- Auto-dismiss after a configurable duration (default 4 s)
- Max 200 log entries kept in memory; duplicates adjacent to each other are de-duplicated

## Log Panel

A collapsible panel at the bottom of the screen showing structured app activity:

- **Levels:** `info`, `success`, `error`, `output` (each with distinct colour coding)
- **Timestamps** (HH:MM:SS) on every entry
- **Clear** button to flush all entries
- Toggled via the titlebar button or `View → Toggle Log` / `Ctrl+L`

## State Management

Global state via React Context + `useReducer` (`store.tsx`). No external state library.

**State shape:**

| Field | Type | Purpose |
|-------|------|---------|
| `projectsJSON` | `string` | Raw JSON from last `ddev list` call |
| `config` | `AppConfig` | Persisted user preferences |
| `logEntries` | `LogEntry[]` | Capped at 200, de-duplicated |
| `toasts` | `ToastEntry[]` | Active toast queue |
| `currentView` | `'list' \| 'detail'` | Navigation state |
| `selectedProject` | `string \| null` | Active project name |
| `terminalActive` | `boolean` | Whether terminal tab is mounted |

**Actions:** `SET_PROJECTS_JSON`, `SET_CONFIG`, `PATCH_CONFIG`, `ADD_LOG`, `CLEAR_LOG`, `ADD_TOAST`, `REMOVE_TOAST`, `NAVIGATE_DETAIL`, `NAVIGATE_LIST`, `SET_VIEW_MODE`, `SET_TERMINAL_ACTIVE`

## Cache Utility

A reusable in-memory cache module (`lib/cache.ts`) with TTL-based expiry:

- **`Cache<T>`** - generic cache class with `get`, `set`, `has`, `delete`, `clear`
- **`dirCache`** - pre-built instance for directory listings (2-minute TTL)
- **`fileCache`** - pre-built instance for file/markdown content (5-minute TTL)
- **`cacheKey(project, path)`** - builds namespaced cache keys

## Add-on Filter Utility

`lib/addonFilter.ts` - fuzzy search for the add-on marketplace picker:

- Tokenises the search query on whitespace
- Matches against repository name segments and description text
- Returns a filtered + ranked list of available add-ons
