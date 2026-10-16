# GPT image provider integration — 2026-10-15

## Production configuration

- Added channel **209**, `Apilio - gpt-image-2 - Images API`.
- Base URL `https://api.apilio.ai/v1`, protocol `openai`, channel type `llm`, model `gpt-image-2`, timeout 600 seconds.
- Reuses existing pool **115**. No upstream credentials were created or changed.
- Added model group **32**, `apilio-gpt-image` / **Apilio GPT 图片**, bound to channel 209.
- Copied billing configuration from existing Apilio channel 154. Prices were not independently reconciled against the provider's current invoice.
- Existing channel 154 and existing user key bindings were preserved. To select this provider, bind a user API key to group 32 and launch the creation center again.
- No API binary, frontend bundle, or container replacement was needed for this configuration change.

## Verification

Apilio's published synchronous endpoint specification: https://gpt-best.apifox.cn/api-447261009.md

1. Direct upstream `POST /v1/images/generations`: HTTP 200 after 50.1 seconds; one inline image.
2. Public AIPic gateway `https://image.midcode.cc/midcode-api/v1/images/generations`, with a temporary MidCode key bound only to group 32: HTTP 200 after 33.5 seconds; one image decoded successfully into a 953,471-byte PNG.
3. PNG SHA-256: `5797e9f6951d59839dd3d1b64edc914bd7c1a4e9774be591c5f3ea5bfe0f94d2`.

This verifies the public HTTP generation chain, not an interactive browser generation or image-editing request. The first Python public probe could not parse the edge response; repeating with the Go HTTP user agent succeeded.

## zzshu limitation

Existing channels 205–208 already target `https://sub.zzshu.cc/v1`, with pools 123/124 and groups 30 (`gpt-image-2.5`) / 31 (`Adobe-GPT原生生图`). Their configuration was preserved.

- Both pools expose their expected GPT image models via authenticated `/v1/models`.
- Temporary MidCode-key requests for group 30 / flare and group 31 / sunburst returned upstream HTTP 502 in under one second (logs 1159773 and 1159774).
- Direct upstream generation with a valid pixel size also returned HTTP 502.
- Diagnostic requests with `auto` and `1:1` returned explicit HTTP 400 validation errors. They do not establish successful generation: this upstream requires pixel sizes for these requests.
- These failures reproduce without AIPic and do not establish an AIPic timeout defect. zzshu generation is **not verified working**.

## Temporary credentials

Temporary MidCode API keys 434, 435, 437, and 438 were deleted through the user API after their tests. Existing production pool keys were reused only with their already-configured provider hosts. No credentials are recorded in this document.
