package hold

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// holdScript is all-or-nothing. Redis runs it as ONE unit, so nothing can
// happen between the check loop and the write loop.
//
// KEYS = one key per seat, ARGV[1] = user id, ARGV[2] = ttl in milliseconds.
// Returns the (1-based) positions of conflicting seats, or an empty list on success.
var holdScript = redis.NewScript(`
local conflicts = {}
for i, key in ipairs(KEYS) do
  local owner = redis.call('GET', key)
  if owner and owner ~= ARGV[1] then
    conflicts[#conflicts + 1] = i
  end
end
if #conflicts > 0 then
  return conflicts
end
for _, key in ipairs(KEYS) do
  redis.call('SET', key, ARGV[1], 'PX', ARGV[2])
end
return {}
`)

// releaseScript deletes only the keys that still belong to this user,
// so we never delete someone else's hold after ours expired.
var releaseScript = redis.NewScript(`
local n = 0
for _, key in ipairs(KEYS) do
  if redis.call('GET', key) == ARGV[1] then
    redis.call('DEL', key)
    n = n + 1
  end
end
return n
`)

type Store struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewStore(rdb *redis.Client, ttl time.Duration) *Store {
	return &Store{rdb: rdb, ttl: ttl}
}

// The {showtimeID} part is a hash tag: all keys of one showtime land in the
// same slot, so the multi-key scripts also work on Redis Cluster.
func seatKey(showtimeID, seatID string) string {
	return "hold:{" + showtimeID + "}:" + seatID
}

func keysFor(showtimeID string, seatIDs []string) []string {
	keys := make([]string, len(seatIDs))
	for i, id := range seatIDs {
		keys[i] = seatKey(showtimeID, id)
	}
	return keys
}

// Hold holds all seats for the user, or none. It returns the conflicting
// seat ids (empty on success).
func (s *Store) Hold(ctx context.Context, showtimeID string, seatIDs []string, userID string) ([]string, error) {
	res, err := holdScript.Run(ctx, s.rdb, keysFor(showtimeID, seatIDs), userID, s.ttl.Milliseconds()).Slice()
	if err != nil {
		return nil, fmt.Errorf("hold script: %w", err)
	}

	conflicts := make([]string, 0, len(res))
	for _, v := range res {
		idx, ok := v.(int64)
		if !ok || idx < 1 || int(idx) > len(seatIDs) {
			return nil, fmt.Errorf("hold script: unexpected reply %v", v)
		}
		conflicts = append(conflicts, seatIDs[idx-1])
	}
	return conflicts, nil
}

// Owned reports whether EVERY seat is currently held by this user.
func (s *Store) Owned(ctx context.Context, showtimeID string, seatIDs []string, userID string) (bool, error) {
	if len(seatIDs) == 0 {
		return false, nil
	}
	vals, err := s.rdb.MGet(ctx, keysFor(showtimeID, seatIDs)...).Result()
	if err != nil {
		return false, fmt.Errorf("check holds: %w", err)
	}
	for _, v := range vals {
		owner, ok := v.(string) // a missing key is nil, so ok is false
		if !ok || owner != userID {
			return false, nil
		}
	}
	return true, nil
}

// Held returns which of the given seats are held by anyone (one MGET).
func (s *Store) Held(ctx context.Context, showtimeID string, seatIDs []string) (map[string]bool, error) {
	held := make(map[string]bool)
	if len(seatIDs) == 0 {
		return held, nil
	}
	vals, err := s.rdb.MGet(ctx, keysFor(showtimeID, seatIDs)...).Result()
	if err != nil {
		return nil, fmt.Errorf("read holds: %w", err)
	}
	for i, v := range vals {
		if v != nil {
			held[seatIDs[i]] = true
		}
	}
	return held, nil
}

// Release removes this user's holds on the given seats.
func (s *Store) Release(ctx context.Context, showtimeID string, seatIDs []string, userID string) error {
	if len(seatIDs) == 0 {
		return nil
	}
	if err := releaseScript.Run(ctx, s.rdb, keysFor(showtimeID, seatIDs), userID).Err(); err != nil {
		return fmt.Errorf("release holds: %w", err)
	}
	return nil
}
