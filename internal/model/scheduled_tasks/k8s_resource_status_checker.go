package scheduled_tasks

import (
	"context"
	"fmt"
	"github.com/ZZGADA/easy-deploy/internal/define"
	"github.com/ZZGADA/easy-deploy/internal/model/conf"
	"github.com/ZZGADA/easy-deploy/internal/model/dao"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// K8sResourceStatusChecker K8s 资源状态检查器
type K8sResourceStatusChecker struct {
	userK8sResourceDao             *dao.UserK8sResourceDao
	userK8sResourceOperationLogDao *dao.UserK8sResourceOperationLogDao
}

// NewK8sResourceStatusChecker 创建 K8s 资源状态检查器
func NewK8sResourceStatusChecker(userK8sResourceDao *dao.UserK8sResourceDao, userK8sResourceOperationLogDao *dao.UserK8sResourceOperationLogDao) *K8sResourceStatusChecker {
	return &K8sResourceStatusChecker{
		userK8sResourceDao:             userK8sResourceDao,
		userK8sResourceOperationLogDao: userK8sResourceOperationLogDao,
	}
}

var k8sResourceStatusChecker *K8sResourceStatusChecker

func Init() {
	k8sResourceStatusChecker = NewK8sResourceStatusChecker(dao.NewUserK8sResourceDao(conf.DB), dao.NewUserK8sResourceOperationLogDao(conf.DB))
	k8sResourceStatusChecker.start()
}

// Start 启动定时任务
func (c *K8sResourceStatusChecker) start() {
	// 每分钟执行一次
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			c.checkResources()
		}
	}()
	logrus.Info("K8s 资源状态检查器已启动")
}

// checkResources 检查所有用户自定义资源的状态
func (c *K8sResourceStatusChecker) checkResources() {
	// 获取所有 K8s 资源
	resources, err := c.userK8sResourceDao.QueryAll()
	if err != nil {
		logrus.Errorf("查询 K8s 资源失败: %v", err)
		return
	}

	for _, resource := range resources {
		// 查询最新的操作日志，获取资源信息
		logs, err := c.userK8sResourceOperationLogDao.QueryByK8sResourceIDFirst(uint(resource.Id))
		if err != nil {
			logrus.Infof("查询资源 %d erros : %v", resource.Id, err)
			return
		}

		if len(logs) == 0 {
			logrus.Infof("查询资源 %d 的操作日志为空: %v，无需check", resource.Id, err)
			continue
		}

		logrus.Infof("查询资源 %d 的操作日志为: %v", resource.Id, logs)

		// 获取最新的操作日志
		latestLog := logs[0]
		if latestLog.OperationType == "delete" {
			return
		}

		namespace := latestLog.Namespace
		metadataName := latestLog.MetadataName

		// 检查资源状态
		var status int
		var command string

		if resource.ResourceType == "deployment" {
			// 检查部署状态
			deployment, err := conf.KubeClient.AppsV1().Deployments(namespace).Get(context.TODO(), metadataName, metav1.GetOptions{})
			if err != nil {
				// 资源不存在，状态为停止
				status = define.K8sResourceStatusStop
				command = fmt.Sprintf("kubectl get deployment %s -n %s", metadataName, namespace)
			} else {
				// 检查部署状态
				if deployment.Status.AvailableReplicas == *deployment.Spec.Replicas {
					status = define.K8sResourceStatusRun // 运行正常
				} else if deployment.Status.UpdatedReplicas < *deployment.Spec.Replicas {
					status = define.K8sResourceStatusRestart // 容器重启
				} else {
					status = define.K8sResourceStatusStop // 运行停止
				}
				command = fmt.Sprintf("kubectl get deployment %s -n %s", metadataName, namespace)
			}
		} else if resource.ResourceType == "service" {
			// 检查服务状态
			_, err := conf.KubeClient.CoreV1().Services(namespace).Get(context.TODO(), metadataName, metav1.GetOptions{})
			if err != nil {
				// 服务不存在，状态为停止
				status = define.K8sResourceStatusStop
				command = fmt.Sprintf("kubectl get service %s -n %s", metadataName, namespace)
			} else {
				// 服务存在，状态为正常
				status = define.K8sResourceStatusRun
				command = fmt.Sprintf("kubectl get service %s -n %s", metadataName, namespace)
			}
		} else {
			// 不支持的资源类型
			continue
		}

		// 记录操作日志
		operationLog := &dao.UserK8sResourceOperationLog{
			K8sResourceID:  uint(resource.Id),
			UserID:         uint(resource.UserID),
			Namespace:      namespace,
			MetadataName:   metadataName,
			MetadataLabels: latestLog.MetadataLabels,
			OperationType:  "check",
			Status:         status,
			Command:        command,
		}

		// 保存操作日志
		if err := c.userK8sResourceOperationLogDao.Create(operationLog); err != nil {
			logrus.Errorf("保存资源 %d 的操作日志失败: %v", resource.Id, err)
		}
	}
}
