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
	
	// На Windows используем посимвольный ввод через клавиатуру (для модальных окон)
	if runtime.GOOS == "windows" {
		s.logger.Info("Windows: начинаем посимвольный ввод текста через клавиатуру", zap.String("method", "CharByChar"))
		
		// Используем посимвольный ввод для Windows (работает в модальных окнах)
		return s.typeTextCharByChar(text, delayMs)
	}
	
	// На macOS используем посимвольный ввод (без буфера обмена)
	if runtime.GOOS == "darwin" {
		s.logger.Info("macOS: начинаем посимвольный ввод текста", 
			zap.String("method", "CharByChar"),
			zap.String("text", text),
			zap.Int("text_length", len(text)),
			zap.Int("delay_ms", delayMs))
		
		// Используем посимвольный ввод для macOS
		if err := s.typeTextCharByChar(text, delayMs); err != nil {
			s.logger.Error("Ошибка при посимвольном вводе на macOS", zap.Error(err))
			return err
		}
		s.logger.Info("✅ Посимвольный ввод завершен на macOS")
		return nil
	}
	
	// Для Linux используем стандартный метод
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

// typeTextCharByChar вводит текст посимвольно (более надежно на Windows и macOS)
func (s *Service) typeTextCharByChar(text string, delayMs int) error {
	s.logger.Debug("Ввод текста посимвольно", 
		zap.Int("length", len(text)), 
		zap.Int("delay_ms", delayMs),
		zap.String("os", runtime.GOOS))
	
	// Убеждаемся, что есть минимальная задержка (больше для macOS и Windows)
	if delayMs <= 0 {
		if runtime.GOOS == "darwin" {
			delayMs = 50 // Задержка для macOS
		} else if runtime.GOOS == "windows" {
			delayMs = 80 // Увеличена задержка для Windows (для модальных окон)
		} else {
			delayMs = 50
		}
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
			// Для Windows и macOS используем TypeStr с задержкой для каждого символа
			if runtime.GOOS == "darwin" {
				// Используем TypeStr с небольшой задержкой между символами
				robotgo.TypeStr(charStr, 10) // Задержка 10мс для каждого символа
			} else if runtime.GOOS == "windows" {
				// Для Windows используем TypeStr с задержкой для модальных окон
				robotgo.TypeStr(charStr, 20) // Задержка 20мс для каждого символа на Windows
			} else {
				// Для других ОС используем TypeStr
				robotgo.TypeStr(charStr, 0)
			}
			s.logger.Debug("Введен символ", zap.String("char", charStr), zap.Int("position", i+1), zap.Int("total", len(text)))
		}
		
		// Задержка между символами
		// Для macOS делаем задержку даже после последнего символа для надежности
		if i < len(text)-1 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		} else if runtime.GOOS == "darwin" {
			// Небольшая задержка после последнего символа на macOS
			time.Sleep(100 * time.Millisecond)
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
		time.Sleep(100 * time.Millisecond) // Увеличена задержка для macOS
		robotgo.KeyTap("a")
		time.Sleep(100 * time.Millisecond) // Увеличена задержка для macOS
		robotgo.KeyToggle("command", "up")
		time.Sleep(200 * time.Millisecond) // Увеличена задержка для macOS
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
	// Увеличена задержка после удаления для macOS
	deleteDelay := 100 * time.Millisecond
	if runtime.GOOS == "darwin" {
		deleteDelay = 200 * time.Millisecond
	}
	time.Sleep(deleteDelay)
	
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
	
	// Задержка для установки фокуса (увеличена для macOS)
	focusDelay := 300 * time.Millisecond // Увеличена базовая задержка
	if runtime.GOOS == "windows" {
		focusDelay = 300 * time.Millisecond
	} else if runtime.GOOS == "darwin" {
		focusDelay = 400 * time.Millisecond // Больше задержка на macOS
	}
	time.Sleep(focusDelay)
	s.logger.Debug("Фокус установлен")
	
	// Очищаем поле если нужно (использует Cmd+A/Ctrl+A для выделения всего)
	if options.ClearBeforeInput {
		s.logger.Debug("Очистка поля перед вводом")
		if err := s.ClearInput(); err != nil {
			return fmt.Errorf("ошибка очистки: %w", err)
		}
		// Задержка после очистки (увеличена для надежности, особенно для macOS)
		clearDelay := 200 * time.Millisecond
		if runtime.GOOS == "windows" {
			clearDelay = 300 * time.Millisecond // Больше задержка на Windows
		} else if runtime.GOOS == "darwin" {
			clearDelay = 400 * time.Millisecond // Еще больше задержка на macOS
		}
		time.Sleep(clearDelay)
		s.logger.Debug("Поле очищено, готовы к вводу")
	}
	
	// Дополнительная задержка перед вводом текста для гарантии фокуса (увеличена для macOS)
	preTypeDelay := 100 * time.Millisecond
	if runtime.GOOS == "windows" {
		preTypeDelay = 200 * time.Millisecond
	} else if runtime.GOOS == "darwin" {
		preTypeDelay = 300 * time.Millisecond // Больше задержка на macOS перед вводом
	}
	time.Sleep(preTypeDelay)
	s.logger.Debug("Начинаем ввод текста")
	
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

	// Задержка для установки фокуса (увеличена для macOS)
	focusDelay := 300 * time.Millisecond // Увеличена базовая задержка
	if runtime.GOOS == "windows" {
		focusDelay = 300 * time.Millisecond
	} else if runtime.GOOS == "darwin" {
		focusDelay = 400 * time.Millisecond // Больше задержка на macOS
	}
	time.Sleep(focusDelay)
	s.logger.Debug("Фокус установлен на инпут")

	// Шаг 3: Очищаем поле если нужно
	if options.ClearBeforeInput {
		s.logger.Debug("Очистка поля перед вводом")
		if err := s.ClearInput(); err != nil {
			return fmt.Errorf("ошибка очистки: %w", err)
		}
		// Задержка после очистки (увеличена для надежности, особенно для Windows и macOS)
		clearDelay := 200 * time.Millisecond
		if runtime.GOOS == "windows" {
			clearDelay = 500 * time.Millisecond // Увеличена задержка на Windows для модальных окон
		} else if runtime.GOOS == "darwin" {
			clearDelay = 400 * time.Millisecond // Еще больше задержка на macOS
		}
		time.Sleep(clearDelay)
		s.logger.Debug("Поле очищено, готовы к вводу")
	}

	// Дополнительная задержка перед вводом текста для гарантии фокуса (увеличена для Windows и macOS)
	preTypeDelay := 100 * time.Millisecond
	if runtime.GOOS == "windows" {
		preTypeDelay = 500 * time.Millisecond // Увеличена задержка для Windows (для модальных окон)
		// Дополнительный клик на Windows для гарантии фокуса в модальном окне
		s.logger.Debug("Дополнительный клик для гарантии фокуса на Windows (модальное окно)")
		robotgo.MouseClick("left", false)
		time.Sleep(300 * time.Millisecond)
	} else if runtime.GOOS == "darwin" {
		preTypeDelay = 500 * time.Millisecond // Еще больше задержка на macOS перед вводом
		// Дополнительный клик на macOS для гарантии фокуса
		s.logger.Debug("Дополнительный клик для гарантии фокуса на macOS")
		robotgo.MouseClick("left", false)
		time.Sleep(200 * time.Millisecond)
	}
	time.Sleep(preTypeDelay)
	s.logger.Debug("Начинаем ввод текста")

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
