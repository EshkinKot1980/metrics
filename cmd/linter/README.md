## Линтер 

Содержит кастомный анализатор __FatalCheckAnalyzer__ проверяющий вызов
__panic()__, __os.Exit()__ и __log.Fatal()__.

__os.Exit()__ и __log.Fatal()__ разрешено вызывать в функции main() пакета main,
__panic()__ — запрещено везде.

### Сборка и запуск
#### Linux

```bash
# cd /path/to/project/dir

# сборка
cd cmd/linter \
    && go build -buildvcs=false -o linter \
    && cd -

# запуск
cmd/linter/linter -- ./...
# запуск из исходников
go run cmd/linter/main.go -- ./...

```