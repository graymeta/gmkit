package lock

import (
	"errors"
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

// RedisLocker is a Redis backed Locker implementation.
type RedisLocker struct {
	pool   *redis.Pool
	prefix string
}

var _ (Locker) = (*RedisLocker)(nil)

// NewRedisLocker initializes a new RedisLocker.
func NewRedisLocker(pool *redis.Pool, prefix string) *RedisLocker {
	return &RedisLocker{
		pool:   pool,
		prefix: prefix,
	}
}

// Lock attempts to acquire the lock. Returns true if the lock was acquired.
func (l *RedisLocker) Lock(name, uniqueID string, duration time.Duration) (bool, error) {
	conn := l.pool.Get()
	defer conn.Close()

	_, err := redis.String(conn.Do("SET", l.key(name), uniqueID, "NX", "PX", fmt.Sprintf("%d", duration/time.Millisecond)))
	if err != nil {
		if err == redis.ErrNil {
			// lock was not acquired
			return false, nil
		}
		return false, err
	}

	// lock acquired
	return true, nil
}

// ErrUnlock is the error thrown when you try to unlock a lock that's expired or
// a lock that was locked by another unique id.
var ErrUnlock = errors.New("unlock failed, name or unique id incorrect")

// Unlock attempts to unlock a lock, but only if the uniqueID is the one that has the lock.
func (l *RedisLocker) Unlock(name, uniqueID string) error {
	const unlockScript = `
	if redis.call("get",KEYS[1]) == ARGV[1] then
		return redis.call("del",KEYS[1])
	else
		return 0
	end`

	conn := l.pool.Get()
	defer conn.Close()

	cmd := redis.NewScript(1, unlockScript)
	if res, err := redis.Int(cmd.Do(conn, l.key(name), uniqueID)); err != nil {
		return err
	} else if res != 1 {
		return ErrUnlock
	}

	return nil
}

func (l *RedisLocker) key(name string) string {
	return l.prefix + name
}
