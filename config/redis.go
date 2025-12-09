package config

import (
	"fmt"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
)

var redisPool *redis.Pool

// InitRedisPool menginisialisasi connection pool untuk Redis
func InitRedisPool() error {
	redisPool = &redis.Pool{
		MaxIdle:     10,
		MaxActive:   100,
		IdleTimeout: 240 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			connStr := fmt.Sprintf("%s:%s", REDISHost, REDISPort)

			log.Printf("Connecting to Redis at %s", connStr)

			conn, err := redis.Dial("tcp", connStr,
				redis.DialConnectTimeout(5*time.Second),
				redis.DialReadTimeout(5*time.Second),
				redis.DialWriteTimeout(5*time.Second),
			)

			if err != nil {
				log.Printf("Failed to connect to Redis: %v", err)
				return nil, fmt.Errorf("failed to connect to redis: %w", err)
			}

			if REDISPass != "" {
				if _, err := conn.Do("AUTH", REDISPass); err != nil {
					conn.Close()
					log.Printf("Failed to authenticate to Redis: %v", err)
					return nil, fmt.Errorf("failed to authenticate: %w", err)
				}
			}

			log.Println("Successfully connected to Redis")
			return conn, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	conn := redisPool.Get()
	defer conn.Close()

	_, err := conn.Do("PING")
	if err != nil {
		log.Printf("Failed to ping Redis: %v", err)
		return fmt.Errorf("redis connection test failed: %w", err)
	}

	log.Println("Redis pool initialized successfully")
	return nil
}

func GetRedisConn() redis.Conn {
	if redisPool == nil {
		log.Println("Warning: Redis pool is nil, initializing...")
		if err := InitRedisPool(); err != nil {
			log.Printf("Failed to initialize Redis pool: %v", err)
			return nil
		}
	}
	return redisPool.Get()
}

func CloseRedisPool() {
	if redisPool != nil {
		log.Println("Closing Redis pool...")
		redisPool.Close()
	}
}

func PingRedis() error {
	conn := GetRedisConn()
	if conn == nil {
		return fmt.Errorf("failed to get redis connection")
	}
	defer conn.Close()

	_, err := conn.Do("PING")
	return err
}

// SetResetToken menyimpan reset token dan account number dengan expiry time
func SetResetToken(token string, accountNumber string, expiryTime time.Time) error {
	conn := GetRedisConn()
	if conn == nil {
		return fmt.Errorf("failed to get redis connection")
	}
	defer conn.Close()

	// Hitung selisih detik dari sekarang sampai expiry time
	expirySeconds := int(time.Until(expiryTime).Seconds())

	// Validasi expiry tidak negatif atau terlalu kecil
	if expirySeconds <= 0 {
		return fmt.Errorf("expiry time must be in the future")
	}

	tokenKey := fmt.Sprintf("reset_token:%s", token)

	_, err := conn.Do("SETEX", tokenKey, expirySeconds, accountNumber)
	if err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	// log.Printf("Token stored with expiry in %d seconds", expirySeconds)
	return nil
}

// SetResetTokenWithDuration menyimpan reset token dengan durasi
func SetResetTokenWithDuration(token string, accountNumber string, duration time.Duration) error {
	expiryTime := time.Now().Add(duration)
	return SetResetToken(token, accountNumber, expiryTime)
}

// SetResetTokenWithSeconds menyimpan reset token dengan detik
func SetResetTokenWithSeconds(token string, accountNumber string, seconds int) error {
	expiryTime := time.Now().Add(time.Duration(seconds) * time.Second)
	return SetResetToken(token, accountNumber, expiryTime)
}

// GetAccountNumberByToken mendapatkan account number dari token
func GetAccountNumberByToken(token string) (string, error) {
	conn := GetRedisConn()
	if conn == nil {
		return "", fmt.Errorf("failed to get redis connection")
	}
	defer conn.Close()

	tokenKey := fmt.Sprintf("reset_token:%s", token)

	accountNumber, err := redis.String(conn.Do("GET", tokenKey))
	if err != nil {
		if err == redis.ErrNil {
			return "", fmt.Errorf("token expired or not found")
		}
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	return accountNumber, nil
}

// DeleteResetToken menghapus token setelah digunakan
func DeleteResetToken(token string) error {
	conn := GetRedisConn()
	if conn == nil {
		return fmt.Errorf("failed to get redis connection")
	}
	defer conn.Close()

	tokenKey := fmt.Sprintf("reset_token:%s", token)

	_, err := conn.Do("DEL", tokenKey)
	return err
}
