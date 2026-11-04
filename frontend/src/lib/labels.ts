const provisionLabels: Record<string, string> = {
  NOT_STARTED: "Belum disiapkan",
  PENDING: "Menunggu",
  QUEUED: "Menunggu",
  RUNNING: "Disiapkan",
  IN_PROGRESS: "Disiapkan",
  SUCCEEDED: "Siap",
  ACTIVE: "Siap",
  FAILED: "Gagal",
  CANCELLED: "Dibatalkan",
};

const titleCase = (value: string) => value.toLowerCase().replaceAll("_", " ").replace(/\b\w/g, char => char.toUpperCase());

/** Friendly Indonesian label for a tenant's assistant provisioning status. */
export const provisionLabel = (value?: string | null) => {
  if (!value) return "Belum disiapkan";
  return provisionLabels[value.toUpperCase().replaceAll(" ", "_")] ?? titleCase(value);
};

/** Friendly label for an audit log action code, e.g. FILE_UPLOADED -> "File diunggah". */
const auditLabels: Record<string, string> = {
  TENANT_CREATED: "Pesantren dibuat",
  TENANT_UPDATED: "Pesantren diperbarui",
  TENANT_SUSPENDED: "Pesantren ditangguhkan",
  TENANT_ONBOARDING: "Pesantren diaktifkan kembali",
  MEMBER_CREATED: "Pengguna ditambahkan",
  MEMBER_ROLE_UPDATED: "Peran pengguna diubah",
  FILE_UPLOADED: "File diunggah",
  FILE_DELETED: "File dihapus",
  FILE_RENAMED: "Nama file diubah",
  FOLDER_CREATED: "Folder dibuat",
  PROFILE_UPDATED: "Profil pesantren diperbarui",
  HERMES_PROVISIONING_QUEUED: "Penyiapan asisten dimulai",
  HERMES_START: "Asisten diaktifkan",
  HERMES_STOP: "Asisten dinonaktifkan",
  HERMES_RESTART: "Asisten dimulai ulang",
  HERMES_GATEWAY_RESTART: "Gateway asisten dimulai ulang",
  HERMES_UPDATE: "Asisten diperbarui",
  HERMES_DELETE: "Asisten dihapus",
  HERMES_GLOBAL_SETTINGS_UPDATED: "Pengaturan global asisten diperbarui",
  KNOWLEDGE_SYNCED: "Pengetahuan asisten diperbarui",
  ASSISTANT_SYNCED: "Pengaturan asisten disinkronkan",
  KNOWLEDGE_IMAGES_READ: "Gambar dokumen selesai dibaca asisten",
  TELEGRAM_ACTIVATED: "Telegram diaktifkan",
  TELEGRAM_DEACTIVATED: "Telegram dinonaktifkan",
  TELEGRAM_BOT_CONNECTED: "Bot Telegram dihubungkan",
  TELEGRAM_BOT_REMOVED: "Bot Telegram dilepas",
  CONVERSATION_HANDOFF: "Percakapan ditandai perlu bantuan",
  CONVERSATION_RESOLVED: "Percakapan diselesaikan",
  CONVERSATION_ACTIVE: "Percakapan dibuka kembali",
};

export const auditActionLabel = (action?: string) => {
  if (!action) return "Aktivitas sistem";
  return auditLabels[action.toUpperCase()] ?? titleCase(action);
};
const actorLabels: Record<string, string> = { SYSTEM: "Sistem", USER: "Pengguna", WORKER: "Layanan otomatis" };
const targetLabels: Record<string, string> = { TENANT: "Pesantren", USER: "Pengguna", FILE: "File", FOLDER: "Folder", CONVERSATION: "Percakapan", TELEGRAM_BOT: "Bot Telegram", HERMES_GLOBAL_SETTINGS: "Pengaturan global asisten" };
export const auditActorLabel = (value?: string) => (value ? actorLabels[value.toUpperCase()] ?? titleCase(value) : "Sistem");
export const auditTargetLabel = (value?: string) => (value ? targetLabels[value.toUpperCase()] ?? titleCase(value) : "Sistem");
