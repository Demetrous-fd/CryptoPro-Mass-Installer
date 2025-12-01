# CryptoPro Mass Installer

Приложение для быстрого добавления множества электронных подписей на станциях с Windows и Linux.

[Ссылка на скачивание последней версии программы](https://github.com/Demetrous-fd/CryptoPro-Mass-Installer/releases/latest)

### Требования для запуска

- КриптоПро ЭЦП 4-5 версии

### Как использовать

1. Перенесите пары сертификат/контейнер в папку certs, если она отсутствует создайте.
2. Если требуется установить корневые сертификаты, создайте папку root в папке certs и перенесите сюда корневые сертификаты (.cer/.p7b).
3. Запустите cpmass, пары сертификат/контейнер найдутся и установятся автоматически

### Как использовать с pfx контейнерами

0. [Экспортируйте контейнер в pfx файл](https://support.kontur.ru/ca/55441-ustanovka_pfxfajla)
1. Перенесите пары сертификат/pfx_контейнер в папку certs
2. Создайте и опишите один из файлов установки:

- Создайте файл excel, заполните поля container(название pfx файла/директория контейнера), cert(название файла), pfx_password и сохраните файл с названием `data.csv` формате CSV UTF-8 (Разделитель - запятая) в папке с приложением.
    | **container**        | **cert**       | **pfx_password**    |
    |----------------|----------------|-----------------|
    | Иванов А.И.pfx | Иванов А.И.cer | SomeStrongPass |
    | akimokyv.000 | Петров А.И.cer |  |
    | # Сидоров А.И.pfx | Сидоров А.И.cer | SomeStrongPass |
  
    Если требуется пропустить установку определенной подписи, добавьте в начало поля container "решетку с пробелом" > "# "

- Создайте `settings.json` и перечислите все пары сертификат/контейнер в поле `items` (Перед использованием удалите все строки с комментариями `//`)
```json
{
    "default": {
        "pfxPassword": "SharePass", // Общий пароль для всех pfx контейнеров
        "namePattern": "#subject.surname #subject.initials - #subject.title до #expire_after", // Общий шаблон имени контейнеров
        "exportable": true // Разрешает или запрещает экспорт контейнеров, по умолчанию false
    },
	"items": [
		{
            "containerPath": "Иванов А.И.pfx",
            "certificatePath": "Иванов А.И.cer"
		},
		{
            "containerPath": "akimokyv.000",
            "certificatePath": "Петров А.И.cer",
            "name": "Петров А.И. - Инженер до #expire_after" 
		},
		{
            "containerPath": "PeterPetrovich.pfx",
            "certificatePath": "PeterPetrovich.cer",
            "PfxPassword": "SomeStrongPass",
            "exportable": false
		}
	]
}
```
3. Запустите cpmass

### Шаблонизатор имени контейнера

Пример шаблона: `#subject.surname #subject.initials - #subject.title до #expire_after` -> `Иванов А.И. - Инженер до 11.11.2025`

| **Тег**        | **Описание**       | **Пример**    |
|----------------|----------------|-----------------|
| #expire_before  | Действителен с | 11.11.2024 |
| #expire_after  | Действителен до | 11.11.2025 |
| #subject.common_name или #issuer.common_name  | Общее имя | Иванов Иван Иванович / www.example.com / Название организации |
| #subject.surname или #issuer.surname | Фамилия | Иванов |
| #subject.country_name или #issuer.country_name | Код страны | RU |
| #subject.locality_name или #issuer.locality_name | Город или населённый пункт | г.Москва |
| #subject.state_or_province_name или #issuer.state_or_province_name | Штат или область | Московская область |
| #subject.street_address или #issuer.street_address | Адрес | Большой Златоустинский переулок, д. 6, строение 1 |
| #subject.organization_name или #issuer.organization_name | Название организации | Казначейство России |
| #subject.organizational_unit_name или #issuer.organizational_unit_name | Название структурного подразделения | АСУ |
| #subject.title или #issuer.title | Должность или звание субъекта | Инженер |
| #subject.telephone_number или #issuer.telephone_number | Телефон | 8-800-555-35-35 |
| #subject.name или #issuer.name | - | - |
| #subject.given_name или #issuer.given_name | - | Иван Иванович |
| #subject.initials или #issuer.initials | Инициалы | И.И. |
| #subject.pseudonym или #issuer.pseudonym | - | - |
| #subject.email_address или #issuer.email_address | Email | ivanovii@example.com |

### Файл настроек `settings.json`

- cpmass может работать без файла настроек
- Если поле `items` отсутствует, то пары сертификат/контейнер будут взяты из `data.csv` или будут найдены автоматически

```json
{
      "default": { // Значения по умолчанию
            "namePattern": "#subject.surname #subject.initials - #subject.title до #expire_after", 
            "pfxPassword": "SharePass",
            "exportable": true
      },
      "args": { // Аргументы запуска
            "skipRoot": false,
            "skipWait": false,
            "debug": false
      },
      "items": [ // Описание пар сертификат/контейнер
            {
            "name": "Петров П.П. - Инженер до 11.11.2025",
            "containerPath": "PeterPetrovich.pfx",
            "certificatePath": "PeterPetrovich.cer",
            "PfxPassword": "SomeStrongPass",
            "exportable": false
            },
            ...
      ]
}
```

### Аргументы запуска

```shell
Использование:
  cpmass [flags] <command> [command flags]

Commands:
  install - Установка электронной подписи

Flags:
  -debug
        Включить отладочную информацию в консоли
  -exportable
        Разрешить экспорт контейнеров
  -skip-root
        Пропустить установку корневых сертификатов
  -skip-wait
        Пропустить ожидание перед выходом
  -version
        Отобразить версию программы

Запустите `cpmass <command> -h` чтобы получить справку по определенной команде
```

```shell
Использование:
  cpmass install -cont "..." -cert "..." [flags]

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

### Поддержка проекта
Если вы обнаружили ошибку или хотите предложить идею для улучшения проекта, создайте issue.

Если у вас есть возможность и желание внести улучшения в проект, отправляйте pull request.
