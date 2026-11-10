# 11 — Matriks ketertelusuran

## 11.1 Requirement fungsional

| ID | Use case | Sumber utama | Uji yang ada di repo |
|---|---|---|---|
| FR-AUTH-001 | UC-001 | `httpapi/auth.go` | tidak khusus login di daftar yang dibaca |
| FR-AUTH-002 | UC-001 | `auth.go` | — |
| FR-AUTH-003 | UC-001 | `auth.go` | — |
| FR-AUTH-004 | UC-001 | `server.go` `authenticate` | — |
| FR-AUTH-005 | — | `auth.go` `health` | — |
| FR-AUTH-006 | UC-001 | `login/page.tsx` | — |
| FR-TENANT-001 | UC-002 | `tenants.go` | — |
| FR-TENANT-002 | — | `CreateTenant` | — |
| FR-TENANT-003 | — | `updateTenant` | — |
| FR-TENANT-004 | UC-003 | `setTenantStatus` | — |
| FR-TENANT-005 | UC-003 | `00001_init.sql` | — |
| FR-TENANT-006 | — | UNIQUE slug | — |
| FR-TENANT-007 | — | default kuota | — |
| FR-TENANT-008 | UC-009 | `getTenant` | — |
| FR-MEM-001 | UC-006 | `members.go` | — |
| FR-MEM-002 | UC-006 | `createMember` | — |
| FR-MEM-003 | UC-006 | `validMemberRole` | — |
| FR-MEM-004 | UC-006 | `updateMemberRole` | — |
| FR-FILE-001 | UC-008, UC-009 | `files.go` | — |
| FR-FILE-002 | UC-008 | `files.go`, `storage.go` | `storage/storage_test.go` ada |
| FR-FILE-003 | UC-008 | `routes.go` | — |
| FR-FILE-004 | UC-009 | `repository.go` query file | — |
| FR-FILE-005 | UC-008 | `createFolder` | — |
| FR-FILE-006 | UC-007 | `logo.go` | — |
| FR-AST-001 | UC-007 | route provision | `hermes/hermes_test.go`, `integration_test.go` (butuh bridge) |
| FR-AST-002 | UC-007 | route action | sama |
| FR-AST-003 | UC-008 | route sync | — |
| FR-AST-004 | UC-009 | route knowledge status | — |
| FR-AST-005 | UC-004 | `hermes_admin.go` | — |
| FR-AST-006 | UC-004 | `00001_init.sql` | — |
| FR-TG-001 | UC-007, UC-012 | `profile_bot.go` | — |
| FR-TG-002 | UC-007 | `getBot` | — |
| FR-TG-003 | UC-007 | `00005_mvp.sql` | — |
| FR-TG-004 | UC-007 | `deleteBot` | — |
| FR-WA-001 | UC-007, UC-012 | `whatsapp.go` | `whatsapp_routing_test.go` |
| FR-WA-002 | UC-007 | `setWhatsAppService` | — |
| FR-WA-003 | UC-007 | `pairingTimeout` | — |
| FR-WA-004 | UC-007 | `00020`, UI `MAX_ADMINS` | — |
| FR-WA-005 | UC-011 | komentar migrasi `00020` | — |
| FR-KB-001 | UC-007 | `links.go` | `links/links_test.go` |
| FR-KB-002 | — | `crawler.go` | `crawler_test.go` |
| FR-KB-003 | — | `guard.go` | `crawler_test.go` |
| FR-KB-004 | — | `routes.go` crawl | — |
| FR-KB-005 | UC-005 | `requireSuperadmin` | — |
| FR-PUB-001 | UC-011 | `routes.go` | — |
| FR-PUB-002 | UC-011 | `kindParam` | — |
| FR-PUB-003 | UC-011 | `app/p/` | — |
| FR-PUB-004 | UC-012 | `app/page.tsx` FAQ | — |
| FR-CONV-001 | UC-012 | `00015_whatsapp.sql` | — |
| FR-CONV-002 | — | `telegram.go` | — |
| FR-CONV-003 | UC-010 | `telegram.go` | — |
| FR-CONV-004 | UC-013 | `routes.go`, `README.md` | `events_app_test.go` |
| FR-QC-001 | UC-010 | `questions.go` | `questions/questions_test.go` |
| FR-QC-002 | UC-010 | `MaxOpen`, `MaxAnswered` | — |
| FR-QC-003 | UC-012 | `Marker` | `questions_test.go` |
| FR-QC-004 | UC-007 | `contacts.go` | `contacts/contacts_test.go` |
| FR-AUD-001 | — | `auditLogs` | — |
| FR-AUD-002 | — | `CreateTenant` | — |

## 11.2 NFR dan aturan bisnis

| ID | Sumber | Use case terkait |
|---|---|---|
| NFR-SEC-001 | `auth.go` | UC-001 |
| NFR-SEC-002 | `auth.go`, `members.go` | UC-001, UC-006 |
| NFR-SEC-003 | `guard.go` | FR-KB-003 |
| NFR-SEC-004 | `README.md`, `bridge_proto.go` | — |
| NFR-PERF-001 | PRD NFR-02 | [TBD] |
| NFR-REL-001 | `README.md` | [INFERRED] |
| NFR-PRIV-001 | `hermes.go` | UC-012 |
| NFR-OBS-001 | `hermesSkills` | UC-004 |
| NFR-MAINT-001 | `README.md` | — |
| NFR-COMP-001 | PRD out-of-scope | [TBD] |
| BR-001 … BR-009, BR-011, BR-012 | migrasi dan handler yang dikutip di bab 09 | see bab 09 |
| BR-010 | PRD | [TBD] |

## 11.3 Cakupan modul kode

| Area kode | Requirement |
|---|---|
| `httpapi/auth.go` | FR-AUTH-* , NFR-SEC-001, NFR-SEC-002 |
| `httpapi/tenants.go` | FR-TENANT-* , FR-AUD-* |
| `httpapi/members.go` | FR-MEM-* |
| `httpapi/files.go`, `storage` | FR-FILE-001 … 005 |
| `httpapi/logo.go` | FR-FILE-006, FR-PUB-002 |
| `httpapi/hermes_admin.go` | FR-AST-005, NFR-OBS-001 |
| route provision/sync | FR-AST-001 … 004 |
| `profile_bot.go` | FR-TG-* |
| `whatsapp.go`, migrasi 00015/00020 | FR-WA-* |
| `links`, `crawler` | FR-KB-* , BR-008, BR-009 |
| `questions`, `contacts` | FR-QC-* |
| `telegram.go` percakapan | FR-CONV-* |
| `routes.go` publik | FR-PUB-001 |
| `authz/permissions.go` | bab 03, BR-012 |

Fitur yang tidak punya FR produk dicatat di [12](12-open-questions-and-risks.md), bukan dipaksa menjadi kewajiban.
