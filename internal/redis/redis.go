package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/JoachimFlottorp/Linnea/internal/auth"
	"github.com/JoachimFlottorp/Linnea/internal/models"
	"github.com/go-redis/redis/v8"
)

type redisInstance struct {
	client *redis.Client
}

func Create(ctx context.Context, options Options) (Instance, error) {
	rds := redis.NewClient(&redis.Options{
		Addr:     options.Address,
		Username: options.Username,
		Password: options.Password,
		DB:       options.DB,
	})

	if err := rds.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	
	inst := &redisInstance{
		client: rds,
	}

	return inst, nil
}

func (r *redisInstance) formatKey(key string) string {
	return fmt.Sprintf("%s%s", r.Prefix(), key)
}

func (r *redisInstance) Prefix() string {
	return "linnea:"
}

func (r *redisInstance) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *redisInstance) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, r.formatKey(key)).Result()
}

func (r *redisInstance) Set(ctx context.Context, key string, value string) error {
	return r.client.Set(ctx, r.formatKey(key), value, 0).Err()
}

func (r *redisInstance) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, r.formatKey(key)).Err()
}

func (r *redisInstance) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, r.formatKey(key), expiration).Err()
}

func (r *redisInstance) Client() *redis.Client {
	return r.client
}

func (r *redisInstance) SetMember(ctx context.Context, key string, member string) error {
	return r.client.SAdd(ctx, r.formatKey(key), member).Err()
}

func (r *redisInstance) GetMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, r.formatKey(key)).Result()
}

func (r *redisInstance) DelMember(ctx context.Context, key string, member string) error {
	return r.client.SRem(ctx, r.formatKey(key), member).Err()
}

func GetUser(ctx context.Context, r Instance, secret, token string) (*models.User, error) {
	decryptedJwt, err := auth.DecryptJWT(secret, token)
	if err != nil {
		return nil, err
	}

	u := &models.User{}
	
	user, err := r.Get(ctx, fmt.Sprintf("user:%s", decryptedJwt.ID))
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(user), u); err != nil {
		return nil, err
	}

	return u, nil
}

func CreateUser(r Instance, ctx context.Context, u *models.User) error {
	user, err := json.Marshal(u)
	if err != nil {
		return err
	}

	return r.Set(ctx, fmt.Sprintf("user:%s", u.TwitchUID), string(user))
}
