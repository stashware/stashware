#!/bin/bash

pwd=`pwd`

root_dir=$( cd $( dirname ${BASH_SOURCE[0]} ) && pwd )
cd $root_dir

#brew install mingw-w64
###brew install FiloSottile/musl-cross/musl-cross
#brew tap SergioBenitez/osxct; brew install x86_64-unknown-linux-gnu
#git tag -a 'v1.0.1' -m 'version 1.0.1'
#git push origin --tags
ver=`git tag -l |tail -1`
if [ "$ver" == "" ]; then
	ver=`git log -1 --date=short --pretty=format:'%h-%cd'`
fi

platform=""
ext=""

export CGO_ENABLED=1
export GOARCH=amd64

if [ "$1" == "linux" ]; then
	ldext="-linkmode external -extldflags -static"
	platform="-nix64"
	export GOOS=linux
	#export CC=x86_64-linux-musl-gcc
	export CC=x86_64-unknown-linux-gnu-gcc
elif [ "$1" == "windows" ]; then
	platform="-win64"
	ldext="-linkmode external -extldflags -static"
	export GOOS=windows
	export CC=x86_64-w64-mingw32-gcc
	ext=".exe"
elif [ "$1" == "osx" ]; then
	# export CGO_LDFLAGS="-linkmode external -extldflags -w"
	# export CC=gcc-7
	# export CXX=g++-7
	#cp librandomx.a $GOPATH/pkg/mod/github.com/ngchain/go-randomx@v0.1.7/build/macos-x86_64/
	platform="-osx"
	# export CGO_LDFLAGS="$CGO_LDFLAGS -mmacosx-version-min=10.13"
	export GOOS=darwin
fi

dflags=""
if [ "$2" = "release" ]; then
	dflags="-s -w"
fi

prjname=stashware-$ver$platform
prjdir=$pwd/$prjname
rm -fr $prjdir $prjdir.tgz
mkdir $prjdir

cd $root_dir/cmd/swr
go build -gcflags "-N -l" -ldflags "$ldext $dflags -X main.GitVersion=$ver" -o $prjdir/swr$ext

cd $root_dir/cmd/swrcli
go build -gcflags "-N -l" -ldflags "$ldext $dflags -X main.GitVersion=$ver" -o $prjdir/swrcli$ext


cp $root_dir/conf/* $prjdir/

cd $pwd
tar -czf $prjname.tgz $prjname
