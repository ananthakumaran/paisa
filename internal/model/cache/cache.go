package cache

import (
	"fmt"
	"time"

	"github.com/kelindar/binary"
	"github.com/mitchellh/hashstructure/v2"
	"gorm.io/gorm"
)

type Cache struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
	HashKey   string    `json:"hash_key"`
	Value     []byte    `gorm:"type:BLOB" json:"item"`
}

func DeleteExpired(db *gorm.DB) error {
	err := db.Exec("DELETE FROM caches WHERE expires_at < ?", time.Now()).Error
	if err != nil {
		return err
	}
	return nil
}

func Lookup[I any, K any](db *gorm.DB, key K, fallback func() I) I {
	var item I
	var cache Cache

	hash, err := hashstructure.Hash(key, hashstructure.FormatV2, nil)
	hashKey := fmt.Sprintf("%d", hash)
	if err == nil {
		err := db.Where("hash_key = ?", hashKey).First(&cache).Error
		if err == nil {
			if time.Now().Before(cache.ExpiresAt) {
				err := binary.Unmarshal(cache.Value, &item)
				if err == nil {
					return item
				}
			} else {
				DeleteExpired(db)
			}
		}
	}

	item = fallback()
	bytes, err := binary.Marshal(item)
	if err == nil {
		cache = Cache{
			ExpiresAt: time.Now().Add(24 * time.Hour),
			HashKey:   hashKey,
			Value:     bytes,
		}
		db.Save(&cache)
	}
	return item
}
