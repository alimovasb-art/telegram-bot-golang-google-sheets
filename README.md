# Telegram Attendance Bot

Проект реализует production-ready Telegram Attendance Bot на Go 1.25+, без базы данных. Все данные хранятся в Google Sheets.

## Функции

- Регистрация пользователя через `/start`
- Сохранение данных в листе `Users`:
  - Telegram ID
  - Username
  - FullName
  - Group
- Главное меню с инлайн-кнопками:
  - ❌ Не приду
  - ⏰ Опоздаю
  - 🏥 Болею
  - 💻 Зайду на онлайн урок
- Автоматическое сохранение заявок в лист `Attendance`
- Команда `/profile` для просмотра профиля
- FSM-логика состояний регистрации и подачи заявок
- Логирование и обработка ошибок Google Sheets
- Запуск локально и через Docker

## Структура проекта

```
attendance-bot/
├── cmd/
│   └── bot/
│       └── main.go
├── configs/
│   └── .gitkeep
├── internal/
│   ├── app/
│   │   └── app.go
│   ├── config/
│   │   └── config.go
│   ├── fsm/
│   │   └── fsm.go
│   ├── handlers/
│   │   └── handlers.go
│   ├── models/
│   │   └── models.go
│   ├── sheets/
│   │   └── sheets.go
│   └── telegram/
│       └── keyboard.go
├── .env.example
├── credentials.json.example
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
├── main.go
└── README.md
```

## Пререквизиты

- Go 1.25+
- Docker и Docker Compose
- Telegram бот от BotFather
- Google Cloud аккаунт
- Google Sheets с двумя листами: `Users` и `Attendance`

## 1. Создание Telegram бота через BotFather

1. Откройте Telegram и найдите @BotFather.
2. Отправьте команду `/newbot`.
3. Укажите имя бота и username.
4. Получите токен API, например `123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11`.

## 2. Настройка Google Cloud

1. Перейдите в Google Cloud Console.
2. Создайте новый проект или используйте существующий.
3. Включите API `Google Sheets API`.

## 3. Создание Service Account

1. Откройте `IAM & Admin` → `Service Accounts`.
2. Создайте новый сервисный аккаунт.
3. Добавьте роль `Editor` или `Sheets API Editor`.
4. Создайте JSON-ключ и скачайте `credentials.json`.

## 4. Получение credentials.json

Сохраните файл `credentials.json` в корневой директории проекта или в другом месте и укажите путь в `.env`.

## 5. Настройка Google Sheets

1. Создайте новую таблицу Google Sheets.
2. Создайте лист `Users` и лист `Attendance`.
3. Вкладка `Users` должна содержать заголовки:
   - TelegramID
   - Username
   - FullName
   - Group
4. Вкладка `Attendance` должна содержать заголовки:
   - Date
   - Time
   - TelegramID
   - Username
   - FullName
   - Group
   - Status
   - Reason
5. Предоставьте доступ сервисному аккаунту (email из `client_email`) на просмотр и редактирование таблицы.

## 6. Настройка .env

Скопируйте `.env.example` в `.env` и заполните:

```env
TELEGRAM_BOT_TOKEN=your-telegram-bot-token
GOOGLE_CREDENTIALS_PATH=credentials.json
GOOGLE_SPREADSHEET_ID=your-google-spreadsheet-id
```

## 7. Запуск локально

1. Убедитесь, что `credentials.json` находится рядом с `.env`.
2. Запустите команду:

```bash
go run ./cmd/bot
```

## 8. Запуск через Docker

1. Постройте образ:

```bash
docker compose up -d --build
```

2. Убедитесь, что файлы `credentials.json` и `.env` доступны контейнеру через тома.

## 9. Деплой на Oracle Cloud Free Tier

1. Зарегистрируйтесь на Oracle Cloud Free Tier.
2. Создайте виртуальную машину Linux.
3. Установите Docker и Docker Compose на VM.
4. Загрузите проект на сервер (git clone / scp).
5. Разместите `credentials.json` и `.env` на сервере.
6. Запустите:

```bash
docker compose up -d --build
```

> На Oracle Cloud сервисы должны иметь доступ к интернету для Telegram и Google Sheets.

## 10. Команды для пользователей

- `/start` — регистрация и начало работы.
- `/profile` — показывает профиль пользователя.
- `/change_name` — изменить ваше ФИО, если оно указано неправильно.
- `/help` — показывает список доступных команд.

## Прием заявок

После выбора статуса бот обязательно запросит причину. Без причины заявка не сохраняется.

## Дополнительные заметки

- Вся информация хранится только в Google Sheets.
- Администратор просматривает данные через Google Sheets, Telegram-команд для администратора нет.
