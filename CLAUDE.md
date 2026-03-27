# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Full-stack wedstrijdplanningssysteem voor een dartclub. Go-backend + Angular 17-frontend, geleverd als één binair bestand via `//go:embed`.

**Key dependencies:** `github.com/go-chi/chi/v5`, `github.com/xuri/excelize/v2`, `github.com/jung-kurt/gofpdf`, `github.com/google/uuid`, `github.com/mattn/go-sqlite3` (CGO).

## Architecture

Clean Architecture met vier hoofdlagen:

```
Angular frontend (ingebed via web/embed.go)
    ↓ HTTP/REST
infra/http/handler/   — Chi-router, handlers, middleware
    ↓
usecase/              — Bedrijfslogica, DTOs, use cases
    ↓
domain/               — Entiteiten, repository-interfaces
    ↓
infra/sqlite/         — Repository-implementaties
infra/excel/          — Excel import/export (excelize)
infra/html/           — HTML wedstrijdformulier (afdruk)
infra/pdf/            — PDF export
scheduler/            — Round-robin + simulated annealing
```

## Commands

### Backend

```bash
go run ./cmd/server/   # Start de server (poort 8080)
go build ./cmd/server/ # Bouw de binary
go test ./...          # Voer alle tests uit
go fmt ./...           # Formatteer Go-code
go vet ./...           # Statische analyse
go mod tidy            # Synchroniseer go.mod
```

### Frontend

```bash
cd frontend
npm install
npm start              # Angular dev server op poort 4200 (proxiet /api → :8080)
npm run build          # Productie-build naar web/dist/dart-scheduler/
```

### Beide samen (via Make)

```bash
make frontend   # Bouw Angular en kopieer dist naar web/
make dev        # Start Go-backend
make test       # Go-tests
make docker     # Bouw en start via docker compose
```

## Key conventions

- **Naamopmaak:** Namen zijn opgeslagen als `"Achternaam, Voornaam"`. Gebruik altijd `domain.FormatDisplayName(name)` voor weergave in UI, Excel en HTML-exports.
- **Buildvolgorde Excel:** Heights → Widths → Merges → Values → Styles (merges wissen stijlen).
- **Repository-interfaces** zitten in `domain/`; implementaties in `infra/sqlite/`.
- **DTOs** voor use-case input/output staan in `usecase/dto.go`.
- **Paginagrootte Excel-avondformulier:** 26 rijen per pagina (`rowsPerPage = 26`).
- **PDF-exporters (`infra/pdf/`):** Twee exporters:
  - `Exporter` — volledige competitie (één pagina per avond, eenvoudige match-tabel).
  - `EveningExporter` — wedstrijdformulier per avond, 1-op-1 replica van de Excel-export: A4 liggend, 25 datarijen per pagina (`rowsPerPage = 25`), zelfde 17 kolommen en breedtes. Inhaalopmerkingen worden als extra pagina's achteraan toegevoegd via `emptyExport`.

## Test coverage

| Pakket | Testbestand(en) |
|---|---|
| `scheduler/` | `scheduler_test.go`, `internals_test.go` (34 tests) |
| `infra/excel/` | `importer_test.go` (7 tests) |
| `usecase/` | `schedule_usecase_test.go`, `score_usecase_test.go` (8 tests) |
| `domain/` | `player_test.go` |
