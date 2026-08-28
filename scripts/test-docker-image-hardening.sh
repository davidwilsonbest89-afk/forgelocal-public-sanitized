#!/usr/bin/env bash
set -euo pipefail

IMAGE=${1:?usage: $0 IMAGE}
DOCKER=(docker)
if ! docker info >/dev/null 2>&1; then
  DOCKER=(sudo docker)
fi

"${DOCKER[@]}" run --rm --network=host --entrypoint /bin/bash "$IMAGE" -lc '
  test ! -e /etc/ssl/private/ssl-cert-snakeoil.key
  test ! -e /etc/ssl/certs/ssl-cert-snakeoil.pem
  test ! -e /etc/systemd/system/multi-user.target.wants/ssl-cert.service
'
printf '%s\n' 'IMAGE_SNAKEOIL_ARTIFACTS=ABSENT'
