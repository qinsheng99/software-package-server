package translationimpl

import (
	"errors"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/region"
	v2 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/nlp/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/nlp/v2/model"

	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
)

var instance *service

func Init(cfg *Config, languages []string) error {
	sl, err := getSupportedLanguage(languages)
	if err != nil {
		return err
	}

	auth := basic.NewCredentialsBuilder().
		WithAk(cfg.AccessKey).
		WithSk(cfg.SecretKey).
		WithProjectId(cfg.Project).
		Build()

	client := v2.NewNlpClient(core.NewHcHttpClientBuilder().
		WithCredential(auth).
		WithRegion(region.NewRegion(cfg.Region, cfg.AuthEndpoint)).
		Build())

	instance = &service{
		cli:                client,
		supportedLanguages: sl,
	}

	return nil
}

func Translation() *service {
	return instance
}

type service struct {
	cli                *v2.NlpClient
	supportedLanguages map[string]model.TextTranslationReqTo
}

func (s *service) Translate(content string, l dp.Language) (string, error) {
	to, ok := s.supportedLanguages[l.Language()]
	if !ok {
		return "", errors.New("unsupported language")
	}

	req := model.TextTranslationReq{
		Text: content,
		From: model.GetTextTranslationReqFromEnum().AUTO,
		To:   to,
	}

	v, err := s.cli.RunTextTranslation(
		&model.RunTextTranslationRequest{Body: &req},
	)
	if err != nil {
		return "", err
	}

	if err = s.checkErrorMag(v.ErrorMsg); err != nil {
		return "", err
	}

	if v.TranslatedText != nil {
		return *v.TranslatedText, nil
	}

	return "", errors.New("no translated text")
}

func (s *service) checkErrorMag(msg *string) (err error) {
	if msg != nil {
		err = errors.New(*msg)
	}

	return
}

func (s *service) LanguageDetection(content string) (string, error) {
	req := model.LanguageDetectionReq{Text: content}

	v, err := s.cli.RunLanguageDetection(&model.RunLanguageDetectionRequest{Body: &req})
	if err != nil {
		return "", err
	}

	if err = s.checkErrorMag(v.ErrorMsg); err != nil {
		return "", err
	}

	if v.DetectedLanguage != nil {
		return *v.DetectedLanguage, nil
	}

	return "", errors.New("no match language type")
}
