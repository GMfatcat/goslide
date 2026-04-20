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

### AI 投影片生成（experimental）

> ⚠️ **實驗性功能。** 產出品質高度取決於所選模型與 prompt。生成的 Markdown 會經過 parser sanity check 並自動修復常見語法錯誤，但語意品質（投影片流程、正確性、風格）無保證，簡報前請人工審閱。API 行為與 CLI flag 可能在沒有預告下變動。

透過任何 OpenAI-compatible 的 LLM endpoint（OpenAI、OpenRouter、Ollama、vllm、sglang 等）從主題生成完整簡報。

在 `goslide.yaml` 加入 `generate:` 區段：

```yaml
generate:
  base_url: https://api.openai.com/v1
  model: gpt-4o
  api_key_env: OPENAI_API_KEY
  timeout: 120s
```

設定 API key 後執行：

```bash
export OPENAI_API_KEY=sk-...
goslide generate "Kubernetes 架構介紹"                    # 簡易模式
goslide generate my-prompt.md -o talk.md                 # 進階模式
goslide generate --dump-prompt > system.txt              # 檢視內建 prompt
```

進階模式讀取 `prompt.md` 檔：

```markdown
---
topic: Kubernetes 架構
audience: Backend engineers
slides: 15
theme: dark
language: zh-TW
---
強調 Pod/Service/Ingress，最後放 Q&A slide。
```

除非加 `--force`，否則不會覆寫既有檔案。生成的 Markdown 會用 parser 做一次 sanity check；常見錯誤（未閉合 code fence、frontmatter 缺結尾 `---`）會自動修復並印出透明報告。

**驗證範例。** 參見 [`examples/ai-generated/`](examples/ai-generated/) — 用 OpenRouter 的 `openai/gpt-oss-120b:free` 實際生成的英文與繁體中文簡報（simple + advanced mode），均於首次 parse 通過。`scripts/test-generate-llm.ps1` 可重現。

### 圖片佔位符與多圖投影片

尚未準備好實際圖檔（或用 `goslide generate` 生成時），可用 `placeholder` component：

```
~~~placeholder
hint: K8s 架構圖
icon: 🗺️
aspect: 16:9
---
Control plane 與 worker 互動示意
~~~
```

搭配 `image-grid` layout 可在同一張投影片並排多個 placeholder、真實圖片、圖表、卡片：

```
<!-- layout: image-grid -->
<!-- columns: 2 -->

<!-- cell -->
~~~placeholder
hint: 架構圖
icon: 🗺️
~~~

<!-- cell -->
![Dashboard](./dashboard.png)

<!-- cell -->
~~~placeholder
hint: 趨勢分析
icon: 📈
~~~

<!-- cell -->
~~~chart
type: bar
...
~~~
```

`columns` 可設 `2`、`3` 或 `4`。每個 `<!-- cell -->` 代表一格；格內可放任何內容。

### api component 的 LLM 轉換器（experimental）

`api` component 新增一種 render item：`type: llm`。它把 fetch 到的 JSON 透過 `{{data}}` 代入作者寫的 prompt，再把 LLM 回應渲染在該位置：

```
~~~api
endpoint: /api/sales
render:
  - type: chart:bar
    label-key: quarter
    data-key: revenue
  - type: llm
    prompt: |
      用 3 個分析師觀點摘要以下數字：
      {{data}}
    model: openai/gpt-oss-120b:free   # 可選；未設則用 generate.model
~~~
```

- **Cache-first。** 相同 `(model, prompt, data)` 組合最多呼叫 LLM 一次，後續從 `.goslide-cache/` 讀。
- **Click-to-call。** `goslide serve` 遇到 cache miss 時顯示 `Generate ✨` 按鈕；頁面開啟本身不會自動呼叫 LLM。
- **Build-lock。** `goslide build` 會把 cache 內容烘進靜態 HTML；匯出後的簡報絕不會在觀看時對外呼叫。

直接重用 `goslide.yaml` 的 `generate:` 區段，不需新增設定。

離線 build workflow：在 `api` component 加 `fixture: ./data.json`，`goslide build` 就會用 fixture 當資料來源，不會呼叫實際 endpoint。

`goslide build --llm-refresh` 會在 cache miss 時真的打 LLM 補 cache。沒加這個 flag 的話，cache miss 會讓 build 失敗並列出缺的 slide。

v1.4.0 的版本只支援手寫 `llm` render item；`goslide generate` 還不會自動生成這類 render item。

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

## 📡 講者同步

GoSlide 內建輕量級講者同步功能。講者目前的投影片頁數會即時廣播給所有觀眾。

**講者** 使用 `?role=presenter` 開啟：
```
http://localhost:3000?role=presenter
```

**觀眾** 開啟普通 URL：
```
http://localhost:3000
```

當講者切換投影片時，觀眾左下角會顯示提示：

```
┌──────────────────────────┐
│  Presenter: 5/12  [Jump] │
└──────────────────────────┘
```

- 觀眾可以點 **Jump** 跳到講者目前的頁面
- 觀眾可以自由瀏覽，不會被強制跟隨
- `serve` 和 `host` 模式皆可使用

### 講者檢視

在任何投影片按 **S** 鍵開啟講者檢視（新視窗），顯示：
- 目前投影片 + 下一張預覽
- 講者筆記（來自 markdown 中的 `<!-- notes -->`）
- 經過時間

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
