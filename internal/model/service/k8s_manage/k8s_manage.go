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
func (s *K8sResourceService) SaveResource(userID uint, repositoryID string, resourceType string, ossURL string) error {
	resource := &dao.UserK8sResource{
		UserID:       userID,
		RepositoryID: repositoryID,
		ResourceType: resourceType,
		OssURL:       ossURL,
	}
	return s.userK8sResourceDao.Create(resource)
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
