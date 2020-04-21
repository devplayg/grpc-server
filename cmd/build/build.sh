#!/bin/sh

build() {
    local target=$1
    go build -o $target ../$target/main.go
    chmod 755 $target
}

build generator
build classifier
build calculator
build generator
build notifier