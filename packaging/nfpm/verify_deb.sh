#!/bin/sh -ue

case $(uname -m) in
  amd64* | x86_64* )
    deb_arch=amd64
    ;;
  arm64* | aarch64* )
    deb_arch=arm64
    ;;
  * )
    >&2 echo "Unsupported hardware: $(uname -m)"
    exit 1;
esac

deb_file=${1:-$(ls dist/func-e_*_linux_${deb_arch}.deb)}

echo installing "${deb_file}"
sudo dpkg -i "${deb_file}"

echo ensuring func-e was installed
test -f /usr/bin/func-e
func-e -version

echo ensuring func-e man page was installed
test -f /usr/local/share/man/man8/func-e.8

echo uninstalling func-e
sudo apt-get remove -yqq func-e

echo ensuring func-e was uninstalled
test -f /usr/bin/func-e && exit 1
exit 0
