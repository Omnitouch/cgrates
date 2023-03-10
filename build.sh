#! /usr/bin/env sh
echo "Building CGRateS ..."

GIT_LAST_LOG=$(git log -1 | tr -d "'")

GIT_TAG_LOG=$(git tag -l --points-at HEAD)

if [ ! -z "$GIT_TAG_LOG" ]
then
    GIT_LAST_LOG=""
fi

go install -ldflags "-X 'github.com/Omnitouch/cgrates/utils.GitLastLog=$GIT_LAST_LOG'" github.com/Omnitouch/cgrates/cmd/cgr-engine
cr=$?
go install -ldflags "-X 'github.com/Omnitouch/cgrates/utils.GitLastLog=$GIT_LAST_LOG'" github.com/Omnitouch/cgrates/cmd/cgr-loader
cl=$?
go install -ldflags "-X 'github.com/Omnitouch/cgrates/utils.GitLastLog=$GIT_LAST_LOG'" github.com/Omnitouch/cgrates/cmd/cgr-console
cc=$?
go install -ldflags "-X 'github.com/Omnitouch/cgrates/utils.GitLastLog=$GIT_LAST_LOG'" github.com/Omnitouch/cgrates/cmd/cgr-migrator
cm=$?
go install -ldflags "-X 'github.com/Omnitouch/cgrates/utils.GitLastLog=$GIT_LAST_LOG'" github.com/Omnitouch/cgrates/cmd/cgr-tester
ct=$?

exit $cr || $cl || $cc || $cm || $ct
