import ShareSlugPageClient from "./ShareSlugPageClient";

export const metadata = {
  title: "Shared Spec | MDFlow Studio",
  description: "View and download shared MDFlow specifications.",
  alternates: {
    canonical: "/s",
  },
};

export default function ShareSlugPage() {
  return <ShareSlugPageClient />;
}
