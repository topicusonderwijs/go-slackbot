package slackbot

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/humsie/log"
	"strconv"
	"strings"
	"sync"
	"time"
)

// callbackTTL is how long a callback is kept before GCCallback removes it.
const callbackTTL = time.Hour

type Callback struct {
	Id      uuid.UUID
	Created time.Time
	mu      sync.RWMutex
	Storage map[string]interface{}
}

func (s *Callback) Get(key string) (value interface{}, err error) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	if value, ok := s.Storage[key]; ok == true {
		return value, nil
	}

	return "", fmt.Errorf("key not found")
}

func (s *Callback) GetString(key string) string {

	value, err := s.Get(key)

	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s", value)

}

func (s *Callback) GetInt(key string) int {

	value, err := s.Get(key)

	if err != nil {
		log.Errorln(err)
		return 0
	}

	switch v := value.(type) {
	case int:
		return v
	case string:
		ret, err := strconv.Atoi(v)
		if err != nil {
			log.Errorln(err)
			return 0
		}
		return ret
	default:
		log.Errorf("value for key %q is not an int: %T", key, value)
		return 0
	}

}

func (s *Callback) Set(key string, value interface{}) {

	s.mu.Lock()
	defer s.mu.Unlock()

	s.Storage[key] = value

}

func (s *Callback) AddUUID() uuid.UUID {

	newId := uuid.New()

	log.Debugln("Added ", newId.String(), "for callback with id", s.Id.String())

	CallbackStorage.Store(newId.String(), s)

	return newId

}

// CallbackStorage holds all active callbacks, keyed by their id string. It is a
// sync.Map because it is accessed concurrently from HTTP handlers, the socket
// listener and the GCCallback goroutine.
var CallbackStorage sync.Map

func NewCallback() *Callback {

	sess := Callback{
		Id:      uuid.New(),
		Created: time.Now(),
		Storage: make(map[string]interface{}),
	}

	log.Debugln("Created callback with id", sess.Id.String())

	CallbackStorage.Store(sess.Id.String(), &sess)

	return &sess

}

func FindCallback(id string) (*Callback, error) {

	if strings.HasPrefix(id, "\"") && strings.HasSuffix(id, "\"") {
		id = id[1 : len(id)-1]
	}

	if value, ok := CallbackStorage.Load(id); ok == true {
		return value.(*Callback), nil
	}

	return nil, fmt.Errorf("could not find a callback with id: %s", id)

}

// gcSweep removes every callback that is older than callbackTTL. It performs a
// single pass; GCCallback calls it repeatedly.
func gcSweep() {

	expiredKeys := make([]string, 0)

	CallbackStorage.Range(func(key, value interface{}) bool {
		callback, ok := value.(*Callback)
		if !ok || callback == nil {
			expiredKeys = append(expiredKeys, key.(string))
		} else if time.Since(callback.Created) >= callbackTTL {
			expiredKeys = append(expiredKeys, key.(string))
		}
		return true
	})

	for _, key := range expiredKeys {
		CallbackStorage.Delete(key)
	}

}

// GCCallback periodically removes expired callbacks, sweeping once every sleep
// interval. Run it in its own goroutine: go GCCallback(15 * time.Minute).
func GCCallback(sleep time.Duration) {

	for {
		gcSweep()
		time.Sleep(sleep)
	}

}
