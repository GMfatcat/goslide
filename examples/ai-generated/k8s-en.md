---  
title: Kubernetes Overview  
theme: dark  
language: en  
---  

# Kubernetes  
*Container orchestration platform that automates deployment, scaling, and management of containerized applications.*

---  
title: What Problems Does Kubernetes Solve?  
---  

- Manual container management is error‑prone.  
- Scaling services across many hosts is complex.  
- Ensuring high availability requires constant monitoring.  
- Networking and storage need consistent abstraction.

---  
title: Core Architecture  
layout: two-column  
---  

<!-- col -->  
## Control Plane  

- API Server (kubectl <-> cluster)  
- Scheduler (places pods)  
- Controller Manager (vital controllers)  
- etcd (consistent key‑value store)  

<!-- col -->  
## Worker Node  

- kubelet (node agent)  
- kube-proxy (network rules)  
- container runtime (Docker, containerd)  

---  
title: Primary Objects  
---  

- **Pod**: smallest deployable unit, one or more containers.  
- **Service**: stable network endpoint, load‑balances pods.  
- **Deployment**: declarative rollout & roll‑back of pod sets.  
- **ConfigMap / Secret**: inject configuration & credentials.  

---  
title: Pod Lifecycle  
layout: two-column  
---  

<!-- col -->  
## States  

1. Pending – scheduled but not running.  
2. Running – all containers started.  
3. Succeeded – all containers exited cleanly.  
4. Failed – one or more containers terminated with error.  
5. Unknown – node communication lost.  

<!-- col -->  
## Controllers  

- **ReplicaSet** maintains desired replica count.  
- **Job** runs pods to completion.  
- **DaemonSet** runs a pod on every node.  

---  
title: Service Types & Networking  
---  

- **ClusterIP** – internal only, default.  
- **NodePort** – exposes on each node’s IP + port.  
- **LoadBalancer** – provisions cloud LB.  
- **ExternalName** – alias to DNS name.  

*Uses kube‑proxy (iptables/ipvs) for virtual IP handling.*

---  
title: Persistent Storage  
layout: two-column  
---  

<!-- col -->  
## Volume Types  

- `emptyDir` – temporary, pod‑lifetime.  
- `hostPath` – mounts host directory (node‑specific).  
- `persistentVolumeClaim` – request from a PV.  

<!-- col -->  
## Storage Classes  

- Dynamically provisioned (e.g., AWS EBS, GCE PD).  
- Define reclaim policy, provisioner, parameters.  

---  
title: Scaling & Autoscaling  
---  

- **Horizontal Pod Autoscaler (HPA)** – scales pods based on CPU/memory or custom metrics.  
- **Vertical Pod Autoscaler (VPA)** – adjusts container resources.  
- **Cluster Autoscaler** – adds/removes nodes per pending pod demands.  

---  
title: Security Foundations  
---  

- **RBAC** – role‑based access control for API permissions.  
- **Network Policies** – whitelist traffic between pods.  
- **Pod Security Standards** – restrict privilege escalation.  
- **Secrets** – base64‑encoded, optionally encrypted at rest.

---  
title: Monitoring & Observability  
---  

```chart
type: line
title: Cluster CPU Utilization
data:
  labels: [0m,5m,10m,15m,20m,25m]
  values: [45,55,60,70,65,80]
```  

- Prometheus scrapes metrics from kube‑state‑metrics & cAdvisor.  
- Grafana dashboards visualize pod health, API latency, etc.  
- Fluentd / Loki aggregate logs for centralized analysis.

---  
title: Best‑Practice Checklist  
layout: two-column  
---  

<!-- col -->  
## Deployments  

- Use declarative YAML with `kubectl apply`.  
- Keep images immutable and scanned.  
- Employ rolling updates with `maxSurge`/`maxUnavailable`.  

<!-- col -->  
## Operations  

- Enable PodDisruptionBudgets for graceful maintenance.  
- Tag resources with labels for selection & cost tracking.  
- Regularly back up etcd snapshots.  

---  
title: Summary & Q&A  
---  

- Kubernetes abstracts compute, storage, and networking into declarative APIs.  
- Control plane + worker nodes deliver resilient, scalable workloads.  
- Security, monitoring, and autoscaling complete the production‑ready stack.  

**Questions?**