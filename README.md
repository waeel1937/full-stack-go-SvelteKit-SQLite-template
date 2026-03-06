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

---

## Green IT Principles

This template is designed from the ground up around Green IT. Every architectural decision has a direct impact on energy consumption, hardware lifespan, and cloud infrastructure load. Sustainability is not a feature added on top — it is the result of the architecture itself.

### Less Cloud

Business logic, aggregation, and rule evaluation run entirely on the edge device. What is computed locally never reaches a data center. This directly reduces server capacity, cooling demand, and energy consumption in the cloud.

A Revolution Pi running this stack draws approximately 5 watts. An equivalent cloud server draws 200-400 watts. The difference is structural, not configurational.

| Metric | Cloud-first Architecture | This Template |
|---|---|---|
| Data sent to cloud | ~1 MB/s raw | ~8 KB/s aggregated |
| Cloud compute required | continuous | minimal |
| Network bandwidth | high | -99% |
| Latency for local rules | 100-500 ms roundtrip | < 10 ms local |

### Less Power

Go compiles to a single static binary with no external runtime dependencies. There is no JVM, no Node.js process, no Python interpreter running in the background. The application idles at under 30 MB RAM and releases CPU when there is nothing to process.

SQLite in WAL mode eliminates the need for a separate database process. There is no PostgreSQL daemon, no connection pooling overhead, no background vacuum process consuming resources continuously.

This means the hardware spends most of its time idle — and idle hardware consumes almost no energy.

### Longer Hardware Life

The binary targets ARMv7 and ARM64 and runs on hardware from 2016 onward without modification. There are no framework dependencies that force hardware upgrades. No new Node.js version that drops support for an older kernel. No container runtime that requires more RAM than the device has.

Every year a device stays in production instead of being replaced avoids the embodied carbon of manufacturing a new unit. For industrial hardware, embodied carbon typically accounts for 150-300 kg CO2 equivalent per device. Extending the lifespan from 5 to 10 years cuts that impact in half.

### Offline First

The application functions completely without internet connectivity. SQLite stores all data locally. The rule engine evaluates conditions and triggers actions without a cloud roundtrip. Alarms fire in under 10 milliseconds regardless of network state.

Cloud sync is selective and asynchronous. Only aggregated values are transmitted when connectivity is available. The sync tracks what has already been sent and resumes automatically after outages. No data is lost, no manual intervention is required.

This design eliminates the energy cost of maintaining a persistent cloud connection and reduces the carbon intensity of the entire system.

### Measurable Impact

Green IT is not a label — it is a number. This template is designed so that its environmental impact can be calculated and reported.
```
Energy saved per device per year:
  Cloud server equivalent:     ~350W x 8,760h = 3,066 kWh
  Edge device with this stack:   ~5W x 8,760h =    44 kWh
  Saving per device:                            ~3,022 kWh

CO2 equivalent (EU grid ~0.4 kg/kWh):
  Saving per device per year:               ~1,209 kg CO2

Hardware embodied carbon avoided (10yr vs 5yr lifecycle):
  ~150-300 kg CO2 per device avoided replacement
```

These numbers can be used directly in ESG reporting, Scope 3 disclosures, and Product Carbon Footprint (PCF) calculations.

---

## Quick Start
```bash
docker compose up --build
```

Wait about 60 seconds for Keycloak first start. Then:

| Service | URL | Credentials |
|---|---|---|
| Frontend | http://localhost:3000 | admin / admin |
| Backend API | http://localhost:8080 | JWT token required |
| Keycloak Admin | http://localhost:8180 | admin / admin |

---

## Project Structure
```
.
├── docker-compose.yml
├── go.mod
├── config/
│   └── keycloak/
│       └── edge-realm.json
├── edge-app/
│   ├── cmd/edge-app/main.go
│   ├── config/app.yaml
│   ├── deploy/
│   │   ├── docker/Dockerfile
│   │   └── systemd/edge-app.service
│   ├── internal/
│   │   ├── aggregator/
│   │   ├── api/
│   │   │   ├── http.go
│   │   │   ├── raw.go
│   │   │   ├── status.go
│   │   │   └── grpc/
│   │   ├── config/
│   │   ├── core/
│   │   │   ├── bus.go
│   │   │   ├── events.go
│   │   │   └── shutdown.go
│   │   ├── logging/
│   │   ├── metrics/
│   │   ├── rules/
│   │   ├── storage/
│   │   │   ├── sqlite.go
│   │   │   └── ringbuffer/
│   │   └── sync/
│   ├── scripts/
│   └── tests/
├── frontend/
│   ├── Dockerfile
│   ├── package.json
│   ├── svelte.config.js
│   ├── vite.config.js
│   └── src/
│       ├── app.html
│       ├── lib/stores/api.js
│       └── routes/
│           ├── +layout.svelte
│           ├── +page.svelte
│           ├── alerts/+page.svelte
│           └── rules/+page.svelte
└── src-tauri/
    ├── Cargo.toml
    ├── build.rs
    ├── tauri.conf.json
    └── src/main.rs
```
