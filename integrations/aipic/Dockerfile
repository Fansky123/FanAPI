FROM node:24-bookworm-slim AS build
WORKDIR /build
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl && rm -rf /var/lib/apt/lists/*
ARG AIPIC_REV=f6cac8328d7ddb66c610a4d3fc89423d73976789
RUN curl -fSL --retry 3 "https://codeload.github.com/simplaj/AIPic/tar.gz/${AIPIC_REV}" | tar -xz --strip-components=1
COPY bootstrap.ts prepare.mjs prepare-optimizer.mjs PromptOptimizer.tsx prompt-optimizer.css prepare-history.mjs HistoryDelete.tsx history-delete.css /integration/
RUN node /integration/prepare.mjs /build && npm ci
ENV VITE_TASK_BACKEND="" VITE_API_PROXY_AVAILABLE=false VITE_API_PROXY_LOCKED=false
RUN npm run build

FROM node:24-bookworm-slim
WORKDIR /app
COPY --from=build /build/dist ./dist
COPY --from=build /build/LICENSE ./LICENSE-AIPic
COPY server.mjs prompt-optimizer.mjs ./
ENV HOST=0.0.0.0 PORT=8788
USER node
EXPOSE 8788
CMD ["node", "server.mjs"]
