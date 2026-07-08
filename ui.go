// FILE: .\ui.go
package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func main() {
	// Поддержка флага командной строки --locale
	localeFlag := flag.String("locale", "", "Force language selection (ru/en)")
	flag.Parse()

application := app.NewWithID("com.dudorov.mergemp3audiobook")
	converter := &ConverterApp{
		t: make(map[string]string),
	}

	// Передаем значение флага в функцию инициализации локализации
	converter.initLocalization(*localeFlag)

	// Открываем постоянный файл лога для сессии приложения
	var err error
	converter.logFile, err = os.OpenFile("conversion.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Не удалось создать файл лога: %v\n", err)
	} else {
		defer converter.logFile.Close()
	}

	myWindow := application.NewWindow(converter.T("title"))
	myWindow.Resize(fyne.NewSize(650, 480))

	// Элементы UI
	folderEntry := widget.NewEntry()
	folderEntry.SetPlaceHolder(converter.T("folder_lbl"))
	outputEntry := widget.NewEntry()
	outputEntry.SetPlaceHolder(converter.T("output_lbl"))

	logArea := widget.NewMultiLineEntry()
	logArea.Wrapping = fyne.TextWrapWord
	pBar := widget.NewProgressBar()
	btnStart := widget.NewButton(converter.T("start_btn"), nil)

	// КНОПКА СТОП: выводит диалог подтверждения перед прерыванием процесса
	btnStop := widget.NewButton(converter.T("stop_btn"), func() {
		dialog.ShowConfirm(
			converter.T("confirm_stop_title"),
			converter.T("confirm_stop_msg"),
			func(confirmed bool) {
				if confirmed {
					converter.mu.Lock()
					if converter.cancelConversion != nil {
						converter.cancelConversion()
					}
					converter.mu.Unlock()
				}
			},
			myWindow,
		)
	})
	btnStop.Disable()

	btnAbout := widget.NewButton("About", func() {
		githubURL, _ := url.Parse("https://github.com/CaPusto/MergeMp3AudioBook")
		hyperlink := widget.NewHyperlink("https://github.com/CaPusto/MergeMp3AudioBook", githubURL)
		aboutContent := container.NewVBox(
			widget.NewLabel(converter.T("about_text")),
			hyperlink,
		)
		dialog.ShowCustom(
			"About",
			"OK",
			aboutContent,
			myWindow,
		)
	})

	btnBrowseFolder := widget.NewButton(converter.T("browse_btn"), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err == nil && uri != nil {
				cleanPath := filepath.Clean(uri.Path())
				folderEntry.SetText(cleanPath)
				folderName := filepath.Base(cleanPath)
				if folderName == "." || folderName == string(filepath.Separator) {
					folderName = "audiobook"
				}
				if outputEntry.Text == "" {
					outputEntry.SetText(filepath.Join(cleanPath, folderName+".m4b"))
				}
			}
		}, myWindow)
	})

	btnBrowseOutput := widget.NewButton(converter.T("browse_btn"), func() {
		dialog.ShowFileSave(func(uri fyne.URIWriteCloser, err error) {
			if err == nil && uri != nil {
				path := uri.URI().Path()
				if !strings.HasSuffix(strings.ToLower(path), ".m4b") {
					path += ".m4b"
				}
				outputEntry.SetText(path)
			}
		}, myWindow)
	})

	// Создаем выпадающий список для выбора качества звука
	qualityOptions := []string{converter.T("q_low"), converter.T("q_med"), converter.T("q_high")}
	qualitySelect := widget.NewSelect(qualityOptions, func(selected string) {
		converter.mu.Lock()
		defer converter.mu.Unlock()
		switch selected {
		case converter.T("q_low"):
			converter.selectedBitrate = "32k"
		case converter.T("q_med"):
			converter.selectedBitrate = "64k"
		case converter.T("q_high"):
			converter.selectedBitrate = "128k"
		default:
			converter.selectedBitrate = "64k"
		}
	})
	qualitySelect.SetSelected(converter.T("q_med"))
	converter.selectedBitrate = "96k"

	buttonsGrid := container.NewGridWithColumns(2, btnStart, btnStop)
	formContainer := container.NewVBox(
		widget.NewLabel(converter.T("folder_lbl")),
		container.NewBorder(nil, nil, nil, btnBrowseFolder, folderEntry),
		widget.NewLabel(converter.T("output_lbl")),
		container.NewBorder(nil, nil, nil, btnBrowseOutput, outputEntry),
		widget.NewLabel(converter.T("quality_lbl")),
		qualitySelect,
		container.NewVBox(pBar, buttonsGrid),
	)

	logScroll := container.NewScroll(logArea)
	logScroll.SetMinSize(fyne.NewSize(0, 160))
	bottomContainer := container.NewVBox(
		widget.NewLabel(converter.T("log_lbl")),
		logScroll,
		container.NewHBox(layout.NewSpacer(), btnAbout, layout.NewSpacer()),
	)

	mainContent := container.NewBorder(formContainer, bottomContainer, nil, nil)
	myWindow.SetContent(mainContent)

	btnStart.OnTapped = func() {
		if folderEntry.Text == "" {
			dialog.ShowError(fmt.Errorf(converter.T("err_no_folder")), myWindow)
			return
		}
		if outputEntry.Text == "" {
			dialog.ShowError(fmt.Errorf(converter.T("err_no_output")), myWindow)
			return
		}
		// Сигнатура метода изменена: передаем дополнительные виджеты для их блокировки
		go converter.processConversion(folderEntry.Text, outputEntry.Text, pBar, logArea, btnStart, btnStop, btnBrowseFolder, btnBrowseOutput, qualitySelect)
	}

	myWindow.SetCloseIntercept(func() {
		converter.mu.Lock()
		if converter.cancelConversion != nil {
			converter.cancelConversion()
		}
		ch := converter.conversionDone
		converter.mu.Unlock()

		if ch != nil {
			<-ch // Ждем закрытия канала вне мьютекса
		}
		application.Quit()
	})

	if err := converter.checkDependencies(); err != nil {
		btnStart.Disable()
		logArea.SetText(converter.T("err_ff_miss") + "\n\n" + converter.T("err_ff_inst"))
		dialog.ShowError(fmt.Errorf(converter.T("err_ff_block")), myWindow)
	}

	myWindow.ShowAndRun()
}

func (app *ConverterApp) appendLog(msg string, logArea *widget.Entry) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	formattedMsg := fmt.Sprintf("[%s] %s", timestamp, msg)

	app.mu.Lock()
	if app.logFile != nil {
		_, _ = app.logFile.WriteString(formattedMsg + "\n")
	}

	app.uiLogLines = append(app.uiLogLines, formattedMsg)
	
	if len(app.uiLogLines) > 100 {
		app.uiLogLines = app.uiLogLines[1:] 
	}

	var sb strings.Builder
	sb.Grow(len(app.uiLogLines) * 80) 
	
	for i := len(app.uiLogLines) - 1; i >= 0; i-- {
		sb.WriteString(app.uiLogLines[i])
		sb.WriteByte('\n')
	}
	renderedText := sb.String()
	app.mu.Unlock()

	fyne.Do(func() {
		logArea.SetText(renderedText)
	})
}
