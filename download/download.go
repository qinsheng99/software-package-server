package main

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/opensourceways/robot-gitee-lib/client"
)

var cli client.Client

type Params struct {
	Org      string
	Repo     string
	Path     string
	FileName string
}

func main() {
	param := parse()

	branches, err := cli.GetRepoAllBranch(param.Org, param.Repo)
	checkErr(err)

	// because files cannot be downloaded directly in gitee, we must use API with some params like
	// org, repo, branch, filePath.
	// the normal url of file in gitee is https://gitee.com/euler-ttttt/babel244/raw/master/babel.spec,
	// the branch is master, and the path is babel.spec
	// but when the branch name contains /, and the file in the directory,
	// the url will becomes https://gitee.com/euler-ttttt/babel244/raw/dada/3/4/5/6/babel.spec
	// we don't know which is branch name and which is file path
	// So we must do something to separate them
	for _, branch := range branches {
		if strings.Contains(param.Path, branch.Name) {
			filePath := strings.Split(param.Path, branch.Name)[1]

			content, err := cli.GetPathContent(param.Org, param.Repo, filePath, branch.Name)
			if err != nil {
				continue
			}

			decodeContent, err := base64.StdEncoding.DecodeString(content.Content)
			checkErr(err)

			err = os.WriteFile(param.FileName, decodeContent, 0644)
			checkErr(err)

			break
		}
	}

	_, err = os.Stat(param.FileName)
	if err != nil && !os.IsExist(err) {
		fmt.Printf("download %s failed", param.FileName)
		os.Exit(1)
	}
}

func parse() *Params {
	if len(os.Args) < 3 {
		exit("it needs 2 params, rpm url and token")
	}

	token := os.Args[2]
	cli = client.NewClient(func() []byte {
		return []byte(token)
	})

	rpmUrl := os.Args[1]
	u, err := url.Parse(rpmUrl)
	checkErr(err)

	v := strings.Split(u.Path, "/")
	if v[3] != "raw" {
		exit("source file must be raw format")
	}

	return &Params{
		Org:      v[1],
		Repo:     v[2],
		Path:     u.Path,
		FileName: v[len(v)-1],
	}
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func exit(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}
