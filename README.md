# Boreas [![GoDoc](https://godoc.org/github.com/baltimore-sun-data/boreas?status.svg)](https://godoc.org/github.com/baltimore-sun-data/boreas) [![Go Report Card](https://goreportcard.com/badge/github.com/baltimore-sun-data/boreas)](https://goreportcard.com/report/github.com/baltimore-sun-data/boreas) [![Build Status](https://travis-ci.org/baltimore-sun-data/boreas.svg?branch=master)](https://travis-ci.org/baltimore-sun-data/boreas)

Boreas is a CloudFront invalidator.  It is named after [the Greek god of the North Wind](https://en.wikipedia.org/wiki/Anemoi#Boreas), who blows all the clouds away.

## Installation

First install [Go](http://golang.org).

If you just want to install the binary to your current directory and don't care about the source code, run

```bash
GOBIN="$(pwd)" GOPATH="$(mktemp -d)" go get github.com/baltimore-sun-data/boreas
```

## Screenshots

```bash
$ boreas -h
Usage of boreas:

    boreas [options] <invalidation path>...

Invalidation path defaults to '/*'.

AWS credentials taken from ~/.aws/ or from "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", and other AWS 
configuration environment variables.


Options:

  -dist string
        CloudFront distribution ID
  -ref string
        CloudFront 'CallerReference', a unique identifier for this invalidation request. (default: Unix timestamp)
  -wait duration
        Time out to wait for invalidation to complete. Set to 0 to exit without waiting. (default 10m0s)

$ boreas -dist EABC123EFG4567
2018/02/15 14:35:43 Invalidation ID: "IQ2JXQ53AYXGBB"
Invalidating...................
```
