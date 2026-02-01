import GalleryPageClient from "./GalleryPageClient";

export const metadata = {
  title: "Public Gallery | MDFlow Studio",
  description: "Browse public MDFlow specifications shared by the community.",
  alternates: {
    canonical: "/gallery",
  },
};

export default function GalleryPage() {
  return <GalleryPageClient />;
}
