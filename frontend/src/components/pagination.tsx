"use client";

import { useState } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";

export const pageSizes = [10, 20, 50];

export type Paged<T> = {
  rows: T[]; page: number; pages: number; total: number; size: number; from: number; to: number;
  setPage: (page: number) => void; setSize: (size: number) => void;
};

/** Client-side paging. Pass `resetKey` (e.g. the active filters) to return to page 1 when it changes. */
export function usePagination<T>(items: T[], initialSize = 10, resetKey = ""): Paged<T> {
  const [state, setState] = useState({ page: 1, size: initialSize, key: resetKey });
  const total = items.length;
  const pages = Math.max(1, Math.ceil(total / state.size));
  const page = Math.min(state.key === resetKey ? state.page : 1, pages);
  const start = (page - 1) * state.size;
  return {
    rows: items.slice(start, start + state.size), page, pages, total, size: state.size,
    from: total ? start + 1 : 0, to: Math.min(start + state.size, total),
    setPage: next => setState(current => ({ ...current, page: Math.min(Math.max(1, next), pages), key: resetKey })),
    setSize: next => setState({ page: 1, size: next, key: resetKey }),
  };
}

function pageList(page: number, pages: number): (number | "gap")[] {
  if (pages <= 7) return Array.from({ length: pages }, (_, i) => i + 1);
  const set = new Set([1, pages, page - 1, page, page + 1]);
  const sorted = [...set].filter(n => n >= 1 && n <= pages).sort((a, b) => a - b);
  return sorted.flatMap((n, i) => (i && n - sorted[i - 1] > 1 ? ["gap" as const, n] : [n]));
}

export function Pagination<T>({ paged, label = "data" }: { paged: Paged<T>; label?: string }) {
  if (!paged.total) return null;
  return (
    <nav className="pagination" aria-label="Navigasi halaman">
      <span className="pagination-info">Menampilkan <strong>{paged.from}–{paged.to}</strong> dari <strong>{paged.total}</strong> {label}</span>
      <div className="pagination-controls">
        <label className="pagination-size"><span>Per halaman</span>
          <select value={paged.size} onChange={event => paged.setSize(Number(event.target.value))} aria-label="Jumlah data per halaman">
            {pageSizes.map(size => <option key={size} value={size}>{size}</option>)}
          </select>
        </label>
        <button type="button" className="pagination-button" disabled={paged.page <= 1} onClick={() => paged.setPage(paged.page - 1)} aria-label="Halaman sebelumnya"><ChevronLeft size={15} /></button>
        {pageList(paged.page, paged.pages).map((item, i) => item === "gap"
          ? <span key={`gap-${i}`} className="pagination-gap">…</span>
          : <button key={item} type="button" className={item === paged.page ? "pagination-button active" : "pagination-button"} aria-current={item === paged.page ? "page" : undefined} onClick={() => paged.setPage(item)}>{item}</button>)}
        <button type="button" className="pagination-button" disabled={paged.page >= paged.pages} onClick={() => paged.setPage(paged.page + 1)} aria-label="Halaman berikutnya"><ChevronRight size={15} /></button>
      </div>
    </nav>
  );
}
