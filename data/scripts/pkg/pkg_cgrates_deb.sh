#! /usr/bin/env sh

WORK_DIR=/tmp/cgrates
rm -rf $WORK_DIR
mkdir -p $WORK_DIR
cp -r debian $WORK_DIR/debian
cd $WORK_DIR
git clone https://github.com/Omnitouch/cgrates.git src/github.com/Omnitouch/cgrates
tar xvzf src/github.com/Omnitouch/cgrates/data/tutorials/fs_evsock/freeswitch/etc/freeswitch_conf.tar.gz -C src/github.com/Omnitouch/cgrates/data/tutorials/fs_evsock/freeswitch/etc/
rm src/github.com/Omnitouch/cgrates/data/tutorials/fs_evsock/freeswitch/etc/freeswitch_conf.tar.gz
dpkg-buildpackage -us -uc
#rm -rf $WORK_DIR
