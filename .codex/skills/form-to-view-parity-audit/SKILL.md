---
name: form-to-view-parity-audit
description: Audit and fix product field parity across create/edit forms, request payloads, backend models or APIs, preview screens, public detail pages, validation, i18n labels, and tests. Use when the user asks whether submitted fields are fully represented downstream, whether a publishing form matches a public detail page, whether fields are missing or duplicated across UI surfaces, or asks for a reusable workflow for form-to-view completeness checks.
---

# Form-to-View Parity Audit

## Purpose

Use this skill to run a field parity audit: confirm that every user-entered field has an intentional downstream outcome, and that every public/preview field is backed by a clear source. Treat this as a product traceability task, not just a code search.

Useful terms:
- `Field parity audit`
- `Form-to-view parity check`
- `Requirements traceability matrix`
- `Content model coverage`
- `Data lineage check`

## Workflow

1. Identify the surfaces.
   - Form or editor UI.
   - Form state/composable.
   - Submit payload builder.
   - API model or backend DTO when present.
   - Preview UI when present.
   - Public detail/list UI.
   - Validation and tests.

2. Build a field matrix.
   Use this table shape:

   ```text
   Field | Form input | Payload/API | Preview | Public detail | Validation | Status | Action
   ```

   Status values:
   - `covered`: captured, persisted, and intentionally displayed or intentionally hidden.
   - `missing-display`: captured/persisted but not shown where product expects it.
   - `missing-submit`: shown in form but not submitted.
   - `derived`: not directly shown/submitted, but intentionally derived from another field.
   - `duplicate`: same meaning appears twice in one UI flow.
   - `intentional-hidden`: retained for backend/admin use but not public.
   - `dead-field`: no clear runtime use.

3. Check product semantics before editing.
   - Do not assume every form field belongs on the public page.
   - Mark private/admin fields separately.
   - If two fields overlap, choose one owner and derive the other when needed.
   - Prefer room/item-level fields over project-level duplicates when users make decisions at item level.

4. Implement only clear fixes.
   - Remove duplicate UI fields when another more specific field owns the meaning.
   - Add missing public detail display only when it helps user decisions.
   - Keep backend compatibility by deriving legacy required fields from canonical fields.
   - Update preview to match public detail for owner review.
   - Update i18n labels when visible copy changes.

5. Validate.
   - Run typecheck or the narrowest relevant test command.
   - If behavior crosses frontend/backend contracts, add or update focused tests.
   - Report only commands and pass/fail unless details are needed.

## Output Format

Start with the audit conclusion, then the matrix, then fixes.

Keep the final response concise:

```text
Conclusion:
<1-3 lines>

Field Parity Matrix:
<table>

Fixes Applied:
<short bullets>

Verification:
<command>: <pass/fail>
```

If no edits were made, replace `Fixes Applied` with `Recommended Fixes`.

## Project-Specific Notes For ajoliving_web

Follow repository instructions first.

For property/service residence work, inspect these files before editing:
- `web/src/pages/property/editor/PropertyEditorPage.vue`
- `web/src/pages/property/detail/PropertyDetailPage.vue`
- `web/src/model/property.ts`
- `web/src/httpapis/properties.ts`
- `web/src/i18n/locales/zh-HK/property.ts`
- `web/src/i18n/locales/en/property.ts`
- Backend DTO/model files under `http_service/internal/handler/` and `http_service/internal/service/` if API contract changes.

For service residences, common canonical ownership:
- Project-level basics: project name, English name, address, district, website, WhatsApp, fax.
- Room-level decision fields: room type, room category, area, rent range, minimum stay.
- Content sections: residence info, services, facilities, extra charges.
- Contact: WhatsApp primary action, unlockable phone or other contact payload.
- Derived summary fields: lowest rent, minimum usable area, project-level minimum stay when backend still requires it.

Do not add decorative copy to public pages. The owner cares about runtime truth, product effect, risk, decisions, and verification status.
