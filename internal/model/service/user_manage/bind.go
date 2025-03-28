package user_manage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ZZGADA/easy-deploy/internal/config"
	"github.com/ZZGADA/easy-deploy/internal/model/dao"
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
func (s *BindService) BindGithub(ctx context.Context, userID uint, code string) error {
	// 1. 使用 code 获取 access token
	accessToken, err := s.getGithubAccessToken(code)
	if err != nil {
		return fmt.Errorf("failed to get GitHub access token: %v", err)
	}

	// 2. 使用 access token 获取用户信息
	githubUser, err := s.getGithubUserInfo(accessToken)
	if err != nil {
		return fmt.Errorf("failed to get GitHub user info: %v", err)
	}

	// 3. 保存用户 GitHub 信息到数据库
	userGithub := &dao.UserGithub{
		UserID:      userID,
		GithubID:    uint(githubUser.ID),
		Login:       githubUser.Login,
		Name:        githubUser.Name,
		Email:       githubUser.Email,
		AvatarURL:   githubUser.AvatarURL,
		AccessToken: accessToken,
	}

	return s.userGithubDao.Create(ctx, userGithub)
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
