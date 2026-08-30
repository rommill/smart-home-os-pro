# Smart Home OS Pro

[![Smart Home OS Pro CI](https://github.com/rommill/smart-home-os-pro/actions/workflows/ci.yml/badge.svg)](https://github.com/rommill/smart-home-os-pro/actions/workflows/ci.yml)

An enterprise-grade, hybrid-architecture Smart Home platform engineered for high-throughput IoT data ingestion, real-time device isolation, and telemetry visualization.
This project demonstrates production-ready microservice synchronization using standard industry patterns, strict type-safety, and secure user-data boundaries.

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
|  |       backend (Go)        | ------------ |     MQTT Broker (Mosquitto)    |    |
|  +---------------------------+              +----------------------+---------+    |
|    - Lightweight MQTT Ingestion                                    ^              |
|    - Fast Real-Time Processing                                     |              |
+--------------------------------------------------------------------|--------------+
                                                                     |
                                              +----------------------+---------+    |
                                              |    Python Climate Emulator     |    |
                                              +--------------------------------+    |

Architectural Decisions (Why this stack?)
Spring Boot (backend-spring): Serves as the core administrative and identity provider. It handles enterprise-level security, JWT parsing, data validation, and core relational business logic.

Go (backend): Functions as an ultra-lightweight, reactive telemetry ingestion engine. Handling high-frequency MQTT message streams from hundreds of IoT sensors via a heavy framework is resource-inefficient; Go provides near-zero memory footprint and optimal throughput.

MQTT (Mosquitto): Utilized as the industrial standard for low-overhead, pub/sub communication, perfectly mimicking real hardware network limitations.

Key Features
Strict Device Isolation (Multi-Tenancy): Devices are securely mapped to verified user identities extracted directly from valid JWT claims. One user can never access or modify another user's IoT fleet.

Dual-Protocol Synchronicity: Synchronous operations (Authentication, Asset management) run over REST API, while asynchronous data streams (Telemetry) are piped via MQTT.

Automated Sensor Simulation: Includes a native Python emulator that mimics real-world climate shifts and broadcasts state-packets.

Internationalization (i18n): Production-ready UI localized with multi-language support from the ground up.

Technology Stack
Frontend: React 18, TypeScript, Vite, Tailwind CSS, i18next

Administrative Backend: Java 21, Spring Boot 4.0, Spring Security (Stateless JWT), Hibernate/JPA

Ingestion Backend: Go (Golang), Eclipse Paho MQTT

Infrastructure & Messaging: Docker, Docker Compose, Eclipse Mosquitto MQTT Broker

Databases: PostgreSQL (Production-grade relational state), Redis, SQLite

Getting Started
Prerequisites
Docker & Docker Compose installed

Java 21 & Go SDK (Optional, for local bare-metal testing)

Installation & Launch
Clone the repository:
```
git clone https://github.com/rommill/smart-home-os-pro.git
cd smart-home-os-pro 
```

Spin up the entire containerized infrastructure (Database, Broker, Backends, Emulator):

```
docker compose up --build
```

API & Documentation
Once the services are active, the fully interactive OpenAPI / Swagger UI configuration is automatically exposed:

URL: http://localhost:8082/swagger-ui/index.html

Target Integration Example: Authenticating Requests
To execute secure requests (POST /api/devices), authenticate first via the auth endpoint to receive your Bearer Token:

```
POST /api/auth/login
Content-Type: application/json

{
  "username": "roman",
  "password": "your_secure_password"
}
```
Response:
```
{
  "token": "eyJhbGciOiJIUzM4NCJ9.eyJzdWIiOiJyb21hbi..."
}
```
Pass this token inside Swagger UI via the top-right Authorize lock button to clear the security filter container.

Testing
Both backend segments include comprehensive suites covering validation rules, business operations, and mock integration points.

Execute Spring Boot Unit/Integration Tests:

```
cd backend-spring
./gradlew test
```
Execute Go Telemetry Tests:

```
cd backend
go test ./...
```

Roadmap & Production Hardening
[ ] Implement Refresh Token mechanics to complement short-lived access JWTs.

[ ] Integrate Role-Based Access Control (RBAC) layers inside Spring Security.

[x] Implement a full CI/CD Pipeline using GitHub Actions for multi-architecture Docker builds.

[ ] Introduce horizontal scaling policies and a resilient persistent storage engine cluster.

