import type { Metadata } from "next";
import { privateSection } from "@/lib/seo";

export const metadata: Metadata = privateSection("Hermes");
export default function Layout({ children }: Readonly<{ children: React.ReactNode }>) { return children; }
