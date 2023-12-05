package week_05

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

/*
实现微信登陆

1.在微信开放平台创建应用并获取AppID和AppSecret
2.构建扫描登陆的URL，参数：
	·AppID：应用的唯一标识
	·RedirectURI：用户授权后重定向的回调URL
	·ResponseType：授权类型，固定为code
	·范围：请求授权的作用范围，一般使用snsapi_login
	·State：用于保持请求和回调的状态，可自定义
3.用户扫码并授权后，将重定向到提供的回调URL
4.在回调处理函数中，获取code函数，然后构建获取访问令牌的URL
5.发送HTTP请求到获取访问令牌的URL，参数：
	·AppID：应用的唯一标识
	·AppSecret：应用的密钥
	·Code：授权码，即上一步获取的code
	·GrantType：授权类型，固定为authorization_code
6.解析响应，获取访问令牌和OpenID
7.使用访问令牌和OpenID，调用微信API获取用户信息
*/

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	OpenID      string `json:"openid"`
}

type UserInfoResponse struct {
	OpenID     string `json:"openid"`
	Nickname   string `json:"nickname"`
	HeadImgURL string `json:"headimgurl"`
}

const (
	AppID       = "YOUR_APP_ID"
	AppSecret   = "YOUR_APP_SECRET"
	RedirectURI = "YOUR_REDIRECT_URI"
)

func main() {
	//1.构建扫码登陆的URL
	authURL := fmt.Sprintf("https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=STATE#wechat_redirect",
		AppID, url.QueryEscape(RedirectURI))
	fmt.Println("请访问以下URL进行扫码登陆")
	fmt.Println(authURL)

	//2.在回调处理函数中获取code和state
	// ...

	//3.构建获取访问令牌的URL
	tokenURL := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		AppID, AppSecret, "YOUR_CODE")

	//4.发送HTTP请求获取访问令牌
	resp, err := http.Get(tokenURL)
	if err != nil {
		fmt.Println("获取访问令牌失败:", err)
		return
	}
	defer resp.Body.Close()

	//5.解析访问令牌响应
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取响应失败:", err)
		return
	}

	var infoResp UserInfoResponse
	err = json.Unmarshal(data, &infoResp)
	if err != nil {
		fmt.Println("解析响应失败:", err)
		return
	}

	nickname := infoResp.Nickname
	headimgURL := infoResp.HeadImgURL

	fmt.Println("用户信息:")
	fmt.Println("昵称:", nickname)
	fmt.Println("头像URL:", headimgURL)
}
