# CryptoPro Mass Installer
Приложение для быстрого добавления множества сертификатов на станциях с Windows и Linux.

### Требования для запуска

- КриптоПро ЭЦП 4-5 версии

### Проблемы с Windows defender

Windows Defender обнаруживает приложение как "Trojan:Win32/Phonzy.C!ml". Это распространённая проблема с программами на Golang: https://www.reddit.com/r/golang/comments/s1bh01/goexecutables_and_windows_defender/

Решение проблемы: Добавьте папку приложения в исключения

### Как использовать

0. [Экспортируйте сертификат с закрытом ключом в pfx файл](https://support.kontur.ru/ca/38782-kopirovanie_kontejnera_s_sertifikatom_na_dr#header_ad9459fa9)
1. Перенесите сертификаты и pfx файлы в папку certs, если она отсутствует создайте
2. Создайте файл excel, заполните поля pfx(название файла), cert(название файла), password и сохраните файл с названием "data" формате CSV UTF-8 (Разделитель - запятая) в папке с приложением.
    | **pfx**        | **cert**       | **password**    |
    |----------------|----------------|-----------------|
    | Иванов А.И.pfx | Иванов А.И.cer | SomeStrongPass |
    | Петров А.И.pfx | Петров А.И.cer | SomeStrongPass |
    | # Сидоров А.И.pfx | Сидоров А.И.cer | SomeStrongPass |
  
    Если требуется пропустить установку определенного сертификата, добавьте в начало поля pfx "решетку с пробелом" > "# "

3. Запустите mass

### Аргументы запуска
- ```--fast``` Запускается множества экземпляров certmgr, ускоряя установку сертификатов, но возможны ошибки во время работы
- `--debug`  Отладочная информация 

### Какие действия выполняются

1. Чтение данных из data.csv
2. Получение отпечатка/thumbprint сертификата из cer файла
3. Поиск сертификата в хранилище, если сертификат существует, то установка пропускается
   ```shell
   certmgr -list -thumbprint 0000000000000000000000000000000000000000
   ```
4. Установка закрытого ключа из pfx файла
   ```shell
   certmgr -inst -pfx -file "certs/Иванов А.И.pfx" -pin SomeStrongPass -silent
   или
   certmgr -inst -pfx -file "certs/Иванов А.И.pfx" -pin SomeStrongPass # Windows 7 и ниже
   ```
5. Установка сертификата в контейнер
   ```
   certmgr -inst -inst_to_cont -file "certs/Иванов А.И.cer" -cont \\.\Container\path -silent
   ```
6. Если возникла ошибка во время установки, то данные удаляются
   ```
   certmgr -delete -certificate -thumbprint 0000000000000000000000000000000000000000
   certmgr -delete -container \\.\Container\path
   ```
