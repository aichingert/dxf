#!/bin/sh

GOARCH=wasm GOOS=js go build -o main.wasm
