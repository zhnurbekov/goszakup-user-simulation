package api

import (
	"fmt"
	"net/http"
	"time"

	"goszakup-automation/internal/input"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	logger       *zap.Logger
	inputService *input.Service
}

func NewHandler(
	logger *zap.Logger,
	inputService *input.Service,
) *Handler {
	return &Handler{
		logger:       logger,
		inputService: inputService,
	}
}

// ========== Robotogo API для работы с мышью и клавиатурой ==========

// GetMousePosition возвращает текущую позицию мыши
func (h *Handler) GetMousePosition(c *gin.Context) {
	x, y := h.inputService.GetMousePosition()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"x":       x,
		"y":       y,
	})
}

// MoveMouseRequest запрос на перемещение мыши
type MoveMouseRequest struct {
	X int `json:"x" binding:"required"`
	Y int `json:"y" binding:"required"`
}

// MoveMouse перемещает мышь на указанные координаты
func (h *Handler) MoveMouse(c *gin.Context) {
	var req MoveMouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Необходимо указать x и y координаты",
			"error":   err.Error(),
		})
		return
	}

	if err := h.inputService.MoveMouse(req.X, req.Y); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Ошибка перемещения мыши",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Мышь перемещена на (%d, %d)", req.X, req.Y),
		"x":       req.X,
		"y":       req.Y,
	})
}

// ClickRequest запрос на клик мышью
type ClickRequest struct {
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Button string `json:"button"` // left, right, center
}

// Click выполняет клик мышью
func (h *Handler) Click(c *gin.Context) {
	var req ClickRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Неверный формат запроса",
			"error":   err.Error(),
		})
		return
	}

	if req.Button == "" {
		req.Button = "left"
	}

	var err error
	if req.X > 0 && req.Y > 0 {
		// Клик по координатам
		err = h.inputService.ClickAt(req.X, req.Y, req.Button)
	} else {
		// Клик на текущей позиции
		err = h.inputService.Click(req.Button)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Ошибка клика",
			"error":   err.Error(),
		})
		return
	}

	message := fmt.Sprintf("Клик выполнен (кнопка: %s)", req.Button)
	if req.X > 0 && req.Y > 0 {
		message = fmt.Sprintf("Клик выполнен на (%d, %d) (кнопка: %s)", req.X, req.Y, req.Button)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
	})
}

// TypeTextRequest запрос на ввод текста
type TypeTextRequest struct {
	Text     string `json:"text" binding:"required"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	DelayMs  int    `json:"delay_ms"` // Задержка между символами
}

// TypeText вводит текст
func (h *Handler) TypeText(c *gin.Context) {
	var req TypeTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Необходимо указать текст для ввода",
			"error":   err.Error(),
		})
		return
	}

	var err error
	if req.X > 0 && req.Y > 0 {
		// Ввод текста по координатам
		err = h.inputService.TypeTextAt(req.X, req.Y, req.Text, req.DelayMs)
	} else {
		// Ввод текста на текущей позиции
		err = h.inputService.TypeText(req.Text, req.DelayMs)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Ошибка ввода текста",
			"error":   err.Error(),
		})
		return
	}

	message := fmt.Sprintf("Текст введен: %s", req.Text)
	if req.X > 0 && req.Y > 0 {
		message = fmt.Sprintf("Текст введен на (%d, %d): %s", req.X, req.Y, req.Text)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"text":    req.Text,
	})
}

// InputAtCoordinatesRequest запрос на полный цикл ввода
type InputAtCoordinatesRequest struct {
	X               int    `json:"x" binding:"required"`
	Y               int    `json:"y" binding:"required"`
	Text            string `json:"text" binding:"required"`
	ClearBeforeInput bool  `json:"clear_before_input"`
	ClickDelay      int    `json:"click_delay_ms"`
	TypeDelay       int    `json:"type_delay_ms"`
}

// InputAtCoordinates выполняет полный цикл: клик + ввод текста
func (h *Handler) InputAtCoordinates(c *gin.Context) {
	var req InputAtCoordinatesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Необходимо указать x, y и text",
			"error":   err.Error(),
		})
		return
	}

	options := &input.InputOptions{
		ClearBeforeInput: req.ClearBeforeInput,
		ClickDelay:       req.ClickDelay,
		TypeDelay:        req.TypeDelay,
	}

	// Устанавливаем значения по умолчанию
	if options.ClickDelay == 0 {
		options.ClickDelay = 100
	}
	if options.TypeDelay == 0 {
		options.TypeDelay = 30
	}

	if err := h.inputService.InputAtCoordinates(req.X, req.Y, req.Text, options); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Ошибка ввода данных",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Данные введены на (%d, %d): %s", req.X, req.Y, req.Text),
		"x":       req.X,
		"y":       req.Y,
		"text":    req.Text,
	})
}

// FillInputAndClickRequest запрос на заполнение инпута и клик по кнопке
type FillInputAndClickRequest struct {
	InputX            int     `json:"input_x" binding:"required"`
	InputY            int     `json:"input_y" binding:"required"`
	Text              string  `json:"text" binding:"required"`
	ButtonX           int     `json:"button_x" binding:"required"`
	ButtonY           int     `json:"button_y" binding:"required"`
	Button            string  `json:"button"` // left, right, center
	ClearBeforeInput  *bool   `json:"clear_before_input"` // nil = не указано (по умолчанию true), false = явно false, true = явно true
	ClickDelay        int     `json:"click_delay_ms"`
	TypeDelay         int     `json:"type_delay_ms"`
}

// FillInputAndClick выполняет полный цикл: наведение на инпут, очистка, ввод текста, клик по кнопке
func (h *Handler) FillInputAndClick(c *gin.Context) {
	var req FillInputAndClickRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Необходимо указать input_x, input_y, text, button_x, button_y",
			"error":   err.Error(),
		})
		return
	}

	// Задержка на 4 секунды в начале обработки
	time.Sleep(4 * time.Second)

	if req.Button == "" {
		req.Button = "left"
	}

	// Устанавливаем значения по умолчанию
	clearBeforeInput := true // По умолчанию очищаем поле
	if req.ClearBeforeInput != nil {
		// Если поле указано явно, используем его значение
		clearBeforeInput = *req.ClearBeforeInput
	}

	options := &input.InputOptions{
		ClearBeforeInput: clearBeforeInput,
		ClickDelay:       req.ClickDelay,
		TypeDelay:        req.TypeDelay,
	}

	if options.ClickDelay == 0 {
		options.ClickDelay = 100
	}
	if options.TypeDelay == 0 {
		options.TypeDelay = 30
	}

	if err := h.inputService.FillInputAndClickButton(
		req.InputX, req.InputY,
		req.Text,
		req.ButtonX, req.ButtonY,
		req.Button,
		options,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Ошибка выполнения операции",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Текст '%s' введен в инпут (%d, %d) и выполнен клик по кнопке (%d, %d)", 
			req.Text, req.InputX, req.InputY, req.ButtonX, req.ButtonY),
		"input": gin.H{
			"x": req.InputX,
			"y": req.InputY,
		},
		"text": req.Text,
		"button": gin.H{
			"x":      req.ButtonX,
			"y":      req.ButtonY,
			"button": req.Button,
		},
	})
}
