# FanAPI 前端重构交接文档（给后端）

本文档用于把当前前端重构成果完整交接给后端同学，目标是：

- 你可以直接在这套新前端上继续开发
- 你可以知道哪些目录是新的、哪些是旧的
- 你可以知道 staging 现在是什么状态
- 你可以知道后续前后端协同时必须遵守的规则

本文档默认面向项目内部开发，不写产品背景，只写接手所需信息。

---

## 1. 当前结论

FanAPI 的新前端已经完成替换主线，后续前端开发应基于：

- [web/app](E:/Mowen开发/FanAPI/web/app)

不要再继续扩展旧前端：

- [web/user](E:/Mowen开发/FanAPI/web/user)

旧前端现在的作用：

- 仅作为历史实现参考
- 仅用于查旧字段、旧接口用法、旧管理页逻辑

不要再往旧 Vue 页面里加新功能。

---

## 2. 当前分支与 PR

当前开发分支：

- `feat/react-frontend-redesign`

当前 PR：

- [PR #3](https://github.com/Fansky123/FanAPI/pull/3)

基线分支：

- `master`

如果你要继续在这套前端上开发，建议直接从这个分支或合并后的 `master` 开始，不要从旧前端分叉。

---

## 3. 新前端技术栈

新前端固定技术栈如下：

- React 19
- Vite
- TypeScript
- Tailwind CSS
- shadcn/ui
- Radix UI
- React Router

关键配置文件：

- [web/app/package.json](E:/Mowen开发/FanAPI/web/app/package.json)
- [web/app/components.json](E:/Mowen开发/FanAPI/web/app/components.json)
- [web/app/vite.config.ts](E:/Mowen开发/FanAPI/web/app/vite.config.ts)
- [web/app/src/app/router.tsx](E:/Mowen开发/FanAPI/web/app/src/app/router.tsx)

---

## 4. 设计规范来源

前端/UI 的最高优先级规范文件是：

- [DESIGN.md](E:/Mowen开发/FanAPI/DESIGN.md)

请把它当成“前端设计合同”。

关键要求：

- 整体风格基于 Cal.com，细节精度参考 Linear
- 使用统一语义色和统一组件模式
- 一个产品，四个角色视图，不允许四套 UI 风格
- 信息密度要够，但不能乱
- 所有核心 UI 必须基于 shadcn/ui 体系扩展

如果后续你要补页面、加字段、做新表单，优先遵守这个文件，而不是模仿旧 Vue 页面。

---

## 5. 目录说明

### 5.1 新前端根目录

- [web/app](E:/Mowen开发/FanAPI/web/app)

### 5.2 主要目录

- [web/app/src/app](E:/Mowen开发/FanAPI/web/app/src/app)
  - 应用入口和路由
- [web/app/src/layouts](E:/Mowen开发/FanAPI/web/app/src/layouts)
  - 四个角色的布局壳
- [web/app/src/pages](E:/Mowen开发/FanAPI/web/app/src/pages)
  - 各页面实现
- [web/app/src/components/shared](E:/Mowen开发/FanAPI/web/app/src/components/shared)
  - 页面级共享组件
- [web/app/src/components/ui](E:/Mowen开发/FanAPI/web/app/src/components/ui)
  - shadcn/ui 基础组件包装
- [web/app/src/lib/api](E:/Mowen开发/FanAPI/web/app/src/lib/api)
  - API 调用封装
- [web/app/src/lib/auth](E:/Mowen开发/FanAPI/web/app/src/lib/auth)
  - token 和模式状态
- [web/app/tests/e2e](E:/Mowen开发/FanAPI/web/app/tests/e2e)
  - Playwright 回归测试

### 5.3 旧前端目录

- [web/user](E:/Mowen开发/FanAPI/web/user)

用途：

- 查旧页面结构
- 查旧接口调用方式
- 查旧后台复杂表单字段

禁止：

- 在旧前端继续新增功能
- 让新功能只存在旧前端

---

## 6. 四个角色的页面归属

当前新前端已经覆盖四个角色：

### User

页面在：

- [web/app/src/pages/user](E:/Mowen开发/FanAPI/web/app/src/pages/user)

已覆盖的主要页面：

- dashboard
- models
- keys
- billing
- logs
- tasks
- profile
- docs
- stats
- exchange
- invite
- playground
- image-gen
- video-gen

### Admin

页面在：

- [web/app/src/pages/admin](E:/Mowen开发/FanAPI/web/app/src/pages/admin)

已覆盖的主要页面：

- dashboard
- channels
- users
- billing
- cards
- key-pools
- tasks
- llm-logs
- settings
- vendors
- withdraw
- ocpc

### Agent

页面在：

- [web/app/src/pages/agent](E:/Mowen开发/FanAPI/web/app/src/pages/agent)

当前已覆盖：

- login
- dashboard

### Vendor

页面在：

- [web/app/src/pages/vendor](E:/Mowen开发/FanAPI/web/app/src/pages/vendor)

当前已覆盖：

- login
- dashboard
- keys

---

## 7. 路由入口

新前端路由总入口：

- [web/app/src/app/router.tsx](E:/Mowen开发/FanAPI/web/app/src/app/router.tsx)

路由规则：

- 用户端走 `/login`、`/dashboard` 等
- 管理端走 `/admin/*`
- 客服端走 `/agent/*`
- 号商端走 `/vendor/*`

鉴权方式：

- 各角色 token 分开存储在 `localStorage`
- 角色之间不要共用 token key

存储逻辑在：

- [web/app/src/lib/auth/storage.ts](E:/Mowen开发/FanAPI/web/app/src/lib/auth/storage.ts)

---

## 8. API 约束

### 8.1 总原则

这次重构阶段没有重做后端协议，新前端默认兼容现有 Go API。

也就是说：

- 前端优先适配后端现有接口
- 后端不要轻易改字段名
- 若必须改接口，先同步前端，不要偷偷改

### 8.2 API 封装位置

- [web/app/src/lib/api/http.ts](E:/Mowen开发/FanAPI/web/app/src/lib/api/http.ts)
- [web/app/src/lib/api/public.ts](E:/Mowen开发/FanAPI/web/app/src/lib/api/public.ts)
- [web/app/src/lib/api/user.ts](E:/Mowen开发/FanAPI/web/app/src/lib/api/user.ts)
- [web/app/src/lib/api/admin.ts](E:/Mowen开发/FanAPI/web/app/src/lib/api/admin.ts)
- [web/app/src/lib/api/vendor.ts](E:/Mowen开发/FanAPI/web/app/src/lib/api/vendor.ts)
- [web/app/src/lib/api/agent.ts](E:/Mowen开发/FanAPI/web/app/src/lib/api/agent.ts)

后续规则：

- 新接口先在这里加类型和方法
- 页面里不要直接散写 `fetch('/api/xxx')`，除非是 `/v1/*` 这类必须直接打的场景

### 8.3 生成类页面的特殊规则

用户端以下页面：

- [UserPlaygroundPage.tsx](E:/Mowen开发/FanAPI/web/app/src/pages/user/UserPlaygroundPage.tsx)
- [UserImageGenPage.tsx](E:/Mowen开发/FanAPI/web/app/src/pages/user/UserImageGenPage.tsx)
- [UserVideoGenPage.tsx](E:/Mowen开发/FanAPI/web/app/src/pages/user/UserVideoGenPage.tsx)

目前会显式带：

- `channel_id`

原因：

- staging 的模型路由状态曾经不稳定
- 显式 `channel_id` 更适合做真实回归和问题定位

后续不要去掉这一点，除非你已经确认模型路由在所有环境都稳定。

---

## 9. 后台渠道管理页说明

这次已经把后台渠道页从“演示版”补到了“可维护真实配置”的版本。

文件：

- [web/app/src/pages/admin/AdminChannelsPage.tsx](E:/Mowen开发/FanAPI/web/app/src/pages/admin/AdminChannelsPage.tsx)

当前支持的关键字段：

- 基础字段：name、model、type、protocol、base_url、method
- 超时字段：timeout_ms、query_timeout_ms
- 认证字段：auth_type、auth_param_name、auth_region、auth_service
- 计费字段：billing_type、billing_config、billing_script
- 脚本字段：request_script、response_script、query_script、error_script
- 轮询字段：query_url、query_method
- 号池字段：key_pool_id
- 负载字段：weight、priority
- 展示字段：icon_url、description

这意味着后端同学后续维护真实上游渠道配置时，已经不需要回旧 Vue 后台。

### 重要约束

如果后端在 `channels` 表里新增了字段，必须同步更新两处：

- [web/app/src/lib/api/admin.ts](E:/Mowen开发/FanAPI/web/app/src/lib/api/admin.ts)
- [web/app/src/pages/admin/AdminChannelsPage.tsx](E:/Mowen开发/FanAPI/web/app/src/pages/admin/AdminChannelsPage.tsx)

否则会出现：

- 前端保存时把新字段丢掉
- 页面编辑一次，把数据库里原有配置覆盖掉

---

## 10. 号池与号商页面说明

### 号池管理

文件：

- [web/app/src/pages/admin/AdminKeyPoolsPage.tsx](E:/Mowen开发/FanAPI/web/app/src/pages/admin/AdminKeyPoolsPage.tsx)

已支持：

- 新建号池
- 选择关联渠道
- 打开号池内 Key 管理
- 添加 Key
- 修改优先级
- 修改启用状态
- 切换号商上传开关
- 删除号池

### 号商端上传 Key

文件：

- [web/app/src/pages/vendor/VendorKeysPage.tsx](E:/Mowen开发/FanAPI/web/app/src/pages/vendor/VendorKeysPage.tsx)

已支持：

- 拉取可上传号池
- 提交 Key
- 成功/失败反馈
- 查看已提交 Key 列表

交接后如果你要继续改这个链路，先查后端：

- [internal/handler/vendor.go](E:/Mowen开发/FanAPI/internal/handler/vendor.go)
- [internal/handler/key_pool.go](E:/Mowen开发/FanAPI/internal/handler/key_pool.go)

---

## 11. OCPC 页面说明

文件：

- [web/app/src/pages/admin/AdminOcpcPage.tsx](E:/Mowen开发/FanAPI/web/app/src/pages/admin/AdminOcpcPage.tsx)

已支持：

- 新建账户
- 编辑账户
- 启用/停用
- 删除账户
- 调度保存
- 立即上报

真实回归已经做过，不是只打开页面。

对应后端：

- [internal/handler/ocpc.go](E:/Mowen开发/FanAPI/internal/handler/ocpc.go)

---

## 12. 用户生成页说明

### Playground

- [web/app/src/pages/user/UserPlaygroundPage.tsx](E:/Mowen开发/FanAPI/web/app/src/pages/user/UserPlaygroundPage.tsx)

### Image

- [web/app/src/pages/user/UserImageGenPage.tsx](E:/Mowen开发/FanAPI/web/app/src/pages/user/UserImageGenPage.tsx)

### Video

- [web/app/src/pages/user/UserVideoGenPage.tsx](E:/Mowen开发/FanAPI/web/app/src/pages/user/UserVideoGenPage.tsx)

已做过的修正：

- 显式带 `channel_id`
- 增加真实错误提示
- 无可用渠道时禁止提交
- 处理新创建 API Key 的明文使用场景

注意：

- 用户历史旧 Key 不一定能直接明文调用
- 新建 Key 后的那一次明文返回最关键

相关文件：

- [web/app/src/pages/user/UserKeysPage.tsx](E:/Mowen开发/FanAPI/web/app/src/pages/user/UserKeysPage.tsx)

---

## 13. Docker / 部署现状

新前端已经接入 Docker 构建，不再是旧 `web/user`。

关键文件：

- [Dockerfile](E:/Mowen开发/FanAPI/Dockerfile)
- [docker-compose.yml](E:/Mowen开发/FanAPI/docker-compose.yml)
- [config.docker.yaml](E:/Mowen开发/FanAPI/config.docker.yaml)
- [docker/nats-server.conf](E:/Mowen开发/FanAPI/docker/nats-server.conf)

现在的部署方式：

- `api` 镜像会构建并打包 `web/app`
- 公共测试站访问的是新 React 前端

### 当前公开测试站

- [https://fanapi-test.chainfish.xyz](https://fanapi-test.chainfish.xyz)

常用入口：

- [用户登录](https://fanapi-test.chainfish.xyz/login)
- [后台登录](https://fanapi-test.chainfish.xyz/admin/login)
- [号商登录](https://fanapi-test.chainfish.xyz/vendor/login)

---

## 14. staging 现状

当前 staging 上保留了几条“可回归”的测试资源，目的是保证后端继续开发时有稳定验证面：

- `Staging Playground Echo`
- `Staging Vendor Echo`
- `Staging Video Echo`
- `Staging Vendor Pool`

这些不是正式业务配置，而是：

- 用于前端真提交流程回归
- 用于避免外部真实 vendor key 不可用时整站不可测

如果你后面把真实上游 key 和真实渠道全部配齐，可以删掉这些 staging-only 资源。

---

## 15. 已修过的关键问题

### 15.1 充值后仍显示余额不足

根因：

- 后台充值写了数据库
- 但运行时扣费走的是 Redis
- Redis 没同步，导致用户端 `/v1/*` 仍然按旧余额扣费

修复位置：

- [internal/service/billing.go](E:/Mowen开发/FanAPI/internal/service/billing.go)

你如果后面再改充值逻辑，必须保证：

- DB 和 Redis 同步更新

### 15.2 生成页模型路由不稳定

修复方式：

- 用户生成页显式带 `channel_id`

不要轻易改回纯 `model` 路由模式，除非你确认环境级路由状态没有问题。

### 15.3 Playwright 本地端口冲突

原因：

- Windows 本机 `4173` 在某些环境里被系统排除

修复位置：

- [web/app/playwright.config.ts](E:/Mowen开发/FanAPI/web/app/playwright.config.ts)

现在默认走可配置端口，不要再写死回 `4173`。

---

## 16. 本地开发命令

前端开发目录：

- [web/app](E:/Mowen开发/FanAPI/web/app)

常用命令：

```bash
pnpm install
pnpm dev
pnpm build
pnpm lint
pnpm test:e2e
```

如果你只改了前端，交付前至少跑：

```bash
pnpm build
pnpm lint
pnpm test:e2e
```

---

## 17. 后端接手开发时的规则

### 必须做

- 所有新前端页面继续写在 `web/app`
- 所有新接口先补 `src/lib/api/*`
- 改渠道/号池/OCPC/用户生成链路时，优先在 staging 上做真提交验证
- 改数据库字段时同步更新前端类型

### 不要做

- 不要回旧 Vue 页面继续加功能
- 不要在页面里散写未经封装的接口调用
- 不要跳过 [DESIGN.md](E:/Mowen开发/FanAPI/DESIGN.md) 自己另起一套风格
- 不要把后台复杂配置又做回旧前端

---

## 18. 建议的后续开发流程

如果你要继续开发新功能，建议顺序是：

1. 先确认后端接口是否已有
2. 若没有，先定接口字段
3. 在 `src/lib/api/*` 补类型和方法
4. 在 `src/pages/*` 补页面
5. 若涉及通用 UI，再补到 `src/components/shared` 或 `src/components/ui`
6. 本地跑 `build + lint + e2e`
7. 部署到 staging 做真流程验证

---

## 19. 你接手时最常用的文件

如果你只想快速定位，优先看这些：

- [DESIGN.md](E:/Mowen开发/FanAPI/DESIGN.md)
- [web/app/src/app/router.tsx](E:/Mowen开发/FanAPI/web/app/src/app/router.tsx)
- [web/app/src/lib/api/admin.ts](E:/Mowen开发/FanAPI/web/app/src/lib/api/admin.ts)
- [web/app/src/lib/api/user.ts](E:/Mowen开发/FanAPI/web/app/src/lib/api/user.ts)
- [web/app/src/lib/api/vendor.ts](E:/Mowen开发/FanAPI/web/app/src/lib/api/vendor.ts)
- [web/app/src/pages/admin/AdminChannelsPage.tsx](E:/Mowen开发/FanAPI/web/app/src/pages/admin/AdminChannelsPage.tsx)
- [web/app/src/pages/admin/AdminKeyPoolsPage.tsx](E:/Mowen开发/FanAPI/web/app/src/pages/admin/AdminKeyPoolsPage.tsx)
- [web/app/src/pages/admin/AdminOcpcPage.tsx](E:/Mowen开发/FanAPI/web/app/src/pages/admin/AdminOcpcPage.tsx)
- [web/app/src/pages/user/UserPlaygroundPage.tsx](E:/Mowen开发/FanAPI/web/app/src/pages/user/UserPlaygroundPage.tsx)
- [web/app/src/pages/user/UserImageGenPage.tsx](E:/Mowen开发/FanAPI/web/app/src/pages/user/UserImageGenPage.tsx)
- [web/app/src/pages/user/UserVideoGenPage.tsx](E:/Mowen开发/FanAPI/web/app/src/pages/user/UserVideoGenPage.tsx)
- [internal/service/billing.go](E:/Mowen开发/FanAPI/internal/service/billing.go)
- [docker-compose.yml](E:/Mowen开发/FanAPI/docker-compose.yml)
- [config.docker.yaml](E:/Mowen开发/FanAPI/config.docker.yaml)

---

## 20. 最后一句

从现在开始，FanAPI 的前端主线是 React 这套，不是旧 Vue。

如果你继续开发，请默认：

- 新功能写在新前端
- 老页面只查、不扩展
- staging 要做真提交验证
- 设计规范以 [DESIGN.md](E:/Mowen开发/FanAPI/DESIGN.md) 为准

