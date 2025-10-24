# LogShelter

[![Golang](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17.0-336791?style=for-the-badge&logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-DC382D?style=for-the-badge&logo=redis&logoColor=white)](https://redis.io)
[![NATS](https://img.shields.io/badge/NATS-2C3D50?style=for-the-badge&logo=nats)](https://nats.io/)
[![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://www.docker.com/)
---

## Overview

Centralized storage, indexing, and retrieval of structured logs from distributed applications with support for filtering, pagination, and batch writing via NATS. A robust log management solution designed for modern microservices architectures.
<div style="text-align: right"> <sub>*Educational Project*</sub> </div>

## Dependencies

### Development Tools
- [direnv](https://direnv.net/) - Environment management
- [just](https://github.com/casey/just) - Command runner

### Build
- [Golang](https://golang.org/) - Main programming language
- [Docker](https://www.docker.com/) - Containerization (optional)

### Integration Tests
- [uv](https://github.com/astral-sh/uv) - Fast Python package installer
- [Python](https://www.python.org/) - For test infrastructure

## Installation

1. After you've installed and hooked `direnv` in root project directory:
```
direnv allow .
```

2. After direnv is set up copy env & config files:
```
cp .env.example .env
cp ./config/example.config.toml ./config/config.toml
```
And fill files with actual data.

## Running

1. Start all infrastructure (NATS, PostgreSQL, Redis, etc.):
```
docker compose up -d
```

2. Start app via just:
```
just run
```