import type { Metadata } from "next";
import { privateSection } from "@/lib/seo";

export const metadata: Metadata = privateSection("Sumber Web Pesantren");
export default function Layout({ children }: Readonly<{ children: React.ReactNode }>) { return children; }
