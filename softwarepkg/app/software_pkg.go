package app

import (
	"errors"

	"github.com/sirupsen/logrus"

	commonrepo "github.com/opensourceways/software-package-server/common/domain/repository"
	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/maintainer"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/message"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/pkgmanager"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/repository"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/sensitivewords"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/service"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/translation"
)

type SoftwarePkgService interface {
	ApplyNewPkg(*CmdToApplyNewSoftwarePkg) (string, error)
	GetPkgReviewDetail(string) (SoftwarePkgReviewDTO, string, error)
	ListPkgs(*CmdToListPkgs) (SoftwarePkgsDTO, error)
	UpdateApplication(*CmdToUpdateSoftwarePkgApplication) (string, error)

	Approve(string, *domain.User) (string, error)
	Reject(string, *domain.User) (string, error)
	Abandon(string, *domain.User) (string, error)
	RerunCI(string, *domain.User) (string, error)
	NewReviewComment(string, *CmdToWriteSoftwarePkgReviewComment) (string, error)

	TranslateReviewComment(*CmdToTranslateReviewComment) (
		dto TranslatedReveiwCommentDTO, code string, err error,
	)
}

var (
	_ SoftwarePkgService = (*softwarePkgService)(nil)

	softwarePkgRobot = "software-pkg-robot"
)

func NewSoftwarePkgService(
	repo repository.SoftwarePkg,
	manager pkgmanager.PkgManager,
	message message.SoftwarePkgMessage,
	sensitive sensitivewords.SensitiveWords,
	maintainer maintainer.Maintainer,
	translation translation.Translation,
) *softwarePkgService {
	robot, _ := dp.NewAccount(softwarePkgRobot)

	return &softwarePkgService{
		repo:        repo,
		robot:       robot,
		message:     message,
		sensitive:   sensitive,
		maintainer:  maintainer,
		translation: translation,
		pkgService:  service.NewPkgService(manager, message),
	}
}

type softwarePkgService struct {
	repo        repository.SoftwarePkg
	robot       dp.Account
	message     message.SoftwarePkgMessage
	sensitive   sensitivewords.SensitiveWords
	maintainer  maintainer.Maintainer
	translation translation.Translation
	pkgService  service.SoftwarePkgService
}

func (s *softwarePkgService) ApplyNewPkg(cmd *CmdToApplyNewSoftwarePkg) (
	code string, err error,
) {
	v := domain.NewSoftwarePkg(&cmd.Importer, cmd.PkgName, &cmd.Application)
	if s.pkgService.IsPkgExisted(cmd.PkgName) {
		err = errors.New("software package already existed")
		code = errorSoftwarePkgExists

		return
	}

	if err = s.repo.AddSoftwarePkg(&v); err != nil {
		if commonrepo.IsErrorDuplicateCreating(err) {
			code = errorSoftwarePkgExists
		}
	} else {
		e := domain.NewSoftwarePkgAppliedEvent(&v)
		if err1 := s.message.NotifyPkgApplied(&e); err1 != nil {
			logrus.Errorf(
				"failed to notify a new applying pkg:%s, err:%s",
				v.Id, err1.Error(),
			)
		} else {
			logrus.Debugf(
				"successfully to notify a new applying pkg:%s", v.Id,
			)
		}
	}

	return
}

func (s *softwarePkgService) ListPkgs(cmd *CmdToListPkgs) (SoftwarePkgsDTO, error) {
	v, total, err := s.repo.FindSoftwarePkgs(*cmd)
	if err != nil || len(v) == 0 {
		return SoftwarePkgsDTO{}, nil
	}

	return toSoftwarePkgsDTO(v, total), nil
}

func (s *softwarePkgService) UpdateApplication(cmd *CmdToUpdateSoftwarePkgApplication) (string, error) {
	pkg, version, err := s.repo.FindSoftwarePkgBasicInfo(cmd.PkgId)
	if err != nil {
		return errorCodeForFindingPkg(err), err
	}

	if err = pkg.UpdateApplication(&cmd.Application, &cmd.Importer); err != nil {
		return domain.ParseErrorCode(err), err
	}

	if err = s.repo.SaveSoftwarePkg(&pkg, version); err == nil {
		s.addOperationLog(cmd.Importer.Account, dp.PackageOperationLogActionUpdate, cmd.PkgId)
	}

	return "", err
}

func (s *softwarePkgService) addOperationLog(
	user dp.Account, action dp.PackageOperationLogAction, pkgId string,
) {
	log := domain.NewSoftwarePkgOperationLog(user, action, pkgId)

	if err := s.repo.AddOperationLog(&log); err != nil {
		logrus.Errorf(
			"add operation log failed, log:%s, err:%s",
			log.String(), err.Error(),
		)
	}
}
