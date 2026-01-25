package input

import (
	"fmt"
	"runtime"
	"time"

	"github.com/go-vgo/robotgo"
	"go.uber.org/zap"
)

type Service struct {
	logger *zap.Logger
}

func NewService(logger *zap.Logger) *Service {
	return &Service{
		logger: logger,
	}
}

// MoveMouse перемещает мышь на указанные координаты
func (s *Service) MoveMouse(x, y int) error {
	s.logger.Info("Перемещение мыши", zap.Int("x", x), zap.Int("y", y))
	robotgo.MoveMouse(x, y)
	return nil
}

// Click выполняет клик мышью на текущей позиции
func (s *Service) Click(button string) error {
	s.logger.Info("Клик мышью", zap.String("button", button))
	
	switch button {
	case "left":
		robotgo.Click("left")
	case "right":
		robotgo.Click("right")
	case "center", "middle":
		robotgo.Click("center")
	default:
		robotgo.Click("left")
	}
	
	return nil
}

// ClickAt выполняет клик мышью на указанных координатах
func (s *Service) ClickAt(x, y int, button string) error {
	s.logger.Info("Клик мышью по координатам", 
		zap.Int("x", x), 
		zap.Int("y", y), 
		zap.String("button", button))
	
	robotgo.MoveMouse(x, y)
	time.Sleep(50 * time.Millisecond) // Небольшая задержка перед кликом
	
	return s.Click(button)
}

// TypeText вводит текст
func (s *Service) TypeText(text string, delayMs int) error {
	s.logger.Info("Ввод текста", 
		zap.String("text", text), 
		zap.Int("delay_ms", delayMs),
		zap.String("os", runtime.GOOS),
		zap.Int("text_length", len(text)))
	
	if text == "" {
		s.logger.Warn("Пустой текст для ввода")
		return nil
	}
	
	// На Windows рекомендуется использовать задержку между символами
	if delayMs <= 0 {
		if runtime.GOOS == "windows" {
			delayMs = 30 // Увеличена задержка для Windows
		} else {
			delayMs = 10 // Минимальная задержка для других ОС
		}
	}
	
	// На Windows пробуем несколько методов
	if runtime.GOOS == "windows" {
		s.logger.Info("Windows: начинаем ввод текста", zap.String("method", "PasteStr_first"))
		
		// Сначала пробуем PasteStr (самый надежный метод на Windows, использует буфер обмена)
		robotgo.PasteStr(text)
		time.Sleep(200 * time.Millisecond) // Даем время на вставку
		s.logger.Info("✅ Текст введен через PasteStr (Windows)", zap.String("text", text))
		return nil
	}
	
	// Для macOS и Linux используем стандартный метод
	robotgo.TypeStr(text, delayMs)
	s.logger.Info("Текст введен через TypeStr", zap.String("text", text))
	
	return nil
}

// typeTextViaClipboard вводит текст через буфер обмена (Ctrl+V)
func (s *Service) typeTextViaClipboard(text string) error {
	s.logger.Debug("Начало ввода через буфер обмена", zap.String("text", text))
	
	// Сохраняем текущий буфер обмена
	oldClip, err := robotgo.ReadAll()
	if err != nil {
		s.logger.Debug("Не удалось прочитать буфер обмена (не критично)", zap.Error(err))
	} else {
		s.logger.Debug("Буфер обмена сохранен", zap.Int("old_length", len(oldClip)))
	}
	
	// Копируем текст в буфер обмена
	s.logger.Debug("Копирование текста в буфер обмена")
	if err := robotgo.WriteAll(text); err != nil {
		s.logger.Error("Ошибка записи в буфер обмена", zap.Error(err))
		return fmt.Errorf("ошибка записи в буфер обмена: %w", err)
	}
	s.logger.Debug("Текст скопирован в буфер обмена")
	
	time.Sleep(50 * time.Millisecond)
	
	// Вставляем через Ctrl+V
	s.logger.Debug("Вставка через Ctrl+V")
	robotgo.KeyToggle("ctrl", "down")
	time.Sleep(20 * time.Millisecond)
	robotgo.KeyTap("v")
	time.Sleep(20 * time.Millisecond)
	robotgo.KeyToggle("ctrl", "up")
	time.Sleep(50 * time.Millisecond)
	s.logger.Debug("Вставка выполнена")
	
	// Восстанавливаем старый буфер обмена (если был)
	if oldClip != "" {
		time.Sleep(100 * time.Millisecond)
		robotgo.WriteAll(oldClip)
		s.logger.Debug("Буфер обмена восстановлен")
	}
	
	return nil
}

// typeTextCharByChar вводит текст посимвольно (более надежно на Windows)
func (s *Service) typeTextCharByChar(text string, delayMs int) error {
	s.logger.Debug("Ввод текста посимвольно", zap.Int("length", len(text)), zap.Int("delay_ms", delayMs))
	
	// Убеждаемся, что есть минимальная задержка
	if delayMs <= 0 {
		delayMs = 50
	}
	
	for i, char := range text {
		charStr := string(char)
		
		// Специальная обработка для некоторых символов
		if char == '\n' {
			robotgo.KeyTap("enter")
			s.logger.Debug("Введен символ: Enter")
		} else if char == '\t' {
			robotgo.KeyTap("tab")
			s.logger.Debug("Введен символ: Tab")
		} else if char == ' ' {
			robotgo.KeyTap("space")
			s.logger.Debug("Введен символ: Space")
		} else {
			// Пробуем ввести символ через TypeStr (один символ)
			robotgo.TypeStr(charStr, 0)
			s.logger.Debug("Введен символ", zap.String("char", charStr), zap.Int("position", i+1), zap.Int("total", len(text)))
		}
		
		// Задержка между символами (кроме последнего)
		if i < len(text)-1 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
	}
	
	s.logger.Info("Текст введен посимвольно", zap.String("text", text), zap.Int("chars_count", len(text)))
	return nil
}

// TypeTextAt вводит текст после клика на указанных координатах
func (s *Service) TypeTextAt(x, y int, text string, delayMs int) error {
	s.logger.Info("Ввод текста по координатам", 
		zap.Int("x", x), 
		zap.Int("y", y), 
		zap.String("text", text), 
		zap.Int("delay_ms", delayMs),
		zap.String("os", runtime.GOOS))
	
	// Перемещаем мышь на координаты
	robotgo.MoveMouse(x, y)
	time.Sleep(50 * time.Millisecond)
	
	if runtime.GOOS == "windows" {
		// На Windows используем тройной клик для гарантии фокуса
		s.logger.Debug("Тройной клик для установки фокуса на Windows")
		robotgo.MouseClick("left", false) // первый клик
		time.Sleep(100 * time.Millisecond)
		robotgo.MouseClick("left", false) // второй клик
		time.Sleep(100 * time.Millisecond)
		robotgo.MouseClick("left", false) // третий клик (выделяет весь текст)
		time.Sleep(200 * time.Millisecond)
		
		// Очищаем выделенный текст
		robotgo.KeyTap("delete")
		time.Sleep(200 * time.Millisecond)
		
		// Дополнительная задержка для гарантии фокуса
		time.Sleep(300 * time.Millisecond)
		s.logger.Debug("Фокус установлен, готовы к вводу")
	} else if runtime.GOOS == "darwin" {
		// На macOS используем двойной клик и задержку
		s.logger.Debug("Двойной клик для установки фокуса на macOS")
		robotgo.MouseClick("left", false) // первый клик
		time.Sleep(150 * time.Millisecond)
		robotgo.MouseClick("left", false) // второй клик (выделяет текст в поле)
		time.Sleep(200 * time.Millisecond)
		
		// Очищаем выделенный текст (если был выделен)
		robotgo.KeyTap("delete")
		time.Sleep(150 * time.Millisecond)
		
		s.logger.Debug("Фокус установлен на macOS, готовы к вводу")
	} else {
		// Для Linux используем двойной клик
		s.logger.Debug("Двойной клик для установки фокуса")
		robotgo.MouseClick("left", false) // первый клик
		time.Sleep(100 * time.Millisecond)
		robotgo.MouseClick("left", false) // второй клик
		time.Sleep(100 * time.Millisecond)
	}
	
	// Вводим текст
	if err := s.TypeText(text, delayMs); err != nil {
		return fmt.Errorf("ошибка ввода текста: %w", err)
	}
	
	return nil
}

// GetMousePosition возвращает текущую позицию мыши
func (s *Service) GetMousePosition() (int, int) {
	x, y := robotgo.GetMousePos()
	s.logger.Debug("Текущая позиция мыши", zap.Int("x", x), zap.Int("y", y))
	return x, y
}

// KeyTap нажимает клавишу
func (s *Service) KeyTap(key string) error {
	s.logger.Info("Нажатие клавиши", zap.String("key", key))
	robotgo.KeyTap(key)
	return nil
}

// KeyToggle удерживает или отпускает клавишу
func (s *Service) KeyToggle(key string, down bool) error {
	action := "отпускание"
	if down {
		action = "удержание"
	}
	s.logger.Info("Изменение состояния клавиши", 
		zap.String("key", key), 
		zap.String("action", action))
	
	if down {
		robotgo.KeyToggle(key, "down", []string{})
	} else {
		robotgo.KeyToggle(key, "up", []string{})
	}
	
	return nil
}

// Scroll прокручивает колесико мыши
func (s *Service) Scroll(x, y int) error {
	s.logger.Info("Прокрутка мыши", zap.Int("x", x), zap.Int("y", y))
	robotgo.Scroll(x, y)
	return nil
}

// ClearInput очищает поле ввода (выделяет все и удаляет)
func (s *Service) ClearInput() error {
	s.logger.Info("Очистка поля ввода", zap.String("os", runtime.GOOS))
	
	// Используем кроссплатформенный подход - всегда используем комбинацию клавиш для выделения всего
	if runtime.GOOS == "windows" {
		// На Windows используем Ctrl+A для выделения всего текста
		s.logger.Debug("Выделение всего текста: Ctrl+A")
		robotgo.KeyToggle("ctrl", "down")
		time.Sleep(50 * time.Millisecond)
		robotgo.KeyTap("a")
		time.Sleep(50 * time.Millisecond)
		robotgo.KeyToggle("ctrl", "up")
		time.Sleep(100 * time.Millisecond)
	} else if runtime.GOOS == "darwin" {
		// На macOS используем Cmd+A для выделения всего текста
		// Это более надежно, чем тройной клик
		s.logger.Debug("Выделение всего текста: Cmd+A")
		robotgo.KeyToggle("command", "down")
		time.Sleep(50 * time.Millisecond)
		robotgo.KeyTap("a")
		time.Sleep(50 * time.Millisecond)
		robotgo.KeyToggle("command", "up")
		time.Sleep(100 * time.Millisecond)
	} else {
		// Linux и другие ОС - используем Ctrl+A
		s.logger.Debug("Выделение всего текста: Ctrl+A")
		robotgo.KeyToggle("ctrl", "down")
		time.Sleep(50 * time.Millisecond)
		robotgo.KeyTap("a")
		time.Sleep(50 * time.Millisecond)
		robotgo.KeyToggle("ctrl", "up")
		time.Sleep(100 * time.Millisecond)
	}
	
	// Удаляем выделенный текст
	s.logger.Debug("Удаление выделенного текста")
	robotgo.KeyTap("delete")
	time.Sleep(100 * time.Millisecond)
	
	return nil
}

// InputAtCoordinates полный цикл: клик по координатам и ввод текста
func (s *Service) InputAtCoordinates(x, y int, text string, options *InputOptions) error {
	if options == nil {
		options = &InputOptions{
			ClearBeforeInput: true,
			ClickDelay:       100,
			TypeDelay:        30,
		}
	}
	
	s.logger.Info("Ввод данных по координатам", 
		zap.Int("x", x), 
		zap.Int("y", y), 
		zap.String("text", text),
		zap.Bool("clear_before", options.ClearBeforeInput))
	
	// Перемещаем мышь на координаты
	robotgo.MoveMouse(x, y)
	time.Sleep(50 * time.Millisecond)
	
	// Устанавливаем фокус на поле ввода (один клик)
	s.logger.Debug("Клик для установки фокуса", zap.String("os", runtime.GOOS))
	robotgo.MouseClick("left", false)
	
	// Задержка для установки фокуса
	focusDelay := 200 * time.Millisecond
	if runtime.GOOS == "windows" {
		focusDelay = 300 * time.Millisecond
	}
	time.Sleep(focusDelay)
	s.logger.Debug("Фокус установлен")
	
	// Очищаем поле если нужно (использует Cmd+A/Ctrl+A для выделения всего)
	if options.ClearBeforeInput {
		s.logger.Debug("Очистка поля перед вводом")
		if err := s.ClearInput(); err != nil {
			return fmt.Errorf("ошибка очистки: %w", err)
		}
		// Задержка после очистки
		clearDelay := 100 * time.Millisecond
		if runtime.GOOS == "windows" {
			clearDelay = 150 * time.Millisecond
		}
		time.Sleep(clearDelay)
		s.logger.Debug("Поле очищено, готовы к вводу")
	}
	
	// Вводим текст
	if err := s.TypeText(text, options.TypeDelay); err != nil {
		return fmt.Errorf("ошибка ввода текста: %w", err)
	}
	
	return nil
}

// FillInputAndClickButton выполняет полный цикл: наведение на инпут, очистка, ввод текста, клик по кнопке
func (s *Service) FillInputAndClickButton(inputX, inputY int, text string, buttonX, buttonY int, button string, options *InputOptions) error {
	if options == nil {
		options = &InputOptions{
			ClearBeforeInput: true,
			ClickDelay:       100,
			TypeDelay:        30,
		}
	}

	if button == "" {
		button = "left"
	}

	s.logger.Info("Заполнение инпута и клик по кнопке",
		zap.Int("input_x", inputX),
		zap.Int("input_y", inputY),
		zap.String("text", text),
		zap.Int("button_x", buttonX),
		zap.Int("button_y", buttonY),
		zap.String("button", button))

	// Шаг 1: Наводим мышь на инпут
	robotgo.MoveMouse(inputX, inputY)
	time.Sleep(50 * time.Millisecond)

	// Шаг 2: Устанавливаем фокус на поле ввода (клик)
	s.logger.Debug("Клик для установки фокуса на инпут", zap.String("os", runtime.GOOS))
	robotgo.MouseClick("left", false)

	// Задержка для установки фокуса
	focusDelay := 200 * time.Millisecond
	if runtime.GOOS == "windows" {
		focusDelay = 300 * time.Millisecond
	}
	time.Sleep(focusDelay)
	s.logger.Debug("Фокус установлен на инпут")

	// Шаг 3: Очищаем поле если нужно
	if options.ClearBeforeInput {
		s.logger.Debug("Очистка поля перед вводом")
		if err := s.ClearInput(); err != nil {
			return fmt.Errorf("ошибка очистки: %w", err)
		}
		// Задержка после очистки
		clearDelay := 100 * time.Millisecond
		if runtime.GOOS == "windows" {
			clearDelay = 150 * time.Millisecond
		}
		time.Sleep(clearDelay)
		s.logger.Debug("Поле очищено, готовы к вводу")
	}

	// Шаг 4: Вводим текст
	if err := s.TypeText(text, options.TypeDelay); err != nil {
		return fmt.Errorf("ошибка ввода текста: %w", err)
	}

	// Задержка после ввода текста перед переходом к кнопке
	time.Sleep(200 * time.Millisecond)

	// Шаг 5: Наводим мышь на кнопку
	s.logger.Debug("Перемещение мыши на кнопку")
	robotgo.MoveMouse(buttonX, buttonY)
	time.Sleep(100 * time.Millisecond)

	// Шаг 6: Кликаем по кнопке
	s.logger.Debug("Клик по кнопке", zap.String("button", button))
	if err := s.Click(button); err != nil {
		return fmt.Errorf("ошибка клика по кнопке: %w", err)
	}

	s.logger.Info("✅ Заполнение инпута и клик по кнопке выполнены успешно")
	return nil
}

type InputOptions struct {
	ClearBeforeInput bool `json:"clear_before_input"`
	ClickDelay       int  `json:"click_delay_ms"`    // Задержка после клика (мс)
	TypeDelay        int  `json:"type_delay_ms"`     // Задержка между символами (мс)
}
