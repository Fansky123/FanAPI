# MidCode → AIPic

当前侧栏仅保留文本对话和创作中心；创作中心及旧 /image-gen 地址直接跳转自建 AIPic，并导入当前用户全部有效、可恢复的 API Key。文本对话仅展示文本模型与对话。

## 配置

MidCode API 进程增加环境变量：

```dotenv
AIPIC_URL=https://YOUR-STUDIO-DOMAIN/
AIPIC_HANDOFF_SECRET=REPLACE_WITH_AT_LEAST_32_RANDOM_CHARACTERS
```

本目录创建未提交的 `.env`，填入相同交接密钥：

```dotenv
MIDCODE_ORIGIN=https://YOUR-MIDCODE-DOMAIN
AIPIC_HANDOFF_SECRET=REPLACE_WITH_THE_SAME_RANDOM_SECRET
AIPIC_PORT=8790
```

执行 `docker compose -f compose.yml up -d --build`。将 AIPic 域名的 HTTPS 反向代理指向 `127.0.0.1:8790`，上传上限应覆盖参考图体积，代理读取超时应覆盖模型耗时（建议 1800 秒）。域名未确定时不要启用生产跳转。HTTP 仅允许 localhost / 127.0.0.1 开发地址。

镜像固定 AIPic commit `f6cac8328d7ddb66c610a4d3fc89423d73976789`，仅构建前端，不包含 FlyReq 后端。保留上游 MIT 许可证。升级上游时需重新构建和验证引导补丁。

## 流程与边界

1. 登录用户请求 `/api/user/image-studio/handoff`，服务端验证密钥所有者、启用状态、账户和分组绑定。
2. Redis 保存 60 秒一次性凭证（用户 ID / 密钥 ID / 模型 / 提示词，不保存明文 API Key），链接仅通过 fragment 携带随机码。
3. AIPic 在渲染前清掉 fragment，经同源 `/midcode/exchange` 兑换。网关携带服务端共享密钥访问 MidCode `/api/public/image-studio/redeem`。
4. Redis GETDEL 原子消费凭证，重新检查账户/密钥，返回解密密钥。需要 Redis 6.2+。
5. AIPic 保存为 MidCode 配置；请求经 `/midcode-api/v1/images/generations` 或 `/midcode-api/v1/images/edits` 转发到固定 MidCode 上游。保留原有鉴权、路由、费用和日志。

MidCode 公网 `/api/` 前缀沿用现有部署：反向代理移除 `/api` 后交给 Go 路由。若直连 Go 服务，需要自行增加此前缀改写。共享密钥仅在服务端配置，不能写入 VITE 变量。网关不接受用户指定上游，不转发共享密钥到生图接口，不自动重试生成 POST。

当前自动选择首个可交接密钥和首个 GPT/Grok 图片模型；进入工作台后可在 API 设置中切换密钥。模型最终权限仍由 MidCode API 校验。旧式不可解密密钥需要用户自行创建新密钥，程序不会自动创建。Gemini 原生生图接口保留，但本版 AIPic 未适配 Gemini 原生协议，因此入口仅列出 GPT/Grok 图片模型。

当前使用同步 Images（关闭流式），没有关闭页面后任务恢复能力。历史图片使用 AIPic 浏览器存储，旧 MidCode 历史不迁移；切换浏览器不会同步新作品。API Key 按 AIPic 原有行为保存在其浏览器设置中，交接码一次性不代表 API Key 一次性；退出 MidCode 不会撤销该 Key，必要时在 MidCode 密钥页禁用。

## 本地验证

```sh
go test ./internal/handler ./internal/router
node --test integrations/aipic/server.test.mjs integrations/aipic/prompt-optimizer.test.mjs
cd web/app && npm run build
```

对解压的固定版本 AIPic 执行 `node integrations/aipic/prepare.mjs /absolute/path/to/AIPic`，然后在 AIPic 目录 `npm ci`，关闭 `VITE_API_PROXY_AVAILABLE` 并 `npm run build`。网关启动需 `MIDCODE_ORIGIN`、`AIPIC_HANDOFF_SECRET`、`AIPIC_DIST` 三项环境变量；使用 `node integrations/aipic/server.mjs` 启动。

上线顺序：先部署可健康访问的 AIPic，验证交接和真实生成/编辑及扣费，再发布 MidCode 新入口。回滚入口版本时保留 API 服务和计费数据。
# 2026-10-14 多密钥接入更新

MidCode 文本对话页仅保留聊天。侧栏“创作中心”、旧 `/image-gen` 地址自动进入 AIPic。

新入口发送 `all_keys: true`，后端将当前用户有效且可恢复的密钥 ID 写入一次性交接票据；兑换时重新校验账户、密钥启用状态与归属，不接受调用方指定其他账户。所有有效密钥（包含未绑定图片模型分组的密钥）导入为独立 AIPic 配置，优先选择有可用 GPT/Grok 图片模型的配置。未绑定图片模型的配置会显示相应标注；原模型分组权限与计费不变。

旧单密钥交接格式继续兼容。密钥不放入跳转 URL，已禁用、无法恢复原值的旧密钥不会导入。


## 工作台增强

提示词旁提供优化入口，可从已导入密钥的文本模型中选择，按正常文本费率计费。原始提示词与画面描述实时同步；优化结果可编辑，确认采用才替换输入。修改原文会取消未完成请求并使旧结果失效。

历史侧栏支持单条删除和批量清空失败记录，确认时再次检查状态，正在生成的记录不会删除。仅清理浏览器历史及不再使用的图片，不删除账户调用日志。

当前仍为同步图片请求；外层 CDN 的连接限制不能通过提高网关超时覆盖。数据库内的渠道、号池、价格和密钥属于运行配置，不随源码提交发布；部署记录见 docs/image-providers-20261015.md。
