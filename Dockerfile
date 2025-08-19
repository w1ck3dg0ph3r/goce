FROM docker.io/node:22-alpine AS ui-builder
ARG ui_mode=production
WORKDIR /src/ui
RUN npm install -g corepack@latest &&\
    corepack enable pnpm
COPY ui/package.json ui/pnpm-lock.yaml ./
RUN --mount=type=cache,id=pnpm,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile
COPY ui ./
RUN pnpm build-only --mode ${ui_mode}

FROM docker.io/golang:1.24-alpine AS api-builder
ARG version=unknown
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,id=golang,target=/go/pkg/mod/ \
    go mod download
COPY . .
COPY --from=ui-builder /src/ui/dist ./ui/dist
RUN --mount=type=cache,id=golang,target=/go/pkg/mod/ \
    --mount=type=cache,id=golang,target=/root/.cache/go-build \
    go build -o /bin/goce -ldflags "-s -w -X main.version=${version}" ./cmd/goce &&\
    go build -o /bin/godl -ldflags "-s -w" ./cmd/tools/godl

FROM docker.io/alpine:3.22
RUN apk add ca-certificates tzdata curl tar git
RUN addgroup -g 1000 goce && adduser -u 1000 -DG goce goce &&\
    mkdir /opt/data && chown goce: /opt/data
USER goce
RUN mkdir -p ~/sdk ~/go/pkg/mod ~/.cache/go-build
WORKDIR /opt
COPY --from=api-builder --chown=goce:goce /bin/goce /bin/godl ./
EXPOSE 9000
CMD ["/opt/goce"]
