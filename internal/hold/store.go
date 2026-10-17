package hold

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

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

func (s *Store) Owned(ctx context.Context, showtimeID string, seatIDs []string, userID string) (bool, error) {
	if len(seatIDs) == 0 {
		return false, nil
	}
	vals, err := s.rdb.MGet(ctx, keysFor(showtimeID, seatIDs)...).Result()
	if err != nil {
		return false, fmt.Errorf("check holds: %w", err)
	}
	for _, v := range vals {
		owner, ok := v.(string)
		if !ok || owner != userID {
			return false, nil
		}
	}
	return true, nil
}

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

func (s *Store) Release(ctx context.Context, showtimeID string, seatIDs []string, userID string) error {
	if len(seatIDs) == 0 {
		return nil
	}
	if err := releaseScript.Run(ctx, s.rdb, keysFor(showtimeID, seatIDs), userID).Err(); err != nil {
		return fmt.Errorf("release holds: %w", err)
	}
	return nil
}
