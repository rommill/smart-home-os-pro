# Smart Home OS Pro

[![Smart Home OS Pro CI](https://github.com/rommill/smart-home-os-pro/actions/workflows/ci.yml/badge.svg)](https://github.com/rommill/smart-home-os-pro/actions/workflows/ci.yml)

An enterprise-grade, hybrid-architecture Smart Home platform engineered for high-throughput IoT data ingestion, real-time device isolation, and telemetry visualization.

This project demonstrates production-ready microservice synchronization using standard industry patterns, strict type-safety, and secure user-data boundaries.

---

## System Architecture & Service Split

The platform is explicitly designed with a CQRS-inspired hybrid architecture to optimize resource utilization and isolate domains:

```text
                      +---------------------------------------+
                      |       Frontend (React + TypeScript)   |
                      +-------------------+-------------------+
                                          |
                        HTTP / REST (JWT) | (Event Streaming)
                                          v
+-----------------------------------------+-----------------------------------------+
| [Command & Control Layer]                                                         |
|                                                                                   |
|  +---------------------------+              +----------------------------------+  |
|  |  backend-spring (Java 21) | -----------> |     PostgreSQL (smart_home_db)   |  |
|  +---------------------------+              +----------------------------------+  |
|    - User Management & Auth (JWT)                                  ^              |
|    - Device CRUD & Metadata                                        |              |
+--------------------------------------------------------------------|--------------+
                                                                     |
+--------------------------------------------------------------------|--------------+
| [High-Throughput Ingestion Layer]                                  |              |
|                                                                    |              |
|  +---------------------------+              +----------------------+---------+    |
|  |       backend (Go)         | ------------ |     MQTT Broker (Mosquitto)   |    |
|  +---------------------------+              +----------------------+---------+    |
|    - Lightweight MQTT Ingestion                                    ^              |
|    - Fast Real-Time Processing                                     |              |
+--------------------------------------------------------------------|--------------+
                                                                     |
                                              +----------------------+---------+    |
                                              |    Python Climate Emulator     |    |
                                              +--------------------------------+    |
```

## Architectural Decisions

### Spring Boot

**Spring Boot (`backend-spring`)** serves as the core administrative and identity provider.

It handles:

- Enterprise-level security
- JWT parsing
- Data validation
- Core relational business logic
- User management
- Device management

### Go

**Go (`backend`)** functions as a lightweight telemetry ingestion engine.

It handles high-frequency MQTT message streams and real-time telemetry processing.

Go was selected for its:

- Low memory overhead
- High concurrency
- Efficient network I/O
- Fast startup
- Lightweight runtime characteristics

### MQTT

**Mosquitto MQTT** provides the asynchronous communication layer between the simulated devices and the telemetry backend.

MQTT is well suited for IoT workloads because of its lightweight publish/subscribe communication model.

---

## Key Features

### Strict Device Isolation

Devices are securely mapped to authenticated user identities extracted from valid JWT claims.

The application is designed so that users cannot access or modify another user's IoT devices.

### Dual-Protocol Communication

Synchronous operations such as authentication and device management use REST APIs.

Asynchronous telemetry data is transported through MQTT.

```text
REST / HTTP
     |
     v
Spring Boot
     |
     v
PostgreSQL
```

```text
Python Emulator
      |
      | MQTT
      v
Mosquitto
      |
      v
Go Backend
      |
      v
Telemetry Processing
```

### Automated Sensor Simulation

The project includes a Python climate emulator that simulates environmental changes and publishes telemetry data through MQTT.

This allows the complete IoT data pipeline to be tested without physical hardware.

### Internationalization

The frontend includes multi-language support using `i18next`.

---

## Technology Stack

### Frontend

- React 18
- TypeScript
- Vite
- Tailwind CSS
- i18next

### Administrative Backend

- Java 21
- Spring Boot 4.0
- Spring Security
- Stateless JWT
- Hibernate / JPA

### Telemetry Backend

- Go (Golang)
- Eclipse Paho MQTT

### Infrastructure & Messaging

- Docker
- Docker Compose
- Eclipse Mosquitto MQTT Broker

### Databases

- PostgreSQL
- Redis
- SQLite

### Testing & API

- JUnit
- Go testing
- Swagger / OpenAPI
- GitHub Actions

---

## Project Structure

```text
smart-home-os-pro/
│
├── backend/
│   ├── api/
│   │   ├── auth.go
│   │   ├── handlers_test.go
│   │   ├── middleware.go
│   │   └── telemetry.go
│   │
│   ├── db/
│   │   └── db.go
│   │
│   ├── models/
│   │   └── telemetry.go
│   │
│   ├── go.mod
│   └── go.sum
│
├── backend-spring/
│   ├── src/
│   │   └── test/
│   │       └── java/
│   │           └── ...
│   │
│   ├── build.gradle
│   ├── gradlew
│   └── settings.gradle
│
├── frontend/
│   ├── js/
│   ├── src/
│   │   ├── api/
│   │   ├── auth/
│   │   ├── i18n/
│   │   ├── render/
│   │   ├── theme/
│   │   └── types/
│   │
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
│
├── climate_emulator.py
├── docker-compose.yml
├── mosquitto.conf
└── README.md
```

---

## Getting Started

### Prerequisites

- Docker
- Docker Compose

Optional for running individual services locally:

- Java 21
- Go SDK
- Node.js
- pnpm

### Installation

Clone the repository:

```bash
git clone https://github.com/rommill/smart-home-os-pro.git
cd smart-home-os-pro
```

Start the complete containerized infrastructure:

```bash
docker compose up --build
```

This starts the application infrastructure including the database, MQTT broker, backend services and climate emulator.

---

## API & Documentation

The project exposes an interactive OpenAPI / Swagger UI.

Once the services are running, open:

```text
http://localhost:8082/swagger-ui/index.html
```

Swagger UI can be used to explore and test the REST API.

---

## Authentication Example

Authenticate using:

```http
POST /api/auth/login
Content-Type: application/json
```

Example request:

```json
{
  "username": "roman",
  "password": "your_secure_password"
}
```

Example response:

```json
{
  "token": "eyJhbGciOiJIUzM4NCJ9..."
}
```

The returned JWT can then be supplied through the **Authorize** button in Swagger UI.

Authenticated requests use:

```text
Authorization: Bearer <JWT_TOKEN>
```

---

## Telemetry Pipeline

The telemetry system follows an asynchronous MQTT-based pipeline:

```text
+--------------------------+
| Python Climate Emulator  |
+------------+-------------+
             |
             | MQTT Publish
             v
+--------------------------+
| Mosquitto MQTT Broker    |
+------------+-------------+
             |
             | MQTT Subscribe
             v
+--------------------------+
| Go Telemetry Backend     |
+------------+-------------+
             |
             | Telemetry Processing
             v
+--------------------------+
| Frontend Visualization   |
+--------------------------+
```

The Go backend contains the telemetry handling logic, including:

```text
backend/api/telemetry.go
```

---

## Testing

The project includes automated tests for both backend components.

### Spring Boot Tests

```bash
cd backend-spring
./gradlew test
```

### Go Tests

```bash
cd backend
go test ./...
```

---

## CI/CD

The project uses GitHub Actions for continuous integration.

The CI pipeline automatically validates the project and builds the required components.

[![Smart Home OS Pro CI](https://github.com/rommill/smart-home-os-pro/actions/workflows/ci.yml/badge.svg)](https://github.com/rommill/smart-home-os-pro/actions/workflows/ci.yml)

---

## Roadmap & Production Hardening

- [ ] Implement refresh-token mechanics to complement short-lived access JWTs.
- [ ] Integrate Role-Based Access Control (RBAC) inside Spring Security.
- [x] Implement CI/CD pipeline using GitHub Actions.
- [ ] Introduce horizontal scaling policies.
- [ ] Introduce resilient persistent storage clustering.
- [ ] Expand observability and monitoring.
- [ ] Improve production secret management.

---

## Engineering Focus

The project focuses on practical backend and full-stack engineering concepts:

- Service separation
- Secure authentication
- User-level device isolation
- REST API design
- Asynchronous MQTT communication
- High-frequency telemetry ingestion
- Containerized infrastructure
- Automated testing
- CI/CD
- Type safety
- Internationalization
- Production-oriented architecture

---

## Current Status

Smart Home OS Pro is an actively developed project.

The current implementation includes:

- React Smart Home dashboard
- TypeScript frontend
- Device management
- JWT authentication
- Spring Boot backend
- Go telemetry backend
- MQTT communication
- PostgreSQL integration
- Python climate emulator
- Docker-based infrastructure
- Swagger / OpenAPI documentation
- Automated backend tests
- GitHub Actions CI

Additional security, scalability and production-hardening features are planned.

---

## Author

**Roman**

Full-stack Engineer focused on React, TypeScript, Java, Spring Boot, Go and modern backend architecture.

GitHub:  
https://github.com/rommill
