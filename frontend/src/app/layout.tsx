import type { Metadata, Viewport } from "next";
import "./globals.css";
import { keywords, siteDescription, siteName, siteTitle, siteUrl, themeColor } from "@/lib/seo";

export const metadata: Metadata = {
  metadataBase: new URL(siteUrl),
  title: { default: siteTitle, template: `%s · ${siteName}` },
  description: siteDescription,
  applicationName: siteName,
  keywords,
  authors: [{ name: siteName }],
  creator: siteName,
  publisher: siteName,
  category: "education",
  formatDetection: { telephone: false, email: false, address: false },
  manifest: "/manifest.webmanifest",
  appleWebApp: { capable: true, title: siteName, statusBarStyle: "default" },
  openGraph: {
    type: "website",
    locale: "id_ID",
    siteName,
    title: siteTitle,
    description: siteDescription,
    images: [{ url: "/og-image.png", width: 1200, height: 630, alt: `${siteName} — chatbot AI untuk penerimaan santri baru`, type: "image/png" }],
  },
  twitter: { card: "summary_large_image", title: siteTitle, description: siteDescription, images: ["/og-image.png"] },
  robots: { index: true, follow: true, googleBot: { index: true, follow: true, "max-image-preview": "large", "max-snippet": -1, "max-video-preview": -1 } },
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  colorScheme: "light",
  themeColor,
};

const jsonLd = {
  "@context": "https://schema.org",
  "@graph": [
    { "@type": "Organization", "@id": `${siteUrl}/#organization`, name: siteName, url: siteUrl, logo: { "@type": "ImageObject", url: `${siteUrl}/icon-512.png`, width: 512, height: 512 } },
    { "@type": "WebSite", "@id": `${siteUrl}/#website`, url: siteUrl, name: siteName, description: siteDescription, inLanguage: "id-ID", publisher: { "@id": `${siteUrl}/#organization` } },
    { "@type": "SoftwareApplication", name: siteName, description: siteDescription, applicationCategory: "EducationalApplication", operatingSystem: "Web", inLanguage: "id-ID", url: siteUrl, image: `${siteUrl}/og-image.png` },
  ],
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="id">
      <body>
        <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd).replace(/</g, "\u003c") }} />
        {children}
      </body>
    </html>
  );
}
