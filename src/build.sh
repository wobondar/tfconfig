#!/usr/bin/env bash

platforms=("linux/amd64" "darwin/amd64" "linux/386" "linux/arm" "linux/arm64" "darwin/arm64")
mkdir -p ./build

rm -r ./build/*

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name=tfconfig'_'$1'_'$GOOS'_'$GOARCH
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    env GOPATH=$(pwd):$GOPATH GOOS=$GOOS GOARCH=$GOARCH go build -o ./build/$output_name $package
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done

cd build

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    input_name=tfconfig'_'$1'_'$GOOS'_'$GOARCH
    output_name=tfconfig
    mv ./$input_name ./$output_name
    zip ./$input_name.zip ./$output_name
    mv ./$output_name ./$input_name
done

shasum -a 256 *.zip > 'tfconfig_'$1'_SHA256SUMS'
shasum -a 256 -c 'tfconfig_'$1'_SHA256SUMS'
cd ..
