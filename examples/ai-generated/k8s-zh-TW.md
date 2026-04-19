---
title: Kubernetes 入門
theme: dark
language: zh-TW
---
# Kubernetes 是什麼？

把它想像成 **一座自動化的快餐廚房**，  
所有料理（應用程式）會在不同的爐子（容器）上同時烹調，  
而 Kubernetes 就是那位負責指揮、分配食材、監控烹飪進度的總廚師。

---

title: 為什麼需要「總廚師」？
---
- 料理需求會突增（流量高峰）或下降
- 食材（資源）有限，不能讓爐子閒置或過載
- 有時爐子會壞掉，需要立即換一道菜
- 總廚師自動調整、補位、擴增，讓客人不會等太久

---

title: 核心概念：容器（Container）
---
- **容器** 就像 **小保鮮盒**，裝著完整的料理配方和材料
- 可以在任何爐子上快速加熱、搬移
- 保證每次味道都一樣（環境一致）

---

title: 核心概念：Pod
---
- **Pod** 是 **一組一起上桌的菜**（可能是一主菜加配菜）
- 同時在同一個爐子裡運行，互相共享水槽與調味料
- 最小可管理單位；Kubernetes 會把 Pod 當成一個「餐點」來安排

---

title: 核心概念：Node（工作站）
---
- **Node** 就是 **快餐店的爐灶區**，有多個爐子（CPU、記憶體）可用
- 每個 Node 會報告自己還剩多少資源，讓總廚師決定下單哪裡

---

title: 核心概念：控制平面（Control Plane）
---
```card
---
title: 控制平面
icon: "🧑‍🍳"
---
* 接收點餐（使用者需求）  
* 判斷哪個 Node 有空位  
* 發布「做菜指令」給 Node  
* 監控菜品狀態，失敗時重新下單
```

---

title: 自動伸縮（Horizontal Pod Autoscaler）
---
- 想像店裡突然來了 100 人排隊，總廚師會 **自動加開更多爐子**（新 Pod）  
- 需求減少時，又會 **關閉多餘的爐子**，省電又省錢

---

title: 自我修復（Self-healing）
---
- 若某道菜（Pod）因爐子故障燒焦，總廚師會 **立刻重新下單**，把同樣的菜做出來  
- 用戶永遠看不到「爐子壞掉」的情況

---

title: 兩欄示例：部署一個簡單網站
---
layout: two-column
---
<!-- col -->
### 步驟 1：寫下料理配方
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: web-pod
spec:
  containers:
  - name: web
    image: nginx
```
<!-- col -->
### 步驟 2：交給總廚師
```bash
kubectl apply -f web-pod.yaml
```
總廚師會把 Pod 安排到合適的 Node，確保網站可以對外服務。

---

title: 小挑戰：自己動手玩 Minikube
---
- **Minikube** 是在筆電上模擬整座快餐廚房的迷你版
- 步驟：
  1. 安裝 VirtualBox（或其他虛擬機）  
  2. `brew install minikube`（Mac）或在官網下載執行檔  
  3. `minikube start` 開啟本地 Kubernetes  
  4. 用上面的小範例部署一個 Nginx 網站  
- 完成後，用瀏覽器打開 `http://$(minikube ip)`，看到「Welcome to nginx!」就成功啦！

祝大家玩得開心，期待看到你們的第一個 Pod 🎉