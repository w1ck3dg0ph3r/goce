#!/bin/bash

version=$1
echo "requested version: ${version}"

installed="none"
if command -v golangci-lint &>/dev/null; then
  installed=v$(golangci-lint version | sed 's/.*version \([^ ]\+\).*/\1/')
fi
echo "installed version: ${installed}"

if [ ! "$installed" = "$version" ]; then
  echo "installing ${version}"
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
    | sh -s -- -b $(go env GOPATH)/bin $version
fi
