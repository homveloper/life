package bullet

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/danghamo/life/internal/domain/shared"
)

// RedisRepository implements Repository using Redis
type RedisRepository struct {
	client *redis.Client
}

// NewRedisRepository creates a new Redis-based bullet repository
func NewRedisRepository(client *redis.Client) Repository {
	repo := &RedisRepository{
		client: client,
	}

	// Initialize JSON search index (non-blocking)
	go repo.initializeSearchIndex()

	return repo
}

// initializeSearchIndex creates the FT.CREATE index for bullet JSON documents
func (r *RedisRepository) initializeSearchIndex() {
	ctx := context.Background()

	// Drop existing index if it exists (ignore errors)
	r.client.Do(ctx, "FT.DROPINDEX", "idx:bullet")

	// Create new search index for JSON documents
	_, err := r.client.Do(ctx, "FT.CREATE", "idx:bullet",
		"ON", "JSON",
		"PREFIX", "1", "bullet:",
		"SCHEMA",
		"$.id", "AS", "id", "TEXT",
		"$.player_id", "AS", "player_id", "TAG",
		"$.weapon_type", "AS", "weapon_type", "TAG",
		"$.state", "AS", "state", "TAG",
		"$.current_position.x", "AS", "current_pos_x", "NUMERIC",
		"$.current_position.y", "AS", "current_pos_y", "NUMERIC",
		"$.max_range", "AS", "max_range", "NUMERIC",
		"$.fired_at", "AS", "fired_at", "NUMERIC",
		"$.expires_at", "AS", "expires_at", "NUMERIC",
	).Result()

	if err != nil {
		// Log error but don't fail - repository will use fallback methods
		fmt.Printf("Warning: Failed to create bullet search index: %v\n", err)
	}
}

// Save saves a bullet to Redis using JSONSet with TTL
func (r *RedisRepository) Save(ctx context.Context, bullet *Bullet) error {
	key := r.bulletKey(bullet.ID)

	// Calculate TTL based on bullet expiry time
	ttl := time.Until(bullet.ExpiresAt)
	if ttl <= 0 {
		ttl = 10 * time.Second // Minimum TTL for expired bullets
	}

	// Store bullet data using JSONSet
	if err := r.client.JSONSet(ctx, key, "$", bullet).Err(); err != nil {
		return fmt.Errorf("failed to save bullet: %w", err)
	}

	// Set TTL
	if err := r.client.Expire(ctx, key, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set TTL: %w", err)
	}

	return nil
}

// Load loads a bullet by ID using JSONGet
func (r *RedisRepository) Load(ctx context.Context, bulletID BulletID) (*Bullet, error) {
	key := r.bulletKey(bulletID)

	result, err := r.client.JSONGet(ctx, key, "$").Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to load bullet: %w", err)
	}

	// JSONGet with "$" returns the root object as string
	bullet := &Bullet{}
	if err := json.Unmarshal([]byte(result), bullet); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bullet: %w", err)
	}

	return bullet, nil
}

// LoadActive loads all active bullets using FT.SEARCH
func (r *RedisRepository) LoadActive(ctx context.Context) ([]*Bullet, error) {
	// Use FT.SEARCH to find bullets with state="active"
	result, err := r.client.Do(ctx, "FT.SEARCH", "idx:bullet", "@state:active").Result()
	if err != nil {
		// Fallback to scanning if index doesn't exist
		return r.loadActiveFallback(ctx)
	}

	return r.parseBulletSearchResults(ctx, result)
}

// loadActiveFallback fallback method when search index is not available
func (r *RedisRepository) loadActiveFallback(ctx context.Context) ([]*Bullet, error) {
	pattern := r.bulletKey(BulletID("*"))

	var bullets []*Bullet
	iter := r.client.Scan(ctx, 0, pattern, 100).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()
		bulletID := BulletID(key[len("bullet:"):])

		bullet, err := r.Load(ctx, bulletID)
		if err != nil {
			continue // Skip failed loads
		}
		if bullet != nil && bullet.IsActive() {
			bullets = append(bullets, bullet)
		}
	}

	return bullets, iter.Err()
}

// parseBulletSearchResults parses FT.SEARCH results into bullets
func (r *RedisRepository) parseBulletSearchResults(ctx context.Context, result any) ([]*Bullet, error) {
	results, ok := result.([]any)
	if !ok || len(results) < 1 {
		return []*Bullet{}, nil
	}

	// First element is the count
	count, ok := results[0].(int64)
	if !ok {
		return []*Bullet{}, nil
	}

	var bullets []*Bullet
	// Results come in pairs: [key, fields, key, fields, ...]
	for i := int64(1); i < count*2+1; i += 2 {
		if i+1 >= int64(len(results)) {
			break
		}

		key, ok := results[i].(string)
		if !ok {
			continue
		}

		// Extract bullet ID from key
		bulletID := BulletID(key[len("bullet:"):])
		bullet, err := r.Load(ctx, bulletID)
		if err != nil {
			continue
		}
		if bullet != nil {
			bullets = append(bullets, bullet)
		}
	}

	return bullets, nil
}

// LoadActiveForPlayer loads all active bullets for a specific player using FT.SEARCH
func (r *RedisRepository) LoadActiveForPlayer(ctx context.Context, playerID PlayerID) ([]*Bullet, error) {
	// Use FT.SEARCH with multiple conditions: state=active AND player_id=playerID
	query := fmt.Sprintf("@state:active @player_id:%s", playerID.String())
	result, err := r.client.Do(ctx, "FT.SEARCH", "idx:bullet", query).Result()
	if err != nil {
		// Fallback to scanning if index doesn't exist
		return r.loadActiveForPlayerFallback(ctx, playerID)
	}

	return r.parseBulletSearchResults(ctx, result)
}

// loadActiveForPlayerFallback fallback method when search index is not available
func (r *RedisRepository) loadActiveForPlayerFallback(ctx context.Context, playerID PlayerID) ([]*Bullet, error) {
	pattern := r.bulletKey(BulletID("*"))

	var bullets []*Bullet
	iter := r.client.Scan(ctx, 0, pattern, 100).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()
		bulletID := BulletID(key[len("bullet:"):])

		bullet, err := r.Load(ctx, bulletID)
		if err != nil {
			continue // Skip failed loads
		}
		if bullet != nil && bullet.IsActive() && bullet.PlayerID == playerID {
			bullets = append(bullets, bullet)
		}
	}

	return bullets, iter.Err()
}

// LoadInArea loads all active bullets within a specific area using geo search
func (r *RedisRepository) LoadInArea(ctx context.Context, topLeft, bottomRight shared.Position) ([]*Bullet, error) {
	// Use FT.SEARCH with numeric range for position
	query := fmt.Sprintf("@state:active @current_pos_x:[%f %f] @current_pos_y:[%f %f]",
		topLeft.X, bottomRight.X, topLeft.Y, bottomRight.Y)

	result, err := r.client.Do(ctx, "FT.SEARCH", "idx:bullet", query).Result()
	if err != nil {
		// Fallback to loading all active bullets and filtering
		return r.loadInAreaFallback(ctx, topLeft, bottomRight)
	}

	return r.parseBulletSearchResults(ctx, result)
}

// loadInAreaFallback fallback method when search index is not available
func (r *RedisRepository) loadInAreaFallback(ctx context.Context, topLeft, bottomRight shared.Position) ([]*Bullet, error) {
	activeBullets, err := r.loadActiveFallback(ctx)
	if err != nil {
		return nil, err
	}

	var bulletsInArea []*Bullet
	for _, bullet := range activeBullets {
		if r.isPositionInArea(bullet.CurrentPos, topLeft, bottomRight) {
			bulletsInArea = append(bulletsInArea, bullet)
		}
	}

	return bulletsInArea, nil
}

// Delete removes a bullet from the repository
func (r *RedisRepository) Delete(ctx context.Context, bulletID BulletID) error {
	key := r.bulletKey(bulletID)

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// Load bullet data for index cleanup
		data := tx.HGetAll(ctx, key)
		if data.Err() != nil && data.Err() != redis.Nil {
			return data.Err()
		}

		if len(data.Val()) == 0 {
			return nil // Already deleted
		}

		bullet := &Bullet{}
		if dataStr, exists := data.Val()["data"]; exists {
			if err := json.Unmarshal([]byte(dataStr), bullet); err != nil {
				return err
			}
		} else {
			return nil // No data to clean up
		}

		// Execute transaction
		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Del(ctx, key)
			// Note: No manual index cleanup needed with Redis JSON search index
			return nil
		})

		return err
	}, key)
}

// DeleteExpired removes all expired bullets
func (r *RedisRepository) DeleteExpired(ctx context.Context, expiredBefore time.Time) (int, error) {
	// Redis TTL should handle most expired bullets automatically
	// This method handles any remaining ones
	activeBullets, err := r.LoadActive(ctx)
	if err != nil {
		return 0, err
	}

	deletedCount := 0
	for _, bullet := range activeBullets {
		if bullet.ShouldExpire(expiredBefore) {
			if err := r.Delete(ctx, bullet.ID); err != nil {
				return deletedCount, err
			}
			deletedCount++
		}
	}

	return deletedCount, nil
}

// Exists checks if a bullet exists
func (r *RedisRepository) Exists(ctx context.Context, bulletID BulletID) (bool, error) {
	key := r.bulletKey(bulletID)

	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return exists > 0, nil
}

// Count returns the total number of bullets
func (r *RedisRepository) Count(ctx context.Context) (int64, error) {
	pattern := "bullet:*"

	var count int64
	iter := r.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		count++
	}

	return count, iter.Err()
}

// CountActiveForPlayer returns the number of active bullets for a player
func (r *RedisRepository) CountActiveForPlayer(ctx context.Context, playerID PlayerID) (int64, error) {
	indexKey := fmt.Sprintf("idx:bullet:player:%s", playerID.String())

	count, err := r.client.SCard(ctx, indexKey).Result()
	if err != nil {
		return 0, err
	}

	return count, nil
}

// SaveBatch saves multiple bullets in a single pipeline transaction
func (r *RedisRepository) SaveBatch(ctx context.Context, bullets []*Bullet) error {
	if len(bullets) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()

	for _, bullet := range bullets {
		key := r.bulletKey(bullet.ID)

		// Calculate TTL
		ttl := time.Until(bullet.ExpiresAt)
		if ttl <= 0 {
			ttl = 10 * time.Second
		}

		// Add to pipeline - use JSONSet for individual bullets in batch
		pipe.JSONSet(ctx, key, "$", bullet)
		pipe.Expire(ctx, key, ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// DeleteBatch deletes multiple bullets in a single pipeline transaction
func (r *RedisRepository) DeleteBatch(ctx context.Context, bulletIDs []BulletID) error {
	if len(bulletIDs) == 0 {
		return nil
	}

	// First load bullets to get data for index cleanup
	bullets := make([]*Bullet, 0, len(bulletIDs))
	for _, bulletID := range bulletIDs {
		bullet, err := r.Load(ctx, bulletID)
		if err != nil {
			return fmt.Errorf("failed to load bullet %s for deletion: %w", bulletID, err)
		}
		if bullet != nil {
			bullets = append(bullets, bullet)
		}
	}

	// Batch delete in pipeline
	pipe := r.client.Pipeline()
	for _, bullet := range bullets {
		key := r.bulletKey(bullet.ID)
		pipe.Del(ctx, key)
		// Note: No manual index cleanup needed with Redis JSON search index
	}

	_, err := pipe.Exec(ctx)
	return err
}

// LoadBatch loads multiple bullets by their IDs
func (r *RedisRepository) LoadBatch(ctx context.Context, bulletIDs []BulletID) ([]*Bullet, error) {
	if len(bulletIDs) == 0 {
		return []*Bullet{}, nil
	}

	pipe := r.client.Pipeline()
	commands := make(map[BulletID]*redis.MapStringStringCmd)

	// Batch load all bullets
	for _, bulletID := range bulletIDs {
		key := r.bulletKey(bulletID)
		commands[bulletID] = pipe.HGetAll(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute batch load pipeline: %w", err)
	}

	// Process results
	bullets := make([]*Bullet, 0, len(bulletIDs))
	for bulletID, cmd := range commands {
		data, err := cmd.Result()
		if err != nil || len(data) == 0 {
			continue // Skip missing bullets
		}

		bullet := &Bullet{}
		if dataStr, exists := data["data"]; exists {
			if err := json.Unmarshal([]byte(dataStr), bullet); err != nil {
				return nil, fmt.Errorf("failed to unmarshal bullet %s: %w", bulletID, err)
			}
		} else {
			continue // Skip bullets without data field
		}

		bullets = append(bullets, bullet)
	}

	return bullets, nil
}

// bulletKey returns the Redis key for a bullet
func (r *RedisRepository) bulletKey(bulletID BulletID) string {
	return fmt.Sprintf("bullet:%s", bulletID.String())
}

// Note: With Redis JSON and FT.CREATE index, we no longer need manual index management
// The search index is automatically maintained by Redis

// isPositionInArea checks if position is within the specified area
func (r *RedisRepository) isPositionInArea(pos, topLeft, bottomRight shared.Position) bool {
	return pos.X >= topLeft.X && pos.X <= bottomRight.X &&
		pos.Y >= topLeft.Y && pos.Y <= bottomRight.Y
}

// RedisPlayerStatsRepository implements PlayerStatsRepository using Redis
type RedisPlayerStatsRepository struct {
	client *redis.Client
}

// NewRedisPlayerStatsRepository creates a new Redis-based player stats repository
func NewRedisPlayerStatsRepository(client *redis.Client) PlayerStatsRepository {
	return &RedisPlayerStatsRepository{
		client: client,
	}
}

// SaveStats saves player firing stats
func (r *RedisPlayerStatsRepository) SaveStats(ctx context.Context, stats *PlayerStats) error {
	key := r.playerStatsKey(stats.PlayerID)

	data, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("failed to marshal player stats: %w", err)
	}

	fields := map[string]any{
		"data": string(data),
	}

	return r.client.HMSet(ctx, key, fields).Err()
}

// LoadStats loads player firing stats
func (r *RedisPlayerStatsRepository) LoadStats(ctx context.Context, playerID PlayerID) (*PlayerStats, error) {
	key := r.playerStatsKey(playerID)

	data, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to load player stats: %w", err)
	}

	if len(data) == 0 {
		return nil, nil // Not found
	}

	stats := &PlayerStats{}
	if dataStr, exists := data["data"]; exists {
		if err := json.Unmarshal([]byte(dataStr), stats); err != nil {
			return nil, fmt.Errorf("failed to unmarshal player stats: %w", err)
		}
	} else {
		return nil, nil // Data field not found, stats don't exist
	}

	return stats, nil
}

// LoadStatsWithDefaults loads stats or creates default if not found
func (r *RedisPlayerStatsRepository) LoadStatsWithDefaults(ctx context.Context, playerID PlayerID, defaultAmmo int, defaultWeapon WeaponType) (*PlayerStats, error) {
	stats, err := r.LoadStats(ctx, playerID)
	if err != nil {
		return nil, err
	}

	if stats == nil {
		// Create default stats
		stats = NewPlayerStats(playerID, defaultAmmo, defaultWeapon)
		if err := r.SaveStats(ctx, stats); err != nil {
			return nil, fmt.Errorf("failed to save default stats: %w", err)
		}
	}

	return stats, nil
}

// DeleteStats removes player stats
func (r *RedisPlayerStatsRepository) DeleteStats(ctx context.Context, playerID PlayerID) error {
	key := r.playerStatsKey(playerID)
	return r.client.Del(ctx, key).Err()
}

// StatsExists checks if player stats exist
func (r *RedisPlayerStatsRepository) StatsExists(ctx context.Context, playerID PlayerID) (bool, error) {
	key := r.playerStatsKey(playerID)

	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return exists > 0, nil
}

// SaveStatsBatch saves multiple player stats in a single pipeline transaction
func (r *RedisPlayerStatsRepository) SaveStatsBatch(ctx context.Context, stats []*PlayerStats) error {
	if len(stats) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()

	for _, stat := range stats {
		key := r.playerStatsKey(stat.PlayerID)

		data, err := json.Marshal(stat)
		if err != nil {
			return fmt.Errorf("failed to marshal player stats for %s: %w", stat.PlayerID, err)
		}

		fields := map[string]any{
			"data": string(data),
		}

		pipe.HMSet(ctx, key, fields)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// LoadStatsBatch loads multiple player stats by their IDs
func (r *RedisPlayerStatsRepository) LoadStatsBatch(ctx context.Context, playerIDs []PlayerID) ([]*PlayerStats, error) {
	if len(playerIDs) == 0 {
		return []*PlayerStats{}, nil
	}

	pipe := r.client.Pipeline()
	commands := make(map[PlayerID]*redis.MapStringStringCmd)

	// Batch load all player stats
	for _, playerID := range playerIDs {
		key := r.playerStatsKey(playerID)
		commands[playerID] = pipe.HGetAll(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute batch load pipeline: %w", err)
	}

	// Process results
	statsList := make([]*PlayerStats, 0, len(playerIDs))
	for playerID, cmd := range commands {
		data, err := cmd.Result()
		if err != nil || len(data) == 0 {
			continue // Skip missing stats
		}

		stats := &PlayerStats{}
		if dataStr, exists := data["data"]; exists {
			if err := json.Unmarshal([]byte(dataStr), stats); err != nil {
				return nil, fmt.Errorf("failed to unmarshal player stats for %s: %w", playerID, err)
			}
		} else {
			continue // Skip stats without data field
		}

		statsList = append(statsList, stats)
	}

	return statsList, nil
}

// ValidateFirePermissionsBatch validates fire permissions for multiple players in a single pipeline
func (r *RedisPlayerStatsRepository) ValidateFirePermissionsBatch(ctx context.Context, playerIDs []PlayerID, currentTime time.Time) (map[PlayerID]bool, error) {
	if len(playerIDs) == 0 {
		return make(map[PlayerID]bool), nil
	}

	pipe := r.client.Pipeline()
	commands := make(map[PlayerID]*redis.MapStringStringCmd)

	// Batch load all player stats
	for _, playerID := range playerIDs {
		key := r.playerStatsKey(playerID)
		commands[playerID] = pipe.HGetAll(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute validation pipeline: %w", err)
	}

	// Process results and validate permissions
	results := make(map[PlayerID]bool)
	for playerID, cmd := range commands {
		data, err := cmd.Result()
		if err != nil || len(data) == 0 {
			results[playerID] = false // No stats = can't fire
			continue
		}

		stats := &PlayerStats{}
		if dataStr, exists := data["data"]; exists {
			if err := json.Unmarshal([]byte(dataStr), stats); err != nil {
				results[playerID] = false
				continue
			}
		} else {
			results[playerID] = false
			continue
		}

		results[playerID] = stats.CanFire(currentTime)
	}

	return results, nil
}

// playerStatsKey returns the Redis key for player stats
func (r *RedisPlayerStatsRepository) playerStatsKey(playerID PlayerID) string {
	return fmt.Sprintf("player:%s:stats", playerID.String())
}

// RedisFireSessionRepository implements FireSessionRepository using Redis
type RedisFireSessionRepository struct {
	client *redis.Client
}

// NewRedisFireSessionRepository creates a new Redis-based fire session repository
func NewRedisFireSessionRepository(client *redis.Client) FireSessionRepository {
	return &RedisFireSessionRepository{
		client: client,
	}
}

// SaveSession saves a fire session with TTL
func (r *RedisFireSessionRepository) SaveSession(ctx context.Context, session *FireSession) error {
	key := r.fireSessionKey(session.SessionID)

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to serialize fire session: %w", err)
	}

	// Set TTL for fire sessions (1 hour)
	ttl := 1 * time.Hour
	if err := r.client.Set(ctx, key, string(data), ttl).Err(); err != nil {
		return err
	}

	// Update player index
	if session.IsActive {
		playerKey := r.playerSessionKey(session.PlayerID)
		return r.client.Set(ctx, playerKey, session.SessionID.String(), ttl).Err()
	}

	return nil
}

// LoadSession loads a fire session by ID
func (r *RedisFireSessionRepository) LoadSession(ctx context.Context, sessionID shared.ID) (*FireSession, error) {
	key := r.fireSessionKey(sessionID)

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to load fire session: %w", err)
	}

	session := &FireSession{}
	if err := json.Unmarshal([]byte(data), session); err != nil {
		return nil, fmt.Errorf("failed to deserialize fire session: %w", err)
	}

	return session, nil
}

// LoadActiveSessionForPlayer loads active session for a player
func (r *RedisFireSessionRepository) LoadActiveSessionForPlayer(ctx context.Context, playerID PlayerID) (*FireSession, error) {
	playerKey := r.playerSessionKey(playerID)

	sessionIDStr, err := r.client.Get(ctx, playerKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No active session
		}
		return nil, fmt.Errorf("failed to load player session ID: %w", err)
	}

	sessionID := shared.ID(sessionIDStr)
	return r.LoadSession(ctx, sessionID)
}

// DeleteSession removes a fire session
func (r *RedisFireSessionRepository) DeleteSession(ctx context.Context, sessionID shared.ID) error {
	key := r.fireSessionKey(sessionID)

	// Load session to get player ID for cleanup
	session, err := r.LoadSession(ctx, sessionID)
	if err != nil || session == nil {
		// Just delete the key anyway
		return r.client.Del(ctx, key).Err()
	}

	pipe := r.client.Pipeline()
	pipe.Del(ctx, key)

	// Clean up player index
	playerKey := r.playerSessionKey(session.PlayerID)
	pipe.Del(ctx, playerKey)

	_, err = pipe.Exec(ctx)
	return err
}

// DeleteExpiredSessions removes expired sessions
func (r *RedisFireSessionRepository) DeleteExpiredSessions(ctx context.Context, expiredBefore time.Time) (int, error) {
	// Redis TTL handles expiration automatically
	// This method is for manual cleanup if needed
	pattern := "fire_session:*"

	deletedCount := 0
	iter := r.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()

		// Check if key exists (might have been auto-expired)
		exists, err := r.client.Exists(ctx, key).Result()
		if err != nil || exists == 0 {
			continue
		}

		// Try to load and check expiry
		data, err := r.client.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		session := &FireSession{}
		if err := json.Unmarshal([]byte(data), session); err != nil {
			continue
		}

		if session.StartedAt.Before(expiredBefore) {
			if err := r.DeleteSession(ctx, session.SessionID); err == nil {
				deletedCount++
			}
		}
	}

	return deletedCount, iter.Err()
}

// fireSessionKey returns the Redis key for a fire session
func (r *RedisFireSessionRepository) fireSessionKey(sessionID shared.ID) string {
	return fmt.Sprintf("fire_session:%s", sessionID.String())
}

// playerSessionKey returns the Redis key for player's active session
func (r *RedisFireSessionRepository) playerSessionKey(playerID PlayerID) string {
	return fmt.Sprintf("player:%s:active_session", playerID.String())
}
