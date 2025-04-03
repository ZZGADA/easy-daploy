package oss_manage

import (
	"errors"
	"github.com/ZZGADA/easy-deploy/internal/model/dao"
	"gorm.io/gorm"
)

type OssService struct {
	userOssDao *dao.UserOssDao
}

func NewOssService(userOssDao *dao.UserOssDao) *OssService {
	return &OssService{
		userOssDao: userOssDao,
	}
}

// SaveOssAccess 保存 OSS 访问信息
func (s *OssService) SaveOssAccess(userID uint, accessKeyID, accessKeySecret, bucket, region string) error {
	userOss, err := s.userOssDao.QueryByUserID(userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if userOss != nil {
		return errors.New("oss account已经存在，请勿重复保存")
	}

	oss := &dao.UserOss{
		UserID:          uint32(userID),
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
		Bucket:          bucket,
		Region:          region,
	}
	return s.userOssDao.Create(oss)
}

// UpdateOssAccess 更新 OSS 访问信息
func (s *OssService) UpdateOssAccess(userID uint, accessKeyID, accessKeySecret, bucket, region string) error {
	return s.userOssDao.Update(userID, accessKeyID, accessKeySecret, bucket, region)
}

// QueryOssAccess 查询 OSS 访问信息
func (s *OssService) QueryOssAccess(userID uint) (*dao.UserOss, error) {
	return s.userOssDao.QueryByUserID(userID)
}

// DeleteOssAccess 删除 OSS 访问信息
func (s *OssService) DeleteOssAccess(userID uint) error {
	return s.userOssDao.DeleteByUserID(userID)
}
