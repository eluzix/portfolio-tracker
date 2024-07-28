#!/bin/zsh

export OPENSSL_DIR=$(brew --prefix openssl)
ROLE="arn:aws:iam::838643176316:role/tracker-api-role"
cargo lambda build --release --arm64 && cargo lambda deploy --include config.json --profile tracker --binary-name tracker_api --iam-role=$ROLE
