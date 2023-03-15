package translation

import "github.com/opensourceways/software-package-server/softwarepkg/domain/dp"

type Translation interface {
	Translate(string, dp.Language) (string, error)
	LanguageDetection(content string) (string, error)
}
