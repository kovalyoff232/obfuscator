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
go build -ldflags="-s -w" -o ./obfuscated_src/obfuscated_payload ./obfuscated_src/
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

