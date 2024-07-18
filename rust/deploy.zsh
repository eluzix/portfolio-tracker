#!/bin/zsh

ROLE="arn:aws:iam::838643176316:role/tracker-api-role"
cargo lambda build --release --arm64 && cargo lambda deploy --include secrets.json --profile tracker --binary-name tracker_api --iam-role=$ROLE
