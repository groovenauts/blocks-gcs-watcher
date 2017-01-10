package main

import (
	"golang.org/x/net/context"
	"time"

	"cloud.google.com/go/storage"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/api/iterator"
)

type Watcher struct {
	config *Watch
	storageClient *storage.Client
	watchKey *datastore.Key
	notifier *GCSProxyNotifier
}

type UploadedFile struct {
	Url string `datastore:"url"`
  Updated time.Time  `datastore:"updated"`
}

func (w *Watcher) process(ctx context.Context) {
	log.Debugf(ctx, "Start processing: %v %v\n", w.watchKey, w.config)
	w.setup(ctx)

	storedFiles := w.loadStoredFiles(ctx)

	// Prepares a new bucket
	bucket := w.storageClient.Bucket(w.config.BucketName)
	it := bucket.Objects(ctx, nil)
	for {
		o, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Errorf(ctx, "Failed to get Object from storage cause of %v\n", err)
		}
		log.Debugf(ctx, "Processing Object %v\n", o)
		url := "gs://" + o.Bucket + "/" + o.Name
		if updated, ok := storedFiles[url]; ok {
			if o.Updated.After(updated) {
				log.Debugf(ctx, "%v was updated at %v but now it's %v", url, updated, o.Updated)
				w.storeUploadedFiles(ctx, url, o, func(uf *UploadedFile) {
					w.notifier.Updated(ctx, uf.Url)
				})
			}
			delete(storedFiles, url)
		} else {
			log.Debugf(ctx, "%v was inserted\n", url)
			w.storeUploadedFiles(ctx, url, o, func(uf *UploadedFile) {
				w.notifier.Created(ctx, uf.Url)
			})
		}
		log.Debugf(ctx, "Process completed Object %v\n", o)
	}
	for url, updated := range storedFiles {
		log.Debugf(ctx, "%v was deleted\n", url)
		k := datastore.NewKey(ctx, "UploadedFiles", url, 0, w.watchKey)
		if err := datastore.Delete(ctx, k); err != nil {
			log.Debugf(ctx, "Failed to delete: %v \n", url)
		} else {
			uf := &UploadedFile{Url: url, Updated: updated}
			w.notifier.Deleted(ctx, uf.Url)
		}
	}
}

func (w *Watcher) setup(ctx context.Context) {
	// Creates a storageClient
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to create storageClient: %v\n", err)
	}
	w.storageClient = storageClient

	n := NewGCSProxyNotifier(ctx, w.config)
	w.notifier = &n
}

func (w *Watcher) loadStoredFiles(ctx context.Context) map[string]time.Time {
	q := datastore.NewQuery("UploadedFiles").Ancestor(w.watchKey)
	var res []UploadedFile
	_, err := q.GetAll(ctx, &res)
	if err != nil {
		log.Errorf(ctx, "Failed to get all uploaded files: %v\n", err)
	}

	storedFiles := make(map[string]time.Time)
	for _, v := range res {
		log.Debugf(ctx, "Stored File: %v\n", v.Url)
		storedFiles[v.Url] = v.Updated
	}
	return storedFiles
}

func (w *Watcher) storeUploadedFiles(ctx context.Context, url string, o *storage.ObjectAttrs, callback func(*UploadedFile)) {
	k := datastore.NewKey(ctx, "UploadedFiles", url, 0, w.watchKey)
	uf := &UploadedFile{Url: url, Updated: o.Updated}
	if _, err := datastore.Put(ctx, k, uf); err != nil {
		log.Debugf(ctx, "Failed to put %v\n", uf)
	} else {
		callback(uf)
	}
}
