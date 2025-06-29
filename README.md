# URL Shortener Microservice

## Overview

This repository contains a lightweight, production-ready URL Shortener microservice written in Go. It provides a simple REST API for generating short URLs from original URLs and redirecting users from the short URL to the original URL. The service supports both plain text and JSON input formats, providing flexibility for various clients, and adheres to modern best practices for scalability, reliability, and maintainability.

## Features

- **Shorten long URLs:** Create compact, easy-to-share short URLs.
- **Fast redirection:** Instant redirection from short URL to the original.
- **REST API:** Provides both text/plain and application/json interfaces.
- **Configurable:** Parameters can be set via command-line flags or environment variables.
- **Extensible:** Clean separation of concerns for easy extension and unit testing.
- **Middleware-friendly:** Integrates with `chi` for HTTP routing and middleware support (logging, error recovery, etc).
- **Graceful error handling** and informative logging.

## Getting Started

### Prerequisites

- Go 1.19 or later
- Git

### Installation

Clone the repository:

```bash
git clone https://github.com/you-humble/shortener.git
cd shortener
```

### Build

```bash
go build -o bin/shortener cmd/shortener/*.go
```

### Configuration

The service can be configured via flags or environment variables.

Flag	Environment	Default	Description
-a	SERVER_ADDRESS	localhost:8080	Listen address/port
-b	BASE_URL	localhost:8080	Base URL for short links

Example with environment variables:

```bash
export SERVER_ADDRESS="0.0.0.0:8080"
export BASE_URL="https://short.my"
./bin/shortener
```

### Running the Service

Start the server:
```bash
./bin/shortener -a "localhost:8080" -b "https://short.my"
```

## API Reference

### 1. Shorten URL via Text

- **Endpoint:** `POST /`
- **Content-Type:** `text/plain`
- **Body:** original URL as plain text

**Response:**  
- `201 Created`  
- Short URL in response body as plain text

---

### 2. Shorten URL via JSON

- **Endpoint:** `POST /api/shorten`
- **Content-Type:** `application/json`
- **Body:**
json
    {
      "url": "https://original.example.com/some-long-link"
    }

**Response:**  
- `201 Created`  
- JSON object:
json {
      "result": "https://short.my/abc123"
    }

---

### 3. Redirect to Original URL

- **Endpoint:** `GET /{short}`
- **Behavior:** Redirects to the original URL corresponding to the short code.
- **Response:**  
  - `307 Temporary Redirect` on success  
  - `404 Not Found` if not found

## Extending & Improving

**Potential Improvements:**
- Use persistent storage (e.g., PostgreSQL/Redis) instead of in-memory for production.
- Add JWT or API Key authentication for rate limiting and security.
- Implement analytics (number of visits per short link, etc).
- Add an admin UI and API for link management.
- Enable HTTPS (TLS) in production settings.
- Write comprehensive unit and integration tests.
- Support for password-protected or expiring short links.
- Add OpenAPI/Swagger documentation for easy API consumption.
