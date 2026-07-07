FROM oven/bun:1.3.5 AS deps
WORKDIR /app
COPY package.json bun.lock* ./
COPY apps/web/package.json ./apps/web/package.json
COPY packages/api-client/package.json ./packages/api-client/package.json
RUN bun install --frozen-lockfile

FROM deps AS build
COPY apps/web ./apps/web
COPY packages/api-client ./packages/api-client
COPY api ./api
RUN bun --filter @cineforge/web build

FROM oven/bun:1.3.5-slim
WORKDIR /app
ENV NODE_ENV=production
COPY --from=build /app ./
EXPOSE 3000
CMD ["bun", "--filter", "@cineforge/web", "start"]
