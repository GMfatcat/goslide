# CLAUDE.md


## Shell Command Rules

On Windows, chained commands (`cd xxx && git add`) trigger extra permission prompts. Always use **separate Bash calls** for:
- `git add`, `git commit` — never chain with `cd`
- `npm install`, `npm run build` — use `cd` only if absolutely needed, prefer separate call
- `go test`, `go build` — same rule

Use `-C` flag for Go commands when possible (e.g., `go build -C lint/`, `go test -C viz/`) to avoid `cd`.
