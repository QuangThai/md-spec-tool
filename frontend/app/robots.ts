import type { MetadataRoute } from "next";

const baseUrl = "https://md-spec-tool.vercel.app";

export default function robots(): MetadataRoute.Robots {
  return {
    rules: [
      {
        userAgent: "*",
        allow: "/",
      disallow: ["/studio", "/share", "/s", "/api"],
      },
    ],
    sitemap: `${baseUrl}/sitemap.xml`,
  };
}
