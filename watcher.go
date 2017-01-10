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
	foundFiles := w.findFiles(ctx)

	diffs := w.calcDifferences(ctx, storedFiles, foundFiles)

	for _, url := range diffs.created {
		w.storeUploadedFiles(ctx, url, foundFiles[url], w.notifier.Created)
	}
	for _, url := range diffs.updated {
		w.storeUploadedFiles(ctx, url, foundFiles[url], w.notifier.Updated)
	}
	for _, url := range diffs.deleted {
		w.removeUploadedFiles(ctx, url, w.notifier.Deleted)
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

func (w *Watcher) findFiles(ctx context.Context) map[string]time.Time {
	result := make(map[string]time.Time)
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
		log.Debugf(ctx, "Object found %v\n", o)
		url := "gs://" + o.Bucket + "/" + o.Name
		result[url] = o.Updated
	}
	return result
}

type differences struct {
	created []string
	updated []string
	deleted []string
}

func (w *Watcher) calcDifferences(ctx context.Context, stored map[string]time.Time, found map[string]time.Time) *differences {
	created := make([]string, 0, 1)
	updated := make([]string, 0, 1)
	deleted := make([]string, 0, 1)
	for url, foundUpdated := range found {
		if storedUpdated, ok := stored[url]; ok {
			if foundUpdated.After(storedUpdated) {
				log.Debugf(ctx, "Updated %v %v => %v\n", url, storedUpdated, foundUpdated)
				updated = append(updated, url)
			}
			delete(stored, url)
		} else {
			log.Debugf(ctx, "Inserted %v\n", url)
			created = append(created, url)
		}
	}
	for url, _ := range stored {
		log.Debugf(ctx, "Deleted %v\n", url)
		deleted = append(deleted, url)
	}
	return &differences {
		created: created,
		updated: updated,
		deleted: deleted,
	}
}

func (w *Watcher) storeUploadedFiles(ctx context.Context, url string, updated time.Time, callback func(ctx context.Context, url string)) {
	k := datastore.NewKey(ctx, "UploadedFiles", url, 0, w.watchKey)
	uf := &UploadedFile{Url: url, Updated: updated}
	if _, err := datastore.Put(ctx, k, uf); err != nil {
		log.Debugf(ctx, "Failed to put %v\n", uf)
	} else {
		callback(ctx, url)
	}
}

func (w *Watcher) removeUploadedFiles(ctx context.Context, url string, callback func(ctx context.Context, url string)) {
	k := datastore.NewKey(ctx, "UploadedFiles", url, 0, w.watchKey)
	if err := datastore.Delete(ctx, k); err != nil {
		log.Debugf(ctx, "Failed to delete: %v \n", url)
	} else {
		callback(ctx, url)
	}
}
