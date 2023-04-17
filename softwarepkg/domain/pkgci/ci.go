package pkgci

import "github.com/opensourceways/software-package-server/softwarepkg/domain"

type PkgCI interface {
	CreateCIPR(*domain.SoftwarePkgBasicInfo) error
}
