package utils

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"

	"github.com/ZZGADA/easy-deploy/internal/config"
)

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

// StartEmailSender 启动邮件发送goroutine
func StartEmailSender(alertChannel <-chan AlertMessage) {
	logrus.Info("邮件发送服务已启动")

	// 持续从通道中获取并发送邮件
	for alertMsg := range alertChannel {
		// 发送邮件
		if err := sendAlertEmail(alertMsg); err != nil {
			logrus.Errorf("发送告警邮件失败: %v", err)
			// 发送失败时，等待一段时间后重试
			time.Sleep(time.Second * 5)
			continue
		}

		logrus.Infof("告警邮件已发送至 %s", alertMsg.Email)
	}
}

// sendAlertEmail 发送告警邮件
func sendAlertEmail(alertMsg AlertMessage) error {
	// 创建邮件消息
	m := gomail.NewMessage()
	m.SetHeader("From", config.GlobalConfig.Smtp.From)
	m.SetHeader("To", alertMsg.Email)
	m.SetHeader("Subject", fmt.Sprintf("Kubernetes Pod告警: %s/%s",
		alertMsg.Namespace,
		alertMsg.PodName))

	// 构建邮件内容
	body := fmt.Sprintf(`
		<h2>Kubernetes Pod告警通知</h2>
		<p><strong>命名空间:</strong> %s</p>
		<p><strong>Pod名称:</strong> %s</p>
		<p><strong>告警级别:</strong> %s</p>
		<p><strong>告警时间:</strong> %s</p>
		<p><strong>告警内容:</strong></p>
		<pre>%s</pre>
	`,
		alertMsg.Namespace,
		alertMsg.PodName,
		alertMsg.LogLevel,
		alertMsg.Timestamp,
		alertMsg.LogMessage)

	m.SetBody("text/html", body)

	// 发送邮件
	d := gomail.NewDialer(
		config.GlobalConfig.Smtp.Host,
		config.GlobalConfig.Smtp.Port,
		config.GlobalConfig.Smtp.User,
		config.GlobalConfig.Smtp.Password,
	)

	// TODO: need revert
	logrus.Infof("totototototot", d.Host)
	//return d.DialAndSend(m)
	return nil
}
