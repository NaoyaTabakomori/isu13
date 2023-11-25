#!/bin/sh
set -ex

echo "Deploy script started."

PROJECT_ROOT=/home/isucon/webapp

LOG_BACKUP_DIR=/var/log/isucon

USER=isucon
KEY_OPTION="-A"

BACKUP_TARGET_LIST="/var/log/nginx/access.log /var/log/nginx/error.log"

BRANCH=$1
if [ -z "$BRANCH" ]; then
  echo "you must set branch"
  exit 1
fi

ENV=$2
if [ -z "$ENV" ]; then
  echo "you must set env"
  exit 1
fi

WEB_SERVERS="${ENV}"
APP_SERVERS="${ENV}"
DB_SERVER="${ENV}"

# sed -n -r 's/^(LogFormat.*)(" combined)/\1 %D\2/p' /etc/httpd/conf/httpd.conf
echo "Stop Web Server"
for WEB_SERVER in $WEB_SERVERS
do
cat <<EOS | ssh $KEY_OPTION $USER@$WEB_SERVER sh
sudo systemctl stop nginx
EOS
done

echo "Stop Application Server"
for APP_SERVER in $APP_SERVERS
do
cat <<EOS | ssh $KEY_OPTION $USER@$APP_SERVER sh
sudo systemctl stop isupipe-go.service
EOS
done

echo "Stop DataBase Server"
cat <<EOS | ssh $KEY_OPTION $USER@$DB_SERVER sh
sudo systemctl stop mysql
EOS

echo "Get Current git hash"
for APP_SERVER in $APP_SERVERS
do
hash=`cat <<EOS | ssh $KEY_OPTION $USER@$APP_SERVER sh
cd $PROJECT_ROOT
git rev-parse --short HEAD
EOS`
echo "Current Hash: $hash"
done

set +e
LOG_DATE=`date +"%H%M%S"`
echo "Backup App Server LOG"
for LOG_PATH in $BACKUP_TARGET_LIST
do
    LOG_FILE=`basename $LOG_PATH`
for APP_SERVER in $APP_SERVERS
do
    cat <<EOS | ssh $KEY_OPTION $USER@$APP_SERVER sh
sudo mkdir -p ${LOG_BACKUP_DIR}
sudo mv $LOG_PATH ${LOG_BACKUP_DIR}/${LOG_FILE}_${LOG_DATE}_${hash}
EOS
done
done

cat <<EOS | ssh $KEY_OPTION $USER@$DB_SERVER sh
sudo mv /var/log/mysql/mysql-slow.log ${LOG_BACKUP_DIR}/mysql-slow_${LOG_DATE}_${hash}
EOS

set -e

echo "Current Hash: $hash"
echo "Update Project"
for APP_SERVER in $APP_SERVERS
do
cat <<EOS | ssh $KEY_OPTION $USER@$APP_SERVER sh
cd $PROJECT_ROOT
git clean -fd
git reset --hard
git fetch -p
git checkout $BRANCH
git pull --rebase
EOS
done

cat <<EOS | ssh $KEY_OPTION $USER@$DB_SERVER sh
cd $PROJECT_ROOT
git clean -fd
git reset --hard
git fetch -p
git checkout $BRANCH
git pull --rebase

cd $PROJECT_ROOT
cat $PROJECT_ROOT/sql/initdb.d/99_drop_create_db.sql | sudo mysql isupipe
cat $PROJECT_ROOT/sql/initdb.d/10_schema.sql | sudo mysql isupipe

cd $PROJECT_ROOT/go
PATH=/home/isucon/local/python/bin:/home/isucon/local/perl/bin:/home/isucon/webapp/perl/local/bin:/home/isucon/local/ruby/bin:/home/isucon/local/php/bin:/home/isucon/local/php/sbin:/home/isucon/.cargo/bin:/home/isucon/local/node/bin:/home/isucon/local/golang/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games:/snap/bin make build
EOS

echo "Get new git hash"
for APP_SERVER in $APP_SERVERS
do
new_hash=`cat <<EOS | ssh $KEY_OPTION $USER@$APP_SERVER sh
cd $PROJECT_ROOT
git rev-parse --short HEAD
EOS`
echo "Current Hash: $new_hash"
done
echo "Start Database Server"
cat <<EOS | ssh $KEY_OPTION $USER@$DB_SERVER sh
sudo swapoff -a && sudo swapon -a
sudo systemctl start mysql
EOS
echo "Start App Server"
for APP_SERVER in $APP_SERVERS
do
cat <<EOS | ssh $KEY_OPTION $USER@$APP_SERVER sh
sudo swapoff -a && sudo swapon -a
sudo systemctl start isupipe-go.service
EOS
done
echo "Start Web Server"
for WEB_SERVER in $WEB_SERVERS
do
cat <<EOS | ssh $KEY_OPTION $USER@$WEB_SERVER sh
sudo swapoff -a && sudo swapon -a
sudo systemctl start nginx
EOS
done
echo "Deploy script finished."