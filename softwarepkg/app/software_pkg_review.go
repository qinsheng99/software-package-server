package app

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	commonrepo "github.com/opensourceways/software-package-server/common/domain/repository"
	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/sensitivewords"
)

func (s *softwarePkgService) GetPkgReviewDetail(pid string) (SoftwarePkgReviewDTO, string, error) {
	v, _, err := s.repo.FindSoftwarePkg(pid)
	if err != nil {
		return SoftwarePkgReviewDTO{}, errorCodeForFindingPkg(err), err
	}

	return toSoftwarePkgReviewDTO(&v), "", nil
}

func (s *softwarePkgService) NewReviewComment(
	pid string, cmd *CmdToWriteSoftwarePkgReviewComment,
) (code string, err error) {
	if err = s.sensitive.CheckSensitiveWords(cmd.Content.ReviewComment()); err != nil {
		if sensitivewords.IsErrorSensitiveInfo(err) {
			code = errorSoftwarePkgCommentIllegal
		}

		return
	}

	pkg, _, err := s.repo.FindSoftwarePkgBasicInfo(pid)
	if err != nil {
		code = errorCodeForFindingPkg(err)

		return
	}

	if !pkg.CanAddReviewComment() {
		code = errorSoftwarePkgCannotComment
		err = errors.New("can't comment now")

		return
	}

	// TODO: there is a critical case that the comment can't be added now
	comment := domain.NewSoftwarePkgReviewComment(cmd.Author, cmd.Content)
	err = s.repo.AddReviewComment(pid, &comment)

	return
}

func (s *softwarePkgService) TranslateReviewComment(
	cmd *CmdToTranslateReviewComment,
) (dto TranslatedReveiwCommentDTO, code string, err error) {
	v, err := s.repo.FindTranslatedReviewComment(cmd)
	if err == nil {
		dto.Content = v.Content

		return
	}

	if !commonrepo.IsErrorResourceNotFound(err) {
		return
	}

	// translate it
	comment, err := s.repo.FindReviewComment(cmd.PkgId, cmd.CommentId)
	if err != nil {
		if commonrepo.IsErrorResourceNotFound(err) {
			code = errorSoftwarePkgCommentNotFound
		}

		return
	}

	content, err := s.translation.Translate(
		comment.Content.ReviewComment(), cmd.Language,
	)
	if err != nil {
		return
	}

	dto.Content = content

	// save the translated one
	translated := domain.NewSoftwarePkgTranslatedReviewComment(
		&comment, content, cmd.Language,
	)
	_ = s.repo.AddTranslatedReviewComment(cmd.PkgId, &translated)

	return
}

func (s *softwarePkgService) Approve(pid string, user *domain.User) (code string, err error) {
	pkg, version, err := s.repo.FindSoftwarePkgBasicInfo(pid)
	if err != nil {
		code = errorCodeForFindingPkg(err)

		return
	}

	isTC, code, err := s.checkPermission(&pkg, user)
	if err != nil {
		return
	}

	approved, err := pkg.ApproveBy(&domain.SoftwarePkgApprover{
		Account: user.Account,
		IsTC:    isTC,
	})
	if err != nil {
		return
	}

	if err = s.repo.SaveSoftwarePkg(&pkg, version); err != nil {
		return
	}

	if approved {
		s.notifyPkgApproved(&pkg)
	}

	s.addOperationLog(user.Account, dp.PackageOperationLogActionApprove, pid)

	return
}

func (s *softwarePkgService) notifyPkgApproved(pkg *domain.SoftwarePkgBasicInfo) {
	e := domain.NewSoftwarePkgApprovedEvent(pkg)
	err := s.message.NotifyPkgApproved(&e)

	msg := fmt.Sprintf(
		"notify that a pkg:%s/%s was approved", pkg.Id, pkg.PkgName.PackageName(),
	)
	if err == nil {
		logrus.Debugf("successfully to %s", msg)
	} else {
		logrus.Errorf("failed to %s, err:%s", msg, err.Error())
	}
}

func (s *softwarePkgService) Reject(pid string, user *domain.User) (code string, err error) {
	pkg, version, err := s.repo.FindSoftwarePkgBasicInfo(pid)
	if err != nil {
		code = errorCodeForFindingPkg(err)

		return
	}

	isTC, code, err := s.checkPermission(&pkg, user)
	if err != nil {
		return
	}

	err = pkg.RejectBy(&domain.SoftwarePkgApprover{
		Account: user.Account,
		IsTC:    isTC,
	})
	if err != nil {
		return
	}

	if err = s.repo.SaveSoftwarePkg(&pkg, version); err == nil {
		s.addOperationLog(user.Account, dp.PackageOperationLogActionReject, pid)
	}

	return
}

func (s *softwarePkgService) Abandon(pid string, user *domain.User) (code string, err error) {
	pkg, version, err := s.repo.FindSoftwarePkgBasicInfo(pid)
	if err != nil {
		code = errorCodeForFindingPkg(err)

		return
	}

	if err = pkg.Abandon(user); err != nil {
		code = domain.ParseErrorCode(err)
	} else {
		err = s.repo.SaveSoftwarePkg(&pkg, version)
	}

	if err == nil {
		s.addOperationLog(user.Account, dp.PackageOperationLogActionAbandon, pid)
	}

	return
}

func (s *softwarePkgService) RerunCI(pid string, user *domain.User) (code string, err error) {
	pkg, version, err := s.repo.FindSoftwarePkgBasicInfo(pid)
	if err != nil {
		code = errorCodeForFindingPkg(err)

		return
	}

	changed, err := pkg.RerunCI(user)
	if err != nil {
		code = domain.ParseErrorCode(err)

		return
	}

	if changed {
		if err = s.repo.SaveSoftwarePkg(&pkg, version); err != nil {
			return
		}
	}

	e := domain.NewSoftwarePkgAppUpdatedEvent(&pkg)
	if err = s.message.NotifyPkgToRerunCI(&e); err != nil {
		logrus.Errorf(
			"failed to notify re-running ci for pkg:%s, err:%s",
			pkg.Id, err.Error(),
		)
	} else {
		logrus.Debugf(
			"successfully to notify re-running ci for pkg:%s", pkg.Id,
		)

		s.addCommentToRerunCI(pid)

		s.addOperationLog(user.Account, dp.PackageOperationLogActionResunci, pid)
	}

	return
}

func (s *softwarePkgService) addCommentToRerunCI(pkgId string) {
	content, _ := dp.NewReviewComment("The CI will rerun now.")
	comment := domain.NewSoftwarePkgReviewComment(s.robot, content)

	if err := s.repo.AddReviewComment(pkgId, &comment); err != nil {
		logrus.Errorf(
			"failed to add a comment when reruns ci for pkg:%s, err:%s",
			pkgId, err.Error(),
		)
	}
}

func (s *softwarePkgService) checkPermission(pkg *domain.SoftwarePkgBasicInfo, user *domain.User) (
	bool, string, error,
) {
	if has, isTC := s.maintainer.HasPermission(pkg, user); has {
		return isTC, "", nil
	}

	return false, errorSoftwarePkgNoPermission, errors.New("no permission")
}
