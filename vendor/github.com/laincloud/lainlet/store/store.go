package store

import (
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"strings"
)

const (
	// ERROR event action, this will be set when geting error from backend(etcd, zk...)
	ERROR Action = iota - 1
	// GET event action, useless
	GET
	// SET event action, same with update, it's useless
	SET
	// UPDATE event action, return when data changed in backend, include set, create, update, compareAndSwap
	UPDATE
	// DELETE event action, return when some key was delete or expired or compareAndDelete
	DELETE
	// INIT event action, this is only retured when calling watch() and only the first correct event.
	INIT
)

var (
	// ErrBackendNotSupported is thrown when the backend k/v store is not supported by libkv
	ErrBackendNotSupported = errors.New("Backend storage not supported yet, please choose one of")
	// ErrCallNotSupported is thrown when a method is not implemented/supported by the current backend
	ErrCallNotSupported = errors.New("The current call is not supported with this backend")
	// ErrNotReachable is thrown when the API cannot be reached for issuing common store operations
	ErrNotReachable = errors.New("Api not reachable")
	// ErrCannotLock is thrown when there is an error acquiring a lock on a key
	ErrCannotLock = errors.New("Error acquiring the lock")
	// ErrKeyModified is thrown during an atomic operation if the index does not match the one in the store
	ErrKeyModified = errors.New("Unable to complete atomic operation, key modified")
	// ErrKeyNotFound is thrown when the key is not found in the store during a Get operation
	ErrKeyNotFound = errors.New("Key not found in store")
	// ErrPreviousNotSpecified is thrown when the previous value is not specified for an atomic operation
	ErrPreviousNotSpecified = errors.New("Previous K/V pair should be provided for the Atomic operation")
	// ErrKeyExists is thrown when the previous value exists in the case of an AtomicPut
	ErrKeyExists = errors.New("Previous K/V pair exists, cannnot complete Atomic operation")

	stores map[string]Initializer

	action2String = map[Action]string{
		GET:    "get",
		SET:    "update",
		UPDATE: "update",
		DELETE: "delete",
		INIT:   "init",
		ERROR:  "error",
	}
)

func init() {
	stores = make(map[string]Initializer)
}

// Action is the event type
type Action int

// Return the string value for a action
func (a Action) String() string {
	return action2String[a]
}

// Initializer is initialization function which can create real store, used when Register()
type Initializer func([]string) (Store, error)

// Store represents the backend K/V storage
// Each store should support every call listed
// here. Or it couldn't be implemented as a K/V
// backend for libkv
type Store interface {
	// Get a value given its key
	Get(key string) (*KVPair, error)

	// Get the content of a given prefix by recursive
	GetTree(key string) ([]*KVPair, error)

	// Set a value for the given key
	Put(key string, value []byte) error

	// Delete a key
	Delete(key string, recursive bool) error

	// Verify if a Key exists in the store
	Exists(key string) (bool, error)

	// Watch for changes on a key
	Watch(key string, ctx context.Context, recursive bool, index uint64) (<-chan *Event, error)

	// WatchTree watches for changes on child nodes under
	// a given directory
	WatchTree(directory string, ctx context.Context, index uint64) (<-chan *Event, error)

	// List all the nodes in a directory
	List(dir string) ([]string, error)

	// Close the store connection
	Close()
}

// New return a new store by given name and initialize addrs
func New(name string, addrs []string) (Store, error) {
	if f, ok := stores[name]; ok {
		return f(addrs)
	}
	return nil, ErrBackendNotSupported
}

// KVPair represents {Key, Value, Lastindex} tuple
type KVPair struct {
	Key       string
	Value     []byte
	LastIndex uint64
}

// Event represents a event get from watch,
// Action is 'delete' or 'update'
// Data is the latest key-value-data of the watched key
type Event struct {
	Action        Action
	Key           string
	ModifiedIndex uint64
	Data          []*KVPair
}

func (e *Event) String() string {
	s := ""
	for _, pair := range e.Data {
		s += fmt.Sprintf("%s => %s,", pair.Key, pair.Value)
	}
	return e.Action.String() + " " + e.Key + ": " + s
}

func (e *Event) Error() string {
	if e.Action != ERROR {
		return ""
	}
	for _, pair := range e.Data {
		if pair.Key == "error" {
			return string(pair.Value)
		}
	}
	return ""
}

/************** helper functions ************/

// Register a new store, the Intializer is a function used to initialize store
func Register(name string, f Initializer) {
	stores[name] = f
}

// CreateEndpoints creates a list of endpoints given the right scheme
func CreateEndpoints(addrs []string, scheme string) (entries []string) {
	for _, addr := range addrs {
		entries = append(entries, scheme+"://"+addr)
	}
	return entries
}

// Normalize the key for each store to the form:
//
//     /path/to/key
//
func Normalize(key string) string {
	return "/" + join(SplitKey(key))
}

// SplitKey splits the key to extract path informations
func SplitKey(key string) (path []string) {
	if strings.Contains(key, "/") {
		path = strings.Split(key, "/")
	} else {
		path = []string{key}
	}
	return path
}

// join the path parts with '/'
func join(parts []string) string {
	return strings.Join(parts, "/")
}
