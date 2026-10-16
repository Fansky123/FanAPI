# AIPic prompt optimizer — 2026-10-15

## Follow-up: editable synchronized source

The original-prompt textarea is now editable and uses the same store value as the outer scene description. Manual changes use the workspace's existing edit callback, preserving its manual-edit marker. Changing either input cancels any pending optimization and clears the previous result; closing the modal retains manual edits. Empty input disables optimization, and inputs over 8000 characters display a length message. Responses are checked against the submitted source before display.

TypeScript/Vite build passed. Browser fixture verified editing inside the modal updates the outer textarea immediately, closing retains edits, outside edits appear on reopening, changing the source clears the prior optimization, and empty input disables the request. No browser error/warning logs. The backend was unchanged, so no new paid model request was needed.

Follow-up release: `/data/code/MIDCODE/releases/aipic-prompt-sync-20261015`; image `midcode-aipic:prompt-sync-20261015`; container `midcode-aipic-promptsync20261015`; port 8793. Image vhost cut over from 8792 after parallel health checks. Prior instance retained; the follow-up release's `image.midcode.cc.conf.before` restores port 8792.

## Behavior

The prompt inputs in RetouchWorkspace, RetouchSubmitBar and InputBar now offer `✨ 优化提示词`. A modal displays the original prompt, lets the user choose a text model from any imported API key, and previews an editable result. Only `采用提示词` replaces the input; closing preserves it. Empty input disables the button. A changed underlying prompt cannot be overwritten by a stale result. Closing aborts pending requests.

Model discovery uses the existing key-scoped `/api/user/channels` endpoint, not the global `/v1/models` list. The gateway exposes only `/midcode/prompt-models` and `/midcode/optimize-prompt`, targets the fixed MidCode origin, forwards the selected caller key, and never forwards the handoff shared secret. Optimization calls `/v1/chat/completions` with a fixed system instruction, nonstreaming output and a 1600-token limit. Standard text billing applies and is disclosed in the modal. Users with only image groups must bind a text group and re-enter the creation center.

Reproducible integration files are under `integrations/aipic`; `prepare.mjs` installs the UI patches via `prepare-optimizer.mjs`. The Dockerfile includes the optimizer server module and UI files.

## Validation

- TypeScript and Vite production build passed.
- Both Node gateway suites passed: image/multipart forwarding and ticket restrictions; optimizer authentication, key-scoped discovery, exclusion of image models, fixed upstream/request structure, input and body limits.
- Browser fixture at `http://127.0.0.1:18992`: open modal, load model, optimize, preview, adopt, reopen, discard, and empty-input disablement verified. Original text remained unchanged before adoption. No browser error/warning logs. Desktop screenshot inspected; mobile not separately exercised.
- Real temporary-key GPT request on the parallel instance: HTTP 200 in 9.5 seconds, Chinese optimized prompt returned.
- Public `https://image.midcode.cc/midcode/optimize-prompt`: HTTP 200 in 19.8 seconds. Public model discovery also returned HTTP 200 with the key's permitted models.
- Public browser renders the optimizer button with no error/warning logs. Both public entry JS/CSS assets match the release bytes. Cloudflare modifies the HTML response; HTML byte equality was not claimed.
- Fifteen public health probes across cutover returned HTTP 200 / `ok`.
- Temporary keys 439, 440, 441 and 442 were deleted. DeepSeek testing returned no final text / upstream failure; not every listed upstream model is guaranteed healthy. The UI preserves the original and reports these errors.

## Production and rollback

- Release directory: `/data/code/MIDCODE/releases/aipic-prompt-20261015`.
- Image: `midcode-aipic:prompt-20261015`; container: `midcode-aipic-prompt20261015`, host network, loopback port 8792, restart policy `unless-stopped`.
- Runtime environment file is mode 0600 and reuses existing handoff secret and MidCode origin. No credentials are in source or this report.
- OpenResty image vhost now targets 8792. Previous container on 8791 remains available. Restore `image.midcode.cc.conf.before` from the release directory to `/opt/1panel/apps/openresty/openresty/conf/conf.d/image.midcode.cc.conf`, validate `nginx -t` in `1Panel-openresty-0oWm`, then reload to roll back.
- MidCode API container/binary was not changed.
- Temporary unpublished GitHub transfer release was removed after checksum-verified transfer. This was deployment, not a source commit/push.
