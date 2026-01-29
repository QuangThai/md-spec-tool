import MDFlowWorkbench from "@/components/MDFlowWorkbench";
import StudioPageHeader from "@/components/StudioPageHeader";

export const metadata = {
  title: "Studio | MDFlow Studio",
  description: "Advanced technical specification transformation engine.",
};

export default function StudioPage() {
  return (
    <div className="space-y-4 sm:space-y-6 lg:space-y-8 px-4 sm:px-6 lg:px-8 max-w-[1600px] mx-auto">
      {/* <StudioPageHeader /> */}
      <MDFlowWorkbench />
    </div>
  );
}
