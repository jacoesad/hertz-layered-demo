#!/bin/bash

set -euo pipefail

CURDIR=$(cd "$(dirname "$0")"; pwd)
ROOTDIR=$(cd "$CURDIR/.."; pwd)
MODULE=${MODULE:-hz-server}
CLIENT_DIR="biz/client"
SERVER_IDL_DIR="idl/service"
DOWNSTREAM_IDL_DIR="idl/downstream"
DEFAULT_INIT_IDL="$SERVER_IDL_DIR/console/task.thrift"

usage() {
  cat <<EOF
Usage:
  script/hz_generate.sh [all|server|client|downstream|init]

Commands:
  all     Run server hz update and downstream hz client generation. Default.
  server  Regenerate server router/handler/model code from service IDLs.
  client  Regenerate downstream HTTP client code.
  downstream
          Alias of client.
  init    Run the first hz new command. Refuses in an existing project unless HZ_FORCE_INIT=1.

Environment:
  MODULE    Go module name passed to hz. Default: hz-server.
  INIT_IDL  IDL used by init mode. Default: idl/service/console/task.thrift.
EOF
}

require_hz() {
  if ! command -v hz >/dev/null 2>&1; then
    echo "hz is required but was not found in PATH." >&2
    echo "Install it with: go install github.com/cloudwego/hertz/cmd/hz@latest" >&2
    exit 1
  fi
}

collect_idls() {
  local dir
  for dir in "$@"; do
    if [[ -d "$dir" ]]; then
      find "$dir" -type f -name '*.thrift' -print
    fi
  done | sort
}

run_server_update() {
  local server_idls=()
  local idl

  while IFS= read -r idl; do
    server_idls+=("$idl")
  done < <(collect_idls "$SERVER_IDL_DIR")

  if [[ ${#server_idls[@]} -eq 0 ]]; then
    echo "No server IDLs found under: $SERVER_IDL_DIR" >&2
    exit 1
  fi

  for idl in "${server_idls[@]}"; do
    echo "hz update --idl $idl"
    hz update --idl "$idl"
  done
}

run_client_update() {
  local downstream_idls=()
  local idl

  while IFS= read -r idl; do
    downstream_idls+=("$idl")
  done < <(collect_idls "$DOWNSTREAM_IDL_DIR")

  if [[ ${#downstream_idls[@]} -eq 0 ]]; then
    echo "No downstream IDLs found under: $DOWNSTREAM_IDL_DIR" >&2
    exit 1
  fi

  for idl in "${downstream_idls[@]}"; do
    echo "hz client --idl $idl --module $MODULE --client_dir $CLIENT_DIR --force_client"
    hz client --idl "$idl" --module "$MODULE" --client_dir "$CLIENT_DIR" --force_client
  done
}

run_init() {
  local server_idls=()
  local init_idl
  local idl

  while IFS= read -r idl; do
    server_idls+=("$idl")
  done < <(collect_idls "$SERVER_IDL_DIR")

  if [[ ${#server_idls[@]} -eq 0 ]]; then
    echo "No server IDLs found under: $SERVER_IDL_DIR" >&2
    exit 1
  fi

  if [[ "${HZ_FORCE_INIT:-}" != "1" && -f "$ROOTDIR/main.go" ]]; then
    echo "This looks like an existing hz project." >&2
    echo "Refusing to run hz new. Set HZ_FORCE_INIT=1 if you really want to run it." >&2
    exit 1
  fi

  init_idl=${INIT_IDL:-$DEFAULT_INIT_IDL}
  if [[ ! -f "$init_idl" ]]; then
    init_idl=${server_idls[0]}
  fi

  echo "hz new --module $MODULE --idl $init_idl"
  hz new --module "$MODULE" --idl "$init_idl"
}

main() {
  local command=${1:-all}

  cd "$ROOTDIR"

  case "$command" in
    -h|--help|help)
      usage
      ;;
    all)
      require_hz
      run_server_update
      run_client_update
      ;;
    server)
      require_hz
      run_server_update
      ;;
    client|downstream)
      require_hz
      run_client_update
      ;;
    init)
      require_hz
      run_init
      ;;
    *)
      usage >&2
      exit 1
      ;;
  esac
}

main "$@"
