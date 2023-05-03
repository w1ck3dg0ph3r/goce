# goce

**Go** **C**ompiler **E**xplorer similar to Matt Godbolt's [godbolt.org](https://godbolt.org), but lighter, easier to set up and with Go specific features.

Additional features include:
- showing inlining decisions (due to recursion, function complexity, etc.)
- showing variables that escape to heap

![Screenshot](/images/screenshot.jpg)

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