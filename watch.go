package main

import (
	"fmt"
	"regexp"
	"sort"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type EntityNotFound struct {
	cause error
}

func (e *EntityNotFound) Error() string {
	return e.cause.Error()
}

type ValidationError struct {
	msg string
}

func (e *ValidationError) Error() string {
	return e.msg
}

type Watch struct {
	ID string      `form:"-",datastore:"-"` // from key
	Seq int        `form:"seq"`
	Pattern string `form:"pattern"`
	Topic string   `form:"topic"`
}

var (
	TOPIC_REGEXP = regexp.MustCompile(`\Aprojects/[^/]+/topics/[^/]+\z`)
)

func (w *Watch) Validate() error {
	_, err := regexp.Compile(w.Pattern)
	if err != nil {
		return &ValidationError{fmt.Sprintf("Invalid pattern: %v cause of %v", w.Pattern, err)}
	}
	if !TOPIC_REGEXP.MatchString(w.Topic) {
		return &ValidationError{fmt.Sprintf("Invalid topic: %v", w.Topic)}
	}
	return nil
}


type Watches []*Watch

func (w Watches) Len() int {
	return len(w)
}

func (w Watches) Less(i, j int) bool {
	return w[i].Seq < w[j].Seq
}

func (w Watches) Swap(i, j int) {
    w[i], w[j] = w[j], w[i]
}


type WatchService struct {
	ctx context.Context
}

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
	sort.Sort(res)
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
	err := w.Validate()
	if err != nil {
		return err
	}
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
	err := w.Validate()
	if err != nil {
		return err
	}
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


func (s *WatchService) topicFor(url string) (string, error) {
	watches, err := s.All()
	if err != nil {
		return "", err
	}
	for _, w := range watches {
		log.Debugf(s.ctx, "Pattern: %v, Topic: %v\n", w.Pattern, w.Topic)
		re, err := regexp.Compile(w.Pattern)
		if err != nil {
			log.Errorf(s.ctx, "Invalid Regexp: %v",  w.Pattern)
			return "", err
		}
		if re.MatchString(url) {
			return w.Topic, nil
		}
	}
	return "", nil
}
