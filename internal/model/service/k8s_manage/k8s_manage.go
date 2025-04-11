package k8s_manage

import (
	"github.com/ZZGADA/easy-deploy/internal/model/dao"
)

type K8sResourceService struct {
	userK8sResourceDao *dao.UserK8sResourceDao
}

func NewK8sResourceService(userK8sResourceDao *dao.UserK8sResourceDao) *K8sResourceService {
	return &K8sResourceService{
		userK8sResourceDao: userK8sResourceDao,
	}
}

// SaveResource 保存 K8s 资源配置
func (s *K8sResourceService) SaveResource(userID uint, repositoryID string, resourceType string, ossURL string, fileName string) error {
	resource := &dao.UserK8sResource{
		UserID:       uint32(userID),
		RepositoryID: repositoryID,
		ResourceType: resourceType,
		OssURL:       ossURL,
		FileName:     fileName,
	}
	return s.userK8sResourceDao.Create(resource)
}

// UpdateResource 保存 K8s 资源配置
func (s *K8sResourceService) UpdateResource(userID uint, id uint32, repositoryID string, resourceType string, ossURL string, fileName string) error {
	resourceById, err := s.userK8sResourceDao.QueryById(id)
	if err != nil {
		return err
	}
	tx := s.userK8sResourceDao.BeginTx()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	resourceById.IsUpdate = true
	err = s.userK8sResourceDao.UpdateIsUpdateTx(tx, &resourceById)
	if err != nil {
		tx.Rollback()
		return err
	}

	resource := &dao.UserK8sResource{
		UserID:           uint32(userID),
		RepositoryID:     repositoryID,
		ResourceType:     resourceType,
		OssURL:           ossURL,
		FileName:         fileName,
		FatherResourceId: resourceById.Id,
	}

	err = s.userK8sResourceDao.CreateTx(tx, resource)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()

	return nil
}

// DeleteResource 删除 K8s 资源配置
func (s *K8sResourceService) DeleteResource(id uint) error {
	return s.userK8sResourceDao.Delete(id)
}

// QueryResources 查询 K8s 资源配置列表
func (s *K8sResourceService) QueryResources(repositoryID string, resourceType string) ([]dao.UserK8sResource, error) {

	if resourceType == "all" {
		return s.userK8sResourceDao.QueryByRepositoryALL(repositoryID)
	} else {
		return s.userK8sResourceDao.QueryByRepositoryAndType(repositoryID, resourceType)
	}

}

// ValidateResourceType 验证资源类型是否有效
func (s *K8sResourceService) ValidateResourceType(resourceType string) bool {
	validTypes := map[string]bool{
		"deployment": true,
		"service":    true,
		"ingress":    true,
	}
	return validTypes[resourceType]
}
