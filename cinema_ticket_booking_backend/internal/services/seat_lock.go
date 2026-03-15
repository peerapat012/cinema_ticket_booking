package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/peerapat012/CinemaTicketBooking/internal/database"
	"github.com/redis/go-redis/v9"
)

const (
	LockDuration = 5 * time.Minute
)

var (
	ErrNotOwner   = errors.New("you do not own this lock")
	ErrNotLocked  = errors.New("seat is not locked")
	ErrLockFailed = errors.New("failed to acquire lock")
)

type SeatLockService struct {
	redis *redis.Client
}

func NewSeatLockService() *SeatLockService {
	redisClient := database.GetRedis()
	return &SeatLockService{
		redis: redisClient,
	}
}

func (s *SeatLockService) IsAvailable() bool {
	return s.redis != nil
}

func (s *SeatLockService) lockKey(movieID, seatNo string) string {
	return fmt.Sprintf("seat_lock:%s:%s", movieID, seatNo)
}

func (s *SeatLockService) LockSeat(ctx context.Context, movieID, seatNo, userID string) (bool, error) {
	if s.redis == nil {
		return true, nil
	}

	key := s.lockKey(movieID, seatNo)

	locked, err := s.redis.SetNX(ctx, key, userID, LockDuration).Result()
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return false, err
		}
		return false, fmt.Errorf("%w: %v", ErrLockFailed, err)
	}

	return locked, nil
}

func (s *SeatLockService) ExtendLock(ctx context.Context, movieID string, seats []string, userID string) error {
	if s.redis == nil {
		return nil
	}

	for _, seatNo := range seats {
		key := s.lockKey(movieID, seatNo)
		owner, err := s.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}
		if owner == userID {
			s.redis.Expire(ctx, key, LockDuration)
		}
	}
	return nil
}

func (s *SeatLockService) UnlockSeat(ctx context.Context, movieID, seatNo, userID string) error {
	if s.redis == nil {
		return nil
	}

	key := s.lockKey(movieID, seatNo)

	script := redis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`)

	result, err := script.Run(ctx, s.redis, []string{key}, userID).Int64()
	if err != nil {
		return fmt.Errorf("failed to execute unlock script: %w", err)
	}

	if result == 0 {
		owner, _ := s.redis.Get(ctx, key).Result()
		if owner == "" {
			return nil
		}
		return ErrNotOwner
	}

	return nil
}

func (s *SeatLockService) IsLocked(ctx context.Context, movieID, seatNo string) (bool, string, error) {
	if s.redis == nil {
		return false, "", nil
	}

	key := s.lockKey(movieID, seatNo)

	owner, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, "", nil
		}
		return false, "", err
	}

	return true, owner, nil
}

func (s *SeatLockService) GetLockTTL(ctx context.Context, movieID, seatNo string) (time.Duration, error) {
	if s.redis == nil {
		return 0, nil
	}

	key := s.lockKey(movieID, seatNo)

	ttl, err := s.redis.TTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	return ttl, nil
}

func (s *SeatLockService) LockSeats(ctx context.Context, movieID string, seats []string, userID string) (map[string]bool, error) {
	results := make(chan struct {
		seatNo string
		locked bool
		err    error
	}, len(seats))

	for _, seatNo := range seats {
		go func(seatNo string) {
			locked, err := s.LockSeat(ctx, movieID, seatNo, userID)
			results <- struct {
				seatNo string
				locked bool
				err    error
			}{seatNo, locked, err}
		}(seatNo)
	}

	lockResults := make(map[string]bool)
	for i := 0; i < len(seats); i++ {
		result := <-results
		if result.err != nil {
			return lockResults, result.err
		}
		lockResults[result.seatNo] = result.locked
	}
	close(results)

	return lockResults, nil
}

var SeatLock *SeatLockService

func init() {
	SeatLock = NewSeatLockService()
}
