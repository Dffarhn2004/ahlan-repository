import type { Metadata } from "next";
import { privateSection } from "@/lib/seo";

export const metadata: Metadata = privateSection("Pengaturan");
export default function Layout({ children }: Readonly<{ children: React.ReactNode }>) { return children; }
