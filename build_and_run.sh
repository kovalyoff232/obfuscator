#!/bin/bash

echo "Сборка проекта..."
go build -ldflags="-s -w" -o obfuscator_cli .
if [ $? -ne 0 ]; then
    echo "Ошибка сборки."
    exit 1
fi
echo "Сборка завершена."

echo -e "\nЗапуск обфускатора..."
./obfuscator_cli -input ./example_src/kvstore -output ./obfuscated_src
if [ $? -ne 0 ]; then
    echo "Ошибка выполнения обфускатора."
    exit 1
fi
echo "Обфускация завершена."

echo -e "\nСодержимое обфусцированного файла (obfuscated_src/main.go):"
cat ./obfuscated_src/main.go

echo -e "\n\nПопытка собрать и запустить обфусцированный код..."
# Сначала удалим старую базу данных, если она есть
rm -f ./kvstore.db
rm -f ./obfuscated_src/kvstore.db

# Собираем и запускаем с командой 'set'
echo "Тест 1: Установка значения"
go build -ldflags="-s -w" -o ./obfuscated_src/obfuscated_payload ./obfuscated_src/main.go
if [ $? -ne 0 ]; then
    echo "Ошибка сборки обфусцированного кода."
    exit 1
fi
echo "Обфусцированный код успешно собран."

# Меняем директорию на ту, где находится бинарник
cd ./obfuscated_src

./obfuscated_payload set mykey 'my secret value'
if [ $? -ne 0 ]; then
    echo "Ошибка запуска обфусцированного кода (set)."
    cd .. # Возвращаемся обратно
    exit 1
fi

# Запускаем с командой 'get'
echo -e "\nТест 2: Получение значения"
./obfuscated_payload get mykey
if [ $? -ne 0 ]; then
    echo "Ошибка запуска обфусцированного кода (get)."
    cd .. # Возвращаемся обратно
    exit 1
fi


cd .. # Возвращаемся в корневую директорию проекта
echo -e "\n\nСкрипт успешно выполнен."

echo -e "\n\n--- Проверка на VirusTotal ---"

if [ -z "$VT_API_KEY" ]; then
    echo "Переменная окружения VT_API_KEY не установлена. Пропускаю шаг с VirusTotal."
    echo "Для проверки установите ключ: export VT_API_KEY='ваш_ключ'"
    exit 0
fi

OBFUSCATED_BINARY="./obfuscated_src/obfuscated_payload"

if ! command -v jq &> /dev/null
then
    echo "Утилита 'jq' не найдена. Пожалуйста, установите ее для парсинга ответа от VirusTotal."
    exit 1
fi

echo "Загружаю файл $OBFUSCATED_BINARY на VirusTotal..."

UPLOAD_RESPONSE=$(curl --silent --request POST \
  --url https://www.virustotal.com/api/v3/files \
  --header "x-apikey: $VT_API_KEY" \
  --form file=@"$OBFUSCATED_BINARY")

ANALYSIS_ID=$(echo "$UPLOAD_RESPONSE" | jq -r .data.id)

if [ -z "$ANALYSIS_ID" ] || [ "$ANALYSIS_ID" == "null" ]; then
    echo "Не удалось получить ID анализа. Ответ от VirusTotal:"
    echo "$UPLOAD_RESPONSE"
    exit 1
fi

echo "Файл успешно загружен. ID анализа: $ANALYSIS_ID"
echo "Ожидание отчета..."

while true; do
    ANALYSIS_REPORT=$(curl --silent --request GET \
      --url "https://www.virustotal.com/api/v3/analyses/$ANALYSIS_ID" \
      --header "x-apikey: $VT_API_KEY")
    
    STATUS=$(echo "$ANALYSIS_REPORT" | jq -r .data.attributes.status)

    if [ "$STATUS" == "completed" ]; then
        echo "Анализ завершен."
        break
    fi
    
    echo "Статус анализа: $STATUS. Ожидаю 15 секунд..."
    sleep 15
done

echo -e "\n--- Отчет VirusTotal ---"
STATS=$(echo "$ANALYSIS_REPORT" | jq .data.attributes.stats)
MALICIOUS=$(echo "$STATS" | jq .malicious)
SUSPICIOUS=$(echo "$STATS" | jq .suspicious)
UNDETECTED=$(echo "$STATS" | jq .undetected)

echo "Результаты сканирования:"
echo "  - Вредоносный: $MALICIOUS"
echo "  - Подозрительный: $SUSPICIOUS"
echo "  - Не обнаружено: $UNDETECTED"

if [ "$MALICIOUS" -gt 0 ] || [ "$SUSPICIOUS" -gt 0 ]; then
    echo -e "\n[ПРЕДУПРЕЖДЕНИЕ] Файл был помечен антивирусами."
else
    echo -e "\n[OK] Файл чист. Ни один антивирус не счел его вредоносным."
fi