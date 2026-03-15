# DartScheduler

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
| Scores invoeren | Per wedstrijd scores registreren via dialoogvenster |
| Stand bijhouden | Automatisch berekende ranglijst (gewonnen/gelijkspel/verloren) |
| Exporteren | Download schema als Excel of PDF |

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
│  POST /api/import   GET /api/players     │
│  POST /api/schedule/generate             │
│  GET  /api/schedule                      │
│  PUT  /api/matches/{id}/score            │
│  GET  /api/stats                         │
│  GET  /api/export/excel|pdf              │
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
- **Response:** `200 OK`
  ```json
  { "imported": 24 }
  ```

#### `GET /api/players`
Haal alle spelers op, gesorteerd op naam.

- **Response:** `200 OK`
  ```json
  [
    {
      "ID": "uuid",
      "Nr": "1",
      "Name": "Jan de Vries",
      "Email": "jan@example.com",
      "Sponsor": "Cafe De Kroeg",
      "Address": "Hoofdstraat 1",
      "PostalCode": "1234AB",
      "City": "Amsterdam",
      "Phone": "020-1234567",
      "Mobile": "06-12345678",
      "MemberSince": "2020"
    }
  ]
  ```

### Schema

#### `POST /api/schedule/generate`
Genereer een nieuw competitieschema. Overschrijft het bestaande schema.

- **Body:**
  ```json
  {
    "competitionName": "Liga 2026",
    "numEvenings": 20,
    "startDate": "2026-04-01",
    "intervalDays": 7
  }
  ```
- **Response:** `200 OK` — volledig `Schedule`-object met alle avonden en wedstrijden.

#### `GET /api/schedule`
Haal het meest recente schema op, inclusief alle avonden en wedstrijden.

- **Response:** `200 OK`
  ```json
  {
    "ID": "uuid",
    "CompetitionName": "Liga 2026",
    "CreatedAt": "2026-03-14T10:00:00Z",
    "Evenings": [
      {
        "ID": "uuid",
        "Number": 1,
        "Date": "2026-04-01T00:00:00Z",
        "Matches": [
          {
            "ID": "uuid",
            "EveningID": "uuid",
            "PlayerA": "uuid",
            "PlayerB": "uuid",
            "ScoreA": null,
            "ScoreB": null,
            "Played": false
          }
        ]
      }
    ]
  }
  ```

#### `GET /api/schedule/evening/{id}`
Haal één avond op met bijbehorende wedstrijden.

### Wedstrijden

#### `PUT /api/matches/{id}/score`
Sla de score op voor een wedstrijd. Een wedstrijd kan slechts één keer worden ingevoerd.

- **Body:**
  ```json
  { "scoreA": 3, "scoreB": 1 }
  ```
- **Response:** `204 No Content`
- **Fout:** `409 Conflict` als de wedstrijd al gespeeld is.

### Statistieken

#### `GET /api/stats`
Haal de ranglijst op voor alle spelers.

- **Response:** `200 OK`
  ```json
  [
    {
      "Player": { ... },
      "Played": 10,
      "Wins": 7,
      "Losses": 2,
      "Draws": 1,
      "PointsFor": 35,
      "PointsAgainst": 18
    }
  ]
  ```

### Export

#### `GET /api/export/excel`
Download het volledige schema als `.xlsx`-bestand.

- **Response:** `200 OK`, `Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`

#### `GET /api/export/pdf`
Download het volledige schema als `.pdf`-bestand.

- **Response:** `200 OK`, `Content-Type: application/pdf`

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

De applicatie is beschikbaar op [http://localhost:8080](http://localhost:8080). Het database-bestand wordt opgeslagen in een Docker-volume (`dartscheduler_data`).

### Handmatig bouwen

```bash
docker build -t dartscheduler .
docker run -p 8080:8080 -v $(pwd)/data:/data dartscheduler
```

### Multi-stage build

| Stage | Image | Doel |
|---|---|---|
| `frontend-builder` | `node:20-alpine` | Angular productie-build |
| `go-builder` | `golang:1.26-alpine` | Go-binary bouwen (CGO uitgeschakeld) |
| Runtime | `alpine:3.20` | Minimaal eindimage (~20 MB) |
