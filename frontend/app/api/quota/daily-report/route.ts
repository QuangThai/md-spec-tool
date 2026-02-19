import { NextRequest, NextResponse } from "next/server";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export async function GET(request: NextRequest) {
  try {
    const searchParams = request.nextUrl.searchParams;
    const days = searchParams.get("days") || "7";
    const aggregateBy = searchParams.get("aggregate_by") || "session";

    const response = await fetch(
      `${API_URL}/api/quota/daily-report?days=${days}&aggregate_by=${aggregateBy}`,
      {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      }
    );

    const data = await response.json();
    return NextResponse.json(data, { status: response.status });
  } catch (error) {
    console.error("Daily report proxy error:", error);
    return NextResponse.json(
      { error: "Failed to fetch daily report" },
      { status: 500 }
    );
  }
}
