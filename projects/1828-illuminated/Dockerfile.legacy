# =====================================================================
# 1828 Illuminated — multi-stage build (Vue SPA → nginx-served static site)
#
# Build:   docker build -t 1828-illuminated .
# Run:     docker run -p 8080:80 1828-illuminated
# Deploy:  push to a registry / let Dokploy auto-build on 1828.ibeco.me
#
# No backend. Modern definitions are pre-fetched into the bundle at build time
# by scripts/fetch_modern_defs.py (run separately, not at docker build time
# because the Free Dictionary API is rate-limited and the data is committed).
# =====================================================================

# ---------------------------------------------------------------------
# Stage 1 — Node build
# ---------------------------------------------------------------------
FROM node:22-alpine AS build

WORKDIR /app

# Install deps first for layer cache
COPY frontend/package.json ./frontend/
COPY frontend/package-lock.json* ./frontend/
WORKDIR /app/frontend
RUN npm install --no-audit --no-fund

# Copy the rest of the frontend
WORKDIR /app
COPY frontend/ ./frontend/

# Build the static bundle
WORKDIR /app/frontend
RUN npm run build

# ---------------------------------------------------------------------
# Stage 2 — nginx serving
# ---------------------------------------------------------------------
FROM nginx:1.27-alpine

# SPA-friendly nginx config (history-mode router → fallback to index.html)
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Static bundle
COPY --from=build /app/frontend/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://localhost/ >/dev/null || exit 1
