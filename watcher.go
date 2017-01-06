package main

import (
	"golang.org/x/net/context"
	"log"
	"time"

	"cloud.google.com/go/storage"
	"cloud.google.com/go/datastore"

	"google.golang.org/api/iterator"
)

type Watcher struct {
	config *Watch
	datastoreClient *datastore.Client
	storageClient *storage.Client
	watchKey *datastore.Key
	notifier Notifier
}

type UploadedFile struct {
	Url string `datastore:"url"`
  Updated time.Time  `datastore:"updated"`
}

func (w *Watcher) process(ctx context.Context) {
	w.setup(ctx)

	storedFiles := w.loadStoredFiles(ctx)

	// Prepares a new bucket
	bucket := w.storageClient.Bucket(w.config.BucketName)
	it := bucket.Objects(ctx, nil)
	for {
		o, err := it.Next()
		if err != nil && err != iterator.Done {
			log.Fatal(err)
		}
		if err == iterator.Done {
			break
		}
		url := "gs://" + o.Bucket + "/" + o.Name
		if updated, ok := storedFiles[url]; ok {
			if o.Updated.After(updated) {
				log.Println(url, " was updated at", updated, " but now it's ", o.Updated)
				w.storeUploadedFiles(ctx, url, o, func(uf *UploadedFile) {
					w.notifier.Updated(ctx, uf)
				})
			}
			delete(storedFiles, url)
		} else {
			log.Println(url, "was inserted")
			w.storeUploadedFiles(ctx, url, o, func(uf *UploadedFile) {
				w.notifier.Created(ctx, uf)
			})
		}
	}
	for url, updated := range storedFiles {
		log.Println(url, "was deleted")
		k := datastore.NameKey("UploadedFiles", url, w.watchKey)
		if err := w.datastoreClient.Delete(ctx, k); err != nil {
			log.Println("Failed to delete ", url)
		} else {
			uf := &UploadedFile{Url: url, Updated: updated}
			w.notifier.Deleted(ctx, uf)
		}
	}
}

func (w *Watcher) setup(ctx context.Context) {
	// Creates a storageClient
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create storageClient: %v", err)
	}
	w.storageClient = storageClient

	// Creates a datastoreClient
	datastoreClient, err := datastore.NewClient(ctx, w.config.ProjectID)
	if err != nil {
		log.Fatalf("Failed to create datastoreClient for %s: %v", w.config.ProjectID, err)
	}
	w.datastoreClient = datastoreClient

	w.notifier = NewGCSProxyNotifier(ctx, w.config)
}

func (w *Watcher) loadStoredFiles(ctx context.Context) map[string]time.Time {
	q := datastore.NewQuery("UploadedFiles").Ancestor(w.watchKey)
	var res []UploadedFile
	_, err := w.datastoreClient.GetAll(ctx, q, &res)
	if err != nil {
		log.Fatalf("Failed to get all uploaded files: %v", err)
	}

	storedFiles := make(map[string]time.Time)
	for _, v := range res {
		storedFiles[v.Url] = v.Updated
	}
	return storedFiles
}

func (w *Watcher) storeUploadedFiles(ctx context.Context, url string, o *storage.ObjectAttrs, callback func(*UploadedFile)) {
	k := datastore.NameKey("UploadedFiles", url, w.watchKey)
	uf := &UploadedFile{Url: url, Updated: o.Updated}
	if _, err := w.datastoreClient.Put(ctx, k, uf); err != nil {
		log.Println("Failed to put ", uf)
	} else {
		callback(uf)
	}
}
