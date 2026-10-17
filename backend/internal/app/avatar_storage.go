// 头像存储适配器：将头像图片写入系统配置的资源存储（本地磁盘或 OSS/S3/R2 等），
// 跟随平台存储设置决定存储位置，与普通资源上传使用同一套存储决策逻辑。

package app

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"yingce/backend/internal/model"
	"yingce/backend/internal/storage"
)

type avatarStorageAdapter struct {
	svc *Service
}

func (a avatarStorageAdapter) StoreAvatar(userID string, data []byte) (model.AvatarStorageLocation, error) {
	setting, storageSettingID, useOSS, err := a.svc.activeAvatarOSSSetting()
	if err != nil {
		return model.AvatarStorageLocation{}, err
	}
	now := time.Now()
	provider := "local"
	objectKey := localObjectKey(userID, "avatar", "avatar.png", "image/png", now)
	if useOSS {
		provider = setting.Provider
		objectKey = ossObjectKey(setting, userID, "avatar", "avatar.png", "image/png", now)
	}
	if useOSS {
		_, err = putOSSObject(a.svc.storageSettingWithRuntimePolicy(setting), objectKey, "image/png", int64(len(data)), bytes.NewReader(data))
	} else {
		err = writeLocalResourceObject(filepath.Join(a.svc.dataDir, "resources", filepath.FromSlash(objectKey)), bytes.NewReader(data))
	}
	if err != nil {
		return model.AvatarStorageLocation{}, err
	}
	loc := model.AvatarStorageLocation{Provider: provider, ObjectKey: objectKey}
	if useOSS {
		loc.Endpoint = setting.Endpoint
		loc.Bucket = setting.Bucket
		loc.StorageSettingID = storageSettingID
		loc.PublicURL = avatarDirectURL(setting, objectKey)
	}
	return loc, nil
}

// avatarDirectURL returns a browser-accessible URL for the avatar object when
// the storage configuration allows direct access (public CDN or public origin).
// Returns empty string when the browser must proxy through the backend API.
func avatarDirectURL(setting ossSettingValue, objectKey string) string {
	setting = normalizeOSSSetting(setting)
	if storage.CDNEnabled(setting) {
		url, err := ossCDNObjectURL(setting.CDNBaseURL, objectKey)
		if err == nil {
			return url
		}
	}
	if storage.PublicOrigin(setting) {
		url, err := signedOriginObjectURL(setting, objectKey, time.Now().Add(7*24*time.Hour))
		if err == nil {
			return url
		}
	}
	return ""
}

func (a avatarStorageAdapter) ReadAvatar(userID string, loc model.AvatarStorageLocation) ([]byte, error) {
	if loc.ObjectKey == "" {
		return nil, errors.New("头像存储路径为空")
	}
	if loc.Provider == "" || loc.Provider == "local" {
		return os.ReadFile(filepath.Join(a.svc.dataDir, "resources", filepath.FromSlash(loc.ObjectKey)))
	}
	resource := &model.Resource{
		UserID:           userID,
		Provider:         loc.Provider,
		Endpoint:         loc.Endpoint,
		Bucket:           loc.Bucket,
		StorageSettingID: loc.StorageSettingID,
		ObjectKey:        loc.ObjectKey,
	}
	setting, err := a.svc.ossSettingForResource(userID, resource)
	if err != nil {
		return nil, err
	}
	setting = a.svc.storageSettingWithRuntimePolicy(setting)
	stream, err := getOriginOSSObjectRange(setting, loc.ObjectKey, "")
	if err != nil {
		return nil, err
	}
	defer stream.Body.Close()
	return io.ReadAll(stream.Body)
}

func (a avatarStorageAdapter) DeleteAvatar(userID string, loc model.AvatarStorageLocation) error {
	if loc.ObjectKey == "" {
		return nil
	}
	if loc.Provider == "" || loc.Provider == "local" {
		return a.svc.deleteLocalResourceObject(loc.ObjectKey)
	}
	resource := &model.Resource{
		UserID:           userID,
		Provider:         loc.Provider,
		Endpoint:         loc.Endpoint,
		Bucket:           loc.Bucket,
		StorageSettingID: loc.StorageSettingID,
		ObjectKey:        loc.ObjectKey,
	}
	setting, err := a.svc.ossSettingForResource(userID, resource)
	if err != nil {
		return err
	}
	switch strings.ToLower(strings.TrimSpace(loc.Provider)) {
	case aliyunOSSProvider:
		return deleteAliyunOSSObject(setting, loc.ObjectKey)
	case tencentCOSProvider:
		return deleteTencentCOSObject(setting, loc.ObjectKey)
	case qiniuKodoProvider:
		return deleteQiniuObject(setting, loc.ObjectKey)
	case s3Provider:
		return deleteS3Object(setting, loc.ObjectKey)
	default:
		return errors.New("不支持的存储类型：" + loc.Provider)
	}
}

// activeAvatarOSSSetting returns the platform OSS setting if enabled, otherwise
// falls back to local storage. Avatars always follow the platform setting,
// not per-user storage configurations.
func (s *Service) activeAvatarOSSSetting() (ossSettingValue, string, bool, error) {
	_, systemValue, err := s.readOSSSetting()
	if err != nil {
		return ossSettingValue{}, "", false, err
	}
	if !systemValue.Enabled {
		return ossSettingValue{}, "", false, nil
	}
	systemValue, err = validateActiveOSSSetting(systemValue, "管理员尚未启用 OSS", "平台 OSS 配置不完整，请联系管理员")
	return systemValue, systemValue.StorageLocationID, true, err
}
