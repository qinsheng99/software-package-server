package app

import (
	"fmt"

	"github.com/sirupsen/logrus"

	commonrepo "github.com/opensourceways/software-package-server/common/domain/repository"
	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/message"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/pkgci"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/pkgmanager"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/repository"
)

type SoftwarePkgMessageService interface {
	HandlePkgCIChecking(CmdToHandlePkgCIChecking) error
	HandlePkgCIChecked(CmdToHandlePkgCIChecked) error
	HandlePkgCodeSaved(CmdToHandlePkgCodeSaved) error
	HandlePkgInitialized(CmdToHandlePkgInitialized) error
	HandlePkgRepoCreated(CmdToHandlePkgRepoCreated) error
	HandlePkgAlreadyExisted(CmdToHandlePkgAlreadyExisted) error
}

func NewSoftwarePkgMessageService(
	ci pkgci.PkgCI,
	repo repository.SoftwarePkg,
	manager pkgmanager.PkgManager,
	message message.SoftwarePkgIndirectMessage,
) softwarePkgMessageService {
	robot, _ := dp.NewAccount(softwarePkgRobot)

	return softwarePkgMessageService{
		ci:      ci,
		repo:    repo,
		robot:   robot,
		manager: manager,
		message: message,
	}
}

type softwarePkgMessageService struct {
	ci      pkgci.PkgCI
	repo    repository.SoftwarePkg
	robot   dp.Account
	manager pkgmanager.PkgManager
	message message.SoftwarePkgIndirectMessage
}

// HandlePkgCIChecking
func (s softwarePkgMessageService) HandlePkgCIChecking(cmd CmdToHandlePkgCIChecking) error {
	pkg, version, err := s.repo.FindSoftwarePkgBasicInfo(cmd.PkgId)
	if err != nil {
		return err
	}

	if err = pkg.HandleCIChecking(); err != nil {
		return err
	}

	if err = s.ci.CreateCIPR(&pkg); err != nil {
		return err
	}

	if err = s.repo.SaveSoftwarePkg(&pkg, version); err != nil {
		logrus.Errorf(
			"save pkg failed when %s, err:%s",
			cmd.logString(), err.Error(),
		)
	}

	return nil
}

// HandlePkgCIChecked
func (s softwarePkgMessageService) HandlePkgCIChecked(cmd CmdToHandlePkgCIChecked) error {
	pkg, version, err := s.repo.FindSoftwarePkgBasicInfo(cmd.PkgId)
	if err != nil {
		return err
	}

	if err := pkg.HandleCIChecked(cmd.Success); err != nil {
		return err
	}

	s.addCIComment(&cmd)

	if err := s.repo.SaveSoftwarePkg(&pkg, version); err != nil {
		logrus.Errorf(
			"save pkg failed when %s, err:%s",
			cmd.logString(), err.Error(),
		)
	}

	return nil
}

func (s softwarePkgMessageService) addCIComment(cmd *CmdToHandlePkgCIChecked) {
	content, _ := dp.NewReviewComment(cmd.Detail)
	comment := domain.NewSoftwarePkgReviewComment(s.robot, content)

	if err := s.repo.AddReviewComment(cmd.PkgId, &comment); err != nil {
		logrus.Errorf(
			"failed to add a comment when %s, err:%s",
			cmd.logString(), err.Error(),
		)
	}
}

// HandlePkgRepoCreated
func (s softwarePkgMessageService) HandlePkgRepoCreated(cmd CmdToHandlePkgRepoCreated) error {
	if !cmd.isSuccess() {
		logrus.Errorf(
			"failed to create repo on platform:%s for pkg:%s, err:%s",
			cmd.Platform.PackagePlatform(), cmd.PkgId, cmd.FiledReason,
		)

		return nil
	}

	pkg, version, err := s.repo.FindSoftwarePkgBasicInfo(cmd.PkgId)
	if err != nil {
		return err
	}

	if err := pkg.HandleRepoCreated(cmd.RepoCreatedInfo); err != nil {
		return err
	}

	if err := s.repo.SaveSoftwarePkg(&pkg, version); err != nil {
		logrus.Errorf(
			"save pkg failed when %s, err:%s",
			cmd.logString(), err.Error(),
		)
	}

	return nil
}

// HandlePkgCodeSaved
func (s softwarePkgMessageService) HandlePkgCodeSaved(cmd CmdToHandlePkgCodeSaved) error {
	if !cmd.isSuccess() {
		logrus.Errorf(
			"failed to create repo on platform:%s for pkg:%s, err:%s",
			cmd.Platform.PackagePlatform(), cmd.PkgId, cmd.FiledReason,
		)

		return nil
	}

	pkg, version, err := s.repo.FindSoftwarePkgBasicInfo(cmd.PkgId)
	if err != nil {
		return err
	}

	if err := pkg.HandleCodeSaved(cmd.RepoCreatedInfo); err != nil {
		return err
	}

	if err := s.repo.SaveSoftwarePkg(&pkg, version); err != nil {
		logrus.Errorf(
			"save pkg failed when %s, err:%s",
			cmd.logString(), err.Error(),
		)
	}

	return nil
}

// HandlePkgInitialized
func (s softwarePkgMessageService) HandlePkgInitialized(cmd CmdToHandlePkgInitialized) error {
	pkg, version, err := s.repo.FindSoftwarePkgBasicInfo(cmd.PkgId)
	if err != nil {
		return err
	}

	if cmd.isSuccess() {
		if err := pkg.HandlePkgInitialized(cmd.RelevantPR); err != nil {
			return err
		}

		if !pkg.Application.PackagePlatform.IsLocalPlatform() {
			s.notifyPkgInitialized(&pkg, &cmd)
		}
	} else {
		if !cmd.isPkgAreadyExisted() {
			logrus.Errorf("pkg init failed, pkgid:%s, err:%s", cmd.PkgId, cmd.FiledReason)

			return nil
		}

		if err := pkg.HandlePkgAlreadyExisted(); err != nil {
			return err
		}

		s.addCommentForExistedPkg(&cmd)
	}

	if err := s.repo.SaveSoftwarePkg(&pkg, version); err != nil {
		logrus.Errorf(
			"save pkg failed when %s, err:%s",
			cmd.logString(), err.Error(),
		)
	}

	return nil
}

func (s softwarePkgMessageService) notifyPkgInitialized(
	pkg *domain.SoftwarePkgBasicInfo, cmd *CmdToHandlePkgInitialized,
) {
	e := domain.NewSoftwarePkgInitializedEvent(pkg)

	if err := s.message.NotifyPkgIndirectlyApproved(&e); err != nil {
		logrus.Errorf(
			"failed to notify the pkg was approved indirectly when %s, err:%s",
			cmd.logString(), err.Error(),
		)
	}
}

func (s softwarePkgMessageService) addCommentForExistedPkg(cmd *CmdToHandlePkgInitialized) {
	str := fmt.Sprintf(
		"I'am sorry to close this application. Because the pkg was imported sometimes ago. The repo address is %s. You can work on that repo.",
		cmd.RepoLink,
	)
	content, _ := dp.NewReviewComment(str)
	comment := domain.NewSoftwarePkgReviewComment(s.robot, content)

	if err := s.repo.AddReviewComment(cmd.PkgId, &comment); err != nil {
		logrus.Errorf(
			"failed to add a comment when %s, err:%s",
			cmd.logString(), err.Error(),
		)
	}
}

// HandlePkgAlreadyExisted
func (s softwarePkgMessageService) HandlePkgAlreadyExisted(cmd CmdToHandlePkgAlreadyExisted) error {
	if b, _ := s.repo.HasSoftwarePkg(cmd.PkgName); b {
		return nil
	}

	v, err := s.manager.GetPkg(cmd.PkgName)
	if err != nil {
		logrus.Errorf("get pkg/%s failed, err:%s", cmd.PkgName.PackageName(), err.Error())

		return err
	}

	if err = s.repo.AddSoftwarePkg(&v); err != nil {
		if commonrepo.IsErrorDuplicateCreating(err) {
			return nil
		}

		logrus.Errorf(
			"failed to add a software pkg, pkgname:%s, err:%s",
			cmd.PkgName.PackageName(), err.Error(),
		)
	}

	return err
}
