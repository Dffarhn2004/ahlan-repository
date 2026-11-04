import type { Metadata } from "next";
import { privateSection } from "@/lib/seo";

export const metadata: Metadata = privateSection("Dokumen Pengetahuan");
export default function Layout({ children }: Readonly<{ children: React.ReactNode }>) { return children; }
