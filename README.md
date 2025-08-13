# Squid Proxy Whitelist Editor

**Repository**: https://github.com/coledykstra7/docker-whitelist

## Project Purpose
This project provides a web-based editor for managing Squid proxy whitelist and blacklist files, viewing access logs, and reloading Squid configuration. It is designed to simplify the administration of Squid proxy rules and monitoring.

## Features
- Edit whitelist and blacklist files via web UI
- View access logs and summary statistics
- Reload Squid configuration from the web interface
- Static file serving for UI assets
- Dockerized deployment with Squid and editor services

## Endpoints
- `GET /summary` — Returns summary statistics from access logs
- `GET /log` — Returns recent access log entries
- `POST /save` — Saves changes to whitelist/blacklist files
- `POST /reload` — Reloads Squid configuration
- `POST /clear` — Sets a setpoint to filter logs after current time
- `GET /static/template.js` — Serves static JS for the UI

## Build & Run Instructions
### Prerequisites
- Docker & Docker Compose
- Go (for local development)

### Build & Run with Docker Compose
```bash
docker-compose up --build
```
- Access the web UI at: http://localhost:8080
- Squid proxy runs on port 3128

### Local Development
```bash
cd src
go build -o ../squid-editor .
../squid-editor
```

## Directory Structure
- `src/` — Go source code
  - `main.go` — Application entry point
  - `handlers.go` — HTTP request handlers  
  - `logs.go` — Log processing and analysis
  - `squid.go` — Squid proxy operations
  - `files.go` — File I/O operations
  - `html.go` — HTML generation
  - `utils.go` — Utility functions
  - `types.go` — Type definitions and constants
  - `main_test.go` — Unit tests
- `html/` — UI templates and JS
- `data/` — Whitelist, blacklist, and log files
- `squid/` — Squid configuration
- `docker-compose.yml`, `Dockerfile` — Container setup

## License
MIT
