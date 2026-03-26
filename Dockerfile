# ---- Go backend build ----
FROM golang:1.22-alpine AS go-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 go build -o /bin/server ./cmd/server
RUN CGO_ENABLED=0 go build -o /bin/ingest ./cmd/ingest
RUN CGO_ENABLED=0 go build -o /bin/compute ./cmd/compute

# ---- Next.js frontend build ----
FROM node:20-alpine AS web-build
WORKDIR /app
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ .
ARG NEXT_PUBLIC_API_URL=""
ENV NEXT_PUBLIC_API_URL=${NEXT_PUBLIC_API_URL}
RUN npm run build

# ---- Backend runtime ----
FROM alpine:3.20 AS backend
RUN apk add --no-cache ca-certificates tzdata
COPY --from=go-build /bin/server /bin/server
COPY --from=go-build /bin/ingest /bin/ingest
COPY --from=go-build /bin/compute /bin/compute
EXPOSE 8080
CMD ["/bin/server"]

# ---- Frontend runtime ----
FROM node:20-alpine AS frontend
WORKDIR /app
COPY --from=web-build /app/.next/standalone ./
COPY --from=web-build /app/.next/static ./.next/static
COPY --from=web-build /app/public ./public
EXPOSE 3000
ENV HOSTNAME="0.0.0.0"
CMD ["node", "server.js"]
