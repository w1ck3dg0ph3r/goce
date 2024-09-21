# goce

**Go** **C**ompiler **E**xplorer similar to Matt Godbolt's [godbolt.org](https://godbolt.org), but lighter, easier to set up and with Go specific features.

You can check it out at [goce.w1ck3dg0ph3r.dev](https://goce.w1ck3dg0ph3r.dev) or run goce locally.

Additional features include:
- showing inlineability analysis (recursion, function complexity cost, etc.)
- showing inlined function calls
- showing variables that escape to heap

![Screenshot](/images/screenshot.webp)

## Building and running locally

### On the host machine

You will need [go](https://go.dev), [node](https://nodejs.org) and [pnpm](https://pnpm.io/) to build goce.

```bash
make
./goce
```

### Inside the docker container

Build the docker image:

```bash
make image
```

Prepare volume mounts for compilation cache, shared code storage and Go toolchains, module and build caches:

```bash
mkdir -p .cache/{goce,sdk,mod,build} && chown -R 1000:1000 .cache
```

Run the container:

```bash
docker run -d \
    --name goce \
    -p 127.0.0.1:9000:9000 \
    -e GOCE_COMPILERS_SEARCH_GO_PATH=false \
    -e GOCE_COMPILERS_SEARCH_SDK_PATH=true \
    -v $PWD/.cache/goce:/opt/data \
    -v $PWD/.cache/sdk:/home/goce/sdk \
    -v $PWD/.cache/mod:/home/goce/go/pkg/mod \
    -v $PWD/.cache/build:/home/goce/.cache/go-build \
    w1ck3dg0ph3r/goce:latest
```

On the first run there will be no Go toolchains set up. To install 3 latest Go toolchains, run the following and wait around 15s for goce to pick up the downloaded toolchains:

```bash
docker exec -it goce /opt/godl -n 3
```

Alternatively, there is an example [compose.yaml](./compose.yaml) file for `docker compose`.

## Configuration

Goce can be configured via:
- `goce.toml` file located in `/etc/goce` or `~/.config/goce`
- environment variables

See [goce.example.toml](./goce.example.toml), [.env.example](./.env.example) and [config.go](./config.go) for details.

### Notes:

- Right now goce supports the following go compilers:
    - the one found in `$PATH`
    - all versions insalled in `~/sdk/go*` (the default location for [multiple go installations](https://go.dev/doc/manage-install#installing-multiple) on *nix systems)
    - explicitly specified binary

- goce stores compilation cache and shared code snippets in `./data/cache.db` and `./data/shared.db` respectively.
    - the format can vary between versions, so you may have to remove these files after upgrading.
