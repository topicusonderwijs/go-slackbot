package slackbot

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/humsie/log"
	"strconv"
	"strings"
	"time"
)

type Callback struct {

	Id      uuid.UUID
	Created time.Time
	Storage map[string]interface{}

}

func (s *Callback) Get(key string) (value interface{}, err error) {

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

	ret, err := strconv.Atoi(fmt.Sprintf("%d", value))
	if err != nil {
		log.Errorln(err)
		return 0
	}

	return ret

}


func (s *Callback) Set(key string, value interface{}) {

	s.Storage[key] = value

}

func (s *Callback) AddUUID() uuid.UUID {

	newId := uuid.New()

 	log.Debugln("Added ", newId.String(), "for callback with id", s.Id.String())

   	CallbackStorage[newId.String()] = s

	return newId

}

var CallbackStorage map[string]*Callback

func NewCallback() *Callback {

	if CallbackStorage == nil {
		log.Debugln("Creating Storage")
		CallbackStorage = make(map[string]*Callback, 0)
	}

	sess := Callback{
		Id:      uuid.New(),
		Created: time.Now(),
		Storage: make(map[string]interface{},0),
	}

	log.Debugln("Created callback with id", sess.Id.String())

	CallbackStorage[sess.Id.String()] = &sess

	return &sess

}



func FindCallback(id string) (*Callback, error) {

	if strings.HasPrefix(id, "\"") && strings.HasSuffix(id, "\""){
		id = id[1:len(id)-1]
	}

	if callback, ok := CallbackStorage[id]; ok == true {
		return callback, nil
	}

	return nil, fmt.Errorf("could not find a callback with id: %s", id)

}

func GCCallback(sleep time.Duration) {

	if CallbackStorage != nil {
		expiredKeys := make([]string,0)
		for i := range CallbackStorage {
			if CallbackStorage[i] == nil {
				expiredKeys = append(expiredKeys, i)
			} else if CallbackStorage[i].Created.Sub(time.Now()).Hours() >= 1 {
				expiredKeys = append(expiredKeys, i)
			}
		}

		for i := range expiredKeys {
			delete(CallbackStorage, expiredKeys[i])
		}

	}

	time.Sleep(sleep)

}