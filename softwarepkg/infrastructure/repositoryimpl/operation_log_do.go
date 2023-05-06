package repositoryimpl

import (
	"github.com/google/uuid"

	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
)

type operationLogDO struct {
	// must set "uuid" as the name of column
	Id        uuid.UUID `gorm:"column:uuid;type:uuid"`
	User      string    `gorm:"column:user"`
	PkgId     string    `gorm:"column:software_pkg_id"`
	Action    string    `gorm:"column:action"`
	CreatedAt int64     `gorm:"column:created_at"`
}

func (t operationLog) toOperationLogDO(v *domain.SoftwarePkgOperationLog, do *operationLogDO) {
	*do = operationLogDO{
		Id:        uuid.New(),
		User:      v.User.Account(),
		PkgId:     v.PkgId,
		Action:    v.Action.PackageOperationLogAction(),
		CreatedAt: v.Time,
	}
}

func (do *operationLogDO) toSoftwarePkgOperationLog() (v domain.SoftwarePkgOperationLog, err error) {
	v.Id = do.Id.String()
	if v.User, err = dp.NewAccount(do.User); err != nil {
		return
	}

	v.Time = do.CreatedAt
	v.Action = dp.NewPackageOperationLogAction(do.Action)
	v.PkgId = do.PkgId

	return
}
