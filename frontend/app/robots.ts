import type { MetadataRoute } from "next";

const baseUrl = "https://md-spec-tool.vercel.app";

export default function robots(): MetadataRoute.Robots {
  return {
    rules: [
      {
        userAgent: "*",
        allow: "/",
        disallow: ["/studio", "/share", "/api"],
      },
    ],
    sitemap: `${baseUrl}/sitemap.xml`,
  };
}
