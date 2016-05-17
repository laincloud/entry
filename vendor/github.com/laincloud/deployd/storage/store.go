package storage

import "errors"

var (
	ErrNoSuchKey = errors.New("No such key")
)

type Store interface {
	Get(key string, v interface{}) error
	Set(key string, v interface{}, force ...bool) error
	KeysByPrefix(prefix string) ([]string, error)
	Remove(key string) error
	TryRemoveDir(key string)
	RemoveDir(key string) error
}
