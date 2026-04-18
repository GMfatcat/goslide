# GoSlide Theme Expansion — 13 Curated Theme Ideas

> Research completed 2026-04-17. For review and discussion.

## Summary Table

| # | Name | Category | Background | Default Accent | New Fonts? | Priority |
|---|---|---|---|---|---|---|
| 1 | `nord-light` | Light | `#ECEFF4` | teal | None | - |
| 2 | `solarized` | Light | `#FDF6E3` | teal | Optional serif | Top 5 |
| 3 | `paper` | Light | `#F7F3EE` | amber | Optional serif | - |
| 4 | `catppuccin-latte` | Light | `#EFF1F5` | pink | None | - |
| 5 | `chalk` | Light | `#E8EAF0` | purple | None | - |
| 6 | `dracula` | Dark | `#282A36` | pink | None | Top 5 |
| 7 | `catppuccin-mocha` | Dark | `#1E1E2E` | pink | None | Top 5 |
| 8 | `gruvbox` | Dark | `#282828` | amber | None | Top 5 |
| 9 | `midnight` | Dark | `#0D1117` | blue | None | Top 5 |
| 10 | `synthwave` | Dark | `#241734` | pink | Optional: Orbitron | - |
| 11 | `forest` | Specialty | `#1C2B1A` | green | None | - |
| 12 | `rose` | Light/Spec | `#FDF0F0` | pink | Optional: Nunito | - |
| 13 | `amoled` | Specialty | `#000000` | blue | None | - |

## Top 5 Recommendations (zero new dependencies)

1. **`dracula`** — instant recognition from developer audiences
2. **`midnight`** — fills "premium dark executive" gap
3. **`gruvbox`** — retro-warm, distinct from all existing themes
4. **`solarized`** — timeless, widely recognized, "academic light" niche
5. **`catppuccin-mocha`** — trendiest dark theme in developer circles

---

## LIGHT THEMES

### 1. `nord-light`

**Vibe:** Arctic daylight — cool blue-grays and slate whites.

**Target:** Technical docs, open-source talks, dev conferences.

| Token | Hex |
|---|---|
| `--slide-bg` | `#ECEFF4` |
| `--slide-text` | `#2E3440` |
| `--slide-heading` | `#2E3440` |
| `--slide-code-bg` | `#D8DEE9` |
| `--slide-code-text` | `#2E3440` |
| `--slide-border` | `rgba(46, 52, 64, 0.12)` |
| `--slide-muted` | `#4C566A` |
| `--slide-card-bg` | `#E5E9F0` |

**Default accent:** teal. **Font:** Noto Sans TC (bundled). **Inspiration:** Nord design system.

---

### 2. `solarized`

**Vibe:** Iconic warm-cream — easy on eyes, academic.

**Target:** Academic lectures, classroom presentations, long-form technical talks.

| Token | Hex |
|---|---|
| `--slide-bg` | `#FDF6E3` |
| `--slide-text` | `#657B83` |
| `--slide-heading` | `#073642` |
| `--slide-code-bg` | `#EEE8D5` |
| `--slide-code-text` | `#586E75` |
| `--slide-border` | `rgba(88, 110, 117, 0.18)` |
| `--slide-muted` | `#93A1A1` |
| `--slide-card-bg` | `#EEE8D5` |

**Default accent:** teal. **Font:** Noto Sans TC (optional serif future). **Inspiration:** Ethan Schoonover's Solarized.

---

### 3. `paper`

**Vibe:** Warm off-white like quality writing paper — editorial, literary.

**Target:** Humanities, writing workshops, storytelling, design critiques.

| Token | Hex |
|---|---|
| `--slide-bg` | `#F7F3EE` |
| `--slide-text` | `#3C3530` |
| `--slide-heading` | `#1E1A17` |
| `--slide-code-bg` | `#EDE8E0` |
| `--slide-code-text` | `#3C3530` |
| `--slide-border` | `rgba(60, 53, 48, 0.12)` |
| `--slide-muted` | `#8C7B6E` |
| `--slide-card-bg` | `#EDE8E0` |

**Default accent:** amber. **Font:** Optional serif (Lora) for headings. **Inspiration:** Bear.app, iA Writer.

---

### 4. `catppuccin-latte`

**Vibe:** Pastel sunrise — soft, warm pastels, modern.

**Target:** Indie hacker demos, creative coding, approachable dev talks.

| Token | Hex |
|---|---|
| `--slide-bg` | `#EFF1F5` |
| `--slide-text` | `#4C4F69` |
| `--slide-heading` | `#1E2030` |
| `--slide-code-bg` | `#E6E9EF` |
| `--slide-code-text` | `#4C4F69` |
| `--slide-border` | `rgba(76, 79, 105, 0.15)` |
| `--slide-muted` | `#9CA0B0` |
| `--slide-card-bg` | `#DCE0E8` |

**Default accent:** pink. **Font:** Noto Sans TC (bundled). **Inspiration:** Catppuccin Latte palette.

---

### 5. `chalk`

**Vibe:** Light grey-blue chalkboard — modern educational.

**Target:** Educational content, science talks, university lectures.

| Token | Hex |
|---|---|
| `--slide-bg` | `#E8EAF0` |
| `--slide-text` | `#2D3142` |
| `--slide-heading` | `#1B1F36` |
| `--slide-code-bg` | `#D8DBE6` |
| `--slide-code-text` | `#2D3142` |
| `--slide-border` | `rgba(45, 49, 66, 0.14)` |
| `--slide-muted` | `#7A7F9A` |
| `--slide-card-bg` | `#D8DBE6` |

**Default accent:** purple. **Font:** Noto Sans TC (bundled). **Inspiration:** Tailwind slate, Material Blue Grey.

---

## DARK THEMES

### 6. `dracula`

**Vibe:** Classic vampire palette — deep purple-blacks, neon highlights.

**Target:** Dev conferences, live coding, security talks, CTF writeups.

| Token | Hex |
|---|---|
| `--slide-bg` | `#282A36` |
| `--slide-text` | `#F8F8F2` |
| `--slide-heading` | `#F8F8F2` |
| `--slide-code-bg` | `#1E1F29` |
| `--slide-code-text` | `#F8F8F2` |
| `--slide-border` | `rgba(98, 114, 164, 0.3)` |
| `--slide-muted` | `#6272A4` |
| `--slide-card-bg` | `#313442` |

**Default accent:** pink. **Font:** Noto Sans TC (bundled). **Inspiration:** Dracula Theme by Zeno Rocha.

---

### 7. `catppuccin-mocha`

**Vibe:** Warm-dark, cozy — deep coffee browns, soft pastels.

**Target:** Developer talks, indie demos, "grown-up Dracula".

| Token | Hex |
|---|---|
| `--slide-bg` | `#1E1E2E` |
| `--slide-text` | `#CDD6F4` |
| `--slide-heading` | `#CDD6F4` |
| `--slide-code-bg` | `#181825` |
| `--slide-code-text` | `#CDD6F4` |
| `--slide-border` | `rgba(49, 50, 68, 0.8)` |
| `--slide-muted` | `#585B70` |
| `--slide-card-bg` | `#313244` |

**Default accent:** pink. **Font:** Noto Sans TC (bundled). **Inspiration:** Catppuccin Mocha palette.

---

### 8. `gruvbox`

**Vibe:** Retro-warm dark — earthy browns, mustards, forest greens.

**Target:** Systems programming, Vim users, retro computing talks.

| Token | Hex |
|---|---|
| `--slide-bg` | `#282828` |
| `--slide-text` | `#EBDBB2` |
| `--slide-heading` | `#FBF1C7` |
| `--slide-code-bg` | `#1D2021` |
| `--slide-code-text` | `#EBDBB2` |
| `--slide-border` | `rgba(168, 153, 132, 0.2)` |
| `--slide-muted` | `#928374` |
| `--slide-card-bg` | `#3C3836` |

**Default accent:** amber. **Font:** JetBrains Mono optional for all text. **Inspiration:** Gruvbox by Pavel Pertsev.

---

### 9. `midnight`

**Vibe:** Deep navy professional — premium, polished, Apple Keynote dark.

**Target:** Executive presentations, investor pitches, product launches.

| Token | Hex |
|---|---|
| `--slide-bg` | `#0D1117` |
| `--slide-text` | `#E6EDF3` |
| `--slide-heading` | `#FFFFFF` |
| `--slide-code-bg` | `#161B22` |
| `--slide-code-text` | `#E6EDF3` |
| `--slide-border` | `rgba(48, 54, 61, 0.8)` |
| `--slide-muted` | `#8B949E` |
| `--slide-card-bg` | `#161B22` |

**Default accent:** blue. **Font:** Noto Sans TC (bundled). **Inspiration:** GitHub Dark mode.

---

### 10. `synthwave`

**Vibe:** Retrowave/vaporwave — deep purple, hot pink and cyan neons.

**Target:** Creative industry, game dev, music/art tech, visual personality talks.

| Token | Hex |
|---|---|
| `--slide-bg` | `#241734` |
| `--slide-text` | `#F0E6FF` |
| `--slide-heading` | `#FF71CE` |
| `--slide-code-bg` | `#1A0F27` |
| `--slide-code-text` | `#B2F7EF` |
| `--slide-border` | `rgba(255, 113, 206, 0.2)` |
| `--slide-muted` | `#9D6FD4` |
| `--slide-card-bg` | `#2E1F45` |

**Default accent:** pink. Heading glow: `text-shadow: 0 0 12px rgba(255, 113, 206, 0.4)`. **Font:** Optional Orbitron for headings. **Inspiration:** Synthwave'84 VS Code theme.

---

## SPECIALTY THEMES

### 11. `forest`

**Vibe:** Organic, earthy — deep forest greens and warm bark browns.

**Target:** Sustainability, environmental science, green-tech, NGO presentations.

| Token | Hex |
|---|---|
| `--slide-bg` | `#1C2B1A` |
| `--slide-text` | `#D4E8C2` |
| `--slide-heading` | `#E8F5D8` |
| `--slide-code-bg` | `#152213` |
| `--slide-code-text` | `#AECFA4` |
| `--slide-border` | `rgba(100, 160, 80, 0.2)` |
| `--slide-muted` | `#7A9E6E` |
| `--slide-card-bg` | `#243523` |

**Default accent:** green. **Font:** Noto Sans TC (bundled). **Inspiration:** Sustainability branding trends.

---

### 12. `rose`

**Vibe:** Warm blush-pink — delicate, inviting, "cute productivity".

**Target:** Product/UX design, marketing decks, brand strategy, community events.

| Token | Hex |
|---|---|
| `--slide-bg` | `#FDF0F0` |
| `--slide-text` | `#3D2B2B` |
| `--slide-heading` | `#2A1A1A` |
| `--slide-code-bg` | `#F5E0E0` |
| `--slide-code-text` | `#3D2B2B` |
| `--slide-border` | `rgba(200, 120, 120, 0.15)` |
| `--slide-muted` | `#C49090` |
| `--slide-card-bg` | `#F5E0E0` |

**Default accent:** pink. **Font:** Optional Nunito for softer look. **Inspiration:** Catppuccin Rosewater, Linear pink mode.

---

### 13. `amoled`

**Vibe:** True-black OLED — absolute zero background, maximum contrast.

**Target:** Mobile-first on OLED screens, dark room keynotes, accessibility (21:1 contrast).

| Token | Hex |
|---|---|
| `--slide-bg` | `#000000` |
| `--slide-text` | `#EEEEEE` |
| `--slide-heading` | `#FFFFFF` |
| `--slide-code-bg` | `#0A0A0A` |
| `--slide-code-text` | `#DDDDDD` |
| `--slide-border` | `rgba(255, 255, 255, 0.08)` |
| `--slide-muted` | `#666666` |
| `--slide-card-bg` | `#111111` |

**Default accent:** blue. **Font:** Noto Sans TC (bundled). **Inspiration:** Android AMOLED themes, GitHub Dark Dimmed.

---

## Observations

- **10 of 13 themes need zero new fonts** — works with existing Noto Sans TC + JetBrains Mono
- `pink` is the most popular default accent (5 themes) — reflects modern dev aesthetic trends
- Top 5 picks (dracula, midnight, gruvbox, solarized, catppuccin-mocha) are all zero-dependency and fill distinct niches
