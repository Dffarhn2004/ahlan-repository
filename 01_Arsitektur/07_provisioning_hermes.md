# 07 — Provisioning Hermes Profile otomatis dengan aman

## Tujuan

Superadmin menambahkan ponpes di web. Go menyiapkan Hermes Profile terpisah dengan persona, tool policy, bot credential, hooks, dan folder knowledge, lalu mengaktifkannya **hanya jika lulus verifikasi**.

## Kontrak teknis

- `HermesProvisioner` adalah adapter Go untuk daftar operasi: `CreateParked`, `Describe`, `WriteApprovedConfig`, `AttachBot`, `Verify`, `Start`, `Stop`, `Restart`, `Delete`.
- Adapter menjalankan **allowlisted executable + argv fixed**, bukan `sh -c` dengan interpolasi string UI.
- Input tenant/profile **server-generated UUID**, bukan path bebas; timeout per command, max stdout/stderr, redact token, context cancellation.
- Runner bekerja dengan OS user service khusus dan hak tulis hanya ke profile homes yang dikelolanya; Go HTTP API tidak punya shell.
- Implementasi profile files dibuat dengan atomic write (temp+fsync+rename), file mode `0600` untuk secrets, apply config dengan safe reload.

## CLI Hermes yang telah didokumentasikan (pin versi sebelum dipakai)

```bash
hermes profile create <profile_name>
hermes -p <profile_name> gateway stop
hermes -p <profile_name> gateway start
hermes -p <profile_name> gateway restart
hermes -p <profile_name> gateway status
```

Pada multiplex mode, `stop` **parks** named profile, bukan menghentikan host gateway. Marker `gateway.parked` dapat ditempatkan saat provisioning agar profil tidak aktif sebelum selesai. Pastikan marker dibuat secepat mungkin setelah profile create dan sebelum gateway mendeteksi profile baru; untuk menghilangkan race di production, provisioner perlu mekanisme create-under-parked/staging atomik atau kontrol gateway yang dibuktikan dalam integration test. Jangan mengandalkan urutan `create; stop` sebagai jaminan tanpa menguji window auto-discovery.

## Step plan (P0)

| Step | Action | Idempotency proof | Gate |
|---|---|---|---|
| 1 | Allocate immutable tenant_id/profile_key | Unique DB constraint | Tenant row exists |
| 2 | Reserve profile registry | Unique `(host_id, profile_key)` | Owner maps to same tenant |
| 3 | Create profile safely offline | Existing profile inspected before reuse | `PARKED` |
| 4 | Create tenant data root | mkdir via root-safe filesystem API | Directory permissions correct |
| 5 | Render SOUL/persona | Template only, escape user text | No unsafe capabilities |
| 6 | Render Hermes config & skill allowlist | Baseline hash/version | No arbitrary tools |
| 7 | Obtain bot credential from admin | BotFather manual → secure Go API | getMe + unique bot ID |
| 8 | Store/mount token to profile | Secret store or encrypted DB, only runner reads | 0600 permissions |
| 9 | Install outbound signed hooks + scoped tool config | Hook policy version/hash | Receiver handshake |
| 10 | Offline smoke check | Full structured report | Negative security checks passed |
| 11 | Unpark/activate profile | Gateway status confirms | ACTIVE only after healthy |

## Pseudocode worker

```go
func (w *Worker) Provision(ctx context.Context, tenantID uuid.UUID) error {
    t := w.tenantRepo.GetForUpdate(ctx, tenantID)
    if t.Status == Active { return nil }
    if err := w.provisioner.EnsureParkedProfile(ctx, t.ProfileKey); err != nil { return err }
    if err := w.storage.EnsureTenantRoot(ctx, t.ID); err != nil { return err }
    if err := w.provisioner.ApplyBaseline(ctx, t.ProfileKey, t.ConfigRevision); err != nil { return err }
    if !t.TelegramVerified { return ErrWaitingInput }
    if err := w.verifier.RunSecurityGates(ctx, t); err != nil { return err }
    if err := w.provisioner.Start(ctx, t.ProfileKey); err != nil { return err }
    if err := w.verifier.AssertHealthy(ctx, t); err != nil { return err }
    return w.tenantRepo.MarkActive(ctx, t.ID)
}
```

*Pseudocode konseptual, bukan kode Go yang siap compile.*

## Profile dan storage layout usulan

```text
<hermes_home>/profiles/t_<tenant_uuid>/
  SOUL.md
  config.yaml
  .env             # secrets; bukan area File Manager
  gateway.parked   # saat offline
  hooks/
  memories/
  sessions/

<hermes_persistent_volume>/DataPonpes/<tenant_uuid>/
  PPDB/
  Profil/
  Akademik/
  .trash/           # internal; tidak bisa browse bebas
```

`<hermes_home>` dan `<hermes_persistent_volume>` di-resolve saat deployment, jangan mengunci path `profiles/` di `/opt/data` bila image Hermes ternyata memakai home berbeda. Folder `DataPonpes` adalah kustom produk.

## Drifts dan lifecycle updates

- Simpan `desired_config_hash` dan `observed_config_hash`, serta `hermes_version` dan `last_verified_at`.
- Cron reconciler mengukur state nyata terhadap DB; jangan selalu rewrite config bila admin mengubah metadata saja.
- Update `SOUL.md` berlaku bersih untuk sesi baru, bukan harus langsung merombak prompt sesi yang sedang berjalan.
- Rotate token: validate new token → park/profile update atomik → restart profile → health verify → revoke old token melalui BotFather jika relevan; rollback jika health gagal.

## Keamanan command

Dilarang: `exec.Command("sh", "-c", userString)`, menjalankan perintah bebas dari dashboard, expose `.env` sebagai response HTTP, mengizinkan profile name berupa input user, dan provisioning paralel dua job untuk tenant sama.
