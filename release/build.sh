#!/bin/sh

cd $(dirname "$0")

version=$(grep 'const Version' ../version.go | cut '-d"' -f2)

build_dir=heraldd

rm -rf "$build_dir"

echo "Current version: $version"

release() {
  goos=$1
  goarch=$2

  target_name="heraldd-$goos-$goarch-$version"
  target_dir="$build_dir/$target_name"
  archive_name="$target_name.tar.gz"

  rm -rf "$target_dir"
  rm -f "$archive_name"

  mkdir -p "$target_dir"

  GOOS=$goos GOARCH=$goarch go build -ldflags '-s -w' -o "$target_dir" ..

  if [ $? -eq 0 ]; then
    echo "Build Herald Daemon $goos-$goarch successfully"
  else
    echo "Build Herald Daemon $goos-$goarch failed"
  fi

  cp ../support/etc/config.yml.example "$target_dir/config.yml"
  if [ $goos = linux ]; then
    cp -r ../support/systemd "$target_dir/systemd"
  fi

  tar -C "$build_dir" -czf "$build_dir/$archive_name" "$target_name"

  if [ $? -eq 0 ]; then
    echo "Pack Herald Daemon $goos-$goarch successfully"
  else
    echo "Pack Herald Daemon $goos-$goarch failed"
  fi
}


release linux amd64
release linux 386
release linux arm64
release linux arm
release darwin amd64
release freebsd amd64
release freebsd 386
release windows amd64
release windows 386
