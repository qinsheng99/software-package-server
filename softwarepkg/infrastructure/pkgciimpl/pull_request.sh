#!/bin/sh

set -euo pipefail

init() {
    if [ -d "$repo" ]; then
       return
    fi


    git clone --depth=1 "$origin"

    cd "$repo"

    git config user.name "$git_user"
    git config user.email "$git_email"
    git config --global pack.threads 1

    cd ..
}

new_branch() {
    cd "$repo"

    git checkout -- .
    git clean -fd

    git checkout master

    git fetch origin master
    git rebase origin/master

    git checkout -b "$branch_name"
}

modify() {
  dn=$(dirname "$new_repo_file")
  if [ ! -d "$dn" ]; then
     mkdir -p "$dn"
  fi

  echo "$new_repo_content" > "$new_repo_file"

  curl -LO "$spec_url"

  if [[ $src_rpm_url == *"gitee.com"* ]]
  then
    /opt/app/download "$src_rpm_url" "${git_token}"
  else
    curl -LO "$src_rpm_url"
  fi

}

commit() {
    git add .

    git commit -m 'apply new package ci pull request'

    git push origin "$branch_name"

    git checkout master

    git branch -D "$branch_name"
}

git_user=$1
git_token=$2
git_email=$3
branch_name=$4
org=$5
repo=$6
new_repo_file=$7
new_repo_content=$8
spec_url=$9
src_rpm_url=${10}

origin=https://${git_user}:${git_token}@gitee.com/${org}/${repo}.git

init

new_branch

modify

commit
