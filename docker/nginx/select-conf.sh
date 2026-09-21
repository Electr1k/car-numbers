#!/bin/sh
# Выбирает конфиг по наличию сертификата: без него nginx с HTTPS не стартует, а без HTTP certbot не выпустит сертификат
set -e

if [ -f "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" ]; then
    tpl=https
else
    tpl=http
fi

envsubst '${DOMAIN}' < "/etc/nginx/site/$tpl.conf.template" > /etc/nginx/conf.d/default.conf
echo "nginx: $tpl для $DOMAIN"
