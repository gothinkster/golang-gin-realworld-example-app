# RealWorld DevOps Project: Golang + Kubernetes + Monitoring 🚀

![Status](https://img.shields.io/badge/Status-In%20Progress-yellow)
![Docker](https://img.shields.io/badge/docker-ready-blue)
![Kubernetes](https://img.shields.io/badge/kubernetes-planned-blueviolet)

> **🚧 Project Status: Active Development / Implementation Phase**
> This repository is a living document of my DevOps learning journey. I am currently transforming a standard Go application into a cloud-native, monitored microservice.

## 📖 About The Project

This is a **DevOps portfolio project**. The application source code is a fork of the [Golang Gin RealWorld Example App](https://github.com/gothinkster/golang-gin-realworld-example-app).

My goal is to demonstrate practical knowledge of:

- **Containerization** (Docker)
- **Orchestration** (Kubernetes)
- **Observability** (Prometheus & Grafana)
- **CI/CD Automation** (GitHub Actions)

## 📍 Implementation Roadmap & Progress

Here is the current status of the project implementation.

### Phase 1: Dockerization 🐳

- [ ] Create a basic `Dockerfile` for the Go application.
- [ ] Optimize image size using **Multi-stage builds** (Builder vs Runner).
- [ ] Create `docker-compose.yaml` to run App + Database (PostgreSQL) locally.
- [ ] Test application connectivity with the database in containers.

### Phase 2: Kubernetes (K8s) ☸️

- [ ] Setup local cluster (Minikube / Kind).
- [ ] Create **ConfigMap** & **Secret** manifests for environment variables.
- [ ] Create **Deployment** manifest for the Go application.
- [ ] Create **Service** manifest (ClusterIP/NodePort) to expose the app.
- [ ] Create **PostgreSQL** deployment (StatefulSet or Deployment) for K8s.

### Phase 3: Observability & Monitoring 📊

- [ ] **Instrumentation:** Add Prometheus client library to Go code (`/metrics` endpoint).
- [ ] Deploy **Prometheus** to the cluster.
- [ ] Configure Prometheus `scrape_configs` to discover application pods.
- [ ] Deploy **Grafana**.
- [ ] Create a Grafana Dashboard visualizing:
  - [ ] HTTP Request Rate (RPS).
  - [ ] Request Duration (Latency).
  - [ ] Error Rates (5xx codes).

### Phase 4: CI/CD Automation 🤖

- [ ] Set up **GitHub Actions** workflow.
- [ ] Automated Linting & Unit Testing on PR.
- [ ] Automated Docker Image Build & Push to DockerHub/GHCR.

---

## 🛠️ Tech Stack

- **App:** Golang (Gin)
- **Container:** Docker
- **Orchestration:** Kubernetes
- **Monitoring:** Prometheus, Grafana
- **CI/CD:** GitHub Actions

---

_Created by [XamDev]_
