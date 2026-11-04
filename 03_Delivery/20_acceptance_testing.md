# 20 — Acceptance criteria, end-to-end dan security test plan

## Prinsip

Tidak cukup uji happy path. Sistem hanya lolos P0 jika test tenant isolation, public Telegram restriction, provisioning idempotency dan file path safety lulus. Dua tenant uji `A` dan `B`, masing-masing dokumen dengan fakta yang **saling bertentangan secara sengaja** untuk mendeteksi kebocoran.

## Test matrix P0

| Test ID | Scenario | Expected | Requirement |
|---|---|---|---|
| AT-001 | Superadmin POST tenant A | HTTP 202, tenant/job committed atomik | FR-001 |
| AT-002 | Provisioning worker retry 3 kali | Satu profile terdaftar, satu tenant root, tidak ada duplicate bot | FR-002 |
| AT-003 | Simultan tenant A/B | Profile/credentials/memory masing-masing terpisah | FR-002,003 |
| AT-004 | Admin A browse folders | Hanya folder A | FR-004,006 |
| AT-005 | Admin A GET file B by UUID | 404, no metadata leaked | FR-004,015 |
| AT-006 | Upload PDF A 2 MB | File fisik di DataPonpes/A, SHA matches, DB metadata | FR-005,006 |
| AT-007 | Upload file unsupported/malformed | Rejected 4xx, no orphan unsafe | FR-016 |
| AT-008 | `../`, encoded traversal, symlink race | Rejected, no outside-root writes/read | FR-015,016 |
| AT-009 | Chat A tentang biaya yang hanya ada di B | No B facts; abstain | FR-007,008 |
| AT-010 | Chat B dengan brosur B | Jawaban grounded hanya dari B | FR-007,008 |
| AT-011 | Prompt injection via PDF A | Policy tidak berubah, tools tidak melebar | FR-015 |
| AT-012 | Telegram user umum mencoba `/model`, terminal, exec, tools | Privileged commands denied | FR-014 |
| AT-013 | Set Telegram token yang sama ke A+B | 409 + security/audit; B tidak aktif | FR-003 |
| AT-014 | Suspend A, chat B | A parked/denied, B normal | FR-009 |
| AT-015 | Create tenant saat Hermes down | Retry, no active false-positive | FR-010 |
| AT-016 | Update file | Retrieval menampilkan revisi setelah READY | FR-005,008 |
| AT-017 | Call internal tool as profile A with tenant B param | Server tetap scope A atau reject | FR-015 |
| AT-018 | Event with invalid signature/replay | 401/409, nothing ingested | FR-012 |
| AT-019 | Duplicate same event_id | Exactly one stored event | FR-012 |
| AT-020 | Mutate tenant/file | Append-only audit created transactionally | FR-011 |
| AT-021 | Recovery from backup snapshot | Files, metadata checksum, roles and parked state consistent | NFR-07 |
| AT-022 | GO offline during knowledge tool call | Agent abstains, does not fabricate | FR-008 |
| AT-023 | Admin simpan `google_maps_url` valid / tidak valid (`javascript:`, host lain, >2048) | Valid tersimpan + audit; tidak valid `422 MAPS_URL_INVALID`; bot menjawab lokasi dengan tautan itu | FR-022 |
| AT-024 | Sumber crawl ke `169.254.169.254`, `localhost`, `10.x`, IPv6 loopback | `CRAWL_URL_FORBIDDEN`, tidak ada koneksi keluar | FR-027 |
| AT-025 | Situs uji me-redirect ke IP privat / DNS rebinding | Diblokir pada validasi tiap hop | FR-027 |
| AT-026 | `robots.txt` melarang path | Path tidak diambil, tercatat di run | FR-027 |
| AT-027 | Jadwal harian + trigger manual bersamaan | Satu run aktif; manual `409`; tidak ada duplikat file | FR-024,025 |
| AT-028 | Halaman berubah / tidak berubah / hilang 2x | Versi baru hanya saat hash beda; hilang → tidak muncul di retrieval | FR-026 |
| AT-029 | Halaman crawl berisi prompt injection; unggahan dan crawl bertentangan | Policy tak berubah; unggahan menang atau bot minta klarifikasi; jawaban bersitasi URL+tanggal | FR-026 |
| AT-031 | Tempel tautan Google Docs publik / privat / folder Drive / `docs.google.com.evil.test` | Publik terbaca; privat `NEEDS_PUBLIC_SHARING`; folder & host palsu ditolak; fetch hanya ke URL ekspor bentukan server | FR-023,027 |
| AT-030 | Admin A akses sumber/run/halaman crawl milik B | 404 | FR-004 |

## Gherkin examples

```gherkin
Feature: Tenant file isolation
  Given tenant A and tenant B exist
  And user AdminA belongs only to tenant A
  And a file exists under DataPonpes/B
  When AdminA requests the file using a valid UUID belonging to B
  Then the API responds with 404
  And no contents, metadata or real path of B is leaked
  And the unauthorized attempt is recorded for security monitoring
```

```gherkin
Feature: Hermes profile provisioning idempotency
  Given tenant A has a QUEUED provisioning job
  When the worker crashes after creating the parked profile
  And the same job is retried
  Then only one profile key exists for A
  And the existing profile is adopted only after owner validation
  And the tenant does not become ACTIVE until security gates pass
```

```gherkin
Feature: Public Telegram safety
  Given tenant A is configured as a public information bot
  When an unknown Telegram user sends a privileged command
  Then the command is rejected
  And no shell, unrestricted file, model configuration, secrets or cross-tenant data are accessible
  And normal FAQ requests continue to work
```

## Load and resilience (pilot)

- Concurrently simulate 10–30 active chats across 2 tenants; record latency and RAM, not as claim for hundreds.
- Kill worker during stage; ensure lease expires/retry and profile isn't falsely ACTIVE.
- Simulate bot token revoked, Telegram downtime, provider 429, Go tool outage, DB restart, disk limit.
- Hook receiver unavailable: document dropped-event behavior; dashboards mark incomplete data rather than guaranteeing completeness.
- Upload many files at quota boundary and parallel rename/delete; ensure safe state after restart.

## Exit criteria / release checklist

- [ ] All AT-001..022 pass with evidence (logs redacted, screenshots, CI results).
- [ ] No known critical/high cross-tenant or public Telegram privilege bugs.
- [ ] Each profile using allowlisted tools only; security spike result signed off.
- [ ] Backup restore demonstration succeeds.
- [ ] Operator can create/suspend/resume 2 tenants via dashboard without SSH.
- [ ] Public privacy notice, consent policy and contact admin available.
- [ ] Hermes version pinned; upgraded release requires regression suite.

**If any security gate fails, pilot must remain non-public or use isolated alternative architecture.**
