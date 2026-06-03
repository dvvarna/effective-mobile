# effective-mobile

REST-сервис для агрегации данных об онлайн-подписках пользователей.

## Запуск

### Требования
- Go 1.25.5+
- PostgreSQL 18.4+

### Установка
```bash
# 1. Клонировать репозиторий
git clone <repo-url>
cd effective-mobile

# 2. Установить зависимости
go mod tidy

# 3. Создать конфиг
cp subscriptions.env.example subscriptions.env
# Отредактировать реквизиты под свою БД

#4. Сборка в терминале
go build -o bin/effective-mobile.exe ./cmd/server/

#5. Запуск в терминале
.\bin\effective-mobile.exe 

P.S.
Два дня воевал с Docker, но так и не получилось победить 
виртуализацию под Windows 10. Хотя по документации Microsoft
настроил.
Со Swagger так же не удалось сгенерить документацию.
Если судить по данным в интернете, то возможно из за русского 
в описаниях или из за тире в пути или проекте.

