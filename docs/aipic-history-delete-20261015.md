# AIPic history deletion — 2026-10-15

Added delete buttons beside completed/failed records in the retouch/generation history sidebar, plus `清空失败记录`. Uses the existing confirmation dialog and `removeMultipleTasks`, preserving shared image references and current inputs. Running records have no deletion button. On confirmation the status of each captured task is rechecked, so restarted tasks and new tasks are excluded. Deleting a selected record uses the workspace's existing selection reset effect. Only browser history and unreferenced local images are removed; account API logs/billing are not changed.

Reproducible integration: `HistoryDelete.tsx`, `history-delete.css`, `prepare-history.mjs`, invoked by `prepare.mjs` and included in the Dockerfile.

Validation: TypeScript/Vite build passed. Browser fixture tested cancellation, single deletion, persistence after reload, and bulk deletion of two failed records. No browser errors/warnings. Tests created only disposable local failure records, not production records. No paid generation requests were made.

Production release: `/data/code/MIDCODE/releases/aipic-history-20261015`; image `midcode-aipic:history-20261015`; container `midcode-aipic-history20261015`, port 8794, restart policy `unless-stopped`. OpenResty image vhost switches from 8793 after health checks. Previous container retained. Restore `image.midcode.cc.conf.before` from this release to the image vhost and validate/reload OpenResty for rollback. MidCode API untouched.
