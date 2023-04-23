package pkgciimpl

import (
	"fmt"

	"github.com/opensourceways/robot-gitee-lib/client"
	"github.com/opensourceways/server-common-lib/utils"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"

	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
	localutils "github.com/opensourceways/software-package-server/utils"
)

var instance *pkgCIImpl

func Init(cfg *Config) error {
	instance = &pkgCIImpl{
		cli: client.NewClient(func() []byte {
			return []byte(cfg.CreateCIPRToken)
		}),
		cfg: *cfg,
	}

	return nil
}

func PkgCI() *pkgCIImpl {
	return instance
}

type softwarePkgInfo struct {
	PkgId   string `json:"pkg_id"`
	PkgName string `json:"pkg_name"`
	Service string `json:"service"`
}

func (s *softwarePkgInfo) toYaml() ([]byte, error) {
	return yaml.Marshal(s)
}

// pkgCIImpl
type pkgCIImpl struct {
	cli client.Client
	cfg Config
}

func (impl *pkgCIImpl) CreateCIPR(info *domain.SoftwarePkgBasicInfo) error {
	branch := impl.branch(info.PkgName)

	if err := impl.createBranch(branch, info); err != nil {
		return err
	}

	pull, err := impl.cli.CreatePullRequest(
		impl.cfg.CIOrg,
		impl.cfg.CIRepo,
		info.PkgName.PackageName(),
		info.PkgName.PackageName(),
		branch,
		impl.cfg.CreateBranch,
		true,
	)
	if err != nil {
		return err
	}

	return impl.createPRComment(pull.Number)
}

func (impl *pkgCIImpl) createPRComment(id int32) (err error) {
	if err = impl.cli.CreatePRComment(
		impl.cfg.CIOrg,
		impl.cfg.CIRepo, id,
		impl.cfg.Comment,
	); err != nil {
		logrus.Errorf("create pr %d comment failed, err:%s", id, err.Error())
	}

	return
}

func (impl *pkgCIImpl) createBranch(branch string, info *domain.SoftwarePkgBasicInfo) error {
	v := &softwarePkgInfo{
		PkgId:   info.Id,
		PkgName: info.PkgName.PackageName(),
		Service: impl.cfg.CIService,
	}

	content, err := v.toYaml()
	if err != nil {
		return err
	}

	params := []string{
		impl.cfg.CIScript,
		impl.cfg.User,
		impl.cfg.CreateCIPRToken,
		impl.cfg.Email,
		branch,
		impl.cfg.CIOrg,
		impl.cfg.CIRepo,
		"pkginfo.yaml",
		string(content),
		info.Application.SourceCode.SpecURL.URL(),
		info.Application.SourceCode.SrcRPMURL.URL(),
	}

	return impl.runcmd(params)
}

func (impl *pkgCIImpl) runcmd(params []string) error {
	out, err, _ := utils.RunCmd(params...)
	if err != nil {
		logrus.Errorf(
			"run create pull request shell, err=%s, out=%s, params=%v",
			err.Error(), out, params[:len(params)-1],
		)
	}

	return err
}

func (impl *pkgCIImpl) branch(pkg dp.PackageName) string {
	return fmt.Sprintf("%s-%d", pkg.PackageName(), localutils.Now())
}
