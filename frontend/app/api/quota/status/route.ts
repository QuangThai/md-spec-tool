import { NextRequest, NextResponse } from "next/server";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export async function GET(request: NextRequest) {
  try {
    const response = await fetch(`${API_URL}/api/quota/status`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    });

    const data = await response.json();
    return NextResponse.json(data, { status: response.status });
  } catch (error) {
    console.error("Quota status proxy error:", error);
    return NextResponse.json(
      { error: "Failed to fetch quota status" },
      { status: 500 }
    );
  }
}
