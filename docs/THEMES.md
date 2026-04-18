# 🎨 GoSlide Theme Catalog

GoSlide ships with **14 built-in themes** across 4 categories. Set via frontmatter `theme:` or CLI `--theme`.

## Basic

| Theme | Background | Text | Best For |
|-------|-----------|------|----------|
| `default` | White `#FFFFFF` | Dark `#1A1A1A` | Daily internal presentations |
| `dark` | Dark blue `#1A1A2E` | Light `#E0E0E0` | Technical demos, code-heavy talks |

## Professional

| Theme | Background | Text | Best For |
|-------|-----------|------|----------|
| `corporate` | Warm grey `#F5F5F0` | Dark `#2D2D2D` | Formal reports to management |
| `minimal` | Pure white `#FFFFFF` | Soft dark `#333333` | Keynote-style, one idea per slide |
| `hacker` | Black `#0A0A0A` | Green `#00FF00` | Live coding, terminal aesthetic |

## Developer

| Theme | Background | Text | Default Accent | Inspiration |
|-------|-----------|------|----------------|-------------|
| `dracula` | Purple-black `#282A36` | White `#F8F8F2` | pink | [Dracula Theme](https://draculatheme.com/) |
| `midnight` | Deep navy `#0D1117` | Light `#E6EDF3` | blue | GitHub Dark |
| `gruvbox` | Warm dark `#282828` | Cream `#EBDBB2` | amber | [Gruvbox](https://github.com/morhetz/gruvbox) |
| `solarized` | Cream `#FDF6E3` | Slate `#657B83` | teal | [Solarized](https://ethanschoonover.com/solarized/) |
| `catppuccin-mocha` | Dark coffee `#1E1E2E` | Lavender `#CDD6F4` | pink | [Catppuccin](https://github.com/catppuccin/catppuccin) |

## Creative

| Theme | Background | Text | Font | Best For |
|-------|-----------|------|------|----------|
| `ink-wash` | Rice paper `#F5F0EB` | Ink `#2C2C2C` | Noto Sans TC | Chinese ink painting aesthetic |
| `instagram` | Gradient (pink→purple→blue) | Dark `#262626` | Noto Sans TC | Marketing, social media style |
| `western` | Leather brown `#2C1810` | Parchment `#D4B896` | Rye (headings) | Western/cowboy aesthetic |
| `pixel` | Retro dark `#0F0F23` | Green `#00CC00` | Press Start 2P | Retro gaming, 8-bit style |

## Accent Colors

All themes support 8 accent colors via `accent:` in frontmatter:

| Accent | Hex | Preview |
|--------|-----|---------|
| `blue` | `#3B82F6` | 🔵 |
| `teal` | `#14B8A6` | 🟢 |
| `purple` | `#A855F7` | 🟣 |
| `coral` | `#F87171` | 🔴 |
| `amber` | `#F59E0B` | 🟡 |
| `green` | `#22C55E` | 🟢 |
| `red` | `#EF4444` | 🔴 |
| `pink` | `#EC4899` | 🩷 |

## Custom Overrides

Override any CSS variable via `goslide.yaml`:

```yaml
theme:
  overrides:
    slide-bg: "#1e1e2e"
    slide-accent: "#f38ba8"
    slide-heading: "#cdd6f4"
```

Available variables: `slide-bg`, `slide-text`, `slide-heading`, `slide-code-bg`, `slide-code-text`, `slide-border`, `slide-muted`, `slide-card-bg`.
