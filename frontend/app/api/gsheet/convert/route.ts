import { NextRequest } from "next/server";
import { proxyToBackend } from "@/lib/gsheetApiHelper";

export const runtime = "nodejs";

/**
 * POST /api/gsheet/convert
 * 
 * Converts a Google Sheet to MDFlow format.
 * This route proxies the request to the Go backend with OAuth token injection
 * for accessing private Google Sheets.
 */
export async function POST(request: NextRequest) {
  return proxyToBackend(request, {
    backendPath: "/api/mdflow/gsheet/convert",
  });
}
