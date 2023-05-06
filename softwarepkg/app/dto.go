package app

import (
	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/repository"
	"github.com/opensourceways/software-package-server/utils"
)

type CmdToApplyNewSoftwarePkg struct {
	PkgName     dp.PackageName
	Importer    domain.User
	Application domain.SoftwarePkgApplication
}

type CmdToUpdateSoftwarePkgApplication struct {
	PkgId string
	CmdToApplyNewSoftwarePkg
}

type CmdToListPkgs = repository.OptToFindSoftwarePkgs

type CmdToWriteSoftwarePkgReviewComment struct {
	Author  dp.Account
	Content dp.ReviewComment
}

// SoftwarePkgBasicInfoDTO
type SoftwarePkgBasicInfoDTO struct {
	Id        string `json:"id"`
	Importer  string `json:"importer"`
	PkgName   string `json:"pkg_name"`
	Phase     string `json:"phase"`
	CIStatus  string `json:"ci_status"`
	AppliedAt string `json:"applied_at"`
	RepoLink  string `json:"repo_link"`
	PkgDesc   string `json:"desc"`
	Sig       string `json:"sig"`
	Platform  string `json:"platform"`
}

func toSoftwarePkgBasicInfoDTO(v *domain.SoftwarePkgBasicInfo) SoftwarePkgBasicInfoDTO {
	app := &v.Application
	dto := SoftwarePkgBasicInfoDTO{
		Id:        v.Id,
		Sig:       app.ImportingPkgSig.ImportingPkgSig(),
		Phase:     v.Phase.PackagePhase(),
		CIStatus:  v.CI.Status.PackageCIStatus(),
		PkgDesc:   app.PackageDesc.PackageDesc(),
		PkgName:   v.PkgName.PackageName(),
		Platform:  app.PackagePlatform.PackagePlatform(),
		Importer:  v.Importer.Account.Account(),
		AppliedAt: utils.ToDate(v.AppliedAt),
	}

	if v.RepoLink != nil {
		dto.RepoLink = v.RepoLink.URL()
	}

	return dto
}

func toSoftwarePkgBasicInfoDTOs(v []domain.SoftwarePkgBasicInfo) (r []SoftwarePkgBasicInfoDTO) {
	if n := len(v); n > 0 {
		r = make([]SoftwarePkgBasicInfoDTO, n)
		for i := range v {
			r[i] = toSoftwarePkgBasicInfoDTO(&v[i])
		}
	}

	return
}

// SoftwarePkgApplicationDTO
type SoftwarePkgApplicationDTO struct {
	SpecURL           string `json:"spec_url"`
	SrcRPMURL         string `json:"src_rpm_url"`
	PackageDesc       string `json:"desc"`
	PackagePlatform   string `json:"platform"`
	ImportingPkgSig   string `json:"sig"`
	ReasonToImportPkg string `json:"reason"`
}

func toSoftwarePkgApplicationDTO(v *domain.SoftwarePkgApplication) SoftwarePkgApplicationDTO {
	return SoftwarePkgApplicationDTO{
		SpecURL:           v.SourceCode.SpecURL.URL(),
		SrcRPMURL:         v.SourceCode.SrcRPMURL.URL(),
		PackageDesc:       v.PackageDesc.PackageDesc(),
		PackagePlatform:   v.PackagePlatform.PackagePlatform(),
		ImportingPkgSig:   v.ImportingPkgSig.ImportingPkgSig(),
		ReasonToImportPkg: v.ReasonToImportPkg.ReasonToImportPkg(),
	}
}

// SoftwarePkgReviewCommentDTO
type SoftwarePkgReviewCommentDTO struct {
	Id            string `json:"id"`
	Author        string `json:"author"`
	Content       string `json:"content"`
	CreatedAt     string `json:"created_at"`
	SinceCreation int64  `json:"since_creation"`
}

func toSoftwarePkgReviewCommentDTO(v *domain.SoftwarePkgReviewComment) SoftwarePkgReviewCommentDTO {
	return SoftwarePkgReviewCommentDTO{
		Id:            v.Id,
		Author:        v.Author.Account(),
		Content:       v.Content.ReviewComment(),
		CreatedAt:     utils.ToDateTime(v.CreatedAt),
		SinceCreation: utils.Now() - v.CreatedAt,
	}
}

func toSoftwarePkgReviewCommentDTOs(v []domain.SoftwarePkgReviewComment) (r []SoftwarePkgReviewCommentDTO) {
	if n := len(v); n > 0 {
		r = make([]SoftwarePkgReviewCommentDTO, n)
		for i := range v {
			r[i] = toSoftwarePkgReviewCommentDTO(&v[i])
		}
	}

	return
}

// SoftwarePkgOperationLogDTO
type SoftwarePkgOperationLogDTO struct {
	User   string `json:"user"`
	Time   string `json:"time"`
	Action string `json:"action"`
}

func toSoftwarePkgOperationLogDTO(v *domain.SoftwarePkgOperationLog) SoftwarePkgOperationLogDTO {
	return SoftwarePkgOperationLogDTO{
		User:   v.User.Account(),
		Time:   utils.ToDateTime(v.Time),
		Action: v.Action.PackageOperationLogAction(),
	}
}

func toSoftwarePkgOperationLogDTOs(v []domain.SoftwarePkgOperationLog) (r []SoftwarePkgOperationLogDTO) {
	if n := len(v); n > 0 {
		r = make([]SoftwarePkgOperationLogDTO, n)
		for i := range v {
			r[i] = toSoftwarePkgOperationLogDTO(&v[i])
		}
	}

	return
}

// SoftwarePkgApproverDTO
type SoftwarePkgApproverDTO struct {
	Account string `json:"account"`
	IsTC    bool   `json:"is_tc"`
}

// SoftwarePkgReviewDTO
type SoftwarePkgReviewDTO struct {
	SoftwarePkgBasicInfoDTO

	Logs        []SoftwarePkgOperationLogDTO  `json:"logs"`
	Comments    []SoftwarePkgReviewCommentDTO `json:"comments"`
	ApprovedBy  []SoftwarePkgApproverDTO      `json:"approved_by"`
	RejectedBy  []SoftwarePkgApproverDTO      `json:"rejected_by"`
	Application SoftwarePkgApplicationDTO     `json:"application"`
}

func toSoftwarePkgReviewDTO(v *domain.SoftwarePkg) SoftwarePkgReviewDTO {
	return SoftwarePkgReviewDTO{
		SoftwarePkgBasicInfoDTO: toSoftwarePkgBasicInfoDTO(&v.SoftwarePkgBasicInfo),
		Logs:                    toSoftwarePkgOperationLogDTOs(v.Logs),
		Comments:                toSoftwarePkgReviewCommentDTOs(v.Comments),
		ApprovedBy:              toSoftwarePkgApproverDTO(v.ApprovedBy),
		RejectedBy:              toSoftwarePkgApproverDTO(v.RejectedBy),
		Application:             toSoftwarePkgApplicationDTO(&v.Application),
	}
}

func toSoftwarePkgApproverDTO(v []domain.SoftwarePkgApprover) (r []SoftwarePkgApproverDTO) {
	if n := len(v); n > 0 {
		r = make([]SoftwarePkgApproverDTO, n)
		for i := range v {
			r[i] = SoftwarePkgApproverDTO{
				Account: v[i].Account.Account(),
				IsTC:    v[i].IsTC,
			}
		}
	}

	return
}

// SoftwarePkgsDTO
type SoftwarePkgsDTO struct {
	Pkgs  []SoftwarePkgBasicInfoDTO `json:"pkgs"`
	Total int                       `json:"total"`
}

func toSoftwarePkgsDTO(v []domain.SoftwarePkgBasicInfo, total int) SoftwarePkgsDTO {
	return SoftwarePkgsDTO{
		Pkgs:  toSoftwarePkgBasicInfoDTOs(v),
		Total: total,
	}
}

// CmdToTranslateReviewComment
type CmdToTranslateReviewComment = repository.TranslatedReviewCommentIndex

type TranslatedReveiwCommentDTO struct {
	Content string `json:"content"`
}
