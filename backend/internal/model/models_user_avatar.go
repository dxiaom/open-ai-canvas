package model

import "time"

// UserAvatar keeps the bounded image payload out of ordinary user queries.
// New avatars store the image in the system-configured resource storage
// (local disk or OSS/S3/R2/etc.) and record the location here. The legacy
// Data column is retained so existing in-database avatars remain readable
// until the user re-uploads.
type UserAvatar struct {
	UserID           string    `gorm:"primaryKey;size:36" json:"userId"`
	Data             []byte    `json:"-"` // backward compat for pre-storage avatars
	Provider         string    `gorm:"size:24" json:"-"`
	ObjectKey        string    `gorm:"size:512" json:"-"`
	Endpoint         string    `gorm:"size:255" json:"-"`
	Bucket           string    `gorm:"size:255" json:"-"`
	StorageSettingID string    `gorm:"size:36" json:"-"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// AvatarStorageLocation describes where the avatar image is physically stored.
type AvatarStorageLocation struct {
	Provider         string
	ObjectKey        string
	Endpoint         string
	Bucket           string
	StorageSettingID string
	PublicURL        string // direct CDN/origin URL for browser access; empty means proxy through API
}

// HasStorageLocation reports whether the avatar is stored externally (not in Data).
func (a *UserAvatar) HasStorageLocation() bool {
	return a.ObjectKey != ""
}

// StorageLocation returns the storage location info.
func (a *UserAvatar) StorageLocation() AvatarStorageLocation {
	return AvatarStorageLocation{
		Provider:         a.Provider,
		ObjectKey:        a.ObjectKey,
		Endpoint:         a.Endpoint,
		Bucket:           a.Bucket,
		StorageSettingID: a.StorageSettingID,
	}
}
