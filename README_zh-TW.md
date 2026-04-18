# 🎯 GoSlide

**Markdown 驅動的互動式簡報系統 — 單一執行檔、離線優先。**

GoSlide 將 `.md` 檔案轉換為 Reveal.js 簡報，支援即時圖表、流程圖、API 儀表板和互動控制元件。不需要 Node.js、不需要 npm — 只要一個 Go binary。

> 📖 [English README](README.md)

---

## ⚡ 快速開始

```bash
# 安裝
go install github.com/GMfatcat/goslide/cmd/goslide@latest

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
| 🎨 **14 種主題** | default、dark、corporate、minimal、hacker、dracula、midnight、gruvbox、solarized、catppuccin-mocha、ink-wash、instagram、western、pixel |
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

14 種內建主題 × 8 種強調色 = **112 種視覺組合**。

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
# API 代理
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

## 📄 授權

MIT
