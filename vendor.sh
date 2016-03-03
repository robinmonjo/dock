#! /bin/bash

set -e

git_clone() {
  pkg=$1
  rev=$2
  
  pkg_url=https://$pkg
	target_dir=vendor/$pkg

	echo -n "$pkg @ $rev: "
  
  if [ -d $target_dir ]; then
		echo -n 'rm old, '
		rm -fr $target_dir
	fi

	echo -n 'clone, '
  git clone --quiet --no-checkout $pkg_url $target_dir
  ( cd $target_dir && git reset --quiet --hard $rev )
  
  echo -n 'rm VCS, '
	( cd $target_dir && rm -rf .{git,hg} )

	echo done
}

git_clone github.com/codegangsta/cli v1.2.0
git_clone github.com/Sirupsen/logrus v0.8.6
git_clone github.com/kr/pty

svn export https://github.com/docker/docker/trunk/pkg/term vendor/github.com/docker/docker/pkg/term --non-interactive --trust-server-cert