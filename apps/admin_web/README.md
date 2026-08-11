# NEXORA Admin Web

Operations command center for the quick-commerce platform. See `ARCHITECTURE.md` for stack, nav tree, and RBAC.

## Prerequisites

- Node.js 20+
- npm (workspace root or app-local)

## Install

From the monorepo root (preferred if using workspaces):

```bash
npm install
```

Or from this app:

```bash
cd apps/admin_web
npm install
```

`@nexora/ui` is linked via `file:../../packages/web/ui`.

## Develop

```bash
cd apps/admin_web
npm run dev
```

Open [http://localhost:3000](http://localhost:3000). Sign in on `/login` with any email/password (demo session). Default role is `admin`; elevate via code or auth store for `super_admin` kill-switch demos.

## Scripts

| Script | Command |
|--------|---------|
| Dev | `npm run dev` |
| Build | `npm run build` |
| Start | `npm run start` |
| Lint | `npm run lint` |
| Typecheck | `npx tsc --noEmit` |
| Unit tests | `npm test` |

## Feature status

See `FEATURES.md`.

## Stack notes

- Next.js App Router + React 19 + TypeScript
- TanStack Query for data hooks (mock APIs under `src/features/*/api.ts`)
- `@nexora/ui` dense admin components
- ECharts via `echarts-for-react` for analytics / monitoring / AI charts
