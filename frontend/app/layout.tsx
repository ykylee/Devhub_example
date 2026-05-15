import type { Metadata } from "next";
import "./globals.css";
import { ToastContainer } from "@/components/ui/Toast";

export const metadata: Metadata = {
  title: "DevHub - Team Integrated Development Hub",
  description: "Next-generation integrated development hub for modern engineering teams.",
};

// Paint 전에 저장된 theme 을 html element 에 반영해 light→dark FOUC 를 회피.
// localStorage 키는 Header dropdown 의 theme toggle 과 공유한다 ("devhub-theme").
const themeBootstrap = `try{var t=localStorage.getItem("devhub-theme");if(t==="dark")document.documentElement.classList.add("theme-dark");}catch(e){}`;

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="h-full antialiased">
      <head>
        <script dangerouslySetInnerHTML={{ __html: themeBootstrap }} />
      </head>
      <body className="min-h-full flex flex-col font-sans">
        {children}
        <ToastContainer />
      </body>
    </html>
  );
}
