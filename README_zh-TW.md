# 🎯 GoSlide

**Markdown 驅動的互動式簡報系統 — 單一執行檔、離線優先。**

GoSlide 將 `.md` 檔案轉換為 Reveal.js 簡報，支援即時圖表、流程圖、API 儀表板和互動控制元件。不需要 Node.js、不需要 npm — 只要一個 Go binary。

> 📖 [English README](README.md)

---

## ⚡ 快速開始

### 方式一：直接下載

從 [GitHub Releases](https://github.com/GMfatcat/goslide/releases) 下載對應平台的執行檔：

| 平台 | 檔案 |
|------|------|
| Windows (x64) | `goslide-windows-amd64.exe` |
| macOS (Intel) | `goslide-darwin-amd64` |
| macOS (Apple Silicon) | `goslide-darwin-arm64` |
| Linux (x64) | `goslide-linux-amd64` |
| Linux (ARM64) | `goslide-linux-arm64` |

重新命名為 `goslide`（Windows 為 `goslide.exe`），放到 PATH 中即可使用。

### 方式二：透過 Go 安裝

```bash
go install github.com/GMfatcat/goslide/cmd/goslide@latest
```

### 開始使用

```bash
# 建立簡報
goslide init

# 啟動即時預覽
goslide serve talk.md

# 匯出為獨立 HTML
goslide build talk.md
```

## ✨ 功能特色

| 功能 | 說明 |
|------|------|
| 📝 **Markdown 撰寫** | 用 `.md` 寫投影片 — frontmatter 設定、`---` 分頁 |
| 🎨 **22 種主題** | default、dark、corporate、minimal、hacker、dracula、midnight、gruvbox、solarized、catppuccin-mocha、ink-wash、instagram、western、pixel、nord-light、paper、catppuccin-latte、chalk、synthwave、forest、rose、amoled |
| 📊 **圖表** | 長條圖、折線圖、圓餅圖、雷達圖、迷你折線圖（Chart.js） |
| 🔀 **流程圖** | Mermaid.js 流程圖、序列圖、ER 圖 |
| 📋 **表格** | 可排序表格，點擊欄位標題排序 |
| 🎛️ **互動控制** | 標籤頁、滑桿、開關 + 響應式 `$variable` 綁定 |
| 🃏 **可展開卡片** | 格狀排列 + 點擊展開詳細內容 |
| 🌐 **API 儀表板** | 從後端 API 即時取得資料並自動重新整理 |
| 🔌 **API 代理** | 內建反向代理，支援認證標頭注入 |
| 📦 **單一執行檔** | 所有資源透過 `go:embed` 打包（約 8MB） |
| 🔄 **即時重載** | 編輯 `.md` → 瀏覽器自動重新整理，保持投影片位置 |
| 🖥️ **講者檢視** | 按 `S` 開啟計時器、筆記、下一張預覽 |
| 📤 **靜態匯出** | `goslide build` → 產出單一 `.html` 檔案，離線也能用 |
| 🏠 **主機模式** | 將整個目錄作為簡報資料庫提供服務 |
| 📡 **講者同步** | 觀眾可看到講者目前的投影片頁數 + 跳轉按鈕 |

## 🎨 主題

22 種內建主題 × 8 種強調色 = **176 種視覺組合**。

```yaml
---
theme: dracula
accent: pink
---
```

👉 [完整主題目錄](docs/THEMES.md)

## 📐 版面配置

12 種投影片版面，透過 HTML 註解設定：

```markdown
---
<!-- layout: two-column -->

# 標題

<!-- left -->
左欄內容

<!-- right -->
右欄內容
```

可用版面：`default`、`title`、`section`、`two-column`、`code-preview`、`three-column`、`image-left`、`image-right`、`quote`、`split-heading`、`top-bottom`、`grid-cards`、`blank`

## 📦 元件

透過圍欄程式碼區塊嵌入互動元件：

```markdown
~~~chart:bar
title: 營收
labels: ["Q1", "Q2", "Q3"]
data: [100, 150, 200]
color: teal
~~~
```

👉 [完整元件參考](docs/COMPONENTS.md)

## ⌨️ CLI 指令

```bash
goslide serve <file.md>     # 啟動即時預覽
goslide host <directory>    # 主機模式（多簡報）
goslide build <file.md>     # 匯出為獨立 HTML
goslide init                # 建立新簡報範本
goslide list [directory]    # 列出目錄中的簡報
```

👉 [完整 CLI 參考](docs/CLI.md)

## ⚙️ 設定檔

在 `.md` 同目錄下建立 `goslide.yaml`（選用）：

```yaml
# API 代理 — 將瀏覽器請求透過 Go server 轉發到上游 API
api:
  proxy:
    /api/backend:
      target: http://localhost:8000
      headers:
        Authorization: "Bearer ${API_TOKEN}"

# 自訂主題覆蓋
theme:
  overrides:
    slide-bg: "#1e1e2e"
    slide-accent: "#f38ba8"
```

> **注意：** 當 `goslide.yaml` 設定了 proxy，GoSlide 會在每次代理請求時嘗試連線上游。如果上游服務未啟動，console 會出現 `proxy error` 且瀏覽器顯示 502 — 這只影響 API 元件的投影片，不影響其他內容。

### 🧪 使用 Mock Server 測試 API 元件

GoSlide 附帶一個 mock API server 用於測試 API 驅動的投影片：

```bash
# Terminal 1：啟動 mock API server
go run examples/mock-api/main.go
# → Mock API running on http://localhost:9999

# Terminal 2：複製範例設定檔並啟動
cp examples/goslide.yaml.example examples/goslide.yaml
go run ./cmd/goslide serve examples/demo.md --no-open
# → 開啟 http://localhost:3000，翻到 API Dashboard 投影片
```

範例設定檔（`goslide.yaml.example`）將 `/api/mock` 代理到 `localhost:9999`。重新命名為 `goslide.yaml` 即可啟用。測試完成後可移除或改名，避免 mock server 未啟動時出現 proxy error。

## 🏗️ 從原始碼建置

```bash
git clone https://github.com/GMfatcat/goslide.git
cd goslide
bash scripts/vendor.sh --update-checksums
go build -o goslide ./cmd/goslide
```

**需求：** Go 1.21+

## 🎬 轉場效果

```yaml
---
transition: perspective  # 3D Y 軸旋轉
---
```

可用：`slide`（預設）、`fade`、`convex`、`concave`、`zoom`、`none`、`perspective`、`flip`

## 📝 講者筆記

```markdown
# 我的投影片

內容在此。

<!-- notes -->

講者筆記 — 按 S 鍵在講者檢視中可見。
```

## 🌐 CJK 支援

GoSlide 內建 Noto Sans TC 字型，完整支援繁體中文、日文漢字等 CJK 字元。

## 🙏 致謝

GoSlide 建立在這些優秀的開源專案之上：

| 專案 | 用途 | 授權 |
|------|------|------|
| [Reveal.js](https://revealjs.com/) | 投影片渲染引擎 | MIT |
| [Chart.js](https://www.chartjs.org/) | 圖表（長條、折線、圓餅、雷達、迷你折線） | MIT |
| [Mermaid](https://mermaid.js.org/) | 流程圖、序列圖、ER 圖 | MIT |
| [goldmark](https://github.com/yuin/goldmark) | Markdown 解析器 | MIT |
| [cobra](https://github.com/spf13/cobra) | CLI 框架 | Apache-2.0 |
| [fsnotify](https://github.com/fsnotify/fsnotify) | 檔案系統監控 | BSD-3 |
| [coder/websocket](https://github.com/coder/websocket) | WebSocket 函式庫 | MIT |
| [Noto Sans TC](https://fonts.google.com/noto/specimen/Noto+Sans+TC) | CJK 字型 | OFL-1.1 |
| [JetBrains Mono](https://www.jetbrains.com/lp/mono/) | 等寬字型 | OFL-1.1 |
| [Press Start 2P](https://fonts.google.com/specimen/Press+Start+2P) | 像素字型（pixel 主題） | OFL-1.1 |
| [Rye](https://fonts.google.com/specimen/Rye) | 西部字型（western 主題） | OFL-1.1 |

主題配色靈感來自：[Dracula](https://draculatheme.com/)、[Catppuccin](https://github.com/catppuccin/catppuccin)、[Gruvbox](https://github.com/morhetz/gruvbox)、[Solarized](https://ethanschoonover.com/solarized/)、[Nord](https://www.nordtheme.com/)。

## 📄 授權

MIT
