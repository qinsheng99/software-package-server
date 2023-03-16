package middleware

import (
	"errors"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/opensourceways/server-common-lib/utils"
	"golang.org/x/time/rate"

	commonstl "github.com/opensourceways/software-package-server/common/controller"
	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
)

const (
	keyCookie                  = "_Y_G_"
	keyUserInfo                = "user_info"
	headerPrivateToken         = "PRIVATE-TOKEN"
	errorBadRequestNoHeader    = "bad_request_no_token"
	errorBadRequestNoCookie    = "bad_request_no_cookie"
	errorBadRequestHaventLogin = "bad_request_havent_login"
)

var instance *userCheckingMiddleware

type Config struct {
	UserInfoURL string `json:"user_info_url" required:"true"`
	LimiterConfig
}

func Init(cfg *Config) {
	instance = &userCheckingMiddleware{
		client:      utils.NewHttpClient(3),
		userInfoURL: cfg.UserInfoURL,
	}

	limiterInstance = &ipRateLimiter{
		ips:   make(map[string]*rate.Limiter),
		mu:    &sync.RWMutex{},
		limit: rate.Limit(cfg.Limit),
		burst: cfg.Burst,
	}
}

func UserChecking() *userCheckingMiddleware {
	return instance
}

type userCheckingMiddleware struct {
	userInfoURL string
	client      utils.HttpClient
}

func (m *userCheckingMiddleware) FetchUser(ctx *gin.Context) (domain.User, error) {
	if v, exists := ctx.Get(keyUserInfo); exists {
		if u, ok := v.(domain.User); ok {
			return u, nil
		}
	}

	return domain.User{}, errors.New("no user info")
}

func (m *userCheckingMiddleware) CheckUser(ctx *gin.Context) {
	if code, err := m.doCheck(ctx); err != nil {
		commonstl.SendFailedResp(ctx, code, err)

		ctx.Abort()
	} else {
		ctx.Next()
	}
}

func (m *userCheckingMiddleware) doCheck(ctx *gin.Context) (string, error) {
	t := m.token(ctx)
	if t == "" {
		return errorBadRequestNoHeader, errors.New("no token")
	}

	c := m.cookie(ctx)
	if c == "" {
		return errorBadRequestNoCookie, errors.New("no cookie")
	}

	v, code, err := m.getUserInfo(t, c)
	if err == nil {
		ctx.Set(keyUserInfo, v)
	}

	return code, err
}

func (m *userCheckingMiddleware) token(ctx *gin.Context) string {
	return ctx.GetHeader(headerPrivateToken)
}

func (m *userCheckingMiddleware) cookie(ctx *gin.Context) string {
	c, _ := ctx.Cookie(keyCookie)

	return c
}

func (m *userCheckingMiddleware) cookieHeader(cookie string) string {
	return keyCookie + "=" + cookie
}

func (m *userCheckingMiddleware) getUserInfo(token, cookie string) (
	r domain.User, code string, err error,
) {
	req, err := http.NewRequest(http.MethodGet, m.userInfoURL, nil)
	if err != nil {
		return
	}

	req.Header.Set("token", token)
	req.Header.Set("Cookie", m.cookieHeader(cookie))

	var result struct {
		Data userInfoData `json:"data"`
	}

	status, err := m.client.ForwardTo(req, &result)
	if err != nil {
		if status == http.StatusUnauthorized {
			code = errorBadRequestHaventLogin
			err = errors.New("no login")
		}
	} else {
		r, err = result.Data.toUser()
	}

	return
}

type userInfoData struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

func (d *userInfoData) toUser() (v domain.User, err error) {
	if v.Account, err = dp.NewAccount(d.Username); err != nil {
		return
	}

	v.Email, err = dp.NewEmail(d.Email)

	return
}
