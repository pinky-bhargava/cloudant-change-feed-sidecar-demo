#!/usr/bin/env bash

# exit safely
set -ex

echo "inside side-car setup script"
# setup private dependencies

# dependencies
yum update -y --setopt=tsflags=nodocs --nobest
yum install -y git gcc
yum clean all -y

git config --global user.name "Pinky Bhargava"
git config --global user.email "pinky.bhargava@in.ibm.com"
#git config --global url."git@github.ibm.com:".insteadOf "https://github.ibm.com"

# install go 1.17 via golang.org
curl https://golang.org/dl/go1.17.3.linux-amd64.tar.gz -S -L -O

tar -C /usr/local -xzf go1.17.3.linux-amd64.tar.gz && rm -f go1.17.3.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

#check go version
go version

# go mod
cd $GOPATH/src/github.com/cloudant-change-feed-sidecar-demo
export GOPRIVATE=github.com/*
export GO111MODULE=on

go mod tidy

# compile
go install $GOPATH/src/github.com/cloudant-change-feed-sidecar-demo/cmd/sidecar



# run golangci-lint with gosec
cd $GOPATH/src/github.com/cloudant-change-feed-sidecar-demo

# install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.42.1

#$GOPATH/bin/golangci-lint run --enable gosec
#/usr/bin/golangci-lint run -v

# convenience
echo "set -o vi" >> $HOME/.bashrc


# add commit log for convenience
git log -4 > /usr/share/doc/git_history.log
git --no-pager log -4 --pretty=format:'{%n  "commit": "%H",%n  "abbreviated_commit": "%h",%n  "tree": "%T",%n  "abbreviated_tree": "%t",%n  "parent": "%P",%n  "abbreviated_parent": "%p",%n  "refs": "%D",%n  "encoding": "%e",%n
"subject": "%s",%n  "sanitized_subject_line": "%f",%n  "body": "%b",%n  "commit_notes": "%N",%n  "verification_flag": "%G?",%n  "signer": "%GS",%n  "signer_key": "%GK",%n  "author": {%n    "name": "%aN",%n    "email": "%aE",%n
"date": "%aD"%n  },%n  "commiter": {%n    "name": "%cN",%n    "email": "%cE",%n    "date": "%cD"%n  }%n},' > /usr/share/doc/git_history.json


