#!/bin/bash

pkgs=$(go list ./... 2> /dev/null | grep -v /src/)
deps=`echo ${pkgs} | tr ' ' ","`
echo "mode: atomic" > cover.out

for pkg in $pkgs; do
    set -e
    go test -cover -coverpkg "$deps" -coverprofile=profile.tmp $pkg
    set +e

    if [ -f profile.tmp ]; then
        tail -n +2 profile.tmp >> cover.out
        rm profile.tmp
    fi
done;

go tool cover -html=cover.out -o cover.html
