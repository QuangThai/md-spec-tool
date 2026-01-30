import type { MetadataRoute } from "next";

export default function robots(): MetadataRoute.Robots {
  return {
    rules: [
      {
        userAgent: "*",
        allow: ["/", "/docs"],
        disallow: ["/studio", "/api"],
      },
    ],
    sitemap: "https://md-spec-tool.vercel.app/sitemap.xml",
  };
}
