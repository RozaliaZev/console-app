# Консольное приложение

## Применение приложения
- отслеживание изменений в различных директориях
- выполнение произвольного набора консольных команд

## Конфигурирование приложения

Конфигурирование приложения происходит через файф config.yaml

- **path** директория, на изменения в которой мы подписываемся

- **commands** команды, которые автоматически запустятся после изменений в репозитории

- **log_file** файл для логирования, куда будут отправляться логи всех действий (изменения и команды)

- **include_regexp/exclude_regexp** макси дают возможность включать/отключать файлы

```sh
path: /rep
log_file: L:/rep/log2.out
include_regexp:
  - .*.go$
  - .*.env$
exclude_regexp:
  - .*.out$
commands:
  - go build -o app L:/rep/main.go
  - L:/console-app/app
```

Данные об изменениях в отслеживаемых файлах, а также команды сохраняются в БД. Поэтому в файле config.yaml также храним переменные для подключения к Postgres.

```sh
db:
  HOST: host
  PORT: 5432
  USER: postgres
  PASSWORD: password
  DBNAME: base
  SSLMODE: disable
```

## Запуск приложения

после заполнения файла config.yaml можно использовать приложение.

можно сделать билд и запустить его, можно просто воспользоваться командой

```sh
go run main.go
```

Кроме логов можно следить за приложением в консоли, как только приложение будет готово к работе, выведется _start!_

## База данных

Данные об изменениях в отслеживаемых файлах, а также команды сохраняются в БД. 

Данные сохраняются в две таблицы - изменения и команды отдельно.

Примеры заполненных таблиц:

![alt text](https://sun9-43.userapi.com/impg/vCctfSob9m-F0zak7YwPeVOfFkIRFOqmcROqpw/lveOVqkg60w.jpg?size=579x119&quality=95&sign=33955f71eb5303de0bcfe6391e7954e1&type=album)

![alt text](https://sun9-56.userapi.com/impg/mGLTsXQ66AeyeTlZzkyFaSk10FtfiYS_p4rK9A/66-F0rZmUEE.jpg?size=433x153&quality=95&sign=77f11e182c75c5d1b785662780fe443f&type=album)

## Логирование

Пример сохраненных логов:

![alt text](https://sun9-76.userapi.com/impg/aQRdMGB83IOV_MpSVK0DX3pPPEIfCmWKGPqyQg/k7Wf1ZmG9AY.jpg?size=618x99&quality=95&sign=969a0688c71bc2d1af4e8584c0b4ccd8&type=album)

## Завершение работы приложения

При нажатии  Ctrl+C программа завершит свою работу и выведет _Exiting program..._.
