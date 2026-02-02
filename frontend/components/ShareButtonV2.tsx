import { Share2 } from "lucide-react";
import React, { Dispatch, SetStateAction } from "react";
import { Tooltip } from "./ui/Tooltip";
import { Select } from "./ui/Select";

interface ShareButtonV2Props {
  mdflowOutput: string;
  shareOptionsRef: React.RefObject<HTMLDivElement>;
  setShowShareOptions: Dispatch<SetStateAction<boolean>>;
  showShareOptions: boolean;
  shareTitle: string;
  setShareTitle: Dispatch<SetStateAction<string>>;
  shareSlug: string;
  setShareSlug: Dispatch<SetStateAction<string>>;
  shareVisibility: "public" | "private";
  setShareVisibility: Dispatch<SetStateAction<"public" | "private">>;
  shareAllowComments: boolean;
  setShareAllowComments: Dispatch<SetStateAction<boolean>>;
  creatingShare: boolean;
  handleCreateShare: () => Promise<void>;
  shareSlugError: string | null;
}

function ShareButtonV2({
  mdflowOutput,
  shareOptionsRef,
  setShowShareOptions,
  showShareOptions,
  shareTitle,
  setShareTitle,
  shareSlug,
  setShareSlug,
  shareVisibility,
  setShareVisibility,
  shareAllowComments,
  setShareAllowComments,
  creatingShare,
  handleCreateShare,
  shareSlugError,
}: ShareButtonV2Props) {
  return (
    <Tooltip content={creatingShare ? "Sharing..." : "Share"}>
      <div className="relative" ref={shareOptionsRef}>
        <button
          type="button"
          onClick={() => {
            if (!mdflowOutput || creatingShare) return;
            setShowShareOptions((prev) => !prev);
          }}
          disabled={!mdflowOutput || creatingShare}
          className={`p-1.5 sm:p-2 rounded-lg border transition-all ${
            mdflowOutput && !creatingShare
              ? "bg-white/5 hover:bg-white/10 border-white/10 hover:border-white/20 text-white/60 hover:text-white"
              : "bg-white/5 border-white/5 text-white/20 cursor-not-allowed"
          }`}
        >
          <Share2 className="w-3.5 h-3.5" />
        </button>

        {showShareOptions && (
          <div className="absolute right-0 top-full mt-2 w-64 rounded-xl border border-white/10 bg-black/90 backdrop-blur-xl p-3 shadow-2xl">
            <div className="space-y-3 text-[10px] text-white/70">
              <div>
                <label className="block text-[9px] uppercase tracking-widest text-white/40 mb-1">
                  Title
                </label>
                <input
                  value={shareTitle}
                  onChange={(event) => setShareTitle(event.target.value)}
                  placeholder="Optional title"
                  className="w-full rounded-md bg-white/5 border border-white/10 px-2 py-1.5 text-[10px] text-white/80 focus:outline-none focus:border-accent-orange/40"
                />
              </div>
              <div>
                <label className="block text-[9px] uppercase tracking-widest text-white/40 mb-1">
                  Custom Slug
                </label>
                <input
                  value={shareSlug}
                  onChange={(event) => setShareSlug(event.target.value)}
                  placeholder="my-spec"
                  className={`w-full rounded-md bg-white/5 border px-2 py-1.5 text-[10px] text-white/80 focus:outline-none ${
                    shareSlugError
                      ? "border-red-400/60"
                      : "border-white/10 focus:border-accent-orange/40"
                  }`}
                />
                {shareSlugError && (
                  <p className="mt-1 text-[9px] text-red-400/80">
                    {shareSlugError}
                  </p>
                )}
              </div>
              <div className="flex items-center justify-between">
                <label className="text-[9px] uppercase tracking-widest text-white/40">
                  Visibility
                </label>
                <Select
                  value={shareVisibility}
                  onValueChange={(value) =>
                    setShareVisibility(value === "public" ? "public" : "private")
                  }
                  options={[
                    { label: "Public", value: "public" },
                    { label: "Private", value: "private" },
                  ]}
                  size="compact"
                  className="h-7 text-[10px] text-white/80 border-white/10"
                  aria-label="Share visibility"
                />
              </div>
              <div className="flex items-center justify-between">
                <label className="text-[9px] uppercase tracking-widest text-white/40">
                  Comments
                </label>
                <button
                  type="button"
                  onClick={() =>
                    setShareAllowComments((prev) => !prev as boolean)
                  }
                  className={`px-2 py-1 rounded-md text-[9px] uppercase tracking-widest border ${
                    shareAllowComments
                      ? "border-emerald-400/40 text-emerald-300"
                      : "border-white/20 text-white/40"
                  }`}
                >
                  {shareAllowComments ? "Enabled" : "Disabled"}
                </button>
              </div>
              <button
                type="button"
                onClick={handleCreateShare}
                disabled={creatingShare}
                className="w-full mt-1 px-3 py-2 rounded-lg bg-accent-orange hover:bg-accent-orange/90 text-[10px] font-bold uppercase tracking-wider text-white disabled:opacity-50"
              >
                {creatingShare ? "Sharing..." : "Create Share Link"}
              </button>
            </div>
          </div>
        )}
      </div>
    </Tooltip>
  );
}

export default ShareButtonV2;
