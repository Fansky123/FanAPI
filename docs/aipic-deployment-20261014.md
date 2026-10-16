# MidCode → AIPic 部署（2026-10-14）

## 当日第二次更新：直接跳转与全部密钥

此节取代下文初次发布的入口与端口描述：

- 文本对话页仅保留聊天；侧栏增加独立“图片创作”，与“创作中心”均自动跳转 AIPic，不再展示模型/密钥选择中间页。
- 全部有效且可恢复的用户密钥作为独立配置导入，包含未绑定图片模型的密钥；优先激活有图片模型的配置。没有图片权限的密钥仍受原模型分组限制。
- 新生产 API `midcode-multikey20261014-api-1` 监听 `8101`；AIPic `midcode-multikey20261014-aipic-1` 监听 `8791`。上一版 `8100/8790` 保留。
- 发布目录 `/data/code/MIDCODE/releases/aipic-multikey-20261014`，Compose `/data/code/MIDCODE/docker-compose.multikey-20261014.yml`，继续使用初次发布目录的 `.env`。
- 真实账户验证 10 个有效密钥全部匹配，其中 2 个有图片模型；默认配置有图片模型。公网交接、重放拒绝、匿名拒绝和模型代理验证通过。
- 33 组连续探测中，主站和 AIPic 均为 HTTP 200；关键公网资源哈希一致，旧 API 与 worker 未重启。
- 后端专项测试、生产基线 handler/router 测试、前端构建与相关 ESLint/TypeScript 检查通过。浏览器确认纯文本页面、两入口自动跳转、全部密钥可切换及默认图片配置。
- 回滚：`python3 /data/code/MIDCODE/releases/aipic-multikey-20261014/deploy.py rollback`，同时恢复主站和 AIPic 上一版入口。
- 验收：发布目录中的 `verify-local.json`、`verify-public.json`、`continuity.json`、`deployed.json`。本地证据 `artifacts/aipic-20261014/multikey-deployed.json`。

## 上线范围

- 图片工作台：https://image.midcode.cc 。用户从 MidCode 的创作中心、聊天页图片模式或 `/image-gen` 进入。
- 交接链接只含 60 秒有效的一次性凭证；AIPic 网关兑换当前用户的 API Key，浏览器立即移除地址栏凭证。
- AIPic 固定版本：`f6cac8328d7ddb66c610a4d3fc89423d73976789`。使用前端及本仓库网关，不部署上游 FlyReq 后端。
- 图片请求仍经过 MidCode 的鉴权、渠道和计费。保留图片转发 API；替换原生图界面，不影响视频和文本对话。
- 当前自动交接选择 GPT Image / Grok 图片模型；未接入 Gemini 或后台任务恢复。

## 生产拓扑与基线

- OpenResty 主站代理从 `127.0.0.1:8099` 切到 `127.0.0.1:8100`。
- 新 API：`midcode-aipic20261014-api-1`，镜像 `fanapi-api:aipic-20261014`。
- AIPic：`midcode-aipic20261014-aipic-1`，镜像 `midcode-aipic:20261014`，宿主机网络，仅监听 `127.0.0.1:8790`。
- AIPic 固定回源 `http://127.0.0.1:8100`，避免内部请求绕行 Cloudflare。
- 旧 API `midcode-ping20261009-api-1` 保留；原任务 worker 未重启。
- 发布目录：`/data/code/MIDCODE/releases/aipic-20261014`。
- Compose：`/data/code/MIDCODE/docker-compose.aipic-20261014.yml`。操作时必须指定发布目录的 `--env-file .env`；共享密钥仅保存在服务器权限为 `0600` 的文件中。

后端基于已核对哈希的生产 `server-ping-20261009/source`，只叠加交接 handler、测试与两个路由。前端基于 `display-points-20261013` 已上线资源，仅替换三个图片入口、增加启动组件并整体更新 JS 资源版本。没有将本地工作区其他未发布修改一并部署。

## 验证与证据

- 基线 Go handler/router 测试通过；新 API 与 AIPic 容器健康。
- 公网交接返回 200，密钥、模型、提示词匹配；再次兑换返回 410；未登录返回 401；无共享密钥兑换返回 403；非允许代理路由返回 404。
- 浏览器验证生产前端补丁：文本对话不会自动跳转，图片模式才启动交接；视频表单仍可访问；创作中心保留历史记录；旧生图链接可跳转。
- 公网 AIPic HTTPS 页面正常渲染，无浏览器脚本错误。
- 切换期间 37 次主站可用性探测全部返回 200；旧 API 与 worker 启动时间、重启次数不变。
- 未发起收费的真实生图请求；鉴权交接、模型代理与本地模拟生成已验证。
- 服务器记录：`verification-local.json`、`verification-public.json`、`continuity-probes.json`、`manifest.json`、`deployed.json`。公网资源哈希验证记录在 `deployed.json`。
- 本地截图：`artifacts/aipic-20261014/aipic-live.jpg`。

## HTTPS 与回滚

新站使用独立 Let's Encrypt 证书，签发时到期日为 2027-01-12。

- 续期配置：`/etc/cron.d/midcode-aipic-certificate`。
- 续期脚本：`/data/code/MIDCODE/renew-aipic-certificate.sh`。
- 已执行 Certbot `renew --dry-run`，模拟续期成功；日志为发布目录下的 `renew-dry-run.log`。
- 新站配置：`/opt/1panel/apps/openresty/openresty/conf/conf.d/image.midcode.cc.conf`。
- 证书复制到 OpenResty 已挂载的 `www/sites/image.midcode.cc/ssl`，私钥权限 `0600`。

恢复主站旧入口：

```sh
python3 /data/code/MIDCODE/releases/aipic-20261014/deploy.py rollback
```

回滚恢复发布前代理配置并校验、重载 OpenResty。不会停止原 API、worker 或删除数据。新域名配置单独保留。

完整操作脚本和生产源码保存在服务器发布目录；传输所用临时 GitHub 草稿 release 已删除。
