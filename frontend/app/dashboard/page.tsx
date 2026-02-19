import type { Metadata } from "next";
import DashboardPageClient from "./DashboardPageClient";

export const metadata: Metadata = {
  title: "Telemetry Dashboard",
  description: "MVP telemetry dashboard for activation, reliability, and conversion flow.",
  robots: {
    index: false,
    follow: false,
  },
};

export default function DashboardPage() {
  return <DashboardPageClient />;
}
