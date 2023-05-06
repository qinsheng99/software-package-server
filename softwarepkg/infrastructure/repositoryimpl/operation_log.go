package repositoryimpl

import (
	"github.com/opensourceways/software-package-server/common/infrastructure/postgresql"
	"github.com/opensourceways/software-package-server/softwarepkg/domain"
)

type operationLog struct {
	commentDBCli dbClient
}

func (t operationLog) AddOperationLog(v *domain.SoftwarePkgOperationLog) error {
	var do operationLogDO
	t.toOperationLogDO(v, &do)

	filter := operationLogDO{Id: do.Id}

	return t.commentDBCli.Insert(&filter, &do)
}

func (t operationLog) findOperationLogs(pid string) ([]domain.SoftwarePkgOperationLog, error) {
	var dos []operationLogDO

	err := t.commentDBCli.GetRecords(
		[]postgresql.ColumnFilter{
			postgresql.NewEqualFilter(fieldSoftwarePkgId, pid),
		},
		&dos,
		postgresql.Pagination{},
		[]postgresql.SortByColumn{
			{Column: fieldCreatedAt},
		},
	)
	if err != nil || len(dos) == 0 {
		return nil, err
	}

	v := make([]domain.SoftwarePkgOperationLog, len(dos))
	for i := range dos {
		if v[i], err = dos[i].toSoftwarePkgOperationLog(); err != nil {
			return nil, err
		}
	}

	return v, nil
}
