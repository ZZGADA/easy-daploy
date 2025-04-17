package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ZZGADA/easy-deploy/internal/config"
	"github.com/ZZGADA/easy-deploy/internal/model/conf"
	"github.com/ZZGADA/easy-deploy/internal/model/dao"
	"github.com/ZZGADA/easy-deploy/internal/utils"
)

// LogMessage 日志消息结构
type LogMessage struct {
	Log    string `json:"log"`
	Stream string `json:"stream"`
	Docker struct {
		ContainerID string `json:"container_id"`
	} `json:"docker"`
	Kubernetes struct {
		ContainerName    string            `json:"container_name"`
		NamespaceName    string            `json:"namespace_name"`
		PodName          string            `json:"pod_name"`
		ContainerImage   string            `json:"container_image"`
		ContainerImageID string            `json:"container_image_id"`
		PodID            string            `json:"pod_id"`
		PodIP            string            `json:"pod_ip"`
		Host             string            `json:"host"`
		Labels           map[string]string `json:"labels"`
		MasterURL        string            `json:"master_url"`
		NamespaceID      string            `json:"namespace_id"`
		NamespaceLabels  map[string]string `json:"namespace_labels"`
	} `json:"kubernetes"`
	Timestamp string `json:"timestamp"`
}

// AlertMessage 告警消息结构
type AlertMessage struct {
	UserID     uint   `json:"user_id"`
	Email      string `json:"email"`
	PodName    string `json:"pod_name"`
	Namespace  string `json:"namespace"`
	LogLevel   string `json:"log_level"`
	LogMessage string `json:"log_message"`
	Timestamp  string `json:"timestamp"`
}

// 消息处理队列
var messageQueue = make(chan LogMessage, 1000)

// 告警消息通道
var alertChannel = make(chan utils.AlertMessage, 1000)
var operationLogDao *dao.UserK8sResourceOperationLogDao

var teamDao *dao.TeamDao

// StartConsumer 启动Kafka消费者
func StartConsumer() {
	// 创建Kafka消费者
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  config.GlobalConfig.Kafka.Brokers,
		Topic:    config.GlobalConfig.Kafka.Topic,
		GroupID:  config.GlobalConfig.Kafka.GroupId,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		MaxWait:  time.Second * 3,
	})

	defer reader.Close()

	logrus.Info("Kafka消费者已启动，正在监听日志消息...")

	// 启动邮件发送goroutine
	go utils.StartEmailSender(alertChannel)

	// 启动worker池
	startWorkerPool(50)

	operationLogDao = dao.NewUserK8sResourceOperationLogDao(conf.DB)

	// 持续消费消息
	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			logrus.Errorf("读取Kafka消息失败: %v", err)
			continue
		}

		// 将消息发送到处理队列
		var logMsg LogMessage
		if err := json.Unmarshal(msg.Value, &logMsg); err != nil {
			logrus.Errorf("解析日志消息失败: %v", err)
			continue
		}

		// 将消息放入处理队列
		messageQueue <- logMsg
	}
}

// 启动worker池
func startWorkerPool(workerCount int) {
	var wg sync.WaitGroup
	wg.Add(workerCount)
	teamDao = dao.NewTeamDao(conf.DB)

	for i := 0; i < workerCount; i++ {
		go func(workerID int) {
			defer wg.Done()
			logrus.Infof("Worker %d 已启动", workerID)

			for msg := range messageQueue {
				processMessage(msg)
			}
		}(i)
	}
}

// processMessage 处理单条消息
func processMessage(logMsg LogMessage) {
	// 检查是否为系统命名空间
	namespaceName := logMsg.Kubernetes.NamespaceName
	if namespaceName == "kube-system" || namespaceName == "kube-public" || namespaceName == "kube-node-lease" {
		return // 忽略系统命名空间的日志
	}

	// 检查日志级别
	logContent := logMsg.Log
	logLevel := ""
	if strings.Contains(strings.ToLower(logContent), "warn") {
		logLevel = "WARN"
	} else if strings.Contains(strings.ToLower(logContent), "error") {
		logLevel = "ERROR"
	} else if strings.Contains(strings.ToLower(logContent), "panic") {
		logLevel = "PANIC"
	} else if strings.Contains(strings.ToLower(logContent), "fatal") {
		logLevel = "FATAL"
	} else if strings.Contains(strings.ToLower(logContent), "warning") {
		logLevel = "WARN"
	}

	// 如果日志级别为空，则不处理
	if logLevel == "" {
		return
	}

	// 获取Pod创建者信息
	userID, err := getPodCreator(logMsg.Kubernetes.PodName, logMsg.Kubernetes.NamespaceName)
	if err != nil {
		logrus.Errorf("获取Pod创建者失败: %v", err)
		return
	}

	// 获取用户邮箱
	user, err := dao.GetUserByID(userID)
	if err != nil {
		logrus.Errorf("获取用户信息失败: %v", err)
		return
	}

	// 创建告警消息
	alertMsg := utils.AlertMessage{
		UserID:     userID,
		Email:      user.Email,
		PodName:    logMsg.Kubernetes.PodName,
		Namespace:  logMsg.Kubernetes.NamespaceName,
		LogLevel:   logLevel,
		LogMessage: logContent,
		Timestamp:  logMsg.Timestamp,
	}

	// 将告警消息发送到告警通道
	alertChannel <- alertMsg

	ctx := context.Background()
	team, err := teamDao.GetByID(ctx, user.TeamID)
	if err != nil {
		logrus.Errorf("消费者中获取团队信息失败,%v", err)
	}

	teamCreator, err := dao.GetUserByID(uint(team.CreatorID))
	if err != nil {
		logrus.Errorf("获取团队管理员信息失败: %v", err)
		return
	}

	// 创建告警消息
	alertMsgCreator := utils.AlertMessage{
		UserID:     uint(teamCreator.Id),
		Email:      teamCreator.Email,
		PodName:    logMsg.Kubernetes.PodName,
		Namespace:  logMsg.Kubernetes.NamespaceName,
		LogLevel:   logLevel,
		LogMessage: logContent,
		Timestamp:  logMsg.Timestamp,
	}

	// 将告警消息发送到告警通道
	alertChannel <- alertMsgCreator
}

// getPodCreator 获取Pod创建者信息
func getPodCreator(podName, namespace string) (uint, error) {
	// 1. 获取Pod的元数据名称
	pod, err := conf.KubeClient.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return 0, fmt.Errorf("获取Pod信息失败: %v", err)
	}

	logrus.Infof("kafka消费者, getPodCreator, 获取Pod信息成功, pod: %s", pod.ObjectMeta.Name)

	// 2. 获取Pod所属的控制器名称
	controllerName := getControllerName(pod)
	if controllerName == "" {
		return 0, fmt.Errorf("无法确定Pod所属的控制器")
	}

	logrus.Infof("kafka消费者, getPodCreator, Pod所属控制器: %s", controllerName)

	// 3. 从操作日志中查找创建者
	logs, err := operationLogDao.QueryByNamespaceAndMetadataName(namespace, controllerName)
	if err != nil {
		return 0, fmt.Errorf("查询操作日志失败: %v", err)
	}

	if len(logs) == 0 {
		return 0, fmt.Errorf("未找到Pod创建者信息")
	}

	// 返回创建者ID
	return logs[0].UserID, nil
}

// getControllerName 获取Pod所属的控制器名称
func getControllerName(pod *corev1.Pod) string {
	// 检查Pod的OwnerReferences
	if len(pod.OwnerReferences) > 0 {
		owner := pod.OwnerReferences[0]

		// 根据控制器类型返回不同的名称
		switch owner.Kind {
		case "ReplicaSet":
			// 如果是ReplicaSet，需要进一步查找其所属的Deployment
			rs, err := conf.KubeClient.AppsV1().ReplicaSets(pod.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
			if err != nil {
				logrus.Errorf("获取ReplicaSet失败: %v", err)
				return owner.Name
			}

			// 查找ReplicaSet所属的Deployment
			if len(rs.OwnerReferences) > 0 && rs.OwnerReferences[0].Kind == "Deployment" {
				return rs.OwnerReferences[0].Name
			}
			return owner.Name

		case "StatefulSet":
			return owner.Name

		case "DaemonSet":
			return owner.Name

		case "Job":
			return owner.Name

		case "CronJob":
			return owner.Name

		default:
			// 对于其他类型的控制器，直接返回名称
			return owner.Name
		}
	}

	// 如果没有OwnerReferences，则使用Pod的名称
	return pod.Name
}
