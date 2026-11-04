# 23 — Sumber teknis dan verifikasi (24 Oktober 2026)

Dokumen resmi Hermes yang menjadi dasar rancangan. Link ini dimasukkan untuk membantu engineering menguji versi yang benar; PRD **bukan** jaminan semua deployment mendukung konfigurasi sama.

1. **Profiles: Running Multiple Agents** — `hermes profile create`, `profile` owns own config/env/SOUL/memory/session, profile **bukan sandbox**.  
   https://hermes-agent.nousresearch.com/docs/user-guide/profiles
2. **Running Many Gateways at Once** — multiplex profile routing, per-profile bot token, parked profile stop/start, `gateway.multiplex_profiles`.  
   https://hermes-agent.nousresearch.com/docs/user-guide/multi-profile-gateways
3. **Telegram Setup** — bot token dari BotFather, allowlists, commands, attachments; security implications untuk public bot.  
   https://hermes-agent.nousresearch.com/docs/user-guide/messaging/telegram
4. **Environment Variables** — `TELEGRAM_BOT_TOKEN`, allowed users dan opsi allow-all yang memerlukan review deployment.  
   https://hermes-agent.nousresearch.com/docs/reference/environment-variables
5. **Event Hooks** — gateway/plugin/shell hooks; `post_llm_call`, outbound signed lifecycle webhook, best-effort delivery.  
   https://hermes-agent.nousresearch.com/docs/user-guide/features/hooks
6. **Document Extraction** — `read_file` PDF via optional converter, beberapa Office format; JPEG/scan memerlukan handling lain.  
   https://hermes-agent.nousresearch.com/docs/user-guide/features/document-extraction
7. **Hermes Security** — command authorization, filesystem write safety, execution/sandbox constraints, plugin trust model.  
   https://hermes-agent.nousresearch.com/docs/user-guide/security
8. **Multiplexing Gateway Internals (developer guide)** — route ownership, scopes, gateway limitations dan internal control plane.  
   https://github.com/NousResearch/hermes-agent/blob/main/website/docs/developer-guide/multiplexing-gateway.md

## Verified vs our proposed contracts

| Perilaku | Status |
|---|---|
| Profile own state/env/SOUL + CLI create | Didokumentasikan resmi |
| Multiplex profiles and park/start/stop | Didokumentasikan resmi |
| Telegram bot manual BotFather + token | Didokumentasikan resmi |
| Outbound webhook best effort + signed | Didokumentasikan resmi |
| PDF read_file converter optional | Didokumentasikan resmi |
| Folder `DataPonpes` | **Rancangan khusus aplikasi** |
| Go APIs `/api/v1/...` | **Rancangan khusus aplikasi** |
| `search_tenant_documents` tool | **Rancangan khusus aplikasi** |
| HMAC canonical envelope `event_id` | **Rancangan khusus aplikasi**; adapt sesuai event nyata |
| Public anonymous messaging yang benar-benar aman | **Harus dibuktikan di spike** |
| Full message logging guaranteed | **Tidak dijamin** oleh webhook bawaan |
| `hermes profile create` create parked atomik | **Belum bisa diasumsikan**; butuh test/adapter |

## Engineering verification protocol

- Catat Hermes git SHA/image digest, host environment dan konfigurasi test.
- Jalankan CLI `hermes profile create`, status, start/stop profile dua tenant.
- Buktikan bot Telegram A tidak menerima file/info B saat tool restrictions aktif.
- Buktikan unknown Telegram user tidak dapat mengakses administrative slash commands, terminal, model settings, secrets, file tools.
- Kirim hooks ke mock receiver, matikan receiver, ukur retry/drop, payload/usage shape.
- Catat dependencies untuk PDF dan OCR (jangan gunakan lazy auto-install di production tanpa pin/approval).
- Attach test evidence ke backlog sebelum menandai dokumen implementable/go-live.
