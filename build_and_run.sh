#!/bin/bash
set -euo pipefail
INPUT=""
OUTPUT="./obfuscated_src"
PROFILE="safe"
LOG="info"
DISABLE_ANTI_VM=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    -i|--input)
      INPUT="$2"; shift 2 ;;
    -o|--output)
      OUTPUT="$2"; shift 2 ;;
    --profile)
      PROFILE="$2"; shift 2 ;;
    --log)
      LOG="$2"; shift 2 ;;
    --disable-anti-vm)
      DISABLE_ANTI_VM="1"; shift 1 ;;
    *)
      echo "Неизвестный аргумент: $1"
      echo "Пример: $0 -i ./some/project -o ./obfuscated_src --profile safe --log info [--disable-anti-vm]"
      exit 1 ;;
  esac
done
RENAME=true
ENCRYPT_STRINGS=true
INSERT_DEAD_CODE=true
OBF_FLOW=true
OBF_EXPR=true
OBF_DATA=true
OBF_CONST=true
ANTI_DEBUG=true
ANTI_VM=true
INDIRECT=true
INTEGRITY=true
META=true
SELF_MOD=true
case "${PROFILE}" in
  fast)
    INSERT_DEAD_CODE=false
    OBF_FLOW=false
    OBF_EXPR=false
    OBF_DATA=false
    OBF_CONST=false
    INDIRECT=false
    INTEGRITY=false
    META=false
    SELF_MOD=false
    ;;
  safe)
    ;;
  aggressive)
    ;;
  *)
    echo "[warn] Неизвестный профиль: ${PROFILE}. Используется safe."
    ;;
esac
echo "Сборка CLI..."
go build -ldflags="-s -w" -o obfuscator_cli .
echo "CLI собран."
if [[ -n "${INPUT}" ]]; then
  echo -e "\nЗапуск обфускатора..."
  DISABLE_FLAG=""
  if [[ "${DISABLE_ANTI_VM}" == "1" ]]; then
    DISABLE_FLAG="-disable-anti-vm"
    export OBF_DISABLE_ANTI_VM=1
  fi
  set -x
  ./obfuscator_cli \
    -input "${INPUT}" -output "${OUTPUT}" \
    -rename="${RENAME}" \
    -encrypt-strings="${ENCRYPT_STRINGS}" \
    -insert-dead-code="${INSERT_DEAD_CODE}" \
    -obfuscate-control-flow="${OBF_FLOW}" \
    -obfuscate-expressions="${OBF_EXPR}" \
    -obfuscate-data-flow="${OBF_DATA}" \
    -obfuscate-constants="${OBF_CONST}" \
    -anti-debug="${ANTI_DEBUG}" \
    -anti-vm="${ANTI_VM}" ${DISABLE_FLAG} \
    -indirect-calls="${INDIRECT}" \
    -weave-integrity="${INTEGRITY}" \
    -metamorphic="${META}" \
    -self-modifying="${SELF_MOD}"
  set +x
  echo "Обфускация завершена."
  if [[ -f "${OUTPUT}/main.go" ]]; then
    echo -e "\nСодержимое обфусцированного файла (${OUTPUT}/main.go):"
    sed -n '1,200p' "${OUTPUT}/main.go" || true
  fi
  echo -e "\nСборка обфусцированного кода (если это самостоятельный модуль)..."
  if go build -ldflags="-s -w" -o "${OUTPUT}/obfuscated_payload" "${OUTPUT}/" 2>/dev/null; then
    echo "Обфусцированный код успешно собран."
    ( cd "${OUTPUT}" && ./obfuscated_payload || true )
  else
    echo "Пропуск сборки обфусцированного кода: нет самостоятельного модуля или зависимости вне контекста."
  fi
else
  echo "Входной путь (-i/--input) не задан — шаг обфускации пропущен."
fi
echo -e "\nДемонстрационный запуск payload:"
PAYLOAD_PASSWORD=${PAYLOAD_PASSWORD:-} go run ./cmd/payload
echo -e "\nСкрипт успешно выполнен."
