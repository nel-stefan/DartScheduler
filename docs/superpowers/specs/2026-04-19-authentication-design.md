# Authentication & Role-Based Access Control — Design Spec

**Date:** 2026-04-19  
**Project:** DartScheduler

---

## Overview

Add authentication and role-based access control (RBAC) to the DartScheduler application. Currently all routes and APIs are publicly accessible. After this change, all routes require a valid identity — either a JWT token (username/password login) or a trusted local network IP (192.168.x.x → auto-admin).

---

## Roles

| Role | Permissions |
|---|---|
| **viewer** | Schedule view only (`/schema`) |
| **maintainer** | Schedule + score entry + afmelden (`/schema`, `/avond/:id`, `/score/:id`) |
| **admin** | Everything: all viewer/maintainer routes + Stand, Info, Beheer, Spelers, Gebruikers, exports |

---

## Authentication Mechanism

**JWT in localStorage** (HS256, signed with a secret from env var `JWT_SECRET`).

- Token lifetime: 7 days
- Claims: `{ sub, username, role, exp }`
- Stored in `localStorage` under key `dart_token`
- Angular HTTP interceptor attaches `Authorization: Bearer <token>` to every `/api/` request
- On 401 response: clear localStorage, redirect to `/login`

---

## Network Trust Bypass

Requests from `192.168.0.0/16` are automatically granted **admin** identity without a JWT token.

- Checked via `RemoteAddr` before JWT validation
- Injected identity: `{ username: "lokaal netwerk", role: "admin" }`
- Frontend: on boot calls `GET /api/auth/me`, receives admin identity, skips login page entirely
- No token is stored; backend trusts the IP on every request

---

## Backend Changes

### New SQLite Table: `users`

```sql
CREATE TABLE users (
  id          TEXT PRIMARY KEY,
  username    TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,  -- bcrypt
  role        TEXT NOT NULL,    -- 'viewer' | 'maintainer' | 'admin'
  created_at  TEXT NOT NULL
);
```

### New Endpoints (no auth required)

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/auth/login` | `{ username, password }` → `{ token, username, role }` |
| `GET` | `/api/auth/me` | Returns current identity (network trust or JWT), or 401 |

### User Management Endpoints (admin only)

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/users` | List all users |
| `POST` | `/api/users` | Create user `{ username, password, role }` |
| `PUT` | `/api/users/:id` | Update role or password |
| `DELETE` | `/api/users/:id` | Delete user (cannot delete own account) |

### Middleware Chain (per request)

1. **Network trust check** — if `RemoteAddr` ∈ `192.168.0.0/16`: inject admin, continue
2. **JWT validation** — parse `Authorization: Bearer` header, verify signature + expiry, inject identity
3. **No identity** → 401

### Chi Route Groups by Role

- **Public** (no auth): `POST /api/auth/login`, `GET /api/auth/me`
- **viewer+** (any valid identity): GET schedule list, GET evening details, GET player list (only what the schedule view needs to render)
- **maintainer+** (role ∈ {maintainer, admin}): score submission, afmelden endpoints
- **admin only** (role = admin): standings, statistics, info endpoints, schedule generation, player management CRUD, exports, user CRUD

### Seed Command

```bash
go run ./cmd/server/ seed-admin <username> <password>
```

Creates the first admin user. Errors if an admin already exists.

---

## Frontend Changes

### New Files

| File | Purpose |
|---|---|
| `auth.service.ts` | Stores JWT, exposes `role`, `isLoggedIn`, `login()`, `logout()`, `currentUser$` |
| `auth.interceptor.ts` | Adds `Authorization: Bearer` header to all `/api/` requests if token exists |
| `auth.guard.ts` | Redirects unauthenticated users to `/login` |
| `role.guard.ts` | Redirects users without required role to `/` |
| `login.component.ts` | Login form, calls `POST /api/auth/login`, stores token, redirects |
| `users.component.ts` | Admin-only user management page |

### Route Protection

```
/login                  → no guard
/                       → AuthGuard (all authenticated users)
  /schema               → viewer+
  /avond/:id            → maintainer+
  /score/:id            → maintainer+
  /stand                → admin only
  /info                 → admin only
  /beheer               → admin only
  /spelers              → admin only
  /gebruikers           → admin only (new)
/mobile/                → AuthGuard
  /mobile/avond         → viewer+
  /mobile/score/:id     → maintainer+
  /mobile/stand         → admin only
  /mobile/stats         → admin only
```

### Bootstrap Flow

On app boot, Angular calls `GET /api/auth/me`:
- **200 (network trust)**: store `{ username: "lokaal netwerk", role: "admin" }` in AuthService, no token in localStorage, proceed to app
- **200 (JWT in localStorage)**: validate client-side expiry, store identity, proceed to app
- **401**: redirect to `/login`

### Navigation

- Navigation items hidden based on role signal from `AuthService`
- Admin sees: Schema, Stand, Info, Beheer, Spelers, Gebruikers + Logout
- Maintainer sees: Schema + Logout
- Viewer sees: Schema + Logout

### Session Behaviour

- Token key in localStorage: `dart_token`
- Client-side expiry check on boot (no API call needed if expired)
- 401 from any API call → auto-logout + redirect to `/login`
- Logout: clear `dart_token` from localStorage, redirect to `/login`

---

## User Management UI (`/gebruikers`)

- Table: Username, Role, Created At, Actions
- "Gebruiker toevoegen" button → inline form: username + password + role selector
- Per row: role dropdown (saves on change), password reset button (admin sets new password), delete button
- Delete: confirmation dialog, cannot delete own account
- Admin is identified by the JWT `sub` claim matching the user's id

---

## Login Page (`/login`)

- Centered card with username + password fields and "Inloggen" button
- Displays error on wrong credentials
- No "forgot password" — admin resets via `/gebruikers`
- Redirects to original requested route after successful login (or `/` as fallback)

---

## Security Notes

- Passwords hashed with `bcrypt` (cost 12)
- `JWT_SECRET` env var must be set in production; dev falls back to a fixed insecure default with a warning
- Network trust uses `RemoteAddr` directly — `X-Forwarded-For` and similar headers are explicitly ignored to prevent spoofing
- No refresh token; users re-login every 7 days
