FROM node:20-alpine AS ui-builder
ARG ui_mode=production
WORKDIR /src/ui
RUN corepack enable pnpm
COPY ui/package.json ui/pnpm-lock.yaml ./
RUN pnpm install
COPY ui ./
RUN pnpm vite build --mode ${ui_mode}

FROM golang:1.23-alpine AS api-builder
ARG version=unknown
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui-builder /src/ui/dist ./ui/dist
RUN \
  go build -o /bin/goce -ldflags "-s -w -X main.version=${version}" . &&\
  go build -o /bin/godl -ldflags "-s -w" ./tools/godl

FROM alpine:3.20
RUN apk add ca-certificates tzdata curl tar git
RUN \
  addgroup -g 1000 goce && adduser -u 1000 -DG goce goce &&\
  mkdir /opt/data && chown goce: /opt/data
USER goce
RUN mkdir -p ~/sdk ~/go/pkg/mod ~/.cache/go-build
WORKDIR /opt
COPY --from=api-builder --chown=goce:goce /bin/goce /bin/godl ./
EXPOSE 9000
CMD ["/opt/goce"]
