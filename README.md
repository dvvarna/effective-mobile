# effective-mobile

REST-сервис для агрегации данных об онлайн-подписках пользователей.

## Запуск

### Требования
- Go 1.25.5+
- PostgreSQL 18.4+

## Установка

### 1. Клонировать репозиторий
```bash

git clone <repo-url>
cd effective-mobile
```
### 2. Установить зависимости
```bash

go mod tidy
```
### 3. Создать конфиг
```bash

cp subscriptions.env.example subscriptions.env
```
Отредактировать реквизиты под свою БД

### 4. Сборка в терминале
```bash

go build -o bin/effective-mobile.exe ./cmd/server/
```
### 5. Запуск в терминале
```bash

.\bin\effective-mobile.exe 
```
----
**P.S.**<br>
Два дня воевал с Docker, но так и не получилось победить 
виртуализацию под Windows 10. Хотя по документации Microsoft
настроил.<br>
Со Swagger так же не удалось сгенерить документацию.
Если судить по данным в интернете, то возможно из за русского 
в описаниях или из за тире в пути или проекте.

