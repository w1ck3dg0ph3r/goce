# goce

**Go** **C**ompiler **E**xplorer similar to Matt Godbolt's [godbolt.org](https://godbolt.org), but lighter, easier to set up and with Go specific features.

You can check it out at [goce.w1ck3dg0ph3r.dev](https://goce.w1ck3dg0ph3r.dev) or run goce locally.

Additional features include:
- showing inlineability analysis (recursion, function complexity cost, etc.)
- showing inlined function calls
- showing variables that escape to heap

![Screenshot](/images/screenshot.webp)

## Building and running locally

You will need [go](https://go.dev), [node](https://nodejs.org) and [pnpm](https://pnpm.io/) to build goce.

1. Build ui:
    ```shell
    cd ui
    pnpm install
    pnpm build-only
    ```

2. Build server:
    ```shell
    # ui will be embedded into the binary
    go build .
    ```

### Notes:

- Right now goce makes available the following go versions:
    - default one found in `$PATH`
    - all versions insalled in `~/sdk/go*` (the default location for [multiple go installations](https://go.dev/doc/manage-install#installing-multiple) on *nix systems)

- goce stores compilation cache and shared code snippets in `cache.db` and `shared.db` respectively.
    - the format can vary between versions, so you may have to remove these files.
