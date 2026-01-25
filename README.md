# Robotogo API

HTTP API для работы с robotgo - библиотекой автоматизации мыши и клавиатуры на Go.

## Установка

```bash
go mod download
```

## Настройка

Создайте файл `.env` (опционально):

```env
PORT=3000
ENVIRONMENT=development
```

## Запуск

```bash
go run cmd/main.go
```

## API Endpoints

Все endpoints находятся под префиксом `/api/robotogo`

### GET /api/robotogo/mouse/position

Возвращает текущую позицию мыши.

**Response:**
```json
{
  "success": true,
  "x": 100,
  "y": 200
}
```

### POST /api/robotogo/mouse/move

Перемещает мышь на указанные координаты.

**Request:**
```json
{
  "x": 100,
  "y": 200
}
```

**Response:**
```json
{
  "success": true,
  "message": "Мышь перемещена на (100, 200)",
  "x": 100,
  "y": 200
}
```

### POST /api/robotogo/mouse/click

Выполняет клик мышью.

**Request:**
```json
{
  "x": 100,
  "y": 200,
  "button": "left"
}
```

**Параметры:**
- `x`, `y` (опционально) - координаты для клика. Если не указаны, клик выполняется на текущей позиции
- `button` (опционально) - кнопка мыши: `left`, `right`, `center` (по умолчанию `left`)

**Response:**
```json
{
  "success": true,
  "message": "Клик выполнен на (100, 200) (кнопка: left)"
}
```

### POST /api/robotogo/keyboard/type

Вводит текст.

**Request:**
```json
{
  "text": "Hello World",
  "x": 100,
  "y": 200,
  "delay_ms": 30
}
```

**Параметры:**
- `text` (обязательно) - текст для ввода
- `x`, `y` (опционально) - координаты для клика перед вводом
- `delay_ms` (опционально) - задержка между символами в миллисекундах

**Response:**
```json
{
  "success": true,
  "message": "Текст введен: Hello World",
  "text": "Hello World"
}
```

### POST /api/robotogo/input

Выполняет полный цикл: клик по координатам и ввод текста.

**Request:**
```json
{
  "x": 100,
  "y": 200,
  "text": "Hello World",
  "clear_before_input": true,
  "click_delay_ms": 100,
  "type_delay_ms": 30
}
```

**Параметры:**
- `x`, `y` (обязательно) - координаты
- `text` (обязательно) - текст для ввода
- `clear_before_input` (опционально) - очистить поле перед вводом (по умолчанию `true`)
- `click_delay_ms` (опционально) - задержка после клика (по умолчанию 100 мс)
- `type_delay_ms` (опционально) - задержка между символами (по умолчанию 30 мс)

**Response:**
```json
{
  "success": true,
  "message": "Данные введены на (100, 200): Hello World",
  "x": 100,
  "y": 200,
  "text": "Hello World"
}
```

### POST /api/robotogo/fill-and-click

Выполняет полный цикл: наведение на инпут, очистка поля, ввод текста, затем клик по кнопке.

**Request:**
```json
{
  "input_x": 100,
  "input_y": 200,
  "text": "Hello World",
  "button_x": 300,
  "button_y": 400,
  "button": "left",
  "clear_before_input": true,
  "click_delay_ms": 100,
  "type_delay_ms": 30
}
```

**Параметры:**
- `input_x`, `input_y` (обязательно) - координаты инпута
- `text` (обязательно) - текст для ввода
- `button_x`, `button_y` (обязательно) - координаты кнопки
- `button` (опционально) - кнопка мыши для клика: `left`, `right`, `center` (по умолчанию `left`)
- `clear_before_input` (опционально) - очистить поле перед вводом. Если не указано, по умолчанию `true`. Чтобы отключить очистку, укажите `false`
- `click_delay_ms` (опционально) - задержка после клика (по умолчанию 100 мс)
- `type_delay_ms` (опционально) - задержка между символами (по умолчанию 30 мс)

**Response:**
```json
{
  "success": true,
  "message": "Текст 'Hello World' введен в инпут (100, 200) и выполнен клик по кнопке (300, 400)",
  "input": {
    "x": 100,
    "y": 200
  },
  "text": "Hello World",
  "button": {
    "x": 300,
    "y": 400,
    "button": "left"
  }
}
```

## Сборка

```bash
go build -o bin/app cmd/main.go
```
