#!/bin/sh
test -f /root/.ssh/id_rsa || ssh-keygen -b 2048 -t rsa -f /root/.ssh/id_rsa -q -N ""
