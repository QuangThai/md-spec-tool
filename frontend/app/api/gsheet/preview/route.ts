import { NextRequest } from "next/server";
import { proxyToBackend } from "@/lib/gsheetApiHelper";

export const runtime = "nodejs";

/**
 * POST /api/gsheet/preview
 * 
 * Returns a preview of Google Sheet data with AI column mapping before conversion.
 * This route proxies the request to the Go backend with OAuth token injection
 * for accessing private Google Sheets.
 */
export async function POST(request: NextRequest) {
  return proxyToBackend(request, {
    backendPath: "/api/mdflow/gsheet/preview",
  });
}
