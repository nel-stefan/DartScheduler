# DartScheduler API - Technical Reference

**Date:** March 29, 2026
**Purpose:** Detailed technical reference for API documentation implementation
**Status:** Research Only

---

## 1. Complete Request/Response Schema Reference

### 1.1 Player Object

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "nr": "5",
  "name": "Jan de Vries",
  "email": "jan@example.com",
  "sponsor": "Local Bar",
  "address": "Straat 1",
  "postalCode": "4811 AB",
  "city": "Breda",
  "phone": "+31123456789",
  "mobile": "+31612345678",
  "memberSince": "2020-01-15",
  "class": "A"
}
```

**JSON Field Mapping (Go → JSON):**
| Go Field | JSON Field | Type | Required | Description |
|----------|-----------|------|----------|-------------|
| `ID` | `id` | uuid | Yes | Unique identifier |
| `Nr` | `nr` | string | Yes | Member number (unique) |
| `Name` | `name` | string | Yes | Display format: FirstName LastName |
| `Email` | `email` | string | No | Email address |
| `Sponsor` | `sponsor` | string | No | Sponsor name |
| `Address` | `address` | string | No | Street address |
| `PostalCode` | `postalCode` | string | No | Postal/ZIP code |
| `City` | `city` | string | No | City name |
| `Phone` | `phone` | string | No | Landline |
| `Mobile` | `mobile` | string | No | Mobile number |
| `MemberSince` | `memberSince` | string | No | Date string |
| `Class` | `class` | string | No | Skill division |

### 1.2 Schedule Object

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "competitionName": "Liga 2026",
  "season": "2025-2026",
  "createdAt": "2026-03-29T14:30:00Z",
  "evenings": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "number": 1,
      "date": "2026-04-01",
      "isInhaalAvond": false,
      "matches": [
        {
          "id": "550e8400-e29b-41d4-a716-446655440003",
          "eveningId": "550e8400-e29b-41d4-a716-446655440002",
          "playerA": "550e8400-e29b-41d4-a716-446655440010",
          "playerB": "550e8400-e29b-41d4-a716-446655440011",
          "scoreA": 2,
          "scoreB": 1,
          "played": true,
          "leg1Winner": "550e8400-e29b-41d4-a716-446655440010",
          "leg1Turns": 15,
          "leg2Winner": "550e8400-e29b-41d4-a716-446655440010",
          "leg2Turns": 18,
          "leg3Winner": "550e8400-e29b-41d4-a716-446655440011",
          "leg3Turns": 20,
          "reportedBy": "John Smith",
          "rescheduleDate": "",
          "secretaryNr": "1",
          "counterNr": "2",
          "playedDate": "2026-04-01",
          "playerA180s": 2,
          "playerB180s": 1,
          "playerAHighestFinish": 120,
          "playerBHighestFinish": 100
        }
      ]
    }
  ]
}
```

**Null/Empty Handling:**
- `scoreA` and `scoreB` are `null` before match is played
- `leg*Winner` are empty strings `""` if leg not played
- `rescheduleDate` is empty string `""` if not rescheduled
- `playedDate` is empty string `""` if match not yet played

### 1.3 Match Result Submission

```json
{
  "leg1Winner": "550e8400-e29b-41d4-a716-446655440010",
  "leg1Turns": 15,
  "leg2Winner": "550e8400-e29b-41d4-a716-446655440010",
  "leg2Turns": 18,
  "leg3Winner": "",
  "leg3Turns": 0,
  "reportedBy": "John Smith",
  "rescheduleDate": "",
  "secretaryNr": "1",
  "counterNr": "2",
  "playerA180s": 2,
  "playerB180s": 1,
  "playerAHighestFinish": 120,
  "playerBHighestFinish": 100,
  "playedDate": "2026-04-01"
}
```

**Validation Rules:**
- `leg1Winner` must be one of the two player UUIDs or empty string
- `leg1Turns` must be > 0 if leg1Winner is set
- If `leg3Winner` is empty, match is best-of-2 (legs 1-2 decide)
- If `leg3Winner` is set, all three legs were played
- `playerA180s`, `playerB180s` are counts (≥ 0)
- `playerAHighestFinish` is checkout score (0 = not recorded, otherwise > 0)
- `playedDate` format: `YYYY-MM-DD` or empty string

### 1.4 Player Statistics

```json
{
  "player": {
    "id": "550e8400-e29b-41d4-a716-446655440010",
    "nr": "5",
    "name": "Jan de Vries",
    "email": "jan@example.com",
    "sponsor": "",
    "address": "",
    "postalCode": "",
    "city": "",
    "phone": "",
    "mobile": "",
    "memberSince": "",
    "class": ""
  },
  "played": 18,
  "wins": 12,
  "losses": 5,
  "draws": 1,
  "pointsFor": 38,
  "pointsAgainst": 20,
  "oneEighties": 47,
  "highestFinish": 160,
  "minTurns": 12,
  "avgTurns": 18.5,
  "avgScorePerTurn": 4.2
}
```

**Calculation Rules:**
- `wins + losses + draws == played`
- `pointsFor` = total legs won
- `pointsAgainst` = total legs lost
- `minTurns` = minimum turns in any single leg (0 if not played)
- `avgTurns` = `totalTurnsAcrossAllLegs / pointsFor`
- `avgScorePerTurn` = `checkoutScore / totalTurns`

### 1.5 Duty Statistics

```json
{
  "player": { /* Player object */ },
  "count": 4,
  "secretaryCount": 2,
  "counterCount": 2,
  "secretaryMatches": [
    {
      "eveningNr": 1,
      "playerANr": "5",
      "playerAName": "Jan de Vries",
      "playerBNr": "7",
      "playerBName": "Piet Janssen"
    }
  ],
  "counterMatches": [
    {
      "eveningNr": 2,
      "playerANr": "3",
      "playerAName": "Lisa Mueller",
      "playerBNr": "4",
      "playerBName": "Anna Schmidt"
    }
  ]
}
```

### 1.6 Schedule Info (for Info page)

```json
{
  "players": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440010",
      "nr": "5",
      "name": "Jan de Vries"
    }
  ],
  "evenings": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "number": 1,
      "date": "2026-04-01"
    }
  ],
  "matrix": [
    {
      "playerId": "550e8400-e29b-41d4-a716-446655440010",
      "eveningId": "550e8400-e29b-41d4-a716-446655440002",
      "count": 2
    }
  ],
  "buddyPairs": [
    {
      "playerAId": "550e8400-e29b-41d4-a716-446655440010",
      "playerANr": "5",
      "playerAName": "Jan de Vries",
      "playerBId": "550e8400-e29b-41d4-a716-446655440011",
      "playerBNr": "7",
      "playerBName": "Piet Janssen",
      "eveningIds": [
        "550e8400-e29b-41d4-a716-446655440002",
        "550e8400-e29b-41d4-a716-446655440003"
      ],
      "eveningNrs": [1, 2]
    }
  ]
}
```

**Interpretation:**
- `matrix[i].count` = how many matches player has on evening
- `buddyPairs[].eveningIds` = evenings where buddy pair appears together
- All IDs are UUIDs in string format

### 1.7 Evening Statistics (per player, per evening)

```json
[
  {
    "playerId": "550e8400-e29b-41d4-a716-446655440010",
    "oneEighties": 2,
    "highestFinish": 120
  },
  {
    "playerId": "550e8400-e29b-41d4-a716-446655440011",
    "oneEighties": 1,
    "highestFinish": 100
  }
]
```

### 1.8 Season Statistics (per player, per schedule)

```json
[
  {
    "playerId": "550e8400-e29b-41d4-a716-446655440010",
    "oneEighties": 47,
    "highestFinish": 160
  },
  {
    "playerId": "550e8400-e29b-41d4-a716-446655440011",
    "oneEighties": 35,
    "highestFinish": 140
  }
]
```

---

## 2. HTTP Status Codes & Error Responses

### 2.1 Success Responses

| Code | Content-Type | Use Cases |
|------|--------------|-----------|
| 200 | `application/json` | GET, POST (returns data) |
| 204 | (none) | PUT, DELETE, POST (no response body) |

**Example 200:**
```
HTTP/1.1 200 OK
Content-Type: application/json

[{"id": "...", "name": "Jan de Vries"}, ...]
```

**Example 204:**
```
HTTP/1.1 204 No Content
(empty body)
```

### 2.2 Error Responses

| Code | Content-Type | Cause | Example |
|------|--------------|-------|---------|
| 400 | `text/plain` | Malformed request, validation failed | `"invalid input: time parsing error"` |
| 404 | `text/plain` | Resource not found | `"not found"` |
| 409 | `text/plain` | Business logic violation, conflict | `"match already played"` |
| 500 | `text/plain` | Unexpected server error | `"internal server error"` |

**Full Example Error Response:**
```
HTTP/1.1 400 Bad Request
Content-Type: text/plain
Content-Length: 42

invalid input: invalid UUID format
```

### 2.3 Error Mapping in Handlers

**File:** `/infra/http/handler/helpers.go` (current)

```go
func httpErrorDomain(w http.ResponseWriter, err error) {
	var code int
	switch {
	case errors.Is(err, domain.ErrNotFound):
		code = http.StatusNotFound                    // 404
	case errors.Is(err, domain.ErrInvalidInput):
		code = http.StatusBadRequest                  // 400
	case errors.Is(err, domain.ErrAlreadyExists):
		code = http.StatusConflict                    // 409
	case errors.Is(err, domain.ErrMatchAlreadyPlayed):
		code = http.StatusConflict                    // 409
	default:
		code = http.StatusInternalServerError         // 500
	}
	log.Printf("[httpErrorDomain] status=%d err=%v", code, err)
	http.Error(w, err.Error(), code)
}
```

---

## 3. Request/Response Content Types

### 3.1 JSON Endpoints

**Headers:**
```
Content-Type: application/json; charset=utf-8
Accept: application/json
```

**Encoding:**
- UTF-8
- No BOM
- Single-line arrays for compact responses
- Pretty-printed (indented) for readability

### 3.2 File Upload (Multipart)

**Headers:**
```
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary
```

**Fields:**
- `file` — binary file content
- Other form fields as strings

**Example:**
```
POST /api/import
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW

------WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="file"; filename="players.xlsx"
Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet

[binary XLSX content]
------WebKitFormBoundary7MA4YWxkTrZu0gW--
```

### 3.3 File Download

**Export as Excel:**
```
GET /api/export/excel
Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
Content-Disposition: attachment; filename="schedule.xlsx"
```

**Export as PDF:**
```
GET /api/export/pdf
Content-Type: application/pdf
Content-Disposition: attachment; filename="schedule.pdf"
```

**Print as HTML:**
```
GET /api/export/evening/{id}/print
Content-Type: text/html; charset=utf-8
(HTML content for browser printing)
```

---

## 4. Path Parameter Formats

### 4.1 UUID Parameters

All IDs are RFC 4122 UUIDs in string format:
- Format: `550e8400-e29b-41d4-a716-446655440000`
- Regex: `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`
- Case-insensitive in parsing (Go uuid.Parse accepts both cases)

**Usage:**
```
GET /api/players/{id}          → id is PlayerID UUID
GET /api/schedules/{id}        → id is ScheduleID UUID
GET /api/schedule/evening/{id} → id is EveningID UUID
PUT /api/matches/{id}/score    → id is MatchID UUID
```

### 4.2 Special Cases

**Evening ID in context:**
```
DELETE /api/schedules/{id}/evenings/{eveningId}
↑ ScheduleID           ↑ EveningID
```

**Player ID in context:**
```
PUT /api/evenings/{id}/player-stats/{playerId}
↑ EveningID         ↑ PlayerID
```

---

## 5. Query Parameter Formats

### 5.1 Schedule ID (Optional)

**Endpoint:** `/api/stats?scheduleId={id}`

- Type: UUID string
- Required: No
- Default: None (uses latest schedule)
- If provided: Returns stats for that specific schedule only

**Examples:**
```
GET /api/stats
→ Returns standings for latest schedule

GET /api/stats?scheduleId=550e8400-e29b-41d4-a716-446655440001
→ Returns standings for schedule with given ID

GET /api/stats/duties?scheduleId=550e8400-e29b-41d4-a716-446655440001
→ Returns duty stats for given schedule
```

---

## 6. Date/Time Formats

### 6.1 ISO 8601 Dates

**Format:** `YYYY-MM-DD`

**Usage:**
```json
{
  "startDate": "2026-04-01",
  "date": "2026-04-15"
}
```

**Accepted:** Only the date portion (no time)

**Parsing:**
- Go: `time.Parse("2006-01-02", dateString)`
- JavaScript: `new Date(dateString)`
- Python: `datetime.strptime(dateString, "%Y-%m-%d")`

### 6.2 ISO 8601 Timestamps

**Format:** RFC 3339 (with 'Z' timezone)

**Example:** `2026-03-29T14:30:00Z`

**Usage:**
```json
{
  "createdAt": "2026-03-29T14:30:00Z"
}
```

**Parsing:**
- Go: `time.Parse(time.RFC3339, timestamp)`
- JavaScript: `new Date(timestamp)`
- Python: `datetime.fromisoformat(timestamp.replace('Z', '+00:00'))`

### 6.3 Valid Date Examples

✅ Valid:
- `"2026-04-01"`
- `"2026-12-31"`
- `"2020-01-15"`

❌ Invalid:
- `"04/01/2026"` (wrong format)
- `"2026-4-1"` (no zero-padding)
- `"2026-13-01"` (invalid month)

---

## 7. Integer Field Constraints

### 7.1 Match Legs (Turns)

| Field | Min | Max | Valid Range | Notes |
|-------|-----|-----|-------------|-------|
| `leg1Turns` | 0 | 999 | 1-100 typical | Darts to finish |
| `leg2Turns` | 0 | 999 | 1-100 typical | Same as leg1 |
| `leg3Turns` | 0 | 999 | 1-100 typical | Only if 3-leg match |

### 7.2 Statistics

| Field | Min | Max | Notes |
|-------|-----|-----|-------|
| `oneEighties` | 0 | 999 | Count of 180s |
| `highestFinish` | 0 | 170 | Valid darts checkout (0-170) |
| `minTurns` | 0 | 999 | Minimum in any leg |
| `played` | 0 | 999 | Matches played |
| `wins` | 0 | 999 | Matches won |
| `losses` | 0 | 999 | Matches lost |
| `draws` | 0 | 999 | 3-leg draws only |

### 7.3 Schedule Generation

| Field | Min | Max | Example |
|-------|-----|-----|---------|
| `numEvenings` | 1 | 999 | 20 |
| `intervalDays` | 1 | 999 | 7 (weekly) |
| `inhaalNrs[]` | 0 | numEvenings | [5, 10, 15] |
| `vrijeNrs[]` | 0 | numEvenings | [8] |

---

## 8. Array Field Constraints

### 8.1 Buddy Preferences

```json
{
  "buddyIds": ["550e8400-e29b-41d4-a716-446655440001"]
}
```

**Constraints:**
- Can be empty array `[]` (clear all buddies)
- Typical size: 1-5 entries
- Max size: unlimited (no documented limit)
- Each entry must be valid UUID

### 8.2 Special Slot Numbers

```json
{
  "inhaalNrs": [5, 10, 15],
  "vrijeNrs": [8]
}
```

**Constraints:**
- Can be empty arrays `[]`
- Values are 1-based slot numbers
- Must be ≤ `numEvenings`
- Should not overlap (slot can't be both catch-up and skip)

---

## 9. Type Compatibility Matrix

### 9.1 Which Endpoints Accept Which Content Types

| Endpoint | GET | POST | PUT | DELETE | Accepts | Returns |
|----------|-----|------|-----|--------|---------|---------|
| `/api/players` | ✓ | ✓ | - | - | JSON | JSON |
| `/api/import` | - | ✓ | - | - | multipart/form-data | JSON |
| `/api/players/{id}` | - | - | ✓ | ✓ | JSON | JSON (PUT), - (DELETE) |
| `/api/schedule/generate` | - | ✓ | - | - | JSON | JSON |
| `/api/schedules/{id}` | ✓ | - | ✓ | ✓ | - (GET), JSON (PATCH) | JSON (GET/PATCH), - (DELETE) |
| `/api/matches/{id}/score` | - | - | ✓ | - | JSON | - (204) |
| `/api/export/*` | ✓ | - | - | - | - | Binary (xlsx/pdf) or HTML |

---

## 10. Common Integration Patterns

### 10.1 Schedule Lifecycle

```
1. POST /api/schedule/generate
   ↓ Get Schedule with empty scores
2. GET /api/schedule
   ↓ Frontend displays evenings & matches
3. PUT /api/matches/{id}/score (multiple times)
   ↓ User enters results
4. GET /api/stats
   ↓ View standings
5. GET /api/export/excel or /api/export/pdf
   ↓ Download final schedule
```

### 10.2 Player Management Flow

```
1. POST /api/import (upload Excel)
   ↓ Get import count
2. GET /api/players
   ↓ Verify import
3. PUT /api/players/{id}/buddies
   ↓ Optional: set buddy preferences
4. POST /api/schedule/generate
   ↓ Uses players & buddy preferences
```

### 10.3 Catch-up Evening Flow

```
1. Match not played initially
   → User clicks "Report Absent" or postpones
2. POST /evenings/{id}/report-absent
   → Sets match.reportedBy
3. Later: POST /api/schedules/{id}/inhaal-avond
   → Creates new catch-up evening
4. PUT /api/matches/{id}/score
   → Can assign to catch-up evening via rescheduleDate
```

---

## 11. Browser Compatibility for Swagger UI

**Swagger UI v5.x requires:**
- ES2015+ support
- Modern browsers only (no IE11)

**Supported:**
- Chrome 60+
- Firefox 55+
- Safari 12+
- Edge 79+

**Fallback for older browsers:**
- Use ReDoc instead (more compatible)
- Or provide curl/Postman examples

---

## 12. CORS Configuration Details

**Current CORS Policy** (from `/infra/http/middleware/cors.go`):

```
Access-Control-Allow-Origin: * (permissive)
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
```

**Implications:**
- ✅ Allows any origin (dev-friendly, not production-recommended)
- ✅ Supports all HTTP methods
- ✅ Allows custom headers
- ✅ No CORS preflight caching configured

**Recommended for production:**
```
Access-Control-Allow-Origin: https://dartscheduler.example.com
Access-Control-Allow-Max-Age: 3600
Access-Control-Allow-Credentials: false (no cookies)
```

---

## 13. Pagination & Filtering

**Current Implementation:** None

**All list endpoints return full results:**
- `GET /api/players` → all players
- `GET /api/schedules` → all schedules
- `GET /api/stats` → all players' stats

**Scalability Note:**
- For 100 players: negligible impact
- For 10,000 players: consider adding `?limit=50&offset=0` pagination

---

## 14. Caching Strategies

### 14.1 Recommended Cache Headers

**Immutable data (profiles, schedules):**
```
Cache-Control: public, max-age=3600
```

**Mutable data (scores, stats):**
```
Cache-Control: public, max-age=60
```

**Real-time data (active scores):**
```
Cache-Control: no-cache
```

**Current:** No cache headers set (all responses are fresh)

---

## 15. API Rate Limiting

**Current Implementation:** None

**Recommended for production:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1659897600
```

**Suggested limits:**
- General endpoints: 100 req/min
- File uploads: 10 req/min
- Export endpoints: 20 req/min

---

## 16. Logging Strategy for API Usage

**Current implementation** (from `/infra/http/middleware/logger.go`):
```
[METHOD] [PATH] [STATUS_CODE] [DURATION]
```

**Example:**
```
GET /api/players 200 15ms
POST /api/import 200 1.2s
PUT /api/matches/{id}/score 204 45ms
```

**For monitoring/analytics, consider adding:**
- Request ID (UUID)
- User agent
- Response size
- Error messages

---

## 17. Backward Compatibility Guidelines

**For future versions:**
- Never remove endpoints
- Never change response field types
- Never change request parameter types
- Always add new fields at end of objects
- Use `x-deprecated: true` in OpenAPI
- Provide migration guide in CHANGELOG

---

## 18. OpenAPI Extension Points

**Custom extensions for DartScheduler:**

```yaml
x-dart-scheduler-rules:
  - leg1Winner must be one of playerA or playerB UUIDs
  - leg1Turns must be > 0 if leg1Winner is set
  - nameFormat: "LastName, FirstName" stored, "FirstName LastName" displayed

x-examples:
  - language: go
    title: "Schedule Generation Example"
    source: "examples/go/schedule-gen.go"

x-performance:
  - endpoint: /api/export/excel
    avgTime: 2.5s
    maxTime: 10s
    memory: "50MB for 500 matches"

x-errors-extended:
  - code: SCHEDULE_CONFLICT
    status: 409
    meaning: "Schedule operation blocked by conflicting data"
    example: "Cannot delete evening with played matches"
```

---

## 19. Multi-Language Example Template

For each endpoint in OpenAPI, provide:

```yaml
x-code-samples:
  - lang: go
    source: |
      resp, _ := http.Post("http://localhost:8080/api/schedule/generate",
        "application/json", body)

  - lang: typescript
    source: |
      await fetch('/api/schedule/generate', {
        method: 'POST',
        body: JSON.stringify(request)
      })

  - lang: python
    source: |
      requests.post('http://localhost:8080/api/schedule/generate',
        json=request)

  - lang: curl
    source: |
      curl -X POST http://localhost:8080/api/schedule/generate \
        -H "Content-Type: application/json" \
        -d @request.json
```

---

## 20. Testing Against OpenAPI Spec

### 20.1 Validation Tools

**Option 1: Swagger CLI**
```bash
swagger-cli validate openapi.yaml
```

**Option 2: openapi-generator**
```bash
openapi-generator validate -i openapi.yaml
```

**Option 3: Dredd (integration testing)**
```bash
dredd openapi.yaml http://localhost:8080
```

Dredd will:
- Make real requests to each endpoint
- Compare responses against spec
- Report mismatches

### 20.2 Automated Testing Setup

**GitHub Actions:**
```yaml
- name: Validate spec
  run: swagger-cli validate openapi.yaml

- name: Test API against spec
  run: dredd openapi.yaml http://localhost:8080

- name: Generate client
  run: openapi-generator-cli generate -i openapi.yaml -g go
```

---

## Conclusion

This technical reference provides the detailed specifications needed to implement comprehensive API documentation. Key takeaways:

1. **All endpoints return plain-text errors on 4xx/5xx**
2. **All data is JSON (except file uploads/downloads)**
3. **All IDs are UUID strings, all dates are ISO 8601**
4. **No pagination/filtering currently implemented**
5. **No rate limiting or caching headers**
6. **CORS is permissive (suitable for dev, not production)**

For documentation implementation, focus on:
- OpenAPI 3.1 spec completeness
- Accurate request/response examples
- Error code documentation
- Language-specific code examples
- Interactive try-it-out functionality
