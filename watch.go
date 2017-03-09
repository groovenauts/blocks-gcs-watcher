package main

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type (
	EntityNotFound struct {
		cause error
	}
)
func (e *EntityNotFound) Error() string {
	return e.cause.Error()
}

type (
	Watch struct {
		ID string      `form:"-",datastore:"-"` // from key
		Seq int        `form:"seq"`
		Pattern string `form:"pattern"`
		Topic string   `form:"topic"`
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
	log.Debugf(s.ctx, "AllWith #0\n")
	iter := q.Run(s.ctx)
	log.Debugf(s.ctx, "AllWith #1\n")
	var res = Watches{}
	for {
		obj := Watch{}
		key, err := iter.Next(&obj)
		if err == datastore.Done {
			break
		}
		if err != nil {
			log.Errorf(s.ctx, "AllWith err: %v\n", err)
			return nil, err
		}
		obj.ID = key.Encode()
		res = append(res, &obj)
	}
	log.Debugf(s.ctx, "AllWith => %v\n", res)
	for i, w := range res {
		log.Debugf(s.ctx, "AllWith %v: %v, %v, %v\n", i, w.Seq, w.Pattern, w.Topic)
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
	obj := Watch{}
	err = datastore.Get(s.ctx, key, &obj)
	switch {
	case err == datastore.ErrNoSuchEntity:
		return nil, &EntityNotFound{err}
	case err != nil:
		log.Errorf(s.ctx, "WatchService.Find(%v) [%T]%v\n", id, err, err)
		return nil, err
	}
	obj.ID = id
	return &obj, nil
}

func (s *WatchService) Create(w *Watch) error {
	key := datastore.NewIncompleteKey(s.ctx, WATCH_KIND, nil)
	res, err := datastore.Put(s.ctx, key, w)
	if err != nil {
		log.Errorf(s.ctx, "WatchService.Create(%v) [%T]%v\n", w, err, err)
		return err
	}
	w.ID = res.Encode()
	return nil
}

func (s *WatchService) Update(w *Watch) error {
	key, err := datastore.DecodeKey(w.ID)
	if err != nil {
		return err
	}
	_, err = datastore.Put(s.ctx, key, w)
	if err != nil {
		log.Errorf(s.ctx, "WatchService.Update(%v) [%T]%v\n", w, err, err)
		return err
	}
	return nil
}

func (s *WatchService) Delete(id string) error {
	key, err := datastore.DecodeKey(id)
	if err != nil {
		return err
	}
	err = datastore.Delete(s.ctx, key)
	if err != nil {
		return err
	}
	return nil
}
