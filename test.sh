#!/bin/bash

expected="github.com/mkusaka/modi"
result="$(go run main.go)"
if [ "$result" != "$expected" ]; then
  echo "unexpected string output, expected: $expected but got: $result"
  exit 1
fi
echo "success"
