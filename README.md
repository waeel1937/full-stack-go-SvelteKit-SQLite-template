# IIoT Edge Full-Stack Template

A full-stack template for industrial micro-applications in the IIoT and automation domain, designed for reliable, resource-efficient operation on edge devices such as Revolution Pi, industrial PCs, and embedded Linux systems.

The backend is built with Go and SQLite and exposes HTTP/JSON APIs and gRPC for efficient, low-latency, strongly typed communication. The frontend is built with SvelteKit and optionally wrapped as a native desktop application with Tauri. Authentication is handled by Keycloak with OAuth2 and JWT tokens.

---

## Stack

| Layer | Technology |
|---|---|
| Frontend | SvelteKit (dark/light mode) |
| Desktop Shell | Tauri |
| Backend | Go (static binary) |
| Database | SQLite (WAL mode, offline first) |
| Auth | Keycloak (OAuth2 / JWT) |
| API | HTTP/JSON + gRPC |
| Metrics | Prometheus |
| Target Hardware | Revolution Pi, IPC, ARM64, ARMv7 |
| Deployment | Docker Compose, systemd |

---

## Architecture
```mermaid
flowchart TB
    subgraph EDGE["Edge Device - Revolution Pi / IPC (5W)"]
        subgraph Frontend["Frontend Layer"]
            SK[SvelteKit UI]
            TA[Tauri Desktop Shell]
            SK --> TA
        end
        subgraph Transport["Transport Layer"]
            HTTP[HTTP / JSON REST API]
            GRPC[gRPC strongly typed]
        end
        subgraph Backend["Backend Layer"]
            GO[Go Core Service static binary]
            RE[Rule Engine local logic]
            AGG[Aggregator 99% reduction]
            DB[(SQLite offline first)]
        end
        SK --> HTTP
        TA --> HTTP
        SK --> GRPC
        TA --> GRPC
        HTTP --> GO
        GRPC --> GO
        GO --> RE
        GO --> AGG
        GO --> DB
    end
    subgraph AUTH["Authentication"]
        KC[Keycloak OAuth2 / JWT]
    end
    subgraph CLOUD["Cloud - minimal footprint"]
        SYNC[Selective Sync aggregates only]
    end
    subgraph GREEN["Green IT Impact"]
        G1[-99% Network Traffic]
        G2[-95% Cloud Compute]
        G3[+10yr Hardware Life]
        G4[< 30MB RAM Idle]
    end
    SK --> KC
    KC --> GO
    AGG -->|8 KB/s instead of 1 MB/s| SYNC
    EDGE --> GREEN
```
