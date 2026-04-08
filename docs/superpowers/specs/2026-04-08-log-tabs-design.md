# Log-tabs in Beheer — Design Spec

**Datum:** 2026-04-08  
**Status:** Goedgekeurd

---

## Doel

De Server-kaart in de Beheer-tab toont alle logs in één platte lijst. We splitsen dit op in drie tabs (Routing/API, Debug/Error, Info) en voegen betekenisvolle log-regels toe aan de Go-handlers.

---

## Categorisatie: prefix-based

Elke log-regel krijgt een tag aan het begin:

| Tag | Categorie | Tab |
|---|---|---|
| `[HTTP]` | HTTP requests (methode, pad, statuscode, duur) | Routing/API |
| `[ERROR]` | Fouten, encode-errors, constraint-schendingen | Debug/Error |
| `[INFO]` | Bedrijfsgebeurtenissen (import, genereren, etc.) | Info |

Regels zonder herkende tag vallen als vangnet in de Info-tab.

Geen wijzigingen aan `logbuf.Buffer` of de `/api/system/logs` response-structuur — de frontend parseert de prefix.

---

## Backend-wijzigingen

### `infra/http/middleware/logger.go`

Huidige log-regel:
```
GET /api/players 200 1.23ms
```
Wordt:
```
[HTTP] GET /api/players 200 1.23ms
```

### `infra/http/handler/helpers.go`

`writeJSON`-fout en `httpError` krijgen `[ERROR]` prefix:
```go
log.Printf("[ERROR] [writeJSON] encode error: %v", err)
log.Printf("[ERROR] status=%d err=%v", code, err)
```

### Handlers — nieuwe `[INFO]` logs

| Handler | Event | Log-regel |
|---|---|---|
| `player_handler.go` | Import voltooid | `[INFO] %d spelers geïmporteerd` |
| `schedule_handler.go` | Schema gegenereerd | `[INFO] schema gegenereerd seizoen=%q avonden=%d` |
| `schedule_handler.go` | Schema herberekend | `[INFO] schema herberekend id=%s` |
| `schedule_handler.go` | Schema verwijderd | `[INFO] schema verwijderd id=%s` |
| `schedule_handler.go` | Seizoen hernoemd | `[INFO] seizoen hernoemd id=%s naam=%q` |
| `schedule_handler.go` | Actief seizoen | `[INFO] actief seizoen ingesteld id=%s` |
| `score_handler.go` | Score ingediend | `[INFO] score ingediend wedstrijd=%s %d-%d` |

### `cmd/server/main.go`

Startup-log krijgt `[INFO]` prefix:
```
[INFO] config: port=8080 ...
[INFO] listening on :8080
```

---

## Frontend-wijzigingen (`beheer.component.ts`)

### Imports

Voeg `MatTabsModule` en `MatBadgeModule` toe aan de imports-array.

### Computed properties

```typescript
get httpLogs(): string[]  { return this.logs().filter(l => l.startsWith('[HTTP]')); }
get errorLogs(): string[] { return this.logs().filter(l => l.startsWith('[ERROR]')); }
get infoLogs(): string[]  { return this.logs().filter(l => !l.startsWith('[HTTP]') && !l.startsWith('[ERROR]')); }
```

### Template — Server-kaart

De huidige enkelvoudige `log-box` div wordt vervangen door een `mat-tab-group` met 3 tabs:

```
[Routing/API (N)] [Debug/Error (N)] [Info (N)]
┌────────────────────────────────────────────┐
│  donkere log-box met gefilterde regels     │
└────────────────────────────────────────────┘
```

### Kleurcodering per tab

**Routing/API tab** — kleur op HTTP-statuscode:
- 2xx → `#4caf50` (groen)
- 4xx → `#ff9800` (oranje)
- 5xx → `#f44336` (rood)
- Overig → standaard `#d4d4d4`

**Debug/Error tab** — alle regels in `#f44336` (rood).

**Info tab** — standaard `#d4d4d4` (lichtgrijs), geen extra kleur.

### Implementatiedetail kleurcodering

De log-box rendert regels individueel (met `@for`) zodat per-regel inline `color` gezet kan worden op basis van statuscode-parsing. Huidig: één `{{ logs().join('\n') }}` string-interpolatie — dit wordt vervangen.

HTTP-statuscode parsing: log-formaat is `[HTTP] GET /api/players 200 1.23ms`. De statuscode staat op positie 3 na `split(' ')`. Een hulpfunctie `httpLogColor(line: string): string` bepaalt de kleur op basis van `parseInt(parts[3])`.

### Tab-badges

Geen `MatBadgeModule` — het aantal wordt in het tab-label zelf gezet: `Routing/API (12)`. Computed als template-expressie: `Routing/API ({{ httpLogs.length }})`.

---

## Geen wijzigingen aan

- `infra/logbuf/logbuf.go` — geen structuurwijziging nodig
- `infra/http/handler/system_handler.go` — API-response blijft `{"logs": ["..."]}`
- Overige Angular-componenten

---

## Verificatie

```bash
go build ./...          # zero errors
go test ./...           # alle tests groen
cd frontend && npm run build   # zero warnings
```

Handmatige check:
- Refreshknop laadt logs opnieuw
- HTTP-tab toont alleen `[HTTP]`-regels met juiste kleur
- Debug/Error-tab toont alleen `[ERROR]`-regels in rood
- Info-tab toont `[INFO]`-regels + niet-getagde regels
- Badges tonen correct aantal per tab
