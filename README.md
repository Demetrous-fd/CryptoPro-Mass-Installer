# CryptoPro Mass Installer
Приложение для быстрого добавления множества электронных подписей на станциях с Windows и Linux.

[Ссылка на скачивание последней версии программы](https://github.com/Demetrous-fd/CryptoPro-Mass-Installer/releases/latest)

### Требования для запуска

- КриптоПро ЭЦП 4-5 версии

### Как использовать

0. Не обязательно: [Экспортируйте контейнер в pfx файл](https://support.kontur.ru/ca/38782-kopirovanie_kontejnera_s_sertifikatom_na_dr#header_ad9459fa9)
1. Перенесите сертификаты и контейнеры(pfx файлы или директории) в папку certs, если она отсутствует создайте.
2. Если требуется установить корневые сертификаты, создайте папку root в папке certs и перенесите сюда корневые сертификаты (.cer/.p7b).
3. Создайте файл excel, заполните поля container(название pfx файла/директория контейнера), cert(название файла), pfx_password и сохраните файл с названием "data" формате CSV UTF-8 (Разделитель - запятая) в папке с приложением.
    | **container**        | **cert**       | **pfx_password**    |
    |----------------|----------------|-----------------|
    | Иванов А.И.pfx | Иванов А.И.cer | SomeStrongPass |
    | akimokyv.000 | Петров А.И.cer |  |
    | # Сидоров А.И.pfx | Сидоров А.И.cer | SomeStrongPass |
  
    Если требуется пропустить установку определенной подписи, добавьте в начало поля container "решетку с пробелом" > "# "

4. Запустите mass

### Аргументы запуска
```shell
Использование:
  mass [flags] <command> [command flags]

Commands:
  install - Установка электронной подписи
  export - Экспортирование электронной подписи в pfx

Flags:
  -debug
        Включить отладочную информацию в консоли
  -exportable
        Разрешить экспорт контейнеров
  -skip-root
        Пропустить установку корневых сертификатов
  -version
        Отобразить версию программы
  -wait
        Перед выходом ожидать нажатия клавиши enter (default true)

Запустите `mass <command> -h` чтобы получить справку по определенной команде
```

```shell
Использование:
  mass install -cont "..." -cert "..." [flags]

Flags:
  -cert string
        [Требуется] Путь до файла сертификата
  -cont string
        [Требуется] Путь до pfx/папки контейнера
  -name string
        Название контейнера
  -pfx_pass string
        Пароль от pfx контейнера
```

```shell
Использование:
  mass export -cont "..." [flags]

Flags:
  -cont string
        [Требуется] Название контейнера или путь до папки
  -cert string
        Путь до сертификата
  -name string
        Новое название контейнера
  -o string
        Путь до нового pfx контейнера
  -pass string
        Пароль от pfx контейнера

```

### Поддержка проекта
Если вы обнаружили ошибку или хотите предложить идею для улучшения проекта, создайте issue.

Если у вас есть возможность и желание внести улучшения в проект, отправляйте pull request.
