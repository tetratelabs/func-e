#!/bin/sh -ue

rpm_file=${1:-$(ls dist/func-e_*_linux_$(uname -m).rpm)}

echo "installing ${rpm_file}"
sudo rpm -i "${rpm_file}"

echo ensuring func-e was installed
test -f /usr/bin/func-e
func-e -version

echo ensuring func-e man page was installed
test -f /usr/local/share/man/man8/func-e.8

echo uninstalling func-e
sudo rpm -e func-e

echo ensuring func-e was uninstalled
test -f /usr/bin/func-e && exit 1
exit 0
