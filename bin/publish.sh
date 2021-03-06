#!/usr/bin/bash

# This script is executed in a special context:
# - When a new PR has been merged to the master branch. It generates the latest version
# - when a decision is made to officially published a new release

# In both case, it assumes forjj git clone is configured as follow:
# - remote origin is pointing out to a fork
# - remote upstream is connected to forj-oss/forjj, with read/write access
# - The local branch must be master and should be connected to origin/master. Not upstream/master.

# If one of this condition is not respected, the script will exit.

if [ "$BUILD_ENV_LOADED" != "true" ]
then
   echo "Please go to your project and load your build environment. 'source build-env.sh'"
   exit 1
fi

if [ "$GITHUB_TOKEN" = '' ]
then
   echo "GITHUB_TOKEN is missing. You need it to publish."
   exit 1
fi

if [ $# -eq 0 ]
then
   echo 'Usage is $0 latest|version'
   exit 0
fi

set -e
if [ ! -f bin/gothub ]
then
   go get -d github.com/itchio/gothub
   go build -o bin/gothub github.com/itchio/gothub
fi
if [ "$(git rev-parse --abbrev-ref HEAD)" != master ]
then
   echo "You must be on master branch."
   exit 1
fi

REMOTE="$(git remote -v | grep "^upstream")"
if [ "$REMOTE" = "" ]
then
    echo "upstream is missing. You must have it configured (git@github.com:forj-oss/forjj.git) and rights given to push"
    exit 1
fi
if [[ ! "$REMOTE" =~ git@github\.com:forj-oss/forjj\.git ]]
then
    echo "upstream is wrongly configured. It must be set with git@github.com:forj-oss/forjj.git"
    exit 1
fi
git stash # Just in case
git fetch upstream
git reset --hard upstream/master

set +e
if [ "$1" = "latest" ]
then
   TAG=latest
   git tag -d $TAG
else
   TAG="$(grep VERSION version.go | sed 's/const VERSION="\(.*\)"/\1/g')"
   PRE_RELEASE="$(grep VERSION version.go | sed 's/const PRERELEASE="\(.*\)"/\1/g')"
   if [ "$(git tag | grep "^$TAG$")" != "" ]
   then
      echo "Unable to publish $TAG. Already published and released."
      exit 1
   fi
   if [[ "$1" != "--auto" ]]
   then
      echo "You are going to publish version $TAG. Ctrl-C to interrupt or press Enter to go on"
      read
   else
      echo "Publishing version $TAG..."
   fi
fi

set -e
git tag $TAG
git push -f upstream $TAG

build.sh
export GITHUB_USER=forj-oss
export GITHUB_REPO=forjj
if [ "$TAG" = latest ]
then
   set +e
   gothub info -t latest
   if [ $? -ne 0 ]
   then
      # TODO: Remove hardcoded binary name.
      gothub release --tag $TAG --name forjj --description "Latest version of forjj." -p
   fi
else
    GOTHUB_PARS=""
   if [ "$PRE_RELEASE" = true ]
   then
      GOTHUB_PARS="-p"
   fi
   gothub release --tag $TAG --name forjj --description "forjj version $TAG." $GOTHUB_PARS
fi

gothub upload --tag $TAG --name forjj --file $GOPATH/bin/forjj --replace
