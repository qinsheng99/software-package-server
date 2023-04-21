package pkgciimpl

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/opensourceways/robot-gitee-lib/client"
	"github.com/opensourceways/server-common-lib/utils"
	"github.com/sirupsen/logrus"

	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
	localutils "github.com/opensourceways/software-package-server/utils"
)

var instance *pkgCIImpl

func Init(cfg *Config) error {
	tmpl, err := template.ParseFiles(cfg.PkgInfoTpl)
	if err != nil {
		return err
	}

	instance = &pkgCIImpl{
		cli: client.NewClient(func() []byte {
			return []byte(cfg.CreateCIPRToken)
		}),
		cfg:        *cfg,
		pkgInfoTpl: tmpl,
	}

	return nil
}

func PkgCI() *pkgCIImpl {
	return instance
}

type softwarePkgInfo struct {
	PkgId   string
	PkgName string
	Service string
}

// pkgCIImpl
type pkgCIImpl struct {
	cli        client.Client
	cfg        Config
	pkgInfoTpl *template.Template
}

func (impl *pkgCIImpl) CreateCIPR(info *domain.SoftwarePkgBasicInfo) error {
	branch := impl.branch(info.PkgName)

	if err := impl.createBranch(branch, info); err != nil {
		return err
	}

	_, err := impl.cli.CreatePullRequest(
		impl.cfg.CIOrg,
		impl.cfg.CIRepo,
		info.PkgName.PackageName(),
		fmt.Sprintf("add package: %s ci record", info.PkgName.PackageName()),
		branch,
		"master",
		true,
	)

	return err
}

func (impl *pkgCIImpl) createBranch(branch string, info *domain.SoftwarePkgBasicInfo) error {
	content, err := impl.genPkgInfo(&softwarePkgInfo{
		PkgId:   info.Id,
		PkgName: info.PkgName.PackageName(),
		Service: impl.cfg.CIService,
	})
	if err != nil {
		return err
	}
	params := []string{
		impl.cfg.CIScript,
		impl.cfg.CIUser,
		impl.cfg.CreateCIPRToken,
		impl.cfg.CIEmail,
		branch,
		impl.cfg.CIOrg,
		impl.cfg.CIRepo,
		"pkginfo.yaml",
		content,
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

func (impl *pkgCIImpl) genPkgInfo(data *softwarePkgInfo) (string, error) {
	buf := new(bytes.Buffer)

	if err := impl.pkgInfoTpl.Execute(buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
