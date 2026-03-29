# DartScheduler REST API Documentation Plan

**Date:** March 29, 2026
**Status:** Research & Planning (No changes made)
**Version:** 1.0

---

## Executive Summary

DartScheduler is a full-stack dart club competition scheduling system with a Go backend (Chi router) and embedded Angular frontend. The API is currently documented only in the README.md with basic endpoint descriptions and no interactive documentation. This plan outlines a comprehensive approach to create enterprise-grade API documentation with OpenAPI 3.1 compliance, interactive features, and code examples.

---

## 1. Complete API Endpoint Inventory

### 1.1 Authentication & Health

| Method | Endpoint | Purpose | Request | Response | Status |
|--------|----------|---------|---------|----------|--------|
| GET | `/health` | Health check | None | `"ok"` (text/plain) | Working |
| GET | `/api/config` | Get app configuration | None | `{appTitle, clubName}` | Working |

### 1.2 Players Management

| Method | Endpoint | Purpose | Request Body | Response | Notes |
|--------|----------|---------|--------------|----------|-------|
| POST | `/api/import` | Bulk import players from Excel | `multipart/form-data: file` | `{imported: number}` | Max 32MB file |
| GET | `/api/players` | List all players (sorted by name) | None | `Player[]` | Returns formatted display names |
| PUT | `/api/players/{id}` | Update player details | `UpdatePlayerRequest` | `Player` | Uses UUID path param |
| DELETE | `/api/players/{id}` | Delete player & related matches | None | `204 No Content` | Cascading delete |
| GET | `/api/players/{id}/buddies` | Get buddy preferences for player | None | `string[]` (UUIDs) | Array of buddy player IDs |
| PUT | `/api/players/{id}/buddies` | Set buddy preferences | `{buddyIds: string[]}` | `204 No Content` | Overwrites existing buddies |

**Player Data Structure:**
```typescript
{
  id: uuid,
  nr: string,          // Member number
  name: string,        // Display format: "FirstName LastName"
  email: string,
  sponsor: string,
  address: string,
  postalCode: string,
  city: string,
  phone: string,
  mobile: string,
  memberSince: string, // Date string
  class: string        // Skill class/division
}
```

### 1.3 Schedule Management

| Method | Endpoint | Purpose | Request Body | Response | Notes |
|--------|----------|---------|--------------|----------|-------|
| POST | `/api/schedule/generate` | Generate new competition schedule | `GenerateScheduleRequest` | `Schedule` | Full round-robin + SA optimization |
| GET | `/api/schedule` | Get latest/active schedule | None | `Schedule` | Shortcut for frontend |
| GET | `/api/schedules` | List all schedules (lightweight) | None | `SeasonSummary[]` | No matches included |
| GET | `/api/schedules/{id}` | Get full schedule by ID | None | `Schedule` | Includes all evenings & matches |
| GET | `/api/schedule/evening/{id}` | Get single evening with matches | None | `Evening` | Filter evenings client-side from schedule |
| GET | `/api/schedules/{id}/info` | Get schedule metadata for info page | None | `ScheduleInfoResult` | Players, evenings, buddy pairs, matrix |
| PATCH | `/api/schedules/{id}` | Rename competition | `{competitionName: string}` | `204 No Content` | PATCH for semantic correctness |
| DELETE | `/api/schedules/{id}` | Delete entire schedule | None | `204 No Content` | Cascading: deletes evenings & matches |
| POST | `/api/schedules/{id}/inhaal-avond` | Add catch-up evening | `{date: "YYYY-MM-DD"}` | `Schedule` | For rescheduled matches |
| DELETE | `/api/schedules/{id}/evenings/{eveningId}` | Delete single evening | None | `204 No Content` | Removes all matches in evening |
| POST | `/api/schedules/import-season` | Import historical season data | `multipart/form-data: file, competitionName, season` | `Schedule` | Creates new schedule from Excel |

**GenerateScheduleRequest:**
```typescript
{
  competitionName: string,     // e.g., "Liga 2026"
  season: string,              // e.g., "2025-2026"
  numEvenings: number,         // Total slots including catch-up & skipped
  startDate: "YYYY-MM-DD",     // ISO 8601 date
  intervalDays: number,        // Days between evenings (default: 7)
  inhaalNrs: number[],         // Slot numbers for catch-up evenings
  vrijeNrs: number[]           // Slot numbers to skip entirely
}
```

**Schedule Data Structure:**
```typescript
{
  id: uuid,
  competitionName: string,
  season: string,
  createdAt: ISO 8601 timestamp,
  evenings: Evening[]
}

Evening: {
  id: uuid,
  number: number,              // 1-based sequential index
  date: ISO 8601 date,
  isInhaalAvond: boolean,     // Catch-up evening (no pre-assigned matches)
  matches: Match[]
}

Match: {
  id: uuid,
  eveningId: uuid,
  playerA: uuid,
  playerB: uuid,
  scoreA: number | null,       // null until played
  scoreB: number | null,
  played: boolean,

  // Leg details (empty string if not played)
  leg1Winner: uuid | "",
  leg1Turns: number,
  leg2Winner: uuid | "",
  leg2Turns: number,
  leg3Winner: uuid | "",
  leg3Turns: number,

  // Administrative
  reportedBy: string,          // Name of who reported the score
  rescheduleDate: string,      // Date match was rescheduled to
  secretaryNr: string,         // Member number of secretary
  counterNr: string,           // Member number of counter/referee
  playedDate: string,          // Actual date match was played

  // Statistics (0 = not recorded)
  playerA180s: number,         // 180s thrown by player A
  playerB180s: number,
  playerAHighestFinish: number,
  playerBHighestFinish: number
}
```

### 1.4 Match Scoring

| Method | Endpoint | Purpose | Request Body | Response | Notes |
|--------|----------|---------|--------------|----------|-------|
| PUT | `/api/matches/{id}/score` | Submit match result | `SubmitScoreRequest` | `204 No Content` | Detailed leg tracking |
| POST | `/evenings/{id}/report-absent` | Mark player absent (cancel open matches) | `{playerId: uuid, reportedBy: string}` | `204 No Content` | Sets all pending matches to reported |

**SubmitScoreRequest:**
```typescript
{
  leg1Winner: uuid,                // UUID of leg 1 winner or ""
  leg1Turns: number,               // Turns taken in leg 1
  leg2Winner: uuid,
  leg2Turns: number,
  leg3Winner: uuid,                // Empty if 2-leg match
  leg3Turns: number,
  reportedBy: string,              // Name of person reporting
  rescheduleDate: "YYYY-MM-DD",    // If match postponed, when to reschedule
  secretaryNr: string,             // Member number
  counterNr: string,               // Member number
  playerA180s: number,             // 180s statistics
  playerB180s: number,
  playerAHighestFinish: number,
  playerBHighestFinish: number,
  playedDate: "YYYY-MM-DD"         // When match actually took place
}
```

### 1.5 Statistics & Standings

| Method | Endpoint | Purpose | Request Query | Response | Notes |
|--------|----------|---------|----------------|----------|-------|
| GET | `/api/stats?scheduleId={id}` | Get player standings | Optional: scheduleId UUID | `PlayerStats[]` | Optional filter by schedule |
| GET | `/api/stats/duties?scheduleId={id}` | Get secretary/counter duty stats | Optional: scheduleId UUID | `DutyStats[]` | Shows who was secretary/counter and how often |

**PlayerStats:**
```typescript
{
  player: Player,
  played: number,              // Matches played
  wins: number,
  losses: number,
  draws: number,               // 3-leg matches can draw
  pointsFor: number,           // Legs won
  pointsAgainst: number,       // Legs lost
  oneEighties: number,         // 180s thrown
  highestFinish: number,       // Best single finish
  minTurns: number,            // Minimum turns in any leg
  avgTurns: number,            // Average turns per leg
  avgScorePerTurn: number      // Average darts/leg score
}

DutyStats: {
  player: Player,
  count: number,               // Total duty assignments
  secretaryCount: number,
  counterCount: number,
  secretaryMatches: DutyMatch[],
  counterMatches: DutyMatch[]
}

DutyMatch: {
  eveningNr: number,
  playerANr: string,
  playerAName: string,         // Display format
  playerBNr: string,
  playerBName: string
}
```

### 1.6 Evening Statistics (Per-evening player records)

| Method | Endpoint | Purpose | Request | Response | Notes |
|--------|----------|---------|---------|----------|-------|
| GET | `/api/evenings/{id}/player-stats` | Get all 180s/HF for evening | None | `{playerId, oneEighties, highestFinish}[]` | Evening-scoped stats |
| PUT | `/api/evenings/{id}/player-stats/{playerId}` | Update evening stats | `{oneEighties, highestFinish}` | `204 No Content` | Upsert operation |

### 1.7 Season Statistics (Per-season/schedule player records)

| Method | Endpoint | Purpose | Request | Response | Notes |
|--------|----------|---------|---------|----------|-------|
| GET | `/api/schedules/{id}/player-stats` | Get all 180s/HF for schedule | None | `{playerId, oneEighties, highestFinish}[]` | Season-scoped stats |
| PUT | `/api/schedules/{id}/player-stats/{playerId}` | Update season stats | `{oneEighties, highestFinish}` | `204 No Content` | Upsert operation |

### 1.8 Export/Download

| Method | Endpoint | Purpose | Query Params | Response | Notes |
|--------|----------|---------|-----------------|----------|-------|
| GET | `/api/export/excel` | Download full schedule as Excel | `?scheduleId={id}` optional | `.xlsx` file | `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet` |
| GET | `/api/export/pdf` | Download full schedule as PDF | `?scheduleId={id}` optional | `.pdf` file | `application/pdf` |
| GET | `/api/export/evening/{id}/excel` | Download evening match form (Excel) | None | `.xlsx` file | Filename: `wedstrijdformulier_YYYY-MM-DD.xlsx` |
| GET | `/api/export/evening/{id}/pdf` | Download evening match form (PDF) | None | `.pdf` file | Filename: `wedstrijdformulier_YYYY-MM-DD.pdf` |
| GET | `/api/export/evening/{id}/print` | Print evening match form (HTML) | None | `text/html` | Browser-printable HTML page |

### 1.9 System

| Method | Endpoint | Purpose | Request | Response | Notes |
|--------|----------|---------|---------|----------|-------|
| GET | `/api/system/logs` | Get recent server logs | None | `{logs: string[]}` | Last 200 lines |

---

## 2. Current Documentation State Assessment

### 2.1 Existing Documentation

**Location:** `/README.md` (lines 181-315)

**Strengths:**
- Basic endpoint overview with HTTP methods and paths
- Correct endpoint listing matches actual router
- Sample request bodies for key endpoints
- Brief descriptions of what each endpoint does
- Good architectural overview section
- Clear explanation of scheduler algorithm

**Weaknesses:**
1. **No OpenAPI/Swagger specification** - No machine-readable API contract
2. **Minimal request/response details** - Missing many field descriptions
3. **No interactive try-it-out functionality** - Must test manually with curl/Postman
4. **No error code documentation** - Error responses not mapped to HTTP status codes
5. **No data type documentation** - Inline structs in handlers not formally documented
6. **No authentication documentation** - No mention of CORS or auth mechanism
7. **Limited code examples** - Only curl shown implicitly; no language-specific examples
8. **Inconsistent parameter documentation** - Some endpoints lack path/query param details
9. **No webhook/event documentation** - Catchup evening mechanics underexplained
10. **No rate limiting info** - No mention of limits or throttling
11. **Scattered across multiple files** - Handler comments don't form cohesive documentation
12. **No TypeScript type documentation** - Angular frontend types not synchronized

### 2.2 Inline Code Documentation

**In handlers:**
- Minimal godoc comments in handlers
- No parameter descriptions in handler methods
- Request/response struct tags lack documentation
- Helper functions `writeJSON`, `httpError` not documented

**In domain:**
- Good domain entity comments (e.g., `Match`, `Evening`, `Player`)
- Error definitions documented in `errors.go`
- Repository interfaces have purpose comments

**In router:**
- Excellent route overview comment in `/infra/http/server.go` (lines 1-17)
- Shows all endpoints with brief descriptions

### 2.3 Frontend Documentation

**Status:** TypeScript models exist in `frontend/src/app/models.ts` but:
- Not accessible to API users
- No OpenAPI generation from Go structs
- No synchronization between backend types and frontend

---

## 3. Recommended Documentation Approach

### 3.1 Solution Architecture

Implement a **hybrid documentation strategy**:

```
┌─────────────────────────────────────────────┐
│    OpenAPI 3.1 Specification (YAML)         │
│  - Machine readable                         │
│  - Generated from Go code + manual fixes    │
│  - Source of truth for API contract        │
└────────┬──────────────────────┬─────────────┘
         │                      │
    ┌────▼─────────┐      ┌────▼──────────────┐
    │ Swagger UI   │      │ ReDoc HTML Docs   │
    │ Interactive  │      │ Beautiful read    │
    │ try-it-out   │      │ API browser       │
    └──────────────┘      └───────────────────┘
         │                      │
         └──────────┬───────────┘
                    │
         ┌──────────▼──────────┐
         │  /docs endpoint     │
         │  Served from /web   │
         └─────────────────────┘
```

### 3.2 Documentation Technology Stack

| Component | Tool | Rationale |
|-----------|------|-----------|
| **Spec Format** | OpenAPI 3.1 (YAML) | Industry standard, Go tooling support, language-agnostic |
| **Interactive UI** | Swagger UI + ReDoc | Both embedded, zero dependencies, works offline |
| **Code Generation** | swag (go-swagger) | Auto-generates from Go code + annotations |
| **Spec Hosting** | `/docs` endpoint (embedded) | Single binary, no external deps, integrated with app |
| **Code Examples** | Embedded in YAML | Multiple languages (Go, TypeScript, curl, Python) |
| **Type Sync** | Manual Go structs → OpenAPI | Annotations + manual schema definitions |

### 3.3 Implementation Strategy

**Phase 1: OpenAPI Specification (Foundation)**
- Create `openapi.yaml` in project root
- Hand-write initial spec with all endpoints
- Use `x-` extensions for custom documentation
- Reference all Go structs with `#/components/schemas`

**Phase 2: Code Annotations**
- Add godoc comments to handlers with `@Router` tags
- Add struct field comments for JSON marshaling
- Use `swag` tool to auto-parse and validate

**Phase 3: Interactive Documentation**
- Embed Swagger UI HTML in `/web/docs`
- Mount at `/docs` endpoint in router
- Enable try-it-out feature with `/api` proxy

**Phase 4: Code Examples & Guides**
- Add code examples to OpenAPI spec (curl, Go, TypeScript, Python)
- Create integration guide markdown
- Create authentication/error handling guide

**Phase 5: Automation & Maintenance**
- CI/CD validation of OpenAPI spec
- Link checking for references
- Version management for API changes

---

## 4. Concrete Implementation Plan

### 4.1 Directory Structure & Files

```
DartScheduler/
├── openapi.yaml                    # Main OpenAPI 3.1 specification
├── docs/
│   ├── index.html                 # Swagger UI entry point
│   ├── swagger-ui.js              # Swagger UI bundle (vendored)
│   ├── redoc.js                   # ReDoc bundle (vendored)
│   └── CHANGELOG.md               # API version history & breaking changes
├── api/
│   ├── examples/
│   │   ├── go/                    # Go client examples
│   │   ├── typescript/            # Angular/TypeScript examples
│   │   ├── python/                # Python examples
│   │   └── curl/                  # Shell script examples
│   └── guides/
│       ├── QUICK_START.md         # 10-minute integration guide
│       ├── AUTHENTICATION.md      # Auth flows (CORS, headers)
│       ├── ERROR_HANDLING.md      # Error codes & recovery
│       ├── RATE_LIMITING.md       # Throttling & quotas
│       └── WEBHOOK_EVENTS.md      # Catch-up evening events
├── infra/http/
│   └── docs.go                    # Swagger endpoint handler
└── README.md                       # Updated with links to /docs

Files to modify:
├── infra/http/server.go           # Add /docs route mount
├── cmd/server/main.go             # Initialize docs handler
└── frontend/src/app/models.ts     # Keep in sync with Go types
```

### 4.2 OpenAPI Specification Structure

**File:** `/openapi.yaml`

```yaml
openapi: 3.1.0
info:
  title: DartScheduler API
  version: 1.0.0
  description: >
    Competition scheduling system for dart clubs.
    Round-robin schedule generation with optimized evening assignments.
  contact:
    name: Support
  x-logo:
    url: /images/logo.png
    altText: DartScheduler Logo

servers:
  - url: http://localhost:8080
    description: Local development
  - url: https://dartscheduler.example.com
    description: Production

tags:
  - name: Players
    description: Player management and buddy preferences
  - name: Schedules
    description: Competition schedule generation and management
  - name: Matches
    description: Match scoring and evening management
  - name: Statistics
    description: Player standings and performance metrics
  - name: Export
    description: Download schedules as Excel or PDF
  - name: System
    description: Health checks and system information

paths:
  /health:
    get:
      summary: Health check
      operationId: healthCheck
      tags: [System]
      responses:
        '200':
          description: OK
          content:
            text/plain:
              schema:
                type: string
                example: 'ok'

  /api/config:
    get:
      summary: Get application configuration
      operationId: getConfig
      tags: [System]
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  appTitle:
                    type: string
                    example: DartScheduler
                  clubName:
                    type: string
                    example: PDC Roosendaal

  /api/players:
    get:
      summary: List all players
      operationId: listPlayers
      tags: [Players]
      responses:
        '200':
          description: List of players
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Player'

    post:
      summary: Import players from Excel
      operationId: importPlayers
      tags: [Players]
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              type: object
              properties:
                file:
                  type: string
                  format: binary
      responses:
        '200':
          description: Import successful
          content:
            application/json:
              schema:
                type: object
                properties:
                  imported:
                    type: integer
                    example: 24

  # ... all other endpoints documented similarly

components:
  schemas:
    Player:
      type: object
      required: [id, nr, name]
      properties:
        id:
          type: string
          format: uuid
          description: Unique player identifier
          example: 550e8400-e29b-41d4-a716-446655440000
        nr:
          type: string
          description: Member number
          example: '5'
        name:
          type: string
          description: Display name (FirstName LastName format)
          example: Jan de Vries
        email:
          type: string
          format: email
        # ... all other fields

    Schedule:
      type: object
      properties:
        id:
          type: string
          format: uuid
        competitionName:
          type: string
        season:
          type: string
        createdAt:
          type: string
          format: date-time
        evenings:
          type: array
          items:
            $ref: '#/components/schemas/Evening'

    # ... all other schemas

  responses:
    NotFound:
      description: Resource not found
      content:
        text/plain:
          schema:
            type: string
            example: "not found"

    BadRequest:
      description: Invalid input
      content:
        text/plain:
          schema:
            type: string
            example: "invalid input"

    Conflict:
      description: Resource conflict
      content:
        text/plain:
          schema:
            type: string
            example: "match already played"

    InternalError:
      description: Server error
      content:
        text/plain:
          schema:
            type: string
            example: "internal server error"

  securitySchemes: {}  # No authentication currently; document CORS
```

### 4.3 Godoc Annotations (Go Code)

**File:** `/infra/http/server.go`

```go
// Package http registers all API routes and mounts the documentation portal.
//
// # API Overview
//
// The DartScheduler API follows RESTful conventions with JSON request/response bodies.
// All timestamps are in RFC 3339 format (ISO 8601).
//
// # Error Handling
//
// All endpoints return appropriate HTTP status codes:
// - 200 OK: Successful request
// - 204 No Content: Successful action with no return body
// - 400 Bad Request: Invalid input validation failed
// - 404 Not Found: Resource does not exist
// - 409 Conflict: Resource conflict (e.g., match already played)
// - 500 Internal Server Error: Unexpected server error
//
// Error responses are plain text error messages in the response body.
//
// # Documentation
//
// Interactive API documentation is available at /docs (Swagger UI) and /docs/redoc (ReDoc).
// OpenAPI specification is available at /docs/openapi.json
//
// # Version History
//
// See docs/CHANGELOG.md for breaking changes and new features.
package http
```

**File:** `/infra/http/handler/player_handler.go`

```go
// PlayerHandler manages player CRUD and buddy preference operations.
type PlayerHandler struct {
	uc *usecase.PlayerUseCase
}

// Import handles POST /api/import
// @Summary Import players from Excel
// @Description Upload an Excel file with player data (NL or EN format)
// @Tags Players
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Excel file containing player list"
// @Success 200 {object} map[string]int{"imported": 24}
// @Failure 400 {string} string "Invalid file format"
// @Failure 500 {string} string "Server error"
// @Router /import [post]
func (h *PlayerHandler) Import(w http.ResponseWriter, r *http.Request) {
	// ...
}
```

### 4.4 Handler Documentation (samples.go)

**File:** `/infra/http/handler/docs.go` (new)

```go
package handler

import (
	"net/http"

	"DartScheduler/web"
)

// DocsHandler serves API documentation.
type DocsHandler struct {
	// embedded Swagger UI and ReDoc assets
}

// SwaggerUI serves the interactive Swagger UI at /docs
func (h *DocsHandler) SwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(web.DocsSwaggerHTML)
}

// ReDoc serves the ReDoc API reference at /docs/redoc
func (h *DocsHandler) ReDoc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(web.DocsRedocHTML)
}

// OpenAPISpec serves the raw OpenAPI spec at /docs/openapi.json
func (h *DocsHandler) OpenAPISpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(web.DocsOpenAPIJSON)
}
```

### 4.5 Router Changes

**File:** `/infra/http/server.go` (updated)

```go
func NewRouter(
	// ... existing handlers ...
	docsH *handler.DocsHandler,
) http.Handler {
	r := chi.NewRouter()
	// ... existing middleware ...

	// Documentation endpoints
	r.Get("/docs", docsH.SwaggerUI)
	r.Get("/docs/redoc", docsH.ReDoc)
	r.Get("/docs/openapi.json", docsH.OpenAPISpec)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api", func(r chi.Router) {
		// ... existing routes ...
	})

	r.Handle("/*", web.SPAHandler())

	return r
}
```

### 4.6 Integration Guide Content

**File:** `/api/guides/QUICK_START.md`

```markdown
# DartScheduler API Quick Start

Get up and running with the DartScheduler API in 10 minutes.

## Authentication

The API uses CORS for browser-based requests. All endpoints are public (no API key required).

### CORS Headers

All responses include:
```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
```

## Your First Request

### 1. Check API Health

```bash
curl http://localhost:8080/health
# Response: ok
```

### 2. Get App Configuration

```bash
curl http://localhost:8080/api/config
# Response:
# {
#   "appTitle": "DartScheduler",
#   "clubName": "PDC Roosendaal"
# }
```

### 3. List Players

```bash
curl http://localhost:8080/api/players
# Response: [...]
```

### 4. Generate a Schedule

```bash
curl -X POST http://localhost:8080/api/schedule/generate \
  -H "Content-Type: application/json" \
  -d '{
    "competitionName": "Liga 2026",
    "season": "2025-2026",
    "numEvenings": 20,
    "startDate": "2026-04-01",
    "intervalDays": 7,
    "inhaalNrs": [5],
    "vrijeNrs": [10]
  }'
# Response: { Schedule object }
```

## Common Tasks

### Import Players from Excel

```typescript
// TypeScript/Angular example
async importPlayers(file: File): Promise<{imported: number}> {
  const formData = new FormData();
  formData.append('file', file);

  const response = await fetch('/api/players', {
    method: 'POST',
    body: formData
  });

  return response.json();
}
```

### Submit Match Score

```go
// Go example
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type SubmitScoreRequest struct {
	Leg1Winner string `json:"leg1Winner"`
	Leg1Turns  int    `json:"leg1Turns"`
	Leg2Winner string `json:"leg2Winner"`
	Leg2Turns  int    `json:"leg2Turns"`
	Leg3Winner string `json:"leg3Winner"`
	Leg3Turns  int    `json:"leg3Turns"`
	// ... other fields
}

func submitMatchScore(matchID string, score SubmitScoreRequest) error {
	body, _ := json.Marshal(score)
	resp, err := http.DefaultClient.Do(&http.Request{
		Method: "PUT",
		URL:    client.parseRequestURL(fmt.Sprintf("/api/matches/%s/score", matchID)),
		Body:   io.NopCloser(bytes.NewReader(body)),
		Header: http.Header{"Content-Type": {"application/json"}},
	})
	defer resp.Body.Close()
	return err
}
```

### Get Standings

```python
# Python example
import requests

response = requests.get('http://localhost:8080/api/stats')
stats = response.json()

for player_stat in stats:
    print(f"{player_stat['player']['name']}: {player_stat['wins']} wins")
```

## Date Format

All dates use ISO 8601 format: `YYYY-MM-DD`

## Error Responses

All errors are returned as plain text:

```
GET /api/players/invalid-uuid
HTTP/400 Bad Request
Content-Type: text/plain

invalid input: invalid UUID format
```

## Next Steps

- Read [Error Handling](./ERROR_HANDLING.md) for complete error code reference
- Check [Authentication & Security](./AUTHENTICATION.md) for CORS details
- See [Integration Guides](../examples) for language-specific examples
- Review [API Reference](/docs) for all endpoints
```

**File:** `/api/guides/ERROR_HANDLING.md`

```markdown
# Error Handling Guide

All API errors are returned as plain text with appropriate HTTP status codes.

## HTTP Status Codes

| Code | Meaning | When | Example |
|------|---------|------|---------|
| 200 | OK | Successful GET/POST/PUT | Query succeeded |
| 204 | No Content | Success with no body | DELETE, PUT (no response) |
| 400 | Bad Request | Invalid input | Malformed JSON, invalid UUID |
| 404 | Not Found | Resource doesn't exist | Player ID not in database |
| 409 | Conflict | Business logic violation | Match already played, buddy already set |
| 500 | Internal Error | Unexpected server error | Database connection failed |

## Domain Error Codes

These errors are mapped from Go domain sentinel errors:

### NotFound (404)
- Player ID does not exist
- Schedule ID does not exist
- Evening ID does not exist
- Match ID does not exist

**Example:**
```
GET /api/players/550e8400-e29b-41d4-a716-446655440000
HTTP/404
Content-Type: text/plain

not found
```

### InvalidInput (400)
- Malformed request body
- Invalid UUID format
- Missing required fields
- Invalid date format (not YYYY-MM-DD)
- Negative interval days

**Example:**
```
POST /api/schedule/generate
Body: {"startDate": "invalid"}
HTTP/400

invalid input: time parsing error
```

### AlreadyExists (409)
- Player already exists with same member number
- Buddy relationship already set

**Example:**
```
PUT /api/players/550e8400-e29b-41d4-a716-446655440000/buddies
Body: {"buddyIds": ["550e8400-e29b-41d4-a716-446655440001"]}
(when already set)
HTTP/409

already exists
```

### MatchAlreadyPlayed (409)
- Score submitted for match that already has a result
- Cannot reschedule completed match

**Example:**
```
PUT /api/matches/550e8400-e29b-41d4-a716-446655440000/score
(after already submitted)
HTTP/409

match already played
```

## Handling Errors in Code

### JavaScript/TypeScript
```typescript
async function apiCall(url: string): Promise<any> {
  const response = await fetch(url);

  if (!response.ok) {
    const error = await response.text();

    switch (response.status) {
      case 404:
        throw new Error(`Not found: ${error}`);
      case 409:
        throw new Error(`Conflict: ${error}`);
      case 400:
        throw new Error(`Bad request: ${error}`);
      default:
        throw new Error(`Server error: ${error}`);
    }
  }

  return response.json();
}
```

### Python
```python
import requests
from requests.exceptions import HTTPError

try:
    response = requests.get('http://localhost:8080/api/players/invalid')
    response.raise_for_status()
except requests.HTTPError as e:
    error_message = response.text
    print(f"Error ({response.status_code}): {error_message}")
```

### Go
```go
resp, err := http.DefaultClient.Get(url)
if err != nil {
    return err
}
defer resp.Body.Close()

if resp.StatusCode >= 400 {
    body, _ := io.ReadAll(resp.Body)
    return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
}

return json.NewDecoder(resp.Body).Decode(&result)
```

## Retry Strategy

Implement exponential backoff for transient errors (5xx):

```
1st attempt: immediate
2nd attempt: after 1 second
3rd attempt: after 2 seconds
4th attempt: after 4 seconds
5th attempt: after 8 seconds
Max 5 retries
```

Do not retry 4xx errors (invalid input, not found, conflict).
```

### 4.7 Code Examples Structure

**File:** `/api/examples/curl/import-players.sh`

```bash
#!/bin/bash
# Import players from Excel

ENDPOINT="http://localhost:8080/api/import"
EXCEL_FILE="players.xlsx"

curl -X POST \
  -F "file=@${EXCEL_FILE}" \
  "${ENDPOINT}"

# Response: { "imported": 24 }
```

**File:** `/api/examples/typescript/schedule-generation.ts`

```typescript
// Generate a new schedule using the DartScheduler API

import { HttpClient } from '@angular/common/http';

interface GenerateScheduleRequest {
  competitionName: string;
  season: string;
  numEvenings: number;
  startDate: string;  // ISO 8601: YYYY-MM-DD
  intervalDays: number;
  inhaalNrs: number[];
  vrijeNrs: number[];
}

interface Schedule {
  id: string;
  competitionName: string;
  season: string;
  createdAt: string;
  evenings: Evening[];
}

interface Evening {
  id: string;
  number: number;
  date: string;
  isInhaalAvond: boolean;
  matches: Match[];
}

interface Match {
  id: string;
  eveningId: string;
  playerA: string;
  playerB: string;
  scoreA: number | null;
  scoreB: number | null;
  played: boolean;
  leg1Winner: string;
  leg1Turns: number;
  leg2Winner: string;
  leg2Turns: number;
  leg3Winner: string;
  leg3Turns: number;
}

export class ScheduleService {
  constructor(private http: HttpClient) {}

  generateSchedule(request: GenerateScheduleRequest): Promise<Schedule> {
    return this.http
      .post<Schedule>('/api/schedule/generate', request)
      .toPromise()
      .then(schedule => {
        console.log(`Generated schedule with ${schedule.evenings.length} evenings`);
        return schedule;
      })
      .catch(err => {
        console.error('Failed to generate schedule', err);
        throw err;
      });
  }

  getLatestSchedule(): Promise<Schedule> {
    return this.http
      .get<Schedule>('/api/schedule')
      .toPromise();
  }

  submitScore(matchId: string, score: SubmitScoreRequest): Promise<void> {
    return this.http
      .put<void>(`/api/matches/${matchId}/score`, score)
      .toPromise();
  }
}
```

### 4.8 CI/CD Integration

**File:** `.github/workflows/api-docs.yml` (new)

```yaml
name: API Documentation

on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install openapi-generator
        run: npm install -g @openapitools/openapi-generator-cli

      - name: Validate OpenAPI spec
        run: |
          openapi-generator-cli validate -i openapi.yaml

      - name: Check spec links
        run: |
          go run ./cmd/lint-openapi ./openapi.yaml

      - name: Generate Go client (dry-run)
        run: |
          openapi-generator-cli generate \
            -i openapi.yaml \
            -g go \
            --package-name=dartscheduler \
            -o /tmp/client-test

      - name: Upload spec artifact
        uses: actions/upload-artifact@v4
        with:
          name: openapi-spec
          path: openapi.yaml
```

---

## 5. Detailed Step-by-Step Implementation Plan

### Step 1: Create OpenAPI Specification (Weeks 1-2)

**Deliverable:** Complete `openapi.yaml` file

**Tasks:**
1. Write OpenAPI 3.1 base structure (info, servers, tags)
2. Document all 27 endpoints with:
   - Summary and description
   - All path/query/body parameters
   - Request/response schemas
   - Example values
   - Error responses
3. Define all 12 data schemas (Player, Schedule, Match, Evening, etc.)
4. Add common response definitions (404, 400, 409, 500)
5. Add security schemes section (CORS documentation)
6. Validate syntax with online OpenAPI validator
7. Version control: commit to `main` branch

**Files to Create:**
- `/openapi.yaml` (1200+ lines)

**Files to Modify:**
- None at this stage

**Testing:**
- Validate with `openapi-generator-cli validate openapi.yaml`
- Verify all endpoints match `server.go` routes
- Check all schemas are referenced

---

### Step 2: Prepare Documentation Assets (Week 2)

**Deliverable:** HTML documentation portal with Swagger UI & ReDoc

**Tasks:**
1. Download Swagger UI v5.x distribution
   - `swagger-ui.js`, `swagger-ui.css`
   - Store in `/web/docs/`
2. Download ReDoc v2.x distribution
   - `redoc.js`
   - Store in `/web/docs/`
3. Create Swagger UI HTML wrapper (`/web/docs/index.html`)
4. Create ReDoc HTML wrapper (`/web/docs/redoc.html`)
5. Embed OpenAPI spec as JSON in Go code
6. Create docs handler in `infra/http/handler/docs.go`

**Files to Create:**
- `/docs/index.html` (Swagger UI entry point)
- `/docs/redoc.html` (ReDoc entry point)
- `/docs/openapi.yaml` (source spec for reference)
- `/infra/http/handler/docs.go` (HTTP handlers)

**Files to Modify:**
- `/web/embed.go` (add //go:embed directive for docs assets)
- `/infra/http/server.go` (add /docs routes)
- `/cmd/server/main.go` (inject docs handler)

---

### Step 3: Add Code Annotations (Week 3)

**Deliverable:** Godoc comments for all handlers

**Tasks:**
1. Add package-level comment to `infra/http/server.go`
   - API overview
   - Error handling guide
   - Version history
   - Documentation links
2. Add handler method comments with `@Router` tags
   - Existing pattern: search for swagger examples in Go community
3. Add struct field comments (already done for domain entities)
4. Add repository interface documentation

**Files to Modify:**
- `/infra/http/server.go` (package comment)
- `/infra/http/handler/player_handler.go`
- `/infra/http/handler/schedule_handler.go`
- `/infra/http/handler/score_handler.go`
- `/infra/http/handler/stats_handler.go`
- `/infra/http/handler/export_handler.go`
- `/infra/http/handler/evening_stat_handler.go`
- `/infra/http/handler/season_stat_handler.go`
- `/infra/http/handler/config_handler.go`
- `/infra/http/handler/system_handler.go`
- `/domain/repository.go`

---

### Step 4: Create Documentation Guides (Week 4)

**Deliverable:** 5 markdown guides

**Tasks:**
1. Write `/api/guides/QUICK_START.md`
   - Import data walkthrough
   - Generate schedule example
   - Submit scores example
   - Export results example
2. Write `/api/guides/ERROR_HANDLING.md`
   - All error codes with examples
   - Language-specific error handling
   - Retry strategies
3. Write `/api/guides/AUTHENTICATION.md`
   - CORS policy
   - No API key required
   - Security considerations
4. Write `/api/guides/WEBHOOK_EVENTS.md` (if applicable)
   - Catch-up evening mechanics
   - Score submission flow
5. Update `/README.md`
   - Add "API Documentation" section
   - Link to `/docs` endpoint
   - Remove old endpoint table

**Files to Create:**
- `/api/guides/QUICK_START.md`
- `/api/guides/ERROR_HANDLING.md`
- `/api/guides/AUTHENTICATION.md`
- `/api/guides/WEBHOOK_EVENTS.md`
- `/docs/CHANGELOG.md`

**Files to Modify:**
- `/README.md` (update with documentation links)

---

### Step 5: Create Code Examples (Week 4-5)

**Deliverable:** 16+ code examples across 4 languages

**Tasks:**
1. Go examples (3 examples):
   - `/api/examples/go/client.go` - Full client wrapper
   - `/api/examples/go/schedule-gen.go` - Schedule generation
   - `/api/examples/go/score-submission.go` - Score submission
2. TypeScript examples (3 examples):
   - `/api/examples/typescript/models.ts` - Type definitions
   - `/api/examples/typescript/schedule-service.ts` - Service wrapper
   - `/api/examples/typescript/score-submission.ts` - Component usage
3. Python examples (3 examples):
   - `/api/examples/python/client.py` - Requests-based client
   - `/api/examples/python/import-players.py` - CSV import
   - `/api/examples/python/standings.py` - Fetch standings
4. Shell/curl examples (7+ examples):
   - `/api/examples/curl/health.sh`
   - `/api/examples/curl/import-players.sh`
   - `/api/examples/curl/generate-schedule.sh`
   - `/api/examples/curl/submit-score.sh`
   - `/api/examples/curl/export-excel.sh`
   - `/api/examples/curl/standings.sh`
   - `/api/examples/curl/buddy-management.sh`

**Files to Create:**
- All example files listed above

---

### Step 6: Embed Documentation in App (Week 5)

**Deliverable:** `/docs` endpoint functional and styled

**Tasks:**
1. Update `infra/http/server.go`:
   - Add routes for `/docs`, `/docs/redoc`, `/docs/openapi.json`
   - Mount docs handler
2. Update `web/embed.go`:
   - Embed Swagger UI assets
   - Embed ReDoc assets
   - Embed OpenAPI spec as JSON
3. Create `infra/http/handler/docs.go`:
   - Implement SwaggerUI handler
   - Implement ReDoc handler
   - Implement OpenAPI spec handler
4. Test in browser:
   - Navigate to `http://localhost:8080/docs`
   - Try-it-out feature should work
   - ReDoc should render at `/docs/redoc`

**Files to Create:**
- `/infra/http/handler/docs.go`

**Files to Modify:**
- `/infra/http/server.go` (add /docs routes)
- `/web/embed.go` (add //go:embed for docs)
- `/cmd/server/main.go` (inject docs handler)

---

### Step 7: CI/CD Automation (Week 6)

**Deliverable:** Automated documentation validation

**Tasks:**
1. Create `.github/workflows/api-docs.yml`
   - Validate OpenAPI spec syntax
   - Check for broken spec references
   - Generate client code (dry-run)
   - Upload spec as artifact
2. Create `cmd/lint-openapi/main.go`
   - Custom validator for DartScheduler-specific rules
   - Check all endpoints are documented
   - Check all schemas are used
3. Add pre-commit hook
   - Validate spec before commit
   - Prevent invalid specs reaching main branch

**Files to Create:**
- `.github/workflows/api-docs.yml`
- `cmd/lint-openapi/main.go`
- `.githooks/pre-commit`

---

### Step 8: Deployment & Maintenance (Week 6+)

**Deliverable:** Automated spec updates and version management

**Tasks:**
1. Document API versioning strategy
   - Semantic versioning for spec (`openapi.yaml` version field)
   - Breaking change policy
   - Deprecation process
2. Create CHANGELOG template
   - What changed (endpoints/schemas)
   - How to migrate
   - When it was released
3. Set up Swagger UI theming
   - Custom logo/branding
   - Dark mode support
   - Custom CSS for club branding
4. Monitor documentation quality
   - Track which endpoints are used most
   - Monitor broken examples
   - Collect user feedback

**Files to Create:**
- `/docs/CHANGELOG.md`
- `/docs/VERSIONING.md`
- `/docs/custom-theme.css`

---

## 6. Effort & Timeline Estimate

### Summary Table

| Phase | Week | Tasks | Effort | Status |
|-------|------|-------|--------|--------|
| 1. OpenAPI Spec | 1-2 | Write complete spec | 40 hours | Not started |
| 2. Assets | 2 | Download & embed docs | 8 hours | Not started |
| 3. Annotations | 3 | Godoc comments | 12 hours | Not started |
| 4. Guides | 4 | 5 markdown docs | 20 hours | Not started |
| 5. Code Examples | 4-5 | 16+ examples | 16 hours | Not started |
| 6. Integration | 5 | Embed in app | 8 hours | Not started |
| 7. CI/CD | 6 | Automation | 12 hours | Not started |
| 8. Maintenance | 6+ | Versioning, theming | 16 hours | Not started |
| **TOTAL** | **6 weeks** | **All** | **~132 hours** | **Research Complete** |

### Critical Path

1. OpenAPI spec is prerequisite for everything else
2. Assets & integration can happen in parallel (days 8-14)
3. Documentation guides & examples can happen in parallel (days 15-28)
4. CI/CD can be done last (days 29+)

### Quick-Win Path (MVP in 2 weeks)

If you want API docs live quickly:
1. Write OpenAPI spec (7 days)
2. Embed Swagger UI (1 day)
3. Write quick-start guide (2 days)
4. Add 3-4 curl examples (1 day)

**Minimal effort:** ~44 hours → production in 2 weeks

---

## 7. Tooling Recommendations

### For Go Development

| Tool | Purpose | Installation |
|------|---------|--------------|
| `swag` | Auto-generate OpenAPI from godoc | `go install github.com/swaggo/swag/cmd/swag@latest` |
| `oapi-codegen` | Generate Go server stubs from spec | `go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest` |
| `openapi-generator` | Multi-language client generation | `npm install -g @openapitools/openapi-generator-cli` |
| `spectacle` | Static API docs generator | `npm install -g spectacle-docs` |

### For Validation & Testing

| Tool | Purpose |
|------|---------|
| `swagger-cli` | Validate OpenAPI specs | `npm install -g @apidevtools/swagger-cli` |
| `openapi-diff` | Detect breaking changes between versions | `npm install -g openapi-diff` |
| `dredd` | Test API against OpenAPI spec | `npm install -g dredd` |
| `postman` | Manual testing & example collection | Free desktop app |

### For Documentation Hosting

| Option | Cost | Notes |
|--------|------|-------|
| Embedded (current plan) | Free | Single binary, no external deps, offline-capable |
| Swagger Hub | $50/month | Cloud hosting, version management, mocking |
| SwaggerUI standalone | Free | Docker image, standalone server |
| Stoplight | $300+/month | Professional API management platform |

---

## 8. Maintenance Guidelines

### Weekly Checklist

- Monitor 404 errors in API usage (broken examples)
- Review GitHub Issues for documentation gaps
- Update CHANGELOG if API changes
- Validate spec with CI/CD pipeline

### Before Each Release

- Update `openapi.yaml` version field
- Add entry to `docs/CHANGELOG.md`
- Verify all endpoints documented
- Regenerate code examples for changed endpoints
- Test Swagger UI try-it-out feature
- Update frontend TypeScript models if schemas changed

### Documentation Debt Prevention

- Require OpenAPI spec updates in PR reviews
- Link spec changes to code changes in commits
- Automated spec validation in CI/CD
- Set up alerts for spec drift (actual behavior ≠ documented)

---

## 9. Glossary of Terms

| Term | Definition |
|------|-----------|
| **OpenAPI** | Industry-standard specification format for REST APIs |
| **Swagger UI** | Interactive UI for exploring and testing APIs |
| **ReDoc** | Beautiful, responsive API documentation viewer |
| **Godoc** | Go's documentation format (comments above functions) |
| **Schema** | Data type definition (JSON object structure) |
| **Path Parameter** | Variable in URL path (e.g., `{id}` in `/players/{id}`) |
| **Query Parameter** | Variable in query string (e.g., `?scheduleId=...`) |
| **Request Body** | JSON data sent in POST/PUT request |
| **Response Schema** | Expected structure of response JSON |
| **Status Code** | HTTP response code (200, 404, 500, etc.) |
| **Sentinel Error** | Go pattern for domain-specific error types |

---

## 10. Success Metrics

### Quality Indicators

- ✅ 100% endpoint coverage in OpenAPI spec
- ✅ 100% request/response examples included
- ✅ <5 second page load time for Swagger UI
- ✅ Zero broken links in documentation
- ✅ All error codes documented with examples
- ✅ Code examples executable and tested

### Business Metrics

- ✅ Reduce support tickets by 50% (fewer API questions)
- ✅ Reduce time-to-integration by 70% (faster onboarding)
- ✅ 90%+ developer satisfaction (survey)
- ✅ <1 day average time to resolve API issues
- ✅ <10 API documentation-related bug reports/month

---

## Conclusion

DartScheduler has a solid REST API but lacks professional-grade documentation infrastructure. This plan provides a roadmap to add:

1. **Machine-readable OpenAPI 3.1 specification** - API contract everyone can trust
2. **Interactive try-it-out portal** - Swagger UI embedded in app, no external dependencies
3. **Comprehensive guides & examples** - Quick start, error handling, multiple languages
4. **Automated validation** - CI/CD ensures spec stays in sync with code
5. **Beautiful reference docs** - ReDoc for reading, Swagger for experimenting

**Recommended approach:** Start with the MVP (2 weeks, ~44 hours) to get interactive docs live, then iterate with examples and guides over the following months.

The hybrid strategy of embedded (Swagger UI) + external reference (ReDoc) provides both interactive testing and offline-accessible documentation in a single Go binary.

---

**Document prepared:** March 29, 2026
**Status:** Research Only - No Code Changes Made
**Next Step:** Review plan and prioritize phases for implementation
