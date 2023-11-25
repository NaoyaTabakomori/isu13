#!/bin/sh
set -ex

sudo apt-get install -y dstat

# alp
cd /var/tmp/ && wget https://github.com/tkuchiki/alp/releases/download/v1.0.21/alp_linux_amd64.tar.gz
tar -xzvf alp_linux_amd64.tar.gz
sudo mv ./alp /usr/local/bin

# pt-query-digest
cd /var/tmp && wget https://www.percona.com/downloads/percona-toolkit/3.5.5/binary/tarball/percona-toolkit-3.5.5_x86_64.tar.gz
tar -zxvf percona-toolkit-3.5.5_x86_64.tar.gz
sudo mv ./percona-toolkit-3.5.5/bin/pt-query-digest /usr/local/bin
