#!/usr/bin/env bash
set -euo pipefail

IMAGE=${1:?image tag required}
WORK=$(mktemp -d "${TMPDIR:-/tmp}/forgelocal-docker.XXXXXX")
NAME="forgelocal-secret-cycle-${RANDOM}-${RANDOM}"
SENTINEL="$(head -c 48 /dev/urandom | base64 | tr -d '\n')"
INVALID="$(head -c 32 /dev/urandom | base64 | tr -d '\n')"

cleanup() {
  docker rm -f "$NAME" >/dev/null 2>&1 || true
  rm -rf "$WORK"
}
trap cleanup EXIT

mkdir -p "$WORK/data"
printf '%s' "$SENTINEL" > "$WORK/data/.api-token"
chmod 0600 "$WORK/data/.api-token"

if docker history --no-trunc "$IMAGE" > "$WORK/history" 2>&1 && grep -Fq -- "$SENTINEL" "$WORK/history"; then
  printf '%s\n' 'IMAGE_HISTORY_SENTINEL=FOUND' >&2
  exit 41
fi

docker save "$IMAGE" > "$WORK/image.tar"
if grep -aFq -- "$SENTINEL" "$WORK/image.tar"; then
  printf '%s\n' 'IMAGE_LAYERS_SENTINEL=FOUND' >&2
  exit 42
fi
printf '%s\n' 'IMAGE_HISTORY_SENTINEL=ABSENT' 'IMAGE_LAYERS_SENTINEL=ABSENT'

docker run -d --rm --network=host --shm-size=2g --name "$NAME" \
  --env BROWSEFORGE_SEED_BROWSERS=0 \
  --mount "type=bind,src=$WORK/data,dst=/app/data" \
  "$IMAGE" > "$WORK/container-id"

ready=false
for _ in $(seq 1 30); do
  if [ "$(docker inspect -f '{{.State.Running}}' "$NAME" 2>/dev/null || true)" = true ]; then
    ready=true
    break
  fi
  sleep 1
done
[ "$ready" = true ]
docker logs "$NAME" > "$WORK/valid.log" 2>&1 || true
if grep -Fq -- "$SENTINEL" "$WORK/valid.log"; then
  printf '%s\n' 'VALID_LOG_SENTINEL=FOUND' >&2
  exit 43
fi
[ "$(docker exec "$NAME" stat -c '%a' /app/data/.api-token)" = 600 ]
if curl -fsS --max-time 10 http://127.0.0.1:19280/api/status > "$WORK/status.json" 2> "$WORK/status.err"; then
  if grep -Fq -- "$SENTINEL" "$WORK/status.json" "$WORK/status.err"; then
    printf '%s\n' 'VALID_STDOUT_STDERR_SENTINEL=FOUND' >&2
    exit 44
  fi
  printf '%s\n' 'API_STATUS=PASS'
else
  printf '%s\n' 'API_STATUS=NOT_CONFIRMED'
fi
printf '%s\n' 'VALID_LOG_SENTINEL=ABSENT' 'VALID_STDOUT_STDERR_SENTINEL=ABSENT' 'TOKEN_FILE_MODE=0600'
docker stop -t 5 "$NAME" >/dev/null
for _ in $(seq 1 20); do
  [ -z "$(docker ps -aq --filter "name=^${NAME}$")" ] && break
  sleep 0.5
done
[ -z "$(docker ps -aq --filter "name=^${NAME}$")" ]
printf '%s\n' 'CONTAINER_CLEANUP=PASS'

rm -f "$WORK/data/.api-token"
docker run -d --rm --network=host --shm-size=2g --name "$NAME" \
  --env BROWSEFORGE_SEED_BROWSERS=0 \
  --mount "type=bind,src=$WORK/data,dst=/app/data" \
  "$IMAGE" > "$WORK/absent-id"
sleep 3
docker logs "$NAME" > "$WORK/absent.log" 2>&1 || true
if grep -Fq 'API Token:' "$WORK/absent.log"; then
  printf '%s\n' 'ABSENT_TOKEN_OUTPUT=FOUND' >&2
  exit 47
fi
printf '%s\n' 'ABSENT_TOKEN_OUTPUT=ABSENT'
docker rm -f "$NAME" >/dev/null 2>&1 || true

printf '%s' "$INVALID" > "$WORK/data/.api-token"
chmod 0600 "$WORK/data/.api-token"
docker run -d --rm --network=host --shm-size=2g --name "$NAME" \
  --env BROWSEFORGE_SEED_BROWSERS=0 \
  --mount "type=bind,src=$WORK/data,dst=/app/data" \
  "$IMAGE" > "$WORK/invalid-id"
sleep 3
docker logs "$NAME" > "$WORK/invalid.log" 2>&1 || true
if grep -Fq -- "$INVALID" "$WORK/invalid.log"; then
  printf '%s\n' 'INVALID_TOKEN_OUTPUT=FOUND' >&2
  exit 48
fi
printf '%s\n' 'INVALID_TOKEN_OUTPUT=ABSENT' 'INVALID_TOKEN_FILE_MODE=0600'
docker rm -f "$NAME" >/dev/null 2>&1 || true
[ -z "$(docker ps -aq --filter "name=^${NAME}$")" ]
printf '%s\n' 'FINAL_CONTAINER_CLEANUP=PASS' 'DOCKER_SECRET_REDACTION_CYCLE=PASS'
