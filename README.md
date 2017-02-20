# blocks-gcs-watcher

[![Build Status](https://secure.travis-ci.org/groovenauts/blocks-gcs-watcher.png)](https://travis-ci.org/groovenauts/blocks-gcs-watcher)

## Overview

`blocks-gcs-watcher` watches your bucket. If files are uploaded,
modified or removed, `blocks-gcs-watcher` detects them and notify
them by publishing message to pubsub topic.

## Setup

1. Install Go
  - https://golang.org/doc/install
  - Or use [goenv](https://github.com/kaneshin/goenv)
    - You can install goenv by [anyenv](https://github.com/riywo/anyenv)
1. [Install the App Engine SDK for Go](https://cloud.google.com/appengine/docs/go/download?hl=ja)
1. `git clone git@github.com:groovenauts/blocks-gcs-watcher.git $GOPATH/src/github.com/groovenauts/blocks-gcs-watcher`
1. [Install glide](https://github.com/Masterminds/glide#install)
1. `glide install`

## Run test

```
goapp test
```

### With coverage

```
goapp test -coverprofile coverage.out
go tool cover -html=coverage.out
```

## Run server locally

```
$ dev_appserver.py \
  --env_var TOPIC=<YOUR_PUBSUB_TOPIC> \
  ./app.yaml
```


## Deploy

```
$ appcfg.py \
  -A <YOUR_GCP_PROJECT> \
  -E GOOGLE_SITE_VERIFICATION:<YOUR_GOOGLE_SITE_VERIFICATION> \
  -E PUBSUB_TOPIC:<YOUR_PUBSUB_TOPIC> \
  -V $(cat VERSION) \
  update .
```

If you want to set it active, run the following command

```
$ gcloud app services set-traffic gcs-watcher --splits=$(cat VERSION)=1
```
