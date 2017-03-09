package main

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type (
	Watch struct {
		id string // from key
		Seq int
		Pattern string
		Topic string
	}

	Watches []*Watch

	WatchService struct {
		ctx context.Context
	}
)

const (
	WATCH_KIND = "Watches"
)

func (s *WatchService) All() (Watches, error) {
	q := datastore.NewQuery(WATCH_KIND)
	return s.AllWith(q)
}

func (s *WatchService) AllWith(q *datastore.Query) (Watches, error) {
	iter := q.Run(s.ctx)
	var res = Watches{}
	for {
		obj := Watch{}
		key, err := iter.Next(&obj)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		obj.id = key.Encode()
		res = append(res, &obj)
	}
	return res, nil
}

func (s *WatchService) Find(id string) (*Watch, error) {
	log.Debugf(s.ctx, "WatchService.Find(%v)\n", id)
	key, err := datastore.DecodeKey(id)
	if err != nil {
		log.Errorf(s.ctx, "WatchService.Find(%v) [%T]%v\n", id, err, err)
		return nil, err
	}
	log.Debugf(s.ctx, "WatchService.Find(%v) key: %v\n", id, key)
	obj := Watch{id: id}
	err = datastore.Get(s.ctx, key, &obj)
	switch {
	case err == datastore.ErrNoSuchEntity:
		return nil, err
	case err != nil:
		log.Errorf(s.ctx, "WatchService.Find(%v) [%T]%v\n", id, err, err)
		return nil, err
	}
	return &obj, nil
}
