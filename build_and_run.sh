#!/bin/bash

# Скрипт для сборки и запуска обфускатора с примером использования.

# 1. Сборка проекта
echo "Сборка проекта..."
go build -o obfuscator_cli .
if [ $? -ne 0 ]; then
    echo "Ошибка сборки."
    exit 1
fi
echo "Сборка завершена."

# 2. Подготовка тестовой директории
echo -e "\nПодготовка тестового исходного кода..."
mkdir -p ./example_src/pkg
cat > ./example_src/main.go << EOL
package main

import "fmt"

func main() {
	secretMessage := "Это секретное сообщение"
	result := add(10, 20)
	fmt.Printf("%s: %d\n", secretMessage, result)
}

func add(a, b int) int {
	return a + b
}
EOL

echo "Тестовый код создан."

# 3. Запуск обфускатора
echo -e "\nЗапуск обфускатора..."
./obfuscator_cli -input ./example_src -output ./obfuscated_src
if [ $? -ne 0 ]; then
    echo "Ошибка выполнения обфускатора."
    exit 1
fi
echo "Обфускация завершена."

# 4. Проверка результата
echo -e "\nСодержимое обфусцированного файла (obfuscated_src/main.go):"
cat ./obfuscated_src/main.go

echo -e "\n\nПопытка собрать и запустить обфусцированный код..."
# Сначала соберем бинарник из обфусцированного кода
go build -o ./obfuscated_src/obfuscated_payload ./obfuscated_src/main.go
if [ $? -ne 0 ]; then
    echo "Ошибка сборки обфусцированного кода."
    exit 1
fi
echo "Обфусцированный код успешно собран."

# Запускаем его для проверки
./obfuscated_src/obfuscated_payload
if [ $? -ne 0 ]; then
    echo "Ошибка запуска обфусцированного кода."
    exit 1
fi

echo -e "\n\nСкрипт успешно выполнен."

# 5. Загрузка на VirusTotal
echo -e "\n\n--- Проверка на VirusTotal ---"
VT_API_KEY="93d38060166cefe2b415667a010a7cedec7be98db5070b3f359c46f06f3ff5a5"
OBFUSCATED_BINARY="./obfuscated_src/obfuscated_payload"

if ! command -v jq &> /dev/null
then
    echo "Утилита 'jq' не найдена. Пожалуйста, установите ее для парсинга ответа от VirusTotal."
    exit 1
fi

echo "Загружаю файл $OBFUSCATED_BINARY на VirusTotal..."

# Шаг 1: Загрузка файла
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

# Шаг 2: Получение отчета по ID анализа
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

# Шаг 3: Вывод результата
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
