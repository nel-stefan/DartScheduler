# API Documentation Plan - Executive Summary

**Project:** DartScheduler
**Date:** March 29, 2026
**Status:** Research & Planning Complete (No Code Changes Made)
**Prepared for:** Development Team Review

---

## Quick Facts

- **API Endpoints:** 27 public endpoints
- **Data Resources:** 5 main (Player, Schedule, Match, Evening, Stats)
- **Current Documentation:** README section (basic coverage)
- **Documentation Gaps:** Interactive docs, error codes, code examples, OpenAPI spec
- **Effort to Complete:** 6 weeks (132 hours) or 2 weeks MVP (44 hours)
- **Recommended Approach:** Hybrid (embedded Swagger UI + ReDoc + guides)

---

## What We Found

### API Structure (Complete Inventory)

**Players (6 endpoints):**
- Import from Excel, list, update, delete, manage buddy preferences

**Schedules (9 endpoints):**
- Generate, list, get by ID, rename, delete, add catch-up evenings
- Import historical season data

**Matches & Scoring (2 endpoints):**
- Submit match scores with leg details, report player absence

**Statistics (2 endpoints):**
- Player standings, duty assignments (secretary/counter roles)

**Evening & Season Stats (4 endpoints):**
- Per-evening and per-season player statistics (180s, highest finish)

**Export (5 endpoints):**
- Download schedules as Excel or PDF, download evening forms

**System (1 endpoint):**
- Health check, app configuration, server logs

### Current Documentation State

**Strengths:**
- ✅ Correct endpoint listing in README
- ✅ Good architectural overview
- ✅ Clear scheduler algorithm explanation
- ✅ Decent godoc in domain layer

**Critical Gaps:**
- ❌ No OpenAPI/Swagger specification
- ❌ No interactive try-it-out functionality
- ❌ Error codes not documented
- ❌ Field-level documentation missing
- ❌ No code examples (except brief README snippets)
- ❌ No multi-language examples
- ❌ No authentication documentation
- ❌ Frontend types not synchronized

---

## Recommended Solution

### Architecture: Hybrid Documentation Stack

```
┌─────────────────────────┐
│   OpenAPI 3.1 Spec      │ Machine-readable contract
│    (openapi.yaml)       │ Single source of truth
└────────────┬────────────┘
             │
    ┌────────┴────────┐
    │                 │
    ▼                 ▼
┌──────────┐   ┌──────────────┐
│Swagger UI│   │ ReDoc Docs   │
│Interactive  │   │ Beautiful    │
│try-it-out   │   │ reference    │
└──────────┘   └──────────────┘
    │                 │
    └────────┬────────┘
             ▼
    ┌──────────────────┐
    │  /docs endpoint  │
    │  (embedded, no   │
    │   external deps) │
    └──────────────────┘
```

### Why This Approach?

| Aspect | Why This Works |
|--------|---|
| **OpenAPI Spec** | Industry standard, tooling support, generates clients, validates code |
| **Swagger UI** | Interactive testing, try-it-out, request builder, no external deps |
| **ReDoc** | Beautiful offline reading, responsive design, search |
| **Embedded** | Single binary delivery, works offline, no cloud dependencies |
| **Go stack** | Leverage swag/openapi-codegen, godoc integration, no Node/Java overhead |

---

## Implementation Roadmap

### Phase 1: OpenAPI Specification (Weeks 1-2, 40 hours)
**Deliverable:** Complete `openapi.yaml` with all endpoints documented
- Write spec from scratch (hand-coded, not generated)
- Document all 27 endpoints with examples
- Define all 5 main schemas + DTOs
- Add error code definitions
- Add request/response examples

### Phase 2: Interactive Portal (Weeks 2-3, 8 hours)
**Deliverable:** `/docs` endpoint with Swagger UI embedded
- Download Swagger UI v5.x assets
- Create HTML wrapper
- Embed OpenAPI spec in Go code
- Mount in router
- Test try-it-out functionality

### Phase 3: Code Annotations (Week 3, 12 hours)
**Deliverable:** Godoc comments in all handlers
- Package-level API overview
- Handler method comments
- Request/response struct documentation
- Repository interface docs

### Phase 4: Documentation Guides (Week 4, 20 hours)
**Deliverable:** 5 comprehensive markdown guides
- Quick Start (10-minute integration)
- Error Handling (all error codes + recovery)
- Authentication & CORS
- Webhook Events & Catch-up Mechanism
- API Changelog

### Phase 5: Code Examples (Weeks 4-5, 16 hours)
**Deliverable:** 16+ working examples in 4 languages
- Go (3): client wrapper, schedule generation, score submission
- TypeScript (3): models, service, component integration
- Python (3): requests client, import, standings
- Shell/curl (7+): all common operations

### Phase 6: Embed & Test (Week 5, 8 hours)
**Deliverable:** Fully integrated documentation portal
- Update router with /docs routes
- Embed assets in Go binary
- Test all UI features
- Verify try-it-out works

### Phase 7: CI/CD Automation (Week 6, 12 hours)
**Deliverable:** Automated validation & testing
- GitHub Actions workflow
- OpenAPI spec validation
- Broken link detection
- Client code generation (dry-run)

### Phase 8: Maintenance Setup (Week 6+, 16 hours)
**Deliverable:** Long-term documentation quality
- Versioning strategy
- CHANGELOG template
- Theme customization
- Monitoring dashboard

---

## Quick-Win Option: MVP in 2 Weeks

If you want to get interactive docs live fast:

**Week 1:** Write OpenAPI spec (7 days, 40 hours)
**Week 2:**
- Day 1: Embed Swagger UI (4 hours)
- Day 2: Write quick-start guide (6 hours)
- Days 3-5: Add 10 curl examples (4 hours)

**Result:** Production-ready interactive docs in 54 hours

---

## Key Deliverables

### Documentation Files
```
/
├── openapi.yaml                    ← Main spec (1200+ lines)
├── docs/
│   ├── index.html                 ← Swagger UI
│   ├── redoc.html                 ← ReDoc reference
│   ├── CHANGELOG.md               ← Version history
│   └── custom-theme.css           ← Branding
├── api/
│   ├── guides/
│   │   ├── QUICK_START.md
│   │   ├── ERROR_HANDLING.md
│   │   ├── AUTHENTICATION.md
│   │   └── WEBHOOK_EVENTS.md
│   └── examples/
│       ├── go/                    ← 3 examples
│       ├── typescript/            ← 3 examples
│       ├── python/                ← 3 examples
│       └── curl/                  ← 7+ examples
└── infra/http/handler/docs.go    ← HTTP handlers
```

### Code Changes
```
infra/http/server.go              ← Add /docs routes
cmd/server/main.go                ← Inject docs handler
web/embed.go                       ← Embed assets
infra/http/handler/docs.go        ← NEW: doc handlers
.github/workflows/api-docs.yml    ← NEW: CI/CD
cmd/lint-openapi/main.go          ← NEW: validator
```

---

## Success Criteria

### Quality Metrics
- ✅ 100% endpoint coverage in spec
- ✅ All request/response schemas documented
- ✅ All error codes with examples
- ✅ All path/query parameters documented
- ✅ 16+ code examples across 4 languages
- ✅ Zero broken links
- ✅ <5 second page load time

### Business Metrics
- ✅ 50% reduction in API support tickets
- ✅ 70% faster time-to-first-integration
- ✅ 90%+ developer satisfaction
- ✅ <10 API documentation bugs per month

---

## Tool Recommendations

### Essential Tools
| Tool | Purpose | Cost |
|------|---------|------|
| **swag** | Auto-generate OpenAPI | Free (Go) |
| **openapi-generator** | Multi-language clients | Free (Java) |
| **swagger-cli** | Spec validation | Free (npm) |

### Optional Tools
| Tool | Purpose | Cost |
|------|---------|------|
| **Swagger Hub** | Cloud hosting | $50+/month |
| **Stoplight** | Professional platform | $300+/month |
| **Postman** | API testing | Free tier available |
| **Dredd** | Integration testing | Free (npm) |

**Recommendation:** Use free/embedded tools. No cloud costs needed.

---

## Timeline & Effort Summary

| Phase | Duration | Effort | Can Parallelize |
|-------|----------|--------|-----------------|
| OpenAPI Spec | 2 weeks | 40 hours | No (blocks phase 2) |
| Assets & Integration | 1 week | 8 hours | Yes (with phase 3) |
| Code Annotations | 1 week | 12 hours | Yes (with phase 4) |
| Guides & Examples | 2 weeks | 36 hours | Yes (with phase 5) |
| CI/CD & Maintenance | 1 week | 28 hours | After phase 2 |
| **TOTAL** | **6 weeks** | **132 hours** | **Can compress to 4 weeks** |

**For MVP (2 weeks):** ~54 hours, covers interactive docs + quick-start

---

## Risk Assessment

### Low Risk
- OpenAPI spec is just documentation (no code changes)
- Swagger UI is battle-tested, stable library
- Examples are self-contained, non-breaking

### Medium Risk
- Keeping spec in sync with code over time
- Maintaining multiple language examples

### Mitigation
- Automated CI/CD validation
- Template-based example updates
- Clear versioning & deprecation policy

---

## Next Steps

### Immediate (This Week)
1. ✅ Review this documentation plan (DONE)
2. ✅ Review technical reference (DONE)
3. Assign owner for documentation project
4. Set up project board with phases
5. Create GitHub issues for each phase

### Week 1
1. Start OpenAPI spec writing
2. Gather stakeholder feedback on plan
3. Set up CI/CD template
4. Create /docs directory structure

### Week 2-6
Follow implementation phases above

---

## FAQ

**Q: Do we need to rewrite our API to follow the spec?**
A: No. The spec documents your API as-is. Only add documentation, don't change API.

**Q: Will this add external dependencies to our binary?**
A: No. Everything is embedded. Swagger UI, ReDoc, and spec are bundled in the single binary.

**Q: What about backward compatibility with existing API clients?**
A: Zero impact. This is documentation-only. No API changes required.

**Q: Can we do this incrementally?**
A: Yes! MVP is 2 weeks. Full implementation is 6 weeks. You can release MVP and iterate.

**Q: Will the docs work without internet?**
A: Yes. Everything is self-contained. Works fully offline once loaded.

**Q: How do we keep docs in sync with code?**
A: CI/CD validation + strict code review. Require OpenAPI updates in PRs.

**Q: Can we generate the spec automatically?**
A: Partially. Use `swag` to parse godoc comments, but best to hand-write spec for clarity.

**Q: What if the API changes?**
A: Update openapi.yaml + docs, run CI/CD validation, release new version.

---

## Conclusion

DartScheduler has a solid, well-designed API that lacks professional documentation. This plan provides a clear, realistic path to world-class documentation that will:

1. **Reduce support burden** - Clear docs answer most questions
2. **Accelerate integration** - Developers can integrate in hours, not days
3. **Improve quality** - Automated validation catches inconsistencies
4. **Enable growth** - Easy for new developers to understand and extend
5. **Zero external costs** - Embedded in single binary, no cloud services

**Recommended timeline:** Start with MVP (2 weeks) to get interactive docs live, then iterate with guides and examples (4 more weeks).

**Recommended owner:** Senior developer or technical writer with Go experience

**Recommended budget:** 132 hours total (or 54 hours for MVP)

---

**Documents Created:**
1. `API_DOCUMENTATION_PLAN.md` - Complete implementation plan (18,000 words)
2. `API_TECHNICAL_REFERENCE.md` - Technical specifications (10,000 words)
3. `API_DOCUMENTATION_SUMMARY.md` - This document (Executive summary)

**Status:** Research Only - Ready for Implementation Approval
