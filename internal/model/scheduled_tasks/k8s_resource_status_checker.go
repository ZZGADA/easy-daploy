package scheduled_tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ZZGADA/easy-deploy/internal/define"
	"github.com/ZZGADA/easy-deploy/internal/model/conf"
	"github.com/ZZGADA/easy-deploy/internal/model/dao"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8sResourceStatusChecker K8s 资源状态检查器
type K8sResourceStatusChecker struct {
	userK8sResourceDao             *dao.UserK8sResourceDao
	userK8sResourceOperationLogDao *dao.UserK8sResourceOperationLogDao
}

// K8sResourceInfo K8s资源信息结构
type K8sResourceInfo struct {
	ResourceID   uint   `json:"resource_id"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	Namespace    string `json:"namespace"`
	UserID       uint   `json:"user_id"`
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

	// 用于存储运行中的资源信息，按用户ID分组
	userResourcesMap := make(map[uint][]K8sResourceInfo)

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
					// 如果资源正在运行，添加到用户资源列表
					userID := uint(resource.UserID)
					resourceInfo := K8sResourceInfo{
						ResourceID:   uint(resource.Id),
						ResourceName: metadataName,
						ResourceType: resource.ResourceType,
						Namespace:    namespace,
						UserID:       userID,
					}
					userResourcesMap[userID] = append(userResourcesMap[userID], resourceInfo)
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
				// 如果资源正在运行，添加到用户资源列表
				userID := uint(resource.UserID)
				resourceInfo := K8sResourceInfo{
					ResourceID:   uint(resource.Id),
					ResourceName: metadataName,
					ResourceType: resource.ResourceType,
					Namespace:    namespace,
					UserID:       userID,
				}
				userResourcesMap[userID] = append(userResourcesMap[userID], resourceInfo)
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

	// 将每个用户的资源信息存入Redis并推送给对应的WebSocket客户端
	for userID, userResources := range userResourcesMap {
		if len(userResources) > 0 {
			// 将用户的资源信息序列化为JSON
			userResourcesJSON, err := json.Marshal(userResources)
			if err != nil {
				logrus.Errorf("序列化用户 %d 的运行中资源信息失败: %v", userID, err)
				continue
			}

			// 使用Redis存储用户的资源信息，键名格式为 k8s:running_resources:{user_id}
			redisKey := fmt.Sprintf(define.K8sRunningResources, userID)
			err = conf.RedisClient.Set(context.Background(), redisKey, userResourcesJSON, time.Hour).Err()
			if err != nil {
				logrus.Errorf("将用户 %d 的运行中资源信息存入Redis失败: %v", userID, err)
				continue
			}

			// 推送给该用户的WebSocket客户端
			c.pushResourceInfoToUser(userID, userResources, redisKey)
		}
	}
}

// pushResourceInfoToUser 将资源信息推送给指定用户的WebSocket客户端
func (c *K8sResourceStatusChecker) pushResourceInfoToUser(userID uint, resources []K8sResourceInfo, redisKey string) {
	// 获取用户的WebSocket连接
	conn, exists := conf.WSServer.Connections[userID]
	if !exists {
		logrus.Infof("用户 %d 没有活动的WebSocket连接", userID)
		return
	}

	// 构建推送消息
	message := map[string]interface{}{
		"type":      "resource_status",
		"redis_key": redisKey,
		"resources": resources,
		"timestamp": time.Now().Unix(),
	}

	response := map[string]interface{}{
		"success": true,
		"message": "resource_status_running",
		"data":    message,
	}

	// 发送消息
	err := conn.WriteJSON(response)
	if err != nil {
		logrus.Errorf("向用户 %d 推送资源状态消息失败: %v", userID, err)
	} else {
		logrus.Infof("成功向用户 %d 推送资源状态消息", userID)
	}
}

// PushRunningResource quick start one method
func PushRunningResource() {
	k8sResourceStatusChecker = NewK8sResourceStatusChecker(dao.NewUserK8sResourceDao(conf.DB), dao.NewUserK8sResourceOperationLogDao(conf.DB))
	k8sResourceStatusChecker.checkResources()
}
