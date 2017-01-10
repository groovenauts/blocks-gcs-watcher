# blocks-gcs-watcher

## Overview

`blocks-gcs-watcher` watches your bucket. If files are uploaded,
modified or removed, `blocks-gcs-watcher` detects them and notify
them by publishing message to pubsub topic.

## Deploy

```
appcfg.py \
  -A <YOUR_GCP_PROJECT> \
  -E WATCH_ID:<STRING ID> \
  -E BUCKET:<YOUR_GCS_BUCKET> \
  -E PROJECT:<YOUR_PUBSUB_PROJECT> \
  -E TOPIC:<YOUR_PUBSUB_TOPIC> \
  -V $(cat VERSION) \
  update .
```
