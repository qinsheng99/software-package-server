package app

import (
	"fmt"

	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
)

// CmdToHandlePkgCIChecking
type CmdToHandlePkgCIChecking struct {
	PkgId string
}

func (cmd *CmdToHandlePkgCIChecking) logString() string {
	return fmt.Sprintf(
		"handling pkg ci checking, pkgid:%s", cmd.PkgId,
	)
}

// CmdToHandlePkgCIChecked
type CmdToHandlePkgCIChecked struct {
	PkgId    string
	Detail   string
	Success  bool
	PRNumber int
}

func (cmd *CmdToHandlePkgCIChecked) logString() string {
	return fmt.Sprintf(
		"handling pkg ci checked, pkgid:%s", cmd.PkgId,
	)
}

// CmdToHandlePkgInitialized
type CmdToHandlePkgInitialized struct {
	PkgId      string
	RelevantPR dp.URL
	// RepoLink is the one of already existed pkg
	RepoLink    dp.URL
	FiledReason string
}

func (cmd *CmdToHandlePkgInitialized) isSuccess() bool {
	return cmd.FiledReason == "" && cmd.RepoLink == nil
}

func (cmd *CmdToHandlePkgInitialized) isPkgAreadyExisted() bool {
	return cmd.RepoLink != nil
}

func (cmd *CmdToHandlePkgInitialized) logString() string {
	return fmt.Sprintf(
		"handling pkg initialized, pkgid:%s, pr:%s",
		cmd.PkgId, cmd.RelevantPR.URL(),
	)
}

// CmdToHandlePkgRepoCreated
type CmdToHandlePkgRepoCreated struct {
	PkgId       string
	FiledReason string

	domain.RepoCreatedInfo
}

func (cmd *CmdToHandlePkgRepoCreated) isSuccess() bool {
	return cmd.FiledReason == ""
}

func (cmd *CmdToHandlePkgRepoCreated) logString() string {
	if !cmd.isSuccess() {
		return ""
	}

	return fmt.Sprintf(
		"handling pkg repo created, pkgid:%s, platform:%s, repo:%s",
		cmd.PkgId, cmd.Platform.PackagePlatform(), cmd.RepoLink.URL(),
	)
}

// CmdToHandlePkgCodeSaved
type CmdToHandlePkgCodeSaved = CmdToHandlePkgRepoCreated

type CmdToHandlePkgAlreadyExisted struct {
	PkgName dp.PackageName
}
