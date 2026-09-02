#!/bin/sh
set -eu

mkdir -p /data/frontend /data/sql /data/uploads

if [ ! -f /data/app.toml ]; then
	cp /opt/app/app.prod.toml /data/app.toml
fi

rm -rf /data/frontend/assets
cp -a /opt/app/frontend/assets /data/frontend/assets

rm -rf /data/sql/migrations
cp -a /opt/app/sql/migrations /data/sql/migrations

/opt/app/goose -dir ./sql/migrations sqlite3 ./data.db up

exec /opt/app/app
