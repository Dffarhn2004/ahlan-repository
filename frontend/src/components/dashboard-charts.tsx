"use client";

import { ArcElement, BarElement, CategoryScale, Chart as ChartJS, Filler, Legend, LinearScale, LineElement, PointElement, Tooltip, type ChartOptions } from "chart.js";
import { Bar, Doughnut, Line } from "react-chartjs-2";

ChartJS.register(ArcElement, BarElement, CategoryScale, Filler, Legend, LinearScale, LineElement, PointElement, Tooltip);
ChartJS.defaults.font.family = "inherit";
ChartJS.defaults.font.size = 11;
ChartJS.defaults.color = "#6b7a71";

export type DailyActivity = { date: string; user: number; bot: number };
export type ConversationStats = {
  days: number; user_messages: number; bot_replies: number; prev_user_messages: number; prev_bot_replies: number;
  visitors: number; prev_visitors: number; daily: DailyActivity[]; hourly: number[]; status_breakdown: Record<string, number>;
  top_tenants?: { tenant_id: string; name: string; user: number; bot: number }[];
};

const BRAND = "#168a5b";
const GOLD = "#e1b85e";
const shortDate = (value: string) => new Intl.DateTimeFormat("id-ID", { day: "numeric", month: "short" }).format(new Date(`${value}T00:00:00`));
const tooltip = { backgroundColor: "#17211c", padding: 10, cornerRadius: 8, boxPadding: 4 };
const grid = { color: "#eef2ef" };

/** Compares the current period with the previous one, e.g. "Naik 12% dari periode lalu". */
export function trendNote(current: number, previous: number) {
  if (!previous) return current ? "Belum ada pembanding" : "Belum ada data";
  const delta = Math.round(((current - previous) / previous) * 100);
  if (delta === 0) return "Sama seperti periode lalu";
  return `${delta > 0 ? "Naik" : "Turun"} ${Math.abs(delta)}% dari periode lalu`;
}

export function ActivityChart({ daily }: { daily: DailyActivity[] }) {
  const options: ChartOptions<"line"> = {
    responsive: true, maintainAspectRatio: false, interaction: { mode: "index", intersect: false },
    plugins: { legend: { position: "bottom", labels: { usePointStyle: true, boxWidth: 8, padding: 16 } }, tooltip },
    scales: { x: { grid: { display: false }, ticks: { maxTicksLimit: 7, maxRotation: 0 } }, y: { beginAtZero: true, grid, border: { display: false }, ticks: { precision: 0 } } },
  };
  const data = {
    labels: daily.map(day => shortDate(day.date)),
    datasets: [
      { label: "Pertanyaan masuk", data: daily.map(day => day.user), borderColor: BRAND, backgroundColor: "rgba(22,138,91,.12)", fill: true, tension: 0.35, pointRadius: 0, pointHoverRadius: 5, borderWidth: 2 },
      { label: "Balasan asisten", data: daily.map(day => day.bot), borderColor: GOLD, backgroundColor: "transparent", tension: 0.35, pointRadius: 0, pointHoverRadius: 5, borderWidth: 2 },
    ],
  };
  return <div className="chart-box"><Line options={options} data={data} /></div>;
}

export function HourlyChart({ hourly }: { hourly: number[] }) {
  const max = Math.max(...hourly);
  const pad = (hour: number) => String(hour).padStart(2, "0");
  const options: ChartOptions<"bar"> = {
    responsive: true, maintainAspectRatio: false,
    plugins: { legend: { display: false }, tooltip: { ...tooltip, callbacks: { title: items => `Pukul ${items[0].label}.00 WIB`, label: item => `${item.parsed.y} pertanyaan` } } },
    scales: { x: { grid: { display: false }, ticks: { maxRotation: 0, callback: (_, index) => (index % 6 === 0 ? pad(index) : "") } }, y: { beginAtZero: true, grid, border: { display: false }, ticks: { precision: 0 } } },
  };
  const data = { labels: hourly.map((_, hour) => pad(hour)), datasets: [{ data: hourly, borderRadius: 4, backgroundColor: hourly.map(value => (max > 0 && value === max ? BRAND : "#bfdccb")) }] };
  return <div className="chart-box"><Bar options={options} data={data} /></div>;
}

const statusMeta: [string, string, string][] = [["ACTIVE", "Aktif", BRAND], ["HANDOFF", "Perlu admin", GOLD], ["RESOLVED", "Selesai", "#c7d9cd"]];

export function StatusDonut({ breakdown }: { breakdown: Record<string, number> }) {
  const total = statusMeta.reduce((sum, [key]) => sum + (breakdown[key] ?? 0), 0);
  const options: ChartOptions<"doughnut"> = {
    responsive: true, maintainAspectRatio: false, cutout: "68%",
    plugins: { legend: { position: "bottom", labels: { usePointStyle: true, boxWidth: 8, padding: 16 } }, tooltip },
  };
  const data = { labels: statusMeta.map(([, label]) => `${label} (${breakdown[statusMeta.find(item => item[1] === label)![0]] ?? 0})`), datasets: [{ data: statusMeta.map(([key]) => breakdown[key] ?? 0), backgroundColor: statusMeta.map(([, , color]) => color), borderWidth: 0 }] };
  return total ? <div className="chart-box"><Doughnut options={options} data={data} /></div> : <div className="empty">Belum ada percakapan.</div>;
}

/** Pesantren ranked by incoming questions; used by the platform-wide superadmin view. */
export function TopTenantsChart({ items }: { items: NonNullable<ConversationStats["top_tenants"]> }) {
  const options: ChartOptions<"bar"> = {
    indexAxis: "y", responsive: true, maintainAspectRatio: false,
    plugins: { legend: { position: "bottom", labels: { usePointStyle: true, boxWidth: 8, padding: 16 } }, tooltip },
    scales: { x: { beginAtZero: true, grid, border: { display: false }, ticks: { precision: 0 } }, y: { grid: { display: false } } },
  };
  const data = { labels: items.map(item => item.name), datasets: [
    { label: "Pertanyaan masuk", data: items.map(item => item.user), backgroundColor: BRAND, borderRadius: 4 },
    { label: "Balasan asisten", data: items.map(item => item.bot), backgroundColor: GOLD, borderRadius: 4 },
  ] };
  return items.length ? <div className="chart-box"><Bar options={options} data={data} /></div> : <div className="empty">Belum ada percakapan pada periode ini.</div>;
}

/** Generic donut for tenant lifecycle distribution. */
export function DistributionDonut({ labels, values }: { labels: string[]; values: number[] }) {
  const palette = [BRAND, "#7ebd99", "#c7d9cd", GOLD, "#d9786f"];
  const options: ChartOptions<"doughnut"> = { responsive: true, maintainAspectRatio: false, cutout: "68%", plugins: { legend: { position: "bottom", labels: { usePointStyle: true, boxWidth: 8, padding: 16 } }, tooltip } };
  const data = { labels: labels.map((label, index) => `${label} (${values[index]})`), datasets: [{ data: values, backgroundColor: values.map((_, index) => palette[index % palette.length]), borderWidth: 0 }] };
  return <div className="chart-box"><Doughnut options={options} data={data} /></div>;
}
