---
title: Kubernetes Architecture Overview
theme: dark
accent: teal
language: en
---

# Kubernetes Architecture Overview

A high‑level look at the key components and flow of a Kubernetes cluster.

---

<!-- layout: two-column -->

# Cluster Architecture

<!-- left -->

~~~placeholder
hint: Kubernetes Control Plane
icon: 🗺️
aspect: 16:9
---
Control Plane components: API Server, Scheduler, Controller Manager, etcd.
~~~

<!-- right -->

~~~placeholder
hint: Worker Nodes
icon: 🗺️
aspect: 16:9
---
Worker nodes run kubelet, kube-proxy, and containers.
~~~

---

<!-- layout: two-column -->

# API Server Flow

<!-- left -->

~~~placeholder
hint: API Request Flow
icon: 📈
aspect: 4:3
---
Client → API Server → etcd → Components
~~~

<!-- right -->

- Authentication & Authorization
- Admission Controllers
- Persistent store (etcd)

---

<!-- layout: image-grid -->
<!-- columns: 2 -->

<!-- cell -->

~~~placeholder
hint: Pods
icon: 🧪
aspect: 4:3
---
Represents one or more containers.
~~~

<!-- cell -->

~~~placeholder
hint: Services
icon: 🌐
aspect: 4:3
---
Stable networking abstraction over Pods.
~~~

<!-- cell -->

~~~placeholder
hint: Deployments
icon: 📦
aspect: 4:3
---
Declarative Pod lifecycle management.
~~~

<!-- cell -->

~~~placeholder
hint: Ingress
icon: 🚪
aspect: 4:3
---
External access to Services.
~~~

---

<!-- layout: two-column -->

# Networking Model

<!-- left -->

~~~placeholder
hint: CNI Plugins
icon: 🔌
aspect: 3:2
---
Container Network Interface integration.
~~~

<!-- right -->

- Pod networking (IP per Pod)
- Cluster‑wide DNS (CoreDNS)
- Network policies

---

<!-- layout: two-column -->

# Storage Basics

<!-- left -->

~~~placeholder
hint: Persistent Volumes
icon: 💾
aspect: 4:3
---
Cluster‑wide storage resources.
~~~

<!-- right -->

- PV vs PVC
- Storage Classes
- Dynamic provisioning

---

<!-- layout: two-column -->

# Security Principles

<!-- left -->

~~~placeholder
hint: RBAC
icon: 🔐
aspect: 4:3
---
Role‑Based Access Control
~~~

<!-- right -->

- Secrets
- Network policies
- Pod Security Policies / OPA

---

<!-- layout: two-column -->

# Monitoring & Logging

<!-- left -->

~~~placeholder
hint: Prometheus & Grafana
icon: 📊
aspect: 4:3
---
Metrics collection and dashboarding.
~~~

<!-- right -->

- EFK stack (Elasticsearch, Fluentd, Kibana)
- Alertmanager

---

<!-- layout: two-column -->

# Helm – Package Management

<!-- left -->

~~~placeholder
hint: Helm Charts
icon: 📦
aspect: 4:3
---
Reusable application packages.
~~~

<!-- right -->

- Repositories
- Templating
- Release history

---

# Summary & Next Steps

- Understand core objects: Pod, Service, Deployment, Ingress.
- Learn Control Plane vs Worker Node responsibilities.
- Explore networking, storage, security, monitoring, and Helm.

---

# Questions?

Feel free to ask anything about Kubernetes architecture or how it applies to your projects.