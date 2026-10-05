# template

Лёгкий Go-пакет для парсинга и рендеринга форм ботов MAX Messenger с динамической подстановкой значений и генерацией inline-клавиатур.

## Обзор

Превращайте текстовые шаблоны в насыщенные сообщения бота с интерполяцией переменных и декларативным описанием клавиатур. Без конфигов, без лишних зависимостей — только чистая работа со строками.

## Установка

```sh
go get github.com/max-messenger/max-message-template-go
```

## Быстрый старт

```go
import "github.com/max-messenger/max-message-template-go"

// Рендер сообщения с динамическими значениями
msg := template.ReplaceValues("Привет, {name}!", template.Values{"name": "Мир"})
// => "Привет, Мир!"

// Парсинг полной формы с телом и клавиатурой
body, kbd := template.ParseForm(`
Добро пожаловать, {user}!
---
[Профиль](cb:profile), [Настройки](cb:settings)
[Посетить сайт](https://example.com)
`, template.Values{"user": "Алиса"})
```

## Синтаксис форм

Формы объединяют тело сообщения и раскладку клавиатуры, разделённые `---`:

```
Текст сообщения с {плейсхолдерами}
---
[кнопка1](вызов1), [кнопка2](вызов2)
[кнопка3](вызов3)
```

### Типы кнопок

| Префикс | Тип | Пример |
|---------|-----|--------|
| `cb:` | Callback | `[Голос](cb:vote_yes)` |
| `http(s)://` | Внешняя ссылка | `[Сайт](https://max.ru)` |
| `geo:true\|false` | Геолокация | `[Местоположение](geo:true)` |
| `contact:` | Отправка контакта | `[Поделиться контактом](contact:)` |
| `app:<bot_id>` | Открытие мини-приложения | `[Приложение](app:123456)` |
| `msg:` | Отправка сообщения | `[Ответить](msg:привет)` |
| `clip:` | Копирование в буфер | `[Копировать](clip:secret_code)` |

## API Reference

### `ReplaceValues(template string, values Getter) string`

Заменяет все `{key}` плейсхолдеры в строке шаблона. Отсутствующие ключи остаются без изменений.

```go
template.ReplaceValues("ID: {id}, Статус: {status}", template.Values{"id": 42})
// => "ID: 42, Статус: {status}"
```

### `ParseForm(template string, values Getter) (string, *model.Keyboard)`

Полный парсер форм. Подставляет значения, разделяет тело/клавиатуру и строит совместимую с MAX модель `Keyboard`.

```go
body, kbd := template.ParseForm(`
Заказ #{order_id} подтверждён.
---
[Отследить](cb:track:{order_id}), [Отменить](cb:cancel:{order_id})
`, template.Values{"order_id": 7890})
```

### `SplitForm(s string, sep string) (string, string)`

Разделяет строку на секции тела и клавиатуры по разделителю или первому переносу строки.

### `ParseButton(s string) (string, string)`

Парсит одну кнопку в формате `[имя](вызов)` в имя и полезную нагрузку.

### `SplitButtons(s string) []string`

Разделяет строку кнопок по запятым, обрезая пробелы и фильтруя пустые элементы.

### Интерфейс `Getter`

```go
type Getter interface {
    Get(key string) (any, bool)
}
```

Реализуйте этот интерфейс для предоставления пользовательских источников значений. `Values` (алиас map) — реализация по умолчанию.

## Примеры

### Динамическое уведомление о заказе

```go
tmpl := `
**Заказ {order_id}** — {status}
Товаров: {count} | Итого: {total}₽
---
[Детали](cb:order:{order_id}), [Поддержка](msg:Помощь по заказу {order_id})
[Отследить доставку](https://track.example.com/{order_id})
`

values := template.Values{
    "order_id": 12345,
    "status":   "отправлен",
    "count":    3,
    "total":    1299,
}

body, kbd := template.ParseForm(tmpl, values)
```

### Простое callback-меню

```go
body, kbd := template.ParseForm(`
Главное меню
---
[Дашборд](cb:menu:dash), [Отчёты](cb:menu:reports)
[Настройки](cb:menu:settings), [Выход](cb:auth:logout)
`, nil)
```

## Требования

- Go 1.24+ (используется `strings.SplitSeq`)
- `github.com/max-messenger/max-bot-api-client-go/v2`

## Тесты

```sh
go test ./...
```
