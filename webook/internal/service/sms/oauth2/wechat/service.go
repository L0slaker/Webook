package wechat

import (
	"Prove/webook/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	uuid "github.com/lithammer/shortuuid/v4"
	"net/http"
	"net/url"
)

// 提前注册好的域名
// PathEscape对字符串进行转义，以便将其安全地放置在URL路径段内
var redirectURI = url.PathEscape("https://meoying.com/oauth2/wechat/callback")

type Service interface {
	AuthURL(ctx context.Context) (string, error)
	VerifyCode(ctx context.Context, code, state string) (domain.WechatInfo, error)
}

type service struct {
	appId     string
	appSecret string
	client    *http.Client
}

func NewService(appId, appSecret string) Service {
	return &service{
		appId:     appId,
		appSecret: appSecret,
		client:    http.DefaultClient,
	}
}

func (s *service) AuthURL(ctx context.Context) (string, error) {
	// appId、redirect_uri、state 需要自己生成
	// scope：snsapi_login
	const urlPattern = "https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect"
	state := uuid.New()
	return fmt.Sprintf(urlPattern, s.appId, redirectURI, state), nil
}

func (s *service) VerifyCode(ctx context.Context, code, state string) (domain.WechatInfo, error) {
	// appId、secret、code 需要自己生成
	const targetPattern = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code"
	target := fmt.Sprintf(targetPattern, s.appId, s.appSecret, code)
	// 获取一个 http 请求
	//resp, err := http.Get(target)
	//req,err := http.NewRequest(http.MethodGet,target,nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	// 发送请求
	resp, err := s.client.Do(req)
	if err != nil {
		return domain.WechatInfo{}, err
	}

	var resBody Result
	err = json.NewDecoder(resp.Body).Decode(&resBody)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	if resBody.ErrCode != 0 {
		return domain.WechatInfo{},
			fmt.Errorf("微信返回错误响应，错误码：%d，错误信息：%s", resBody.ErrCode, resBody.ErrMsg)
	}

	return domain.WechatInfo{
		UnionId: resBody.UnionId,
		OpenId:  resBody.OpenId,
	}, nil
}

type Result struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`

	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`

	OpenId  string `json:"openid"`
	UnionId string `json:"unionid"`

	Scope string `json:"scope"`
}
