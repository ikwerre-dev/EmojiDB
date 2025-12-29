import type { Metadata } from "next";
import { Geist, Geist_Mono, Inter, Barlow, Bebas_Neue, Saira, Roboto_Slab } from "next/font/google";
import SmoothScroll from "@/components/SmoothScroll";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
  weight: ["400", "500", "600", "700", "800", "900"],
});

const bartle = Barlow({
  variable: "--font-bartle",
  subsets: ["latin"],
  weight: ["900"],
});

const hegarty = Bebas_Neue({
  variable: "--font-hegarty",
  subsets: ["latin"],
  weight: ["400"],
});

const sekuya = Saira({
  variable: "--font-sekuya",
  subsets: ["latin"],
  weight: ["400", "500", "600", "700", "800", "900"],
});

const robotoSlab = Roboto_Slab({
  variable: "--font-roboto-slab",
  subsets: ["latin"],
  weight: ["400", "500", "600", "700", "800", "900"],
});

export const metadata: Metadata = {
  title: "EmojiDB - Encrypted Emoji-Encoded Database",
  description: "Fast, secure, emoji-encoded database with encryption built-in",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body
        className={`${geistSans.variable} ${geistMono.variable} ${inter.variable} ${bartle.variable} ${hegarty.variable} ${sekuya.variable} ${robotoSlab.variable} antialiased`}
      >
        <SmoothScroll>{children}</SmoothScroll>
      </body>
    </html>
  );
}
