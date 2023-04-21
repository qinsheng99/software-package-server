package pkgciimpl

type Config struct {
	CIOrg           string `json:"ci_org"             required:"true"`
	CIRepo          string `json:"ci_repo"            required:"true"`
	CIUser          string `json:"ci_user"            required:"true"`
	CIEmail         string `json:"ci_email"           required:"true"`
	CIScript        string `json:"ci_script"          required:"true"`
	CIService       string `json:"ci_service"         required:"true"`
	PkgInfoTpl      string `json:"pkg_info_tpl"       required:"true"`
	CreateCIPRToken string `json:"create_ci_pr_token" required:"true"`
}

func (cfg *Config) SetDefault() {
	if cfg.CIScript == "" {
		cfg.CIScript = "/opt/app/pull_request.sh"
	}

	if cfg.CIEmail == "" {
		cfg.CIEmail = "software-pkg-robot@openeuler.org"
	}

	if cfg.PkgInfoTpl == "" {
		cfg.PkgInfoTpl = "/opt/app/pkginfo.yaml"
	}
}
