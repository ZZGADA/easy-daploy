package k8s_manage

import (
	"github.com/ZZGADA/easy-deploy/internal/model/dao"
)

// K8sResourceOperationLogService K8s 资源操作日志服务
type K8sResourceOperationLogService struct {
	userK8sResourceOperationLogDao *dao.UserK8sResourceOperationLogDao
}

// NewK8sResourceOperationLogService 创建 K8s 资源操作日志服务
func NewK8sResourceOperationLogService(userK8sResourceOperationLogDao *dao.UserK8sResourceOperationLogDao) *K8sResourceOperationLogService {
	return &K8sResourceOperationLogService{
		userK8sResourceOperationLogDao: userK8sResourceOperationLogDao,
	}
}

// QueryByK8sResourceID 根据 K8s 资源 ID 查询操作日志
func (s *K8sResourceOperationLogService) QueryByK8sResourceID(k8sResourceID uint) ([]*dao.UserK8sResourceOperationLog, error) {
	return s.userK8sResourceOperationLogDao.QueryByK8sResourceID(k8sResourceID)
}
