#!/bin/bash
set -euo pipefail



INPUT=""
OUTPUT="./obfuscated_src"
PROFILE="safe"
LOG="info"

while [[ $# -gt 0 ]]; do
  case "$1" in
    -i|--input)
      INPUT="$2"
      shift 2
      ;;
    -o|--output)
      OUTPUT="$2"
      shift 2
      ;;
    --profile)
      PROFILE="$2"
      shift 2
      ;;
    --log)
      LOG="$2"
      shift 2
      ;;
    *)
      echo "Неизвестный аргумент: $1"
      echo "Пример: $0 -i ./some/project -o ./obfuscated_src --profile safe --log info"
      exit 1
      ;;
  esac
done

echo "Сборка CLI..."
go build -ldflags="-s -w" -o obfuscator_cli .
echo "CLI собран."

if [[ -n "${INPUT}" ]]; then
  echo -e "\nЗапуск обфускатора..."
  ./obfuscator_cli -input "${INPUT}" -output "${OUTPUT}" -profile "${PROFILE}" -log "${LOG}" || {
    echo "Ошибка выполнения обфускатора."
    exit 1
  }
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
