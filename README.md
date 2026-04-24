# DartScheduler


[![CI](https://github.com/nel-stefan/DartScheduler/actions/workflows/ci.yml/badge.svg)](https://github.com/nel-stefan/DartScheduler/actions/workflows/ci.yml)

Wedstrijdplanning voor een dartclub. Importeer de ledenlijst, genereer een competitieschema op basis van round-robin met simulated annealing, voer scores in en exporteer het resultaat naar Excel of PDF.

## Inhoud

- [Overzicht](#overzicht)
- [Architectuur](#architectuur)
- [Snel starten](#snel-starten)
- [Ontwikkeling](#ontwikkeling)
- [API-referentie](#api-referentie)
- [Scheduler-algoritme](#scheduler-algoritme)
- [Excel-importformaat](#excel-importformaat)
- [Docker](#docker)

---

## Overzicht

DartScheduler bestaat uit een Go-backend en een Angular-frontend die samen als één binair bestand worden geleverd. De frontend wordt ingesloten in het Go-binary via `//go:embed` en er is geen aparte webserver nodig.

**Functionaliteit:**

| Functie | Beschrijving |
|---|---|
| Spelers importeren | Upload een Excel-ledenlijst (NL of EN formaat) |
| Schema genereren | Round-robin schema met simulated annealing optimalisatie |
| Scores invoeren | Per wedstrijd leg-winnaar, beurten, 180's en highest finish registreren |
| Stand bijhouden | Automatisch berekende ranglijst (gewonnen/gelijkspel/verloren) per klasse |
| Statistieken | Beurten-records, 180's, highest finish, schrijver/teller-overzicht per speler |
| Exporteren | Schema als Excel of PDF; avondformulier als Excel én HTML (afdruk) |
| Mobiele UI | Aparte mobiele interface op `/m/*` voor score-invoer onderweg |
| Naam­opmaak | Namen altijd weergegeven als "Voornaam Achternaam" (opgeslagen als "Achternaam, Voornaam") |

---

## Architectuur

Het project volgt een gelaagde Clean Architecture:

```
┌─────────────────────────────────────────┐
│              Angular Frontend            │
│  (ingebed in Go-binary via web/embed.go) │
└────────────────────┬────────────────────┘
                     │ HTTP / REST
┌────────────────────▼────────────────────┐
│           infra/http (handlers)          │
│  POST /api/import      GET /api/players  │
│  POST /api/schedule/generate             │
│  GET  /api/schedules   GET /api/schedule │
│  PUT  /api/matches/{id}/score            │
│  GET  /api/stats       /api/stats/duties │
│  GET  /api/export/excel|pdf|evening/…    │
└────────────────────┬────────────────────┘
                     │
┌────────────────────▼────────────────────┐
│              usecase (bedrijfslogica)    │
│  PlayerUseCase  ScheduleUseCase          │
│  ScoreUseCase   ExportUseCase            │
└──────────┬─────────────────┬────────────┘
           │                 │
┌──────────▼──────┐  ┌───────▼────────────┐
│  domain         │  │  scheduler          │
│  (entiteiten &  │  │  (round-robin +     │
│   interfaces)   │  │   simul. annealing) │
└──────────┬──────┘  └────────────────────┘
           │
┌──────────▼────────────────────────────────┐
│  infra/sqlite   infra/excel   infra/pdf    │
│  (repositories) (importer/exporter)        │
└────────────────────────────────────────────┘
```

### Laagbeschrijvingen

| Laag | Pakket | Verantwoordelijkheid |
|---|---|---|
| Domain | `domain/` | Entiteiten (`Player`, `Match`, `Evening`, `Schedule`), repository-interfaces, domeinfouten |
| Scheduler | `scheduler/` | Round-robin pairings genereren en optimaliseren via simulated annealing |
| Use cases | `usecase/` | Orchestratie van domeinlogica; onafhankelijk van infrastructuur |
| HTTP | `infra/http/` | Chi-router, handlers, CORS- en logging-middleware |
| SQLite | `infra/sqlite/` | Repository-implementaties, schema-migraties |
| Excel/PDF | `infra/excel/`, `infra/pdf/` | Import en export van bestanden |
| Web | `web/` | Inbedden van Angular-dist via `//go:embed` |
| Server | `cmd/server/` | Applicatie-entry point; dependency injection |

---

## Snel starten

### Vereisten

- Go 1.26+
- Node.js 20+ en npm (alleen voor frontend-ontwikkeling)

### Backend draaien (met ingebedde frontend)

```bash
# Bouw de frontend en kopieer naar web/dist
make frontend

# Start de server (poort 8080)
make dev
# of direct:
go run ./cmd/server/
```

Open daarna [http://localhost:8080](http://localhost:8080) in de browser.

### Omgevingsvariabelen

| Variabele | Standaard | Beschrijving |
|---|---|---|
| `DATABASE_PATH` | `dartscheduler.db` | Pad naar het SQLite-databasebestand |
| `PORT` | `8080` | Luisterpoort voor de HTTP-server |

---

## Ontwikkeling

### Frontend apart draaien (hot reload)

```bash
cd frontend
npm install
npm start          # Angular dev server op poort 4200
```

De Angular dev server proxiet `/api`-verzoeken naar `localhost:8080`. Zorg dat de Go-backend ook draait.

### Handige make-targets

```bash
make dev        # Start Go-backend
make build      # Bouw dartscheduler-binary
make test       # Voer Go-tests uit
make fmt        # Formatteer Go-code
make vet        # Statische analyse
make tidy       # Synchroniseer go.mod
make frontend   # Bouw Angular en kopieer dist naar web/
make docker     # Bouw en start via docker compose
```

### Tests draaien

```bash
go test ./...
# Of alleen de scheduler-tests:
go test ./scheduler/...
```

### Projectstructuur

```
DartScheduler/
├── cmd/server/         # main.go — entry point en dependency injection
├── domain/             # Domeinentiteiten en repository-interfaces
├── scheduler/          # Round-robin + simulated annealing algoritme
├── usecase/            # Bedrijfslogica (use cases en DTOs)
├── infra/
│   ├── http/           # Chi-router, handlers, middleware
│   ├── sqlite/         # SQLite repository-implementaties en schema
│   ├── excel/          # Excel importer en exporter
│   └── pdf/            # PDF exporter
├── web/                # go:embed wrapper voor de Angular-dist
├── frontend/           # Angular 17-applicatie
│   └── src/app/
│       ├── components/ # overview, upload, standings, score-entry, evening-view
│       ├── services/   # player, schedule, score, export
│       └── models.ts   # TypeScript interfaces
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

---

## API-referentie

Alle endpoints beginnen met `/api`.

### Spelers

#### `POST /api/import`
Importeer spelers vanuit een Excel-bestand.

- **Content-Type:** `multipart/form-data`
- **Veld:** `file` — `.xlsx` of `.xls` bestand
- **Response:** `200 OK` — `{ "imported": 24 }`

#### `GET /api/players`
Haal alle spelers op, gesorteerd op naam.

#### `PUT /api/players/{id}`
Werk een speler bij.

#### `DELETE /api/players/{id}`
Verwijder een speler en al zijn gekoppelde wedstrijden.

#### `GET /api/players/{id}/buddies`
Haal de buddyvoorkeur op voor een speler.

#### `PUT /api/players/{id}/buddies`
Stel de buddyvoorkeur in voor een speler.

### Schema

#### `POST /api/schedule/generate`
Genereer een nieuw competitieschema.

- **Body:**
  ```json
  {
    "competitionName": "Liga 2026",
    "season": "2025-2026",
    "numEvenings": 20,
    "startDate": "2026-04-01",
    "intervalDays": 7,
    "inhaalNrs": [5],
    "vrijeNrs": [10]
  }
  ```
- **Response:** `200 OK` — volledig `Schedule`-object.

#### `GET /api/schedules`
Overzicht van alle schema's (zonder wedstrijden).

#### `GET /api/schedules/{id}`
Haal één schema op met alle avonden en wedstrijden.

#### `GET /api/schedules/{id}/info`
Haal schema-info op: spelermatrix, buddy-paren (voor de Info-pagina).

#### `GET /api/schedule`
Haal het meest recente schema op (shortcut voor de frontend).

#### `GET /api/schedule/evening/{id}`
Haal één avond op met bijbehorende wedstrijden.

#### `DELETE /api/schedules/{id}`
Verwijder een schema inclusief alle avonden en wedstrijden.

#### `POST /api/schedules/{id}/inhaal-avond`
Voeg een inhaalavond toe aan een bestaand schema.

#### `DELETE /api/schedules/{id}/evenings/{eveningId}`
Verwijder één avond uit een schema.

#### `POST /api/schedules/import-season`
Importeer historische seizoensdata vanuit een Excel-bestand.

### Wedstrijden

#### `PUT /api/matches/{id}/score`
Sla de score op voor een wedstrijd.

- **Body:**
  ```json
  {
    "leg1Winner": "uuid-speler-a",
    "leg1Turns": 15,
    "leg2Winner": "uuid-speler-a",
    "leg2Turns": 18,
    "leg3Winner": "",
    "leg3Turns": 0,
    "reportedBy": "5 Jan",
    "rescheduleDate": "",
    "secretaryNr": "5",
    "counterNr": "7",
    "playerA180s": 1,
    "playerB180s": 0,
    "playerAHighestFinish": 120,
    "playerBHighestFinish": 60
  }
  ```
- **Response:** `204 No Content`

#### `POST /api/evenings/{id}/report-absent`
Markeer alle openstaande wedstrijden voor een speler op een avond als afgemeld.

- **Body:** `{ "playerId": "uuid", "reportedBy": "5 Jan" }`
- **Response:** `204 No Content`

### Statistieken

#### `GET /api/stats?scheduleId={id}`
Haal de ranglijst op per speler (gewonnen/verloren/gelijkspel, legs, 180's, highest finish).

#### `GET /api/stats/duties?scheduleId={id}`
Haal schrijver/teller-statistieken op per speler, inclusief per-wedstrijd detail.

### Export

#### `GET /api/export/excel?scheduleId={id}`
Download het volledige schema als `.xlsx`-bestand.

#### `GET /api/export/pdf?scheduleId={id}`
Download het volledige schema als `.pdf`-bestand.

#### `GET /api/export/evening/{id}/excel`
Download het wedstrijdformulier voor één avond als `.xlsx` (26 rijen per pagina, afdrukbaar).

#### `GET /api/export/evening/{id}/print`
Open een HTML-afdrukpagina voor één avond (zelfde layout als het Excel-formulier).

### Systeem

#### `GET /api/system/logs`
Haal de laatste 200 serverlogregels op.

#### `GET /health`
Eenvoudige health-check: retourneert `200 ok`.

---

## Scheduler-algoritme

Het schema wordt in vier stappen opgebouwd:

### 1. Round-robin pairings

De canonieke cirkel-methode genereert voor `N` spelers exact `N-1` rondes waarbij elke speler precies één keer per ronde speelt. Bij een oneven aantal spelers wordt een virtuele BYE-speler toegevoegd; de speler die tegen BYE is ingedeeld, speelt die ronde niet.

### 2. Greedy initiële toewijzing

Alle `N*(N-1)/2` wedstrijden worden gelijkmatig verdeeld over de beschikbare avonden. Dit dient als startpunt voor de optimalisatie.

### 3. Simulated annealing optimalisatie

Het algoritme verbetert de toewijzing door willekeurig twee wedstrijden tussen avonden te verwisselen en de energie te berekenen:

```
energie = wViolation × overtredingen
        + wBuddy    × (1 − buddy_tevredenheid)
        + wVariance × variantie_wedstrijden_per_avond
```

| Parameter | Waarde | Betekenis |
|---|---|---|
| `wViolation` | 1000 | Zwaarste straf: speler speelt > 3 wedstrijden per avond |
| `wBuddy` | 100 | Beloning voor buddy-koppels die op dezelfde avond spelen |
| `wVariance` | 1 | Straf voor ongelijke verdeling van wedstrijden per avond |
| `steps` | 50 000 | Aantal iteraties |
| `T₀ → Tend` | 10 → 0.001 | Temperatuurschema (geometrische afkoeling) |

Slechte verwisselingen worden geaccepteerd met kans `exp(-Δenergy / T)`, waardoor het algoritme lokale optima kan ontsnappen.

### 4. Schema opbouwen

Uit de geoptimaliseerde toewijzing worden `domain.Evening`- en `domain.Match`-objecten aangemaakt met UUIDs en datums.

---

## Excel-importformaat

De importer herkent twee formaten automatisch op basis van de kolomnamen in de eerste rij:

### Nederlands formaat (ledenlijst)

| Kolomnaam | Veld |
|---|---|
| `nr` | Lidnummer |
| `Naam` | Naam (verplicht) |
| `Adres` | Adres |
| `Pc` / `Postcode` | Postcode |
| `Woonpl.` / `Woonplaats` | Woonplaats |
| `Telefoon` | Telefoonnummer |
| `Mobiel` | Mobiel |
| `E-mail adres` | E-mailadres |
| `Lid Sinds` | Lid sinds |

### Engels formaat (legacy)

| Kolomnaam | Veld |
|---|---|
| `Name` | Naam (verplicht) |
| `Email` | E-mailadres |
| `Sponsor` | Sponsor |

Rijen zonder naam worden overgeslagen. Een nieuw import vervangt alle bestaande spelers.

---

## Docker

### Starten met docker compose

```bash
docker compose up --build
```

### Herstart (stop → build → start)

```bash
docker compose -f docker-compose.yml down && \
  docker compose -f docker-compose.yml build && \
  docker compose -f docker-compose.yml up
```

De applicatie is beschikbaar op [http://localhost:8080](http://localhost:8080). Het database-bestand wordt opgeslagen in een Docker-volume (`dartscheduler_data`).

### Handmatig bouwen

```bash
docker build -t dartscheduler .
docker run -p 8080:8080 -v $(pwd)/data:/data dartscheduler
```

### Database backup

De database staat in het Docker volume `dart_data` op `/data/dartscheduler.db`.

**Bestand kopiëren (terwijl de container draait):**
```bash
docker compose cp dartscheduler:/data/dartscheduler.db ./backup-$(date +%Y%m%d).db
```

**SQL-dump (tekst, makkelijk te lezen/importeren):**
```bash
docker compose exec dartscheduler sh -c "sqlite3 /data/dartscheduler.db .dump" > backup-$(date +%Y%m%d).sql
```

**Terugzetten:**
```bash
# Vanuit bestand
docker compose cp ./backup-20260316.db dartscheduler:/data/dartscheduler.db

# Vanuit SQL-dump
docker compose exec -T dartscheduler sh -c "sqlite3 /data/dartscheduler.db" < backup-20260316.sql
```

### Multi-stage build

| Stage | Image | Doel |
|---|---|---|
| `frontend-builder` | `node:20-alpine` | Angular productie-build |
| `go-builder` | `golang:1.26-alpine` | Go-binary bouwen (CGO uitgeschakeld) |
| Runtime | `alpine:3.20` | Minimaal eindimage (~20 MB) |
