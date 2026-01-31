import type { MetadataRoute } from "next";

export default function robots(): MetadataRoute.Robots {
  return {
    rules: [
      {
        userAgent: "*",
        allow: ["/", "/docs", "/batch"],
        disallow: ["/studio", "/share", "/api"],
      },
    ],
    sitemap: "https://md-spec-tool.vercel.app/sitemap.xml",
  };
}
