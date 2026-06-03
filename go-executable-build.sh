#!/usr/bin/bash

set -e

package="tsetmc"
GO_BIN="${GO_BIN:-/home/h-mousavi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.3.linux-amd64/bin/go}"

if [[ -z "$package" ]]; then
  echo "usage: $0 <package-name>"
  exit 1
fi
if [[ ! -x "$GO_BIN" ]]; then
    if command -v go1.26.3 >/dev/null 2>&1; then
        GO_BIN="$(command -v go1.26.3)"
    elif command -v go >/dev/null 2>&1; then
        GO_BIN="$(command -v go)"
    else
        GO_BIN=""
    fi
fi
if [[ -z "$GO_BIN" ]]; then
    echo "Go 1.26.3 was not found. Set GO_BIN to a valid go binary path."
    exit 1
fi

package_split=(${package//\// })
package_name=${package_split[-1]}

platforms=("windows/amd64" "windows/386" "darwin/amd64" "linux/amd64" "linux/386")
mkdir -p ./build

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    echo $platform_split
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name=$package_name'-'$GOOS'-'$GOARCH
    echo $output_name
    if [ $GOOS = "windows" ]
    then
        output_name+='.exe'
    fi

    env -u GOROOT GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 "$GO_BIN" build -o ./build/$output_name $package
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done
