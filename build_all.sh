#!/bin/bash

pwd=`pwd`

root_dir=$( cd $( dirname ${BASH_SOURCE[0]} ) && pwd )
cd $root_dir

cd $root_dir/web_wallet
packr2

cd $root_dir/..
./stashware/build.sh osx release
./stashware/build.sh linux release
./stashware/build.sh windows release
