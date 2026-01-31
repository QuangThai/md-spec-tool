import type { MetadataRoute } from "next";

const baseUrl = "https://md-spec-tool.vercel.app";

export default function sitemap(): MetadataRoute.Sitemap {
  const lastModified = new Date();

  return [
    {
      url: `${baseUrl}/`,
      lastModified,
      changeFrequency: "weekly",
      priority: 1,
    },
    {
      url: `${baseUrl}/docs`,
      lastModified,
      changeFrequency: "weekly",
      priority: 0.8,
    },
    {
      url: `${baseUrl}/batch`,
      lastModified,
      changeFrequency: "monthly",
      priority: 0.7,
    },
  ];
}
