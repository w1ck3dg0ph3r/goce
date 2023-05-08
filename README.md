# goce

**Go** **C**ompiler **E**xplorer similar to Matt Godbolt's [godbolt.org](https://godbolt.org), but lighter, easier to set up and with Go specific features.

Additional features include:
- showing inlineability analysis (recursion, function complexity cost, etc.)
- showing inlined function calls
- showing variables that escape to heap

![Screenshot](/images/screenshot.webp)

## Building Local Version

1. Build ui:
    ```shell
    cd ui
    pnpm install
    pnpm build-only
    ```

2. Build server:
    ```shell
    # ui will be embedded into server binary
    go build .
    ```