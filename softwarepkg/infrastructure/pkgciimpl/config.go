package pkgciimpl

type Config struct {
	User            string `json:"user"               required:"true"`
	Email           string `json:"email"              required:"true"`
	CIOrg           string `json:"ci_org"             required:"true"`
	CIRepo          string `json:"ci_repo"            required:"true"`
	Comment         string `json:"comment"            required:"true"`
	CIScript        string `json:"ci_script"          required:"true"`
	CIService       string `json:"ci_service"         required:"true"`
	CreateBranch    string `json:"create_branch"      required:"true"`
	CreateCIPRToken string `json:"create_ci_pr_token" required:"true"`
}

func (cfg *Config) SetDefault() {
	if cfg.CIScript == "" {
		cfg.CIScript = "/opt/app/pull_request.sh"
	}

	if cfg.CreateBranch == "" {
		cfg.CreateBranch = "master"
	}

	if cfg.Comment == "" {
		cfg.Comment = "/retest"
	}
}
