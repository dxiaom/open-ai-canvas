package repository

import (
	"errors"
	"time"

	"yingce/backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrProfileAvatarUserUnavailable = errors.New("avatar owner is not active")

func (r *Repository) UserAvatar(userID string) (*model.UserAvatar, error) {
	var avatar model.UserAvatar
	err := r.db.First(&avatar, "user_id = ?", userID).Error
	return &avatar, err
}

// UpdateUserAvatar serializes avatar writes on the user row so the payload and
// its cache version always belong to the same committed selection. For custom
// avatars the image is stored externally; only the storage location is persisted
// here. The data parameter is non-nil only for custom avatars and is not written
// to the database — it is used solely to distinguish custom from default.
func (r *Repository) UpdateUserAvatar(userID string, data []byte, loc model.AvatarStorageLocation, version string, preset string, now time.Time) (*model.User, error) {
	var user model.User
	err := r.db.Transaction(func(tx *gorm.DB) error {
		mode := "default"
		if len(data) > 0 {
			mode = "custom"
		}
		updates := map[string]any{"avatar_mode": mode, "avatar_version": version, "updated_at": now}
		if mode == "default" {
			updates["avatar_preset"] = preset
		}
		result := tx.Model(&model.User{}).Where("id = ? AND status = ?", userID, model.UserStatusActive).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrProfileAvatarUserUnavailable
		}
		if mode == "custom" {
			avatar := model.UserAvatar{
				UserID:           userID,
				Provider:         loc.Provider,
				ObjectKey:        loc.ObjectKey,
				Endpoint:         loc.Endpoint,
				Bucket:           loc.Bucket,
				StorageSettingID: loc.StorageSettingID,
				UpdatedAt:        now,
			}
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}}, DoUpdates: clause.AssignmentColumns([]string{"provider", "object_key", "endpoint", "bucket", "storage_setting_id", "updated_at"})}).Create(&avatar).Error; err != nil {
				return err
			}
		} else if err := tx.Delete(&model.UserAvatar{}, "user_id = ?", userID).Error; err != nil {
			return err
		}
		return tx.First(&user, "id = ?", userID).Error
	})
	return &user, err
}
