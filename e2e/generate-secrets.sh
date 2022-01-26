#!/bin/sh
test -f /root/.ssh/id_rsa || ssh-keygen -b 2048 -t rsa -f /root/.ssh/id_rsa -q -N ""
test -f /root/.ssh/password || tr -cd "[:alnum:]" < /dev/urandom | fold -w30 | head -n1 | tr -d '\n' > /root/.ssh/password
