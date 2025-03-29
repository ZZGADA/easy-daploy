package user_manage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ZZGADA/easy-deploy/internal/config"
	"github.com/ZZGADA/easy-deploy/internal/model/dao"
	"gorm.io/gorm"
)

type BindService struct {
	userGithubDao *dao.UserGithubDao
}

func NewBindService(userGithubDao *dao.UserGithubDao) *BindService {
	return &BindService{
		userGithubDao: userGithubDao,
	}
}

type GithubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// BindGithub 处理 GitHub 账号绑定
func (s *BindService) BindGithub(ctx context.Context, userID uint32, code string) (*dao.UserGithub, error) {
	// 1. 使用 code 获取 access token
	accessToken, err := s.getGithubAccessToken(code)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub access token: %v", err)
	}

	// 2. 使用 access token 获取用户信息
	githubUser, err := s.getGithubUserInfo(accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub user info: %v", err)
	}

	// 3. 保存用户 GitHub 信息到数据库
	userGithub := &dao.UserGithub{
		UserId:      userID,
		GithubId:    uint32(githubUser.ID),
		Login:       githubUser.Login,
		Name:        githubUser.Name,
		Email:       githubUser.Email,
		AvatarUrl:   githubUser.AvatarURL,
		AccessToken: accessToken,
	}

	errI := s.userGithubDao.Create(ctx, userGithub)
	if errI != nil {
		return nil, fmt.Errorf("failed to insert github user: %v", errI)
	}
	return userGithub, nil
}

// getGithubAccessToken 获取 GitHub access token
func (s *BindService) getGithubAccessToken(code string) (string, error) {
	clientID := config.GlobalConfig.Github.ClientID
	clientSecret := config.GlobalConfig.Github.ClientSecret

	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("GitHub OAuth 配置未设置")
	}

	url := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s",
		clientID, clientSecret, code)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.AccessToken, nil
}

// getGithubUserInfo 获取 GitHub 用户信息
func (s *BindService) getGithubUserInfo(accessToken string) (*GithubUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user GithubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// CheckGithubBinding 检查用户是否已绑定 GitHub
func (s *BindService) CheckGithubBinding(ctx context.Context, userID uint) (bool, *dao.UserGithub, error) {
	// 从数据库中查询用户的 GitHub 绑定信息
	userGithub, err := s.userGithubDao.GetByUserID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, userGithub, nil
		}
		return false, userGithub, fmt.Errorf("查询 GitHub 绑定状态失败: %w", err)
	}

	return userGithub.Id != 0, userGithub, nil
}

// UnbindGithub 解绑 GitHub 账号
func (s *BindService) UnbindGithub(ctx context.Context, userID uint) error {
	// 1. 获取用户的 GitHub 绑定信息
	userGithub, err := s.userGithubDao.GetByUserID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("用户未绑定 GitHub 账号")
		}
		return fmt.Errorf("获取 GitHub 绑定信息失败: %w", err)
	}

	// 2. 调用 GitHub API 撤销应用的访问权限（可选，因为用户可以直接在 GitHub 设置中撤销）
	if err := s.revokeGithubAccess(userGithub.AccessToken); err != nil {
		// 记录错误但继续执行，因为即使 GitHub API 调用失败，我们仍然要解除本地绑定
		log.Printf("撤销 GitHub 访问令牌失败: %v", err)
	}

	// 3. 软删除用户的 GitHub 绑定记录
	if err := s.userGithubDao.Delete(ctx, userID); err != nil {
		return fmt.Errorf("删除 GitHub 绑定记录失败: %w", err)
	}

	return nil
}

// revokeGithubAccess 撤销 GitHub 应用的访问权限
func (s *BindService) revokeGithubAccess(accessToken string) error {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")

	// GitHub API 文档：https://docs.github.com/en/rest/apps/oauth-applications#delete-an-app-authorization
	req, err := http.NewRequest("DELETE",
		fmt.Sprintf("https://api.github.com/applications/%s/token", clientID),
		strings.NewReader(fmt.Sprintf(`{"access_token":"%s"}`, accessToken)))
	if err != nil {
		return err
	}

	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("撤销访问令牌失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// DeveloperTokenRequest 开发者令牌请求
type DeveloperTokenRequest struct {
	Token      string    `json:"developer_token"`
	ExpireTime time.Time `json:"expire_time"`
	Comment    string    `json:"comment"`
}

// SaveDeveloperToken 保存开发者令牌
func (s *BindService) SaveDeveloperToken(ctx context.Context, userID uint, token, comment string, expireTime time.Time, repositoryName string) error {
	// 检查用户是否已绑定 GitHub
	userGithub, err := s.userGithubDao.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if userGithub == nil {
		return errors.New("用户未绑定 GitHub 账号")
	}

	// 更新开发者令牌信息
	userGithub.DeveloperToken = token
	userGithub.DeveloperTokenComment = comment
	userGithub.DeveloperTokenExpireTime = &expireTime
	userGithub.DeveloperRepositoryName = repositoryName

	// 保存到数据库
	return s.userGithubDao.Update(ctx, userGithub)
}

// UpdateDeveloperToken 更新开发者令牌
func (s *BindService) UpdateDeveloperToken(ctx context.Context, userID uint, token, comment string, expireTime time.Time, repositoryName string) error {
	// 检查用户是否已绑定 GitHub
	userGithub, err := s.userGithubDao.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if userGithub == nil {
		return errors.New("用户未绑定 GitHub 账号")
	}

	// 更新开发者令牌信息
	userGithub.DeveloperToken = token
	userGithub.DeveloperTokenComment = comment
	userGithub.DeveloperTokenExpireTime = &expireTime
	userGithub.DeveloperRepositoryName = repositoryName

	// 保存到数据库
	return s.userGithubDao.Update(ctx, userGithub)
}

// GetDeveloperToken 获取开发者令牌信息
func (s *BindService) GetDeveloperToken(ctx context.Context, userID uint) (*dao.UserGithub, error) {
	// 检查用户是否已绑定 GitHub
	userGithub, err := s.userGithubDao.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if userGithub == nil {
		return nil, nil
	}

	return userGithub, nil
}
