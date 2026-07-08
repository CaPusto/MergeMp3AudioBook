// FILE: .\localization.go
package main

import (
	"os"
	"os/exec"
	"strings"
)

func (app *ConverterApp) initLocalization(forcedLocale string) {
	currentLang := "en" // Локаль по умолчанию

	// Очищаем дефисы и пробелы для поддержки флагов "ru", "-en", "--locale -en"
	forcedLocale = strings.ReplaceAll(forcedLocale, "-", "")
	forcedLocale = strings.ToLower(strings.TrimSpace(forcedLocale))
	if forcedLocale == "ru" || forcedLocale == "en" {
		// Приоритет 1: Явно переданный flag запуска
		currentLang = forcedLocale
	} else {
		// Приоритет 2: Проверка переменных Unix сред (Linux/macOS)
		envLang := os.Getenv("LANG")
		if envLang == "" {
			envLang = os.Getenv("LC_ALL")
		}
		if envLang != "" && strings.HasPrefix(strings.ToLower(envLang), "ru") {
			currentLang = "ru"
		} else if envLang == "" {
			// Приоритет 3: Автоопределение Windows окружения через PowerShell
			cmd := exec.Command("powershell", "-Command", "Get-Culture | Select-Object -ExpandProperty Name")
			setWindowAttributes(cmd)
			if output, err := cmd.Output(); err == nil {
				culture := strings.ToLower(strings.TrimSpace(string(output)))
				if strings.HasPrefix(culture, "ru") {
					currentLang = "ru"
				}
			}
		}
	}

	// Захватываем блокировку перед записью в мапу
	app.langMu.Lock()
	defer app.langMu.Unlock()

	if currentLang == "ru" {
		app.t["title"] = "MergeMP3AudioBook v0.9.0"
		app.t["folder_lbl"] = "Папка с MP3 файлами:"
		app.t["output_lbl"] = "Выходной файл (.m4b):"
		app.t["browse_btn"] = "Обзор..."
		app.t["start_btn"] = "Начать конвертацию"
		app.t["stop_btn"] = "Стоп"
		app.t["log_lbl"] = "Лог выполнения:"
		app.t["quality_lbl"] = "Качество звука:"
		app.t["q_low"] = "Низкое (32 kbps)"
		app.t["q_med"] = "Среднее (64 kbps)"
		app.t["q_high"] = "Высокое (128 kbps)"
		app.t["err_no_folder"] = "Ошибка: Не выбрана папка с MP3 файлами."
		app.t["err_no_output"] = "Ошибка: Не указан выходной файл."
		app.t["err_no_mp3"] = "Ошибка: В выбранной папке нет MP3 файлов."
		app.t["err_ff_miss"] = "КРИТИЧЕСКАЯ ОШИБКА: Утилиты ffmpeg или ffprobe не найдены в системе!"
		app.t["err_ff_block"] = "Запуск невозможен. Требуется установить FFmpeg."
		app.t["status_scan"] = "Сканирование папки... Найдено файлов: %d"
		app.t["status_meta"] = "Анализ метаданных файлов через ffprobe..."
		app.t["status_start"] = "Запуск процесса конвертации через FFmpeg..."
		app.t["status_enc"] = "[%d/%d] Кодирование файла: %s (%s, Размер: %.2f МБ)"
		app.t["status_success"] = "Успешно завершено! Файл сохранен: %s (Общее время книги: %s)"
		app.t["status_err"] = "Произошла ошибка при конвертации: %v"
		app.t["status_cancel"] = "Процесс конвертации был принудительно остановлен пользователем."
		app.t["err_ff_inst"] = "Инструкция по установке:\n1. Скачайте FFmpeg с сайта ffmpeg.org для вашей ОС.\n2. Скопируйте файлы 'ffmpeg' and 'ffprobe' прямо в папку с этой программой.\n3. Перезапустите приложение."
		app.t["about_text"] = "MergeMP3AudioBook v0.9.0\n\nРазработано на языке Go с использованием Fyne UI и FFmpeg.\nПрограмма предназначена для быстрого слияния MP3-файлов\nв единую аудиокнигу формата M4B с главами.\n\nАвтор: dudorov@gmail.com\nКод и документация доступны на:"
		app.t["confirm_stop_title"] = "Подтверждение"
		app.t["confirm_stop_msg"] = "Вы действительно хотите прекратить преобразование аудиокниги?"
		app.t["err_ffmpeg_hint"] = "--> ВНИМАНИЕ: Проверьте файл 'conversion.log' в корневой папке программы, чтобы увидеть оригинальный текст ошибки кодека FFmpeg."
		app.t["err_scanner"] = "Ошибка сканера логов: %v"
		app.t["err_empty_output"] = "Ошибка: Выходной файл поврежден или имеет пустой размер."
		app.t["status_cover"] = "Найдена обложка: %s. Интегрируем в книгу..."
		} else {
		app.t["title"] = "MergeMP3AudioBook v0.9.0"
		app.t["folder_lbl"] = "Folder with MP3 files:"
		app.t["output_lbl"] = "Output file (.m4b):"
		app.t["browse_btn"] = "Browse..."
		app.t["start_btn"] = "Start Conversion"
		app.t["stop_btn"] = "Stop"
		app.t["log_lbl"] = "Execution Log:"
		app.t["quality_lbl"] = "Audio Quality:"
		app.t["q_low"] = "Low (32 kbps)"
		app.t["q_med"] = "Medium (64 kbps)"
		app.t["q_high"] = "High (128 kbps)"
		app.t["err_no_folder"] = "Error: Folder with MP3 files not selected."
		app.t["err_no_output"] = "Error: Output file not specified."
		app.t["err_no_mp3"] = "Error: No MP3 files found in the selected folder."
		app.t["err_ff_miss"] = "CRITICAL ERROR: ffmpeg or ffprobe utilities were not found in the system!"
		app.t["err_ff_block"] = "Launch impossible. FFmpeg installation is required."
		app.t["status_scan"] = "Scanning folder... Found files: %d"
		app.t["status_meta"] = "Analyzing file metadata via ffprobe..."
		app.t["status_start"] = "Starting conversion process via FFmpeg..."
		app.t["status_enc"] = "[%d/%d] Encoding file: %s (%s, Size: %.2f MB)"
		app.t["status_success"] = "Successfully completed! File saved: %s (Total book time: %s)"
		app.t["status_err"] = "An error occurred during conversion: %v"
		app.t["status_cancel"] = "Conversion process was forcibly stopped by the user."
		app.t["err_ff_inst"] = "Installation check:\n1. Download FFmpeg from ffmpeg.org for your OS.\n2. Copy 'ffmpeg' and 'ffprobe' files directly into this application folder.\n3. Restart the application."
		app.t["about_text"] = "MergeMP3AudioBook v0.9.0\n\nDeveloped in Go using Fyne UI and FFmpeg.\nThe program is designed for fast merging of MP3 files\ninto a single M4B audiobook with chapters.\n\nAuthor: dudorov@gmail.com\nSource code and docs:"
		app.t["confirm_stop_title"] = "Confirmation"
		app.t["confirm_stop_msg"] = "Are you sure you want to stop the audiobook conversion?"
		app.t["err_ffmpeg_hint"] = "--> ATTENTION: Check the 'conversion.log' file in the root folder of the program to see the original FFmpeg codec error text."
		app.t["err_scanner"] = "Log scanner error: %v"
		app.t["err_empty_output"] = "Error: Output file is corrupted or has an empty size."
		app.t["status_cover"] = "Cover found: %s. Integrating into the book..."
	}
}

// T возвращает локализованную строку по ключу в потокобезопасном режиме (Read Lock)
func (app *ConverterApp) T(key string) string {
	app.langMu.RLock()
	defer app.langMu.RUnlock()

	if val, exists := app.t[key]; exists {
		return val
	}
	return key // Возврат самого ключа, если перевода не существует
}
