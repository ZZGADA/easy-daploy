package utils

import (
	"fmt"
	"github.com/ZZGADA/easy-deploy/internal/define"
	"github.com/ZZGADA/easy-deploy/internal/model/dao"
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

// SendJoinTeamEmail 发送加入团队申请邮件
func SendJoinTeamEmail(applicant *dao.UserWithGithubInfo, teamCreatorEmail string, requestId uint32, requestType int) error {
	subject := "新的团队申请"
	// 假设后端接口的 URL
	//field := fmt.Sprintf("http://%s:%s/api/team/request/check", config.GlobalConfig.Server.Host, config.GlobalConfig.Server.Port)
	//acceptURL := fmt.Sprintf("%s?request_id=%d&status=1", field, requestId)
	//rejectURL := fmt.Sprintf("%s?request_id=%d&status=1", field, requestId)
	//
	/*
	   <p>
	       <a href="%s" style="display: inline-block; background-color: #4CAF50; color: white; padding: 10px 20px; text-align: center; text-decoration: none; border-radius: 4px;">接受</a>
	       <a href="%s" style="display: inline-block; background-color: #f44336; color: white; padding: 10px 20px; text-align: center; text-decoration: none; border-radius: 4px;">拒绝</a>
	   </p>*/

	mention := ""
	if requestType == define.TeamRequestTypeIn {
		mention = "<h2>您好，有新的用户申请 加入 您的团队</h2>"
	} else {
		mention = "<h2>您好，团队成员申请 离开 您的团队</h2>"
	}
	body := fmt.Sprintf(`
    %s
    <p>申请者信息如下：</p>
    <p><strong>用户 ID:</strong> %d</p>
    <p><strong>用户邮箱:</strong> %s</p>
    <p><strong>GitHub ID:</strong> %d</p>
    <p><strong>GitHub 名称:</strong> %s</p>
	<p><strong>请前往团队管理页面审批</strong></p>`, mention, applicant.ID, applicant.Email, applicant.GithubID, applicant.Name)

	m := gomail.NewMessage()
	m.SetHeader("From", config.GlobalConfig.Smtp.From)
	m.SetHeader("To", teamCreatorEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(
		config.GlobalConfig.Smtp.Host,
		config.GlobalConfig.Smtp.Port,
		config.GlobalConfig.Smtp.User,
		config.GlobalConfig.Smtp.Password,
	)

	return d.DialAndSend(m)
}

// SendTeamRequestEmail 发送审批结果邮件
func SendTeamRequestEmail(team *dao.Team, requesterEmail string, requestType int, status int) error {
	subjectFormat := "团队%s申请已完成"
	subject := ""

	mentionFormat := "<h2>您好，您申请 %s 团队%s的审批已经%s</h2>"
	mention := ""
	if requestType == define.TeamRequestTypeIn {
		subject = fmt.Sprintf(subjectFormat, "加入")
		switch status {
		case define.TeamRequestStatusApproval:
			mention = fmt.Sprintf(mentionFormat, team.TeamName, "加入", "通过")
		case define.TeamRequestStatusReject:
			mention = fmt.Sprintf(mentionFormat, team.TeamName, "加入", "拒绝")
		}
	} else {
		subject = fmt.Sprintf(subjectFormat, "离开")
		switch status {
		case define.TeamRequestStatusApproval:
			mention = fmt.Sprintf(mentionFormat, team.TeamName, "离开", "通过")
		case define.TeamRequestStatusReject:
			mention = fmt.Sprintf(mentionFormat, team.TeamName, "离开", "拒绝")
		}
	}
	body := fmt.Sprintf(`
    %s
    <p>申请者信息如下：</p>
    <p><strong>团队 ID:</strong> %d</p>
	<p><strong>团队 UUID:</strong> %d</p>
	<p><strong>团队 名称:</strong> %s</p>
    <p><strong>团队 简介:</strong> %s</p>
	<p><strong>请前往团队管理页面查看</strong></p>`, mention, team.ID, team.TeamUUID, team.TeamName, team.TeamDescription)

	m := gomail.NewMessage()
	m.SetHeader("From", config.GlobalConfig.Smtp.From)
	m.SetHeader("To", requesterEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(
		config.GlobalConfig.Smtp.Host,
		config.GlobalConfig.Smtp.Port,
		config.GlobalConfig.Smtp.User,
		config.GlobalConfig.Smtp.Password,
	)

	return d.DialAndSend(m)
}
