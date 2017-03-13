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
  --env_var PUBSUB_TOPIC=<YOUR_PUBSUB_TOPIC> \
  ./app.yaml
```

1. open http://localhost:8080/_ah/login
2. Check `Sign in as Administrator`
3. Click `Login`
4. Open http://localhost:8080/admin/watches
5. Add watch settings


## Production Envirionment

### Setup Pubsub

1. Create your topic
2. Create your subscription for the topic

You can use console or `gcloud` command.

#### example

```
$ export PUBSUB_TOPIC=test-topic1
$ export PUBSUB_SUBSCRIPTION=test-subscription1
$ gcloud beta pubsub topics create $PUBSUB_TOPIC
$ gcloud beta pubsub subscriptions create $PUBSUB_SUBSCRIPTION --topic=$PUBSUB_TOPIC
```

### Verify your blocks-gcs-watcher

Follow [the official instruction](https://cloud.google.com/storage/docs/object-change-notification)

1. Create your service account
2. Activate the service account
3. [Search Console](https://www.google.com/webmasters/tools/)
    - It must be `https://gcs-watcher-dot-<YOUR GCP Project ID>.appspot.com`
    - Pick `YOUR_GOOGLE_SITE_VERIFICATION` value of meta tag named `google-site-verification`
        - `<meta name="google-site-verification" content="AAAmjn9inYA1SBBB-LPtDT4wvDuPGeGdxF3hECrmZZZ" />`
4. Deploy your `blocks-gcs-watcher`
5. Start watching your bucket
    - `gsutil notification watchbucket <Your blocs-gcs-watcher URL> gs://<Your bucket name>`

See also [Object Change NotificationをApp Engineで受け取る設定](http://qiita.com/sinmetal/items/0438203034a0cb448448)

### Deploy

```
$ appcfg.py \
  -A <YOUR_GCP_PROJECT> \
  -E GOOGLE_SITE_VERIFICATION:<YOUR_GOOGLE_SITE_VERIFICATION> \
  -V $(cat VERSION) \
  update .
```

If you want to set it active soon, run the following command

```
$ gcloud app services set-traffic gcs-watcher --splits=$(cat VERSION)=1
```

1. open https://<YOUR_HOST>/admin/watches
2. Add watch settings



### Test

Subscribe messages by using [pubsub-devsub](https://github.com/akm/pubsub-devsub).

```
$ pubsub-devsub --project YOUR_GCP_PROJECT --subscription $PUBSUB_SUBSCRIPTION
```

After you upload some files to the bucket, you can see the messages like this:

```
2017-02-20 05:48:55.645 +0000 UTC 55501196589600: map[download_files:gs://test-bucket1/dir1/file-20170220-1448.yml]
2017-02-20 05:49:17.827 +0000 UTC 55504914109200: map[download_files:gs://test-bucket1/dir1/file-20170220-1456.yml]
```
