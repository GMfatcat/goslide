# ⌨️ GoSlide CLI Reference

## Commands

### `serve`

Serve a single Markdown file as a presentation.

```bash
goslide serve <file.md> [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--port` | `-p` | `3000` | Port number |
| `--theme` | `-t` | | Override theme |
| `--accent` | `-a` | | Override accent color |
| `--no-open` | | `false` | Don't auto-open browser |
| `--no-watch` | | `false` | Disable live reload |

### `host`

Serve a directory of presentations with an index page.

```bash
goslide host <directory> [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--port` | `-p` | `8080` | Port number |
| `--no-open` | | `false` | Don't auto-open browser |

Index page at `/`. Presentations at `/talks/{filename}`.

### `build`

Export a presentation as a self-contained HTML file.

```bash
goslide build <file.md> [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `{name}.html` | Output file path |
| `--theme` | `-t` | | Override theme |
| `--accent` | `-a` | | Override accent color |

Output is a single `.html` file (~6-7MB) with all assets inlined. Works offline.

### `init`

Scaffold a new presentation with template content.

```bash
goslide init [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--template` | `-t` | `basic` | Template: `basic`, `demo`, `corporate` |

Creates `talk.md` in the current directory.

### `list`

List presentations in a directory.

```bash
goslide list [directory]
```

Outputs a table with filename, title, theme, and slide count.

## Global Flags

| Flag | Description |
|------|-------------|
| `--verbose` | Verbose logging |
| `--version` | Print version |
| `--help` | Print help |

## Keyboard Shortcuts (in browser)

| Key | Action |
|-----|--------|
| `→` / `Space` / `Enter` | Next slide |
| `←` / `Backspace` | Previous slide |
| `Esc` | Toggle overview / close overlay |
| `F` | Fullscreen |
| `S` | Speaker view |
| `Home` | First slide |
| `End` | Last slide |
