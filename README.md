# GoFlow

Учебный симулятор городского движения на четырёх перекрёстках. Сервер на Go отвечает за симуляцию и JSON API. Интерфейс на go-app компилируется в WebAssembly и работает в браузере.

В каркасе первого дня есть стартовая страница с кнопкой проверки WebAssembly и `GET /api/state` с фиксированным снимком. Движение машин и команды управления появятся после подключения модулей команды.

## Запуск на Windows

Требуются Go 1.26 или новее, браузер с поддержкой WebAssembly и Git для совместной разработки. В `go.mod` закреплён go-app `v11.0.5`.

В корне проекта в PowerShell выполните следующее.

```powershell
go version
go mod tidy
if ($LASTEXITCODE -ne 0) { throw 'Не удалось загрузить зависимости Go' }

$oldOS = $env:GOOS
$oldArch = $env:GOARCH
try {
    $env:GOOS = 'js'
    $env:GOARCH = 'wasm'
    go build -o web/app.wasm ./cmd/goflow
    if ($LASTEXITCODE -ne 0) { throw 'Не удалось собрать WebAssembly' }
} finally {
    $env:GOOS = $oldOS
    $env:GOARCH = $oldArch
}

go run ./cmd/goflow
```

Откройте `http://127.0.0.1:8080/`. После загрузки нажмите «Проверка WebAssembly». Если кнопка показывает «WebAssembly работает», обработчик Go выполнился в браузере. При изменении кода интерфейса пересоберите `web/app.wasm` и перезапустите сервер.

В другой консоли можно проверить тестовый снимок.

```powershell
Invoke-RestMethod http://127.0.0.1:8080/api/state | ConvertTo-Json -Depth 5
```

## Работа команды

- [Архитектура](docs/ARCHITECTURE.md)
- [HTTP API и пример Snapshot](docs/API.md)
- [План и статус команды](docs/SPRINT.md)

Ветки для работы команды — `feat/api`, `feat/simulation`, `feat/control`, `feat/ui`. Общие структуры и API меняем после согласования с капитаном. Сборку `web/app.wasm` хранить в Git не требуется.
