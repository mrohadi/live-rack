# syntax=docker/dockerfile:1.7

# ── Build stage ──────────────────────────────────────────────────────────────
# Node 22 + pinned pnpm. Node 20 + pnpm 11.x triggered:
#   ERR_UNKNOWN_BUILTIN_MODULE inside corepack's pnpm
FROM node:22-alpine AS build
WORKDIR /src

# Pin pnpm via npm (skip corepack — known signature/builtin issues with 11.x)
RUN npm install -g pnpm@11.5.0

# Workspace files (no root package.json — repo uses workspace root only)
COPY pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/web/package.json ./apps/web/

# Install with frozen lockfile (workspace-aware)
RUN pnpm install --frozen-lockfile

# Copy source + build
COPY apps/web ./apps/web

# Build-time public vars (must be passed via --build-arg)
ARG VITE_OIDC_ISSUER
ARG VITE_OIDC_CLIENT_ID
ARG VITE_OIDC_REDIRECT_URI
# Code reads import.meta.env.VITE_API_URL (lib/api.ts, useLogin.ts, ws.ts, …).
# Name must match exactly or the bundle falls back to http://localhost:8080.
ARG VITE_API_URL
ENV VITE_OIDC_ISSUER=${VITE_OIDC_ISSUER} \
    VITE_OIDC_CLIENT_ID=${VITE_OIDC_CLIENT_ID} \
    VITE_OIDC_REDIRECT_URI=${VITE_OIDC_REDIRECT_URI} \
    VITE_API_URL=${VITE_API_URL}

RUN pnpm -F web build

# ── Runtime stage (nginx — light, ~25MB) ─────────────────────────────────────
FROM nginx:1.27-alpine
COPY --from=build /src/apps/web/dist /usr/share/nginx/html
COPY apps/web/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://127.0.0.1/ >/dev/null || exit 1
