package domain

import (
	"fmt"

	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
	"github.com/opensourceways/software-package-server/utils"
)

// SoftwarePkgOperationLog
type SoftwarePkgOperationLog struct {
	Id     string
	PkgId  string
	Time   int64
	User   dp.Account
	Action dp.PackageOperationLogAction
}

func (log *SoftwarePkgOperationLog) String() string {
	return fmt.Sprintf(
		"%s %s %s at %s",
		log.User.Account(),
		log.Action.PackageOperationLogAction(),
		log.PkgId,
		utils.ToDateTime(log.Time),
	)
}

func NewSoftwarePkgOperationLog(
	user dp.Account, action dp.PackageOperationLogAction, pkgId string,
) SoftwarePkgOperationLog {
	return SoftwarePkgOperationLog{
		PkgId:  pkgId,
		Time:   utils.Now(),
		User:   user,
		Action: action,
	}
}
