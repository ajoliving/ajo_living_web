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
| Frontend ad placement | `web-new/src/shared/components/ads/ListingSideAds.spec.ts` | Listing side ad display behavior. |
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
