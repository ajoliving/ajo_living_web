# Project Prompt Index

## Prompt Surfaces
| Surface | File | Purpose | Protected by |
| --- | --- | --- | --- |
| Repository operating rules | `AGENTS.md` | Project-wide agent rules, product boundaries, module ownership, communication style, and deployment guardrails. | Manual review before edits |
| Active frontend rules | `web-new/AGENTS.md` | Vue 3 frontend structure, page singleton layout, API boundaries, component rules, and implementation style for the active web project. | `npm run typecheck`, `npm run test:unit` when frontend behavior changes |
| Backend rules | `http_service/AGENTS.md` | Go service layering, model/API ownership, legacy integration rules, comments, tests, and deployment safeguards. | `go test ./...` when backend behavior changes |
| Archived frontend rules | `web/AGENTS.md` | Historical frontend reference; not the default target for new work. | Manual review only when user explicitly asks to modify `web` |
| Docs memory map | `docs/AGENTS.md` | Local ownership map for documentation, prompts, and maintenance protocol. | Manual review |
| Prompt directory map | `docs/prompts/AGENTS.md` | Local ownership and maintenance rules for reusable UI prompt files. | Manual review |
| Deployment prompt map | `docs/deployment/AGENTS.md` | Local ownership and safety protocol for deployment prompt files. | Manual server inspection before deployment |
| Frontend page memory map | `web-new/src/pages/AGENTS.md` | Local page tree ownership, module boundaries, and UI prompt landing rules. | `npm run typecheck`, `npm run test:unit` when frontend behavior changes |
| Backend internal memory map | `http_service/internal/AGENTS.md` | Local backend layer ownership, domain boundaries, and API contract impact rules. | `go test ./...` when backend behavior changes |
| Public UI baseline prompt | `docs/prompts/ajo-living-public-ui-style-prompt.md` | Reusable baseline for formal AJO Living public page UI, copy, layout, and visual restrictions. | Manual visual review and responsive check |
| Page UI prompts | `docs/prompts/home-page-prompt.md`, `docs/prompts/login-page-prompt.md`, `docs/prompts/marketplace-discover-page-prompt.md`, `docs/prompts/marketplace-filter-page-prompt.md`, `docs/prompts/marketplace-publish-page-prompt.md` | Reusable page-specific UI prompts for frontend implementation or redesign work. | Manual visual review and relevant frontend checks |
| Form dialog prompt | `docs/prompts/global-form-dialog-prompt.md` | Reusable form modal structure, field, state, and acceptance rules. | Manual review and relevant form tests |
| Deployment prompts | `docs/deployment/server-deployment-ai-prompt.md`, `docs/deployment/oss-cdn-letsencrypt-renewal-prompt.md` | Server deployment, existing infrastructure, SSL renewal, and operational safety instructions. | Manual server inspection before deployment |
| API documentation prompt | `docs/development/api-documentation-guide.md` | AI-facing structure and execution prompt for generating HTTP API documentation. | Manual contract review |
| Domestic SMS OTP setup | `http_service/doc/sms.md` | Alibaba Cloud mainland OTP configuration, security boundary, and runtime limits. | `go test ./...` and manual production credential verification |
| Project skill prompt | `.codex/skills/form-to-view-parity-audit/SKILL.md`, `.codex/skills/form-to-view-parity-audit/agents/openai.yaml` | Local Codex skill for form-to-view parity audits and its default prompt. | Manual audit output review |
| GitHub assistant instructions | `.github/copilot-instructions.md` | Reserved assistant instruction surface; currently empty. | Manual review if populated |

## Runtime Prompt Chain
- Product runtime code currently has no LLM request chain and no model-facing prompt builder.
- Repository work starts from `AGENTS.md`, then follows the nearest child `AGENTS.md` for touched files.
- UI prompt work uses `docs/prompts/AGENTS.md`, the relevant prompt file, `web-new/AGENTS.md`, and `web-new/src/pages/AGENTS.md`, then verifies against the relevant page and responsive states.
- Deployment work uses `docs/deployment/AGENTS.md` and `docs/deployment/server-deployment-ai-prompt.md`, then must inspect the live server configuration before changing deployment state.
- Backend contract work uses `http_service/AGENTS.md` and `http_service/internal/AGENTS.md` before changing handler, service, model, router, or API docs.
- Form parity work uses `.codex/skills/form-to-view-parity-audit/SKILL.md`, then traces form input, payload, API model, preview, public display, validation, and tests.

## Tool And Schema Prompts
- `.codex/skills/form-to-view-parity-audit/SKILL.md` defines the local parity-audit workflow and output contract.
- `.codex/skills/form-to-view-parity-audit/agents/openai.yaml` defines the default prompt shown to the related skill interface.
- `docs/development/api-documentation-guide.md` contains an AI execution prompt for API documentation generation.
- No product runtime tool schema or hidden model policy file is present as of 2026-07-08.

## Contract Tests
| Area | File | Coverage |
| --- | --- | --- |
| Frontend payment page | `web-new/src/pages/payments/pay/Page.spec.ts` | POS payment page behavior and state contract. |
| Property publishing flow | `web-new/src/pages/property/editor/PropertyEditorPage.spec.ts`, `http_service/internal/service/property_publish_flow_test.go`, `http_service/internal/service/auth_service_test.go` | Account-type-derived owner or approved-agent publishing identity, backend-authorized public agent snapshots, shared applicable property-field ordering across residential, car park, industrial, shop, and land forms, incomplete draft saving, atomic charged draft creation and update, direct-publish draft fee exclusion, image payload update, formal publish call, missing-field guidance, photo requirement, wallet-backed publish transition, and land listing validation. |
| Frontend ad placement | `web-new/src/shared/components/ads/ListingSideAds.spec.ts` | Listing side ad display behavior. |
| Frontend mobile baseline | `web-new/src/styles/mobile-baseline.spec.js`, `web-new/src/shared/components/navigation/AppHeader.spec.ts`, `web-new/scripts/audit-mobile.mjs` | iPhone 14 Pro Max safe-area, centered mobile Logo, left menu control, right theme and locale controls, signed-out header login entry, partial-width mobile drawer with backdrop close, property and furniture layouts with the supermarket-offer page control-band spacing, an independent search field, and horizontally scrollable quick filters that include sorting, filter sheets anchored to the viewport bottom with visible action controls in standard, short, and landscape viewports, bottom navigation, hidden desktop footer, route-component mobile coverage, legacy redirect coverage, live API guest-route coverage with real listing IDs, management-tab coverage, guest protected-route redirect and nav-visibility coverage, member-to-staff route interception and nav-visibility coverage, sectioned browser audit coverage, body-level dialog reachability, body-level touch targets, visible touch-target sizing, document-level touch-drag scrolling, focused control clearance, bottom-of-page action clearance, chat composer clearance, iOS viewport-height guardrails, browser route audit, browser interaction-flow audit, console and network failure capture, mobile drawer flows, and auth-aware mobile navigation. |
| Frontend session guard | `web-new/src/stores/session.spec.ts` | Concurrent session hydrate behavior used by authenticated and staff route guards. |
| Staff authorization boundary | `web-new/src/router/index.spec.ts`, `http_service/internal/service/auth_service_test.go`, `http_service/internal/service/pos_payment_security_test.go` | Staff guards remain limited to AJO management surfaces, iSmart staff status cannot promote AJO staff access, and regular members can use POS features within resident building and unit permissions. |
| Frontend member building and unit display | `web-new/src/pages/account/my/composables/unit-display.spec.ts`, `web-new/src/pages/building/BuildingPage.spec.ts`, `web-new/src/pages/building/composables/building-display.spec.ts`, `http_service/internal/service/ismart_external_service_test.go`, `http_service/internal/service/user_service_test.go` | POS unit-detail precedence, latest saved building binding selection, member refresh on building-page entry, official building-name fallback, rejection of raw building-ID placeholder names, and shared iSmart building cache permission isolation. |
| Frontend i18n and theme tokens | `web-new/src/i18n/index.spec.ts`, `web-new/src/utils/theme.spec.ts` | English and zh-HK message-key parity, serviced-residence channel copy, locale synchronization, and semantic-to-legacy theme-token synchronization including dark primary-surface contrast. |
| Backend services | `http_service/internal/service/*_test.go` | Auth, chat, iSmart, trend, notification, POS payment, property, secondhand, security CCTV, and wallet service behavior. |
| Backend seeds and migrations | `http_service/internal/database/*_test.go` | Admin seed, media asset migration, and system notification seed behavior. |

No direct prompt regression test currently guards prompt wording or localization leakage. Prompt changes require manual review plus the nearest runtime tests when behavior is affected.

## Maintenance Protocol
- Add or move prompt-bearing docs only under the owning directory, then update this index.
- Update the nearest `AGENTS.md` when a directory gains, loses, moves, or changes responsibility.
- Add dated change-log lines in `YYYY-MM-DD: summary` format to changed memory files.
- Keep prompt docs in 繁體中文 or English unless preserving an existing legacy document.
- Do not index generated output, `dist`, `node_modules`, `.claude/worktrees`, `docs/archive`, database dumps, or backups as stable prompt surfaces.

## Change Log
2026-07-08: Established the project prompt/context index and connected docs, AGENTS files, deployment prompts, and local Codex skill prompts.
2026-07-08: Added local memory maps for reusable UI prompts, deployment prompts, frontend pages, and backend internal layers.
2026-07-08: Archived docs product, prototypes, and backups under `docs/archive` and excluded them from prompt/context indexing.
2026-07-08: Recorded POS iSmart relay v2 deployment chain and verification rules in the deployment prompt.
2026-07-08: Recorded AJO backend iSmart integration base URL deployment requirements alongside the legacy external app API.
2026-07-09: Recorded `我的大廈` finance and owner-account AJO member API usage in the iSmart integration docs.
2026-07-09: Added the default post-change prompt/context memory audit rule to the repository instructions.
2026-07-09: Recorded the `web-new` mobile baseline for iPhone 14 Pro Max safe-area, bottom navigation, single-column pages, scrollable tables, and reachable mobile form actions.
2026-07-09: Added route-aware frontend mobile baseline tests for protected, management, payment, editor, settings, chat, and body-level dialog coverage.
2026-07-09: Added session hydrate guard coverage so staff-only mobile pages wait for current member data before permission checks.
2026-07-09: Tightened the frontend mobile baseline contract for body-level dialog touch targets and chat composer bottom-navigation clearance.
2026-07-09: Extended the frontend mobile baseline to keep focused controls and form actions clear of the fixed bottom navigation.
2026-07-09: Extended the frontend mobile baseline to cover every component route and reject `100vh` or Tailwind screen-height utilities in page, shared, and style sources.
2026-07-09: Added `npm run audit:mobile` as the repeatable iPhone 14 Pro Max browser audit for public, member, staff, payment, chat, settings, and 404 routes.
2026-07-09: Extended `npm run audit:mobile` to cover login, forgot password, chat, payment, wallet recharge, editor, notice, hero, and display-ad interaction flows.
2026-07-09: Extended `npm run audit:mobile` to cover legacy route redirects, all management tabs, bottom navigation switching, and management-tab switching.
2026-07-09: Extended `npm run audit:mobile` to fail undersized visible touch targets and added mobile touch baselines for footer, breadcrumb, pagination, favorite, and toast controls.
2026-07-09: Extended `npm run audit:mobile` to fail console warnings/errors, failed audited requests, API 4xx/5xx responses, and mobile drawer flows.
2026-07-09: Extended `npm run audit:mobile` to cover iPhone 14 Pro Max short-viewport and landscape focused routes with 1023px touch-size baselines.
2026-07-09: Extended `npm run audit:mobile` to fail actionable content covered by the fixed bottom navigation after scrolling to the page bottom.
2026-07-09: Extended `npm run audit:mobile` to verify signed-out protected routes redirect to login and mobile navigation does not expose private entries.
2026-07-09: Extended `npm run audit:mobile` to verify regular members are redirected from staff management and settings routes to the member center, with mobile navigation hiding management entries.
2026-07-10: Extended mobile browser audit live API mode to derive public detail routes from real list data and reserve mutation or OTP flows for mock or explicitly allowed runs.
2026-07-10: Recorded legacy public detail redirects as list fallbacks so live datasets without sample IDs do not send mobile users to 404 detail pages.
2026-07-10: Extended the frontend mobile baseline to reject document-level touch-scroll blockers and probe real touch-drag scrolling on key mobile routes.
2026-07-10: Extended the frontend mobile baseline to hide the desktop footer at the mobile navigation breakpoint and fail browser audits when it remains visible.
2026-07-10: Extended the frontend mobile baseline to verify the left menu, centered Logo, right theme and locale controls, and the mobile locale toggle flow.
2026-07-10: Extended the frontend mobile baseline to verify the partial-width mobile drawer and backdrop-close interaction.
2026-07-10: Extended the frontend mobile baseline to verify the signed-out header login entry routes directly to the login page.
2026-07-10: Recorded compact mobile filter sheets for property and furniture lists, with the filter control placed before the search field and browser interaction coverage for opening and closing each sheet.
2026-07-10: Extended filter-sheet mobile checks to require viewport-bottom anchoring and visible result actions in standard, short, and landscape iPhone 14 Pro Max viewports.
2026-07-10: Added property publishing contract tests for draft creation, image persistence, formal publishing, precise missing-field guidance, photo enforcement, and land listings without an estate name.
2026-07-12: Recorded the property contract that unfinished data may be saved as a draft while full validation runs only at formal publication.
2026-07-12: Recorded the public property and furniture mobile control band alignment with supermarket offers, including single-layer margins, independent search, and horizontally scrollable quick filters with sorting.
2026-07-12: Recorded the fixed mobile quick-filter sets for property region, property type, bedrooms, renovation, and sorting, plus furniture category, area, condition, and sorting.
2026-07-12: Recorded the shared compact mobile workspace baseline for member, notification, management, and settings shells, including horizontal navigation, member action grids, two-column wallet summaries, and horizontal listing status filters.
2026-07-14: Recorded the account-derived property publisher identity contract: owner and agent accounts can publish, and the publishing page cannot switch identity.
2026-07-15: Replaced member-managed publisher identity with registered personal, individual-agent, and agency-company account types plus approval-gated publishing and company subaccounts.
2026-07-14: Recorded the shared applicable building-field order across residential, car park, industrial, shop, and land publishing forms.
2026-07-14: Extended shared property-field ordering to applicable direction and management fee fields, and removed the publishing identity helper sentence.
2026-07-15: Recorded atomic 1,000 AJO Points charging for explicit property draft saves and excluded direct-publish staging saves from duplicate draft charges.
2026-07-16: Registered email-password accounts now create and link an iSmart account before local persistence; iSmart failure leaves no local AJO account, and account type derives the iSmart legal-entity value.
2026-07-16: Recorded iSmart registration-profile snapshot ownership: use upstream values first, backfill only missing member fields from AJO registration, exclude passwords, and keep the member-center iSmart view separate from AJO-local account data.
2026-07-16: Recorded the `ismart_raw` contract: retain upstream registration and POS business fields, merge only usable updates, recursively remove credentials and session material, and keep the stable `ismart_msg` summary separate.
2026-07-16: Recorded personal registration OwnerReg submission after local and iSmart account creation; HTTP 2xx creates only a pending residence application and never an approved AJO property binding.
2026-07-16: Recorded member unit display precedence: resolve `floor` and `unit` by POS `unit_id`, using encoded flat-unit permissions only for authorization and failed-detail fallback.
2026-07-16: Recorded resident-only member-center display and binding for Staff and non-Staff accounts; managed buildings stay outside the personal binding UI.
2026-07-16: Recorded the Staff authorization boundary: AJO Staff protects management surfaces only, iSmart Staff remains external metadata, and ordinary features use member and resource visibility checks.
2026-07-16: Recorded the latest member-building selection and display-name contract, including session refresh, saved-binding precedence, POS name fallback, and placeholder-ID correction.
2026-07-15: Recorded the backend contract for restricted agency registration, versioned individual and company review profiles, company subaccounts, and account-derived property publisher snapshots.
2026-07-15: Recorded company subaccount publish/manage enforcement and private signed-download handling for agency licence evidence.
2026-07-15: Recorded atomic property contact snapshot refresh after agency profile revision approval, including linked company subaccounts.
2026-07-15: Recorded local-first account login for restricted pending or rejected agency accounts and non-blocking review result email delivery with rejection reasons.
2026-07-15: Added frontend i18n parity and theme-token checks for public copy and dark-mode surface contrast.
2026-07-15: Recorded Alibaba Cloud mainland SMS OTP provider configuration, `+86` delivery boundary, cooldown, and in-memory OTP runtime limit.
2026-07-16: Recorded Redis-backed five-minute iSmart building-data caching, permission-before-cache enforcement, fail-open upstream behavior, and local or production loopback-only Docker deployment.
2026-07-16: Recorded production verification for iSmart building caching: first-request fill, second-request Redis hit without a new miss, countdown TTL, equal business data, and exclusion of member metadata from shared cache.
