---
title: Q4 Review
theme: dark
---

# Sales Overview

~~~api
endpoint: /api/sales
fixture: sales.json
render:
  - type: chart:bar
    label-key: q
    data-key: revenue
  - type: llm
    prompt: |
      Using the provided sales data, write exactly 2 analyst-style bullet
      observations. Be concrete: name the best and worst quarter with
      their revenue numbers, and note the growth trajectory. Format as a
      Markdown list, no preamble, no closing remarks.

      Data:
      {{data}}
~~~
