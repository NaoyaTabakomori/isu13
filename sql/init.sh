#!/usr/bin/env bash

set -eux
cd $(dirname $0)

if test -f /home/isucon/env.sh; then
	. /home/isucon/env.sh
fi

ISUCON_DB_HOST=${ISUCON13_MYSQL_DIALCONFIG_ADDRESS:-127.0.0.1}
ISUCON_DB_PORT=${ISUCON13_MYSQL_DIALCONFIG_PORT:-3306}
ISUCON_DB_USER=${ISUCON13_MYSQL_DIALCONFIG_USER:-isucon}
ISUCON_DB_PASSWORD=${ISUCON13_MYSQL_DIALCONFIG_PASSWORD:-isucon}
ISUCON_DB_NAME=${ISUCON13_MYSQL_DIALCONFIG_DATABASE:-isupipe}

ISUCON_DNS_HOST='127.0.0.1'
ISUCON_DNS_PORT='3306'
ISUCON_DNS_USER='isudns'
ISUCON_DNS_PASSWORD='isudns'
ISUCON_DNS_NAME='isudns'


# MySQLを初期化
mysql -u"$ISUCON_DB_USER" \
		-p"$ISUCON_DB_PASSWORD" \
		--host "$ISUCON_DB_HOST" \
		--port "$ISUCON_DB_PORT" \
		"$ISUCON_DB_NAME" < init.sql

mysql -u"$ISUCON_DB_USER" \
		-p"$ISUCON_DB_PASSWORD" \
		--host "$ISUCON_DB_HOST" \
		--port "$ISUCON_DB_PORT" \
		"$ISUCON_DB_NAME" < initial_users.sql

mysql -u"$ISUCON_DB_USER" \
		-p"$ISUCON_DB_PASSWORD" \
		--host "$ISUCON_DB_HOST" \
		--port "$ISUCON_DB_PORT" \
		"$ISUCON_DB_NAME" < initial_livestreams.sql

mysql -u"$ISUCON_DB_USER" \
		-p"$ISUCON_DB_PASSWORD" \
		--host "$ISUCON_DB_HOST" \
		--port "$ISUCON_DB_PORT" \
		"$ISUCON_DB_NAME" < initial_tags.sql

mysql -u"$ISUCON_DB_USER" \
		-p"$ISUCON_DB_PASSWORD" \
		--host "$ISUCON_DB_HOST" \
		--port "$ISUCON_DB_PORT" \
		"$ISUCON_DB_NAME" < initial_livestream_tags.sql

mysql -u"$ISUCON_DB_USER" \
		-p"$ISUCON_DB_PASSWORD" \
		--host "$ISUCON_DB_HOST" \
		--port "$ISUCON_DB_PORT" \
		"$ISUCON_DB_NAME" < initial_reservation_slots.sql

mysql -u"$ISUCON_DB_USER" \
		-p"$ISUCON_DB_PASSWORD" \
		--host "$ISUCON_DB_HOST" \
		--port "$ISUCON_DB_PORT" \
		"$ISUCON_DB_NAME" < initial_reactions.sql

mysql -u"$ISUCON_DB_USER" \
		-p"$ISUCON_DB_PASSWORD" \
		--host "$ISUCON_DB_HOST" \
		--port "$ISUCON_DB_PORT" \
		"$ISUCON_DB_NAME" < initial_ngwords.sql

mysql -u"$ISUCON_DB_USER" \
		-p"$ISUCON_DB_PASSWORD" \
		--host "$ISUCON_DB_HOST" \
		--port "$ISUCON_DB_PORT" \
		"$ISUCON_DB_NAME" < initial_livecomments.sql

bash ../pdns/init_zone.sh

mysql -u"$ISUCON_DNS_USER" \
		-p"$ISUCON_DNS_PASSWORD" \
		--host "$ISUCON_DNS_HOST" \
		--port "$ISUCON_DNS_PORT" \
		"$ISUCON_DNS_NAME" < ../pdns/init_dns.sql
