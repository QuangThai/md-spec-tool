import { expect, test, type Page } from "@playwright/test";

type PreviewPayload = {
  paste_text?: string;
};

type ConvertPayload = {
  pasteText?: string;
  paste_text?: string;
};

type GoogleSheetPayload = {
  url: string;
  gid?: string;
  range?: string;
};

type GoogleSheetConvertPayload = {
  url: string;
  gid?: string;
  range?: string;
};

type MockState = {
  googleAuthConnected: boolean;
  gsheetPreviewServerErrorRemaining: number;
};

const mockState: MockState = {
  googleAuthConnected: false,
  gsheetPreviewServerErrorRemaining: 0,
};

const ONBOARDING_STORAGE = {
  state: {
    hasSeenTour: true,
    isActive: false,
    currentStep: 0,
  },
  version: 0,
};

const PREVIEW_BASE = {
  headers: ["feature_name", "owner"],
  rows: [
    ["Login", "Alice"],
    ["Checkout", "Bob"],
  ],
  total_rows: 2,
  preview_rows: 2,
  header_row: 1,
  confidence: 92,
  column_mapping: {
    feature_name: "feature",
    owner: "owner",
  },
  unmapped_columns: [],
  mapping_quality: {
    score: 0.92,
    header_score: 0.9,
    mapped_ratio: 1,
    core_coverage: 1,
    core_mapped: 2,
    recommended_format: "spec",
    low_confidence_columns: [] as string[],
    column_confidence: {
      feature_name: 0.95,
      owner: 0.9,
    },
  },
  input_type: "table",
  ai_available: false,
};

const META_BASE = {
  header_row: 1,
  column_map: { feature_name: 0, owner: 1 },
  total_rows: 2,
  ai_mode: "off",
  ai_used: false,
  quality_report: {
    strict_mode: false,
    validation_passed: true,
    header_confidence: 90,
    min_header_confidence: 50,
    source_rows: 2,
    converted_rows: 2,
    row_loss_ratio: 0,
    max_row_loss_ratio: 0.1,
    header_count: 2,
    mapped_columns: 2,
    mapped_ratio: 1,
    core_field_coverage: { feature_name: true, owner: true },
  },
};

async function installWorkbenchMocks(page: Page, state: MockState) {
  let convertCount = 0;

  await page.addInitScript((storage) => {
    window.localStorage.setItem("mdflow-onboarding", JSON.stringify(storage));
  }, ONBOARDING_STORAGE);

  await page.route("**/api/quota/status", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        session_id: "e2e-session",
        used_tokens: 100,
        limit_tokens: 100000,
        remaining_tokens: 99900,
        reset_at: new Date(Date.now() + 86_400_000).toISOString(),
        status: "ok",
        daily_conversions: 2,
      }),
    });
  });

  await page.route("**/api/mdflow/templates", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        templates: [
          { name: "spec", description: "Spec", format: "markdown" },
          { name: "table", description: "Table", format: "markdown" },
        ],
      }),
    });
  });

  await page.route("**/api/mdflow/preview**", async (route) => {
    const payload = (route.request().postDataJSON() ?? {}) as PreviewPayload;
    const lowReview = payload.paste_text?.includes("LOW_REVIEW") ?? false;

    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        ...PREVIEW_BASE,
        mapping_quality: {
          ...PREVIEW_BASE.mapping_quality,
          low_confidence_columns: lowReview ? ["feature_name", "owner"] : [],
        },
      }),
    });
  });

  await page.route("**/api/mdflow/paste", async (route) => {
    convertCount += 1;
    const payload = (route.request().postDataJSON() ?? {}) as ConvertPayload;
    const pasteText = payload.paste_text ?? payload.pasteText ?? "";
    const lowReview = pasteText.includes("LOW_REVIEW");

    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        mdflow: `# Converted ${convertCount}\n\n${pasteText}`,
        warnings: [],
        meta: META_BASE,
        format: "spec",
        template: "spec",
        needs_review: lowReview,
      }),
    });
  });

  await page.route("**/api/mdflow/diff", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        format: "unified",
        added_lines: 1,
        removed_lines: 1,
        text: "-old\n+new",
        hunks: [
          {
            old_start: 1,
            old_count: 1,
            new_start: 1,
            new_count: 1,
            lines: [
              { type: "remove", line_num: 1, content: "old" },
              { type: "add", line_num: 1, content: "new" },
            ],
          },
        ],
      }),
    });
  });

  await page.route("**/api/oauth/google/status", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        connected: state.googleAuthConnected,
        email: state.googleAuthConnected ? "qa@example.com" : undefined,
      }),
    });
  });

  await page.route("**/api/oauth/google/logout", async (route) => {
    state.googleAuthConnected = false;
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ ok: true }),
    });
  });

  await page.route("**/api/gsheet/sheets", async (route) => {
    const payload = (route.request().postDataJSON() ?? {}) as GoogleSheetPayload;
    const url = payload.url ?? "";
    if (url.includes("private-401")) {
      await route.fulfill({
        status: 401,
        contentType: "application/json",
        body: JSON.stringify({
          error: "Authentication required",
          code: "UNAUTHORIZED",
          request_id: "req-401-sheets",
        }),
      });
      return;
    }
    if (url.includes("private-403")) {
      await route.fulfill({
        status: 403,
        contentType: "application/json",
        body: JSON.stringify({
          error: "Sheet access denied",
          code: "FORBIDDEN",
          request_id: "req-403-sheets",
        }),
      });
      return;
    }

    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        sheets: [
          { title: "Backlog", gid: "gid-a" },
          { title: "Roadmap", gid: "gid-b" },
        ],
        active_gid: "gid-a",
      }),
    });
  });

  await page.route("**/api/gsheet/preview", async (route) => {
    const payload = (route.request().postDataJSON() ?? {}) as GoogleSheetPayload;
    const url = payload.url ?? "";
    if (url.includes("server-500") && state.gsheetPreviewServerErrorRemaining > 0) {
      state.gsheetPreviewServerErrorRemaining -= 1;
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({
          error: "Internal processing failure",
          code: "INTERNAL_ERROR",
          request_id: "req-500-preview",
        }),
      });
      return;
    }
    if (url.includes("private-401")) {
      await route.fulfill({
        status: 401,
        contentType: "application/json",
        body: JSON.stringify({
          error: "Authentication required",
          code: "UNAUTHORIZED",
          request_id: "req-401-preview",
        }),
      });
      return;
    }
    if (url.includes("private-403")) {
      await route.fulfill({
        status: 403,
        contentType: "application/json",
        body: JSON.stringify({
          error: "Sheet access denied",
          code: "FORBIDDEN",
          request_id: "req-403-preview",
        }),
      });
      return;
    }

    const gid = payload.gid ?? "gid-a";
    const effectiveRange = payload.range ?? "";
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        ...PREVIEW_BASE,
        rows: [[`tab:${gid}`, `range:${effectiveRange || "none"}`]],
        total_rows: 1,
        preview_rows: 1,
      }),
    });
  });

  await page.route("**/api/gsheet/convert", async (route) => {
    const payload =
      (route.request().postDataJSON() ?? {}) as GoogleSheetConvertPayload;
    const gid = payload.gid ?? "gid-a";
    const effectiveRange = payload.range ?? "";

    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        mdflow: `# GSheet Convert\n\n${gid}\n${effectiveRange || "no-range"}`,
        warnings: [],
        meta: META_BASE,
        format: "spec",
        template: "spec",
        needs_review: false,
      }),
    });
  });
}

async function pressAnyShortcut(page: Page, shortcuts: string[]) {
  for (const shortcut of shortcuts) {
    await page.keyboard.press(shortcut);
    await page.waitForTimeout(120);
  }
}

test.beforeEach(async ({ page, context }) => {
  mockState.googleAuthConnected = false;
  mockState.gsheetPreviewServerErrorRemaining = 0;
  await context.grantPermissions(["clipboard-read", "clipboard-write"]);
  await installWorkbenchMocks(page, mockState);
  await page.goto("/studio");
});

test("core flow and keyboard shortcuts work", async ({ page }) => {
  const input = page.getByLabel("Paste TSV or CSV data");
  const runButton = page.getByRole("button", { name: /run/i });

  await input.fill("feature_name\towner\nLogin\tAlice");
  await expect(runButton).toBeEnabled();
  await runButton.click();

  await expect(page.getByText("Conversion complete")).toBeVisible();
  await expect(page.getByText("# Converted 1")).toBeVisible();

  const downloadPromise = page.waitForEvent("download");
  await pressAnyShortcut(page, ["Control+Shift+KeyE", "Meta+Shift+KeyE"]);
  await downloadPromise;

  await pressAnyShortcut(page, ["Control+KeyK", "Meta+KeyK"]);
  await expect(page.getByPlaceholder("Type a command or search...")).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(page.getByPlaceholder("Type a command or search...")).not.toBeVisible();

  await input.fill("feature_name\towner\nCheckout\tBob");
  await pressAnyShortcut(page, ["Control+Enter", "Meta+Enter"]);
  await expect(page.getByText("# Converted 2")).toBeVisible();
});

test("review gate locks copy/export until confirmed", async ({ page }) => {
  const input = page.getByLabel("Paste TSV or CSV data");
  const runButton = page.getByRole("button", { name: /run/i });
  const copyButton = page.getByRole("button", { name: "Copy output" });
  const exportButton = page.getByRole("button", { name: "Export output" });

  await input.fill("LOW_REVIEW\towner\nLogin\tAlice");
  await page.waitForTimeout(700);
  await runButton.click();

  await expect(page.getByText("Review Required").first()).toBeVisible();
  await expect(copyButton).toBeDisabled();
  await expect(exportButton).toBeDisabled();

  const featureNameButton = page.getByRole("button", { name: /feature_name/i });
  const ownerButton = page.getByRole("button", { name: /owner/i });
  if ((await featureNameButton.count()) > 0 && (await ownerButton.count()) > 0) {
    await featureNameButton.click();
    await ownerButton.click();
  }
  await page.getByRole("button", { name: "Confirm Review" }).click();

  await expect(page.getByText("Review confirmed")).toBeVisible();
  await expect(copyButton).toBeEnabled();
  await expect(exportButton).toBeEnabled();
});

test("diff snapshots and history modal can open and close with escape", async ({ page }) => {
  const input = page.getByLabel("Paste TSV or CSV data");
  const runButton = page.getByRole("button", { name: /run/i });

  await input.fill("feature_name\towner\nLogin\tAlice");
  await runButton.click();
  await page.getByRole("button", { name: "Save snapshot" }).click();

  await input.fill("feature_name\towner\nLogin v2\tAlice");
  await runButton.click();
  await page.getByRole("button", { name: "Save snapshot" }).click();

  await page.getByRole("button", { name: "Compare snapshots" }).click();
  await expect(page.getByText("MDFlow Diff Viewer")).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(page.getByText("MDFlow Diff Viewer")).not.toBeVisible();

  await page.getByRole("button", { name: "Show history" }).click();
  await expect(page.getByText("Conversion History")).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(page.getByText("Conversion History")).not.toBeVisible();
});

test("google sheets tab + range changes refresh preview and convert uses latest values", async ({
  page,
}) => {
  const input = page.getByLabel("Paste TSV or CSV data");
  await input.fill("https://docs.google.com/spreadsheets/d/mock-sheet-id/edit#gid=0");

  await expect(page.getByText("Google Sheet")).toBeVisible();
  await expect(page.getByText("tab:gid-a")).toBeVisible();
  await expect(page.getByText("range:none")).toBeVisible();

  const sheetSelectTrigger = page
    .getByRole("combobox", { name: /Backlog|Choose sheet/i })
    .first();
  await expect(sheetSelectTrigger).toBeVisible();
  await sheetSelectTrigger.click();
  await page.getByRole("option", { name: "Roadmap" }).click();
  await expect(sheetSelectTrigger).toContainText("Roadmap");
  await expect(page.getByText("tab:gid-b")).toBeVisible();

  const rangeInput = page.getByPlaceholder("A1:F200 or Sheet1!A1:F200");
  await rangeInput.fill("A1:B2");
  await expect(page.getByText("range:Roadmap!A1:B2")).toBeVisible();

  await page.getByRole("button", { name: /run/i }).click();
  await expect(page.getByText("# GSheet Convert")).toBeVisible();
  await expect(page.locator("pre", { hasText: "Roadmap!A1:B2" })).toBeVisible();
});

test("google auth connect and disconnect states are reflected in UI", async ({ page }) => {
  const input = page.getByLabel("Paste TSV or CSV data");
  await input.fill("https://docs.google.com/spreadsheets/d/mock-sheet-id/edit#gid=0");

  await expect(page.getByRole("button", { name: /Connect Google/i })).toBeVisible();

  mockState.googleAuthConnected = true;
  await page.evaluate(() => {
    window.postMessage({ type: "google-oauth-success" }, window.location.origin);
  });

  await expect(
    page.getByText("Google connected. You can access private sheets without sharing.")
  ).toBeVisible();
  await expect(page.getByRole("button", { name: /Disconnect/i })).toBeVisible();

  await page.getByRole("button", { name: /Disconnect/i }).click();
  await expect(page.getByRole("button", { name: /Connect Google/i })).toBeVisible();
});

test("google sheets 401 shows auth error details and recovers after input change", async ({
  page,
}) => {
  const input = page.getByLabel("Paste TSV or CSV data");
  await input.fill("https://docs.google.com/spreadsheets/d/private-401/edit#gid=0");

  await expect(page.getByText("Unauthorized")).toBeVisible();
  await expect(page.getByText("request_id=req-401-sheets")).toBeVisible();
  await expect(page.getByRole("button", { name: "Retry Preview" })).toHaveCount(0);

  await input.fill("https://docs.google.com/spreadsheets/d/mock-sheet-id/edit#gid=0");
  await expect(page.getByText("Unauthorized")).toHaveCount(0);
  await expect(page.getByText("tab:gid-a")).toBeVisible();
});

test("google sheets 403 shows forbidden error details", async ({ page }) => {
  const input = page.getByLabel("Paste TSV or CSV data");
  await input.fill("https://docs.google.com/spreadsheets/d/private-403/edit#gid=0");

  await expect(page.getByText("Access Denied")).toBeVisible();
  await expect(page.getByText("request_id=req-403-sheets")).toBeVisible();
  await expect(page.getByRole("button", { name: "Retry Preview" })).toHaveCount(0);
});

test("google sheets 500 shows Retry Preview and succeeds on retry", async ({ page }) => {
  // React Query preview has retry enabled; force enough failures to surface ErrorBanner.
  mockState.gsheetPreviewServerErrorRemaining = 4;

  const input = page.getByLabel("Paste TSV or CSV data");
  await input.fill("https://docs.google.com/spreadsheets/d/server-500/edit#gid=0");

  await expect(page.getByRole("button", { name: "Retry Preview" })).toBeVisible();
  await expect(page.getByText("request_id=req-500-preview")).toBeVisible();

  await page.getByRole("button", { name: "Retry Preview" }).click();
  await expect(page.getByRole("button", { name: "Retry Preview" })).toHaveCount(0);
  await page.getByRole("button", { name: /run/i }).click();
  await expect(page.getByText("# GSheet Convert")).toBeVisible();
});
