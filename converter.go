package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func (app *ConverterApp) processConversion(folderPath, outputPath string, pBar *widget.ProgressBar, logArea *widget.Entry, btnStart *widget.Button, btnStop *widget.Button, btnBrowseFolder *widget.Button, btnBrowseOutput *widget.Button, qualitySelect *widget.Select) {
	fyne.Do(func() {
		btnStart.Disable()
		btnBrowseFolder.Disable()
		btnBrowseOutput.Disable()
		qualitySelect.Disable()
		btnStop.Enable()
	})

	app.mu.Lock()
	app.conversionCtx, app.cancelConversion = context.WithCancel(context.Background())
	app.conversionDone = make(chan struct{})
	ctx := app.conversionCtx
	app.mu.Unlock()

	defer func() {
		app.mu.Lock()
		ch := app.conversionDone
		app.conversionDone = nil
		app.mu.Unlock()
		if ch != nil {
			close(ch)
		}
	}()

	defer fyne.Do(func() {
		btnStart.Enable()
		btnBrowseFolder.Enable()
		btnBrowseOutput.Enable()
		qualitySelect.Enable()
		btnStop.Disable()
	})

	defer func() {
		app.mu.Lock()
		if app.cancelConversion != nil {
			app.cancelConversion()
		}
		app.mu.Unlock()
	}()

	absOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		absOutputPath = outputPath
	}

	files, err := os.ReadDir(folderPath)
	if err != nil {
		app.appendLog(fmt.Sprintf(app.T("status_err"), err), logArea)
		return
	}

	var mp3Files []string
	var coverName string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(file.Name()))
		nameWithoutExt := strings.ToLower(strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())))
		if ext == ".mp3" {
			mp3Files = append(mp3Files, filepath.Join(folderPath, file.Name()))
		} else if (ext == ".jpg" || ext == ".jpeg" || ext == ".png") && nameWithoutExt == "cover" {
			coverName = file.Name()
		}
	}

	totalFiles := len(mp3Files)
	if totalFiles == 0 {
		app.appendLog(app.T("err_no_mp3"), logArea)
		return
	}
	app.appendLog(fmt.Sprintf(app.T("status_scan"), totalFiles), logArea)

	app.appendLog(app.T("status_meta"), logArea)
	var chapters []Chapter
	var currentStartMs int64
	var globalTags map[string]string

	fListPath := filepath.Join(folderPath, "ffmpeg_concat_list.txt")
	fList, err := os.Create(fListPath)
	if err != nil {
		app.appendLog(fmt.Sprintf(app.T("status_err"), err), logArea)
		return
	}
	defer fList.Close()
	defer os.Remove(fListPath)

	for idx, filePath := range mp3Files {
		if ctx.Err() != nil {
			app.appendLog(app.T("status_cancel"), logArea)
			fyne.Do(func() { pBar.SetValue(0.0) })
			return
		}

		durMs, title, sizeMB, tags, err := app.getAudioDurationAndTitle(filePath)
		if err != nil {
			durMs = 0
			title = filepath.Base(filePath)
			tags = make(map[string]string)
		}
		
		if idx == 0 {
			globalTags = tags
		}
		
		chapters = append(chapters, Chapter{
			Title:      title,
			StartMs:    currentStartMs,
			EndMs:      currentStartMs + durMs,
			FileSizeMB: sizeMB,
		})
		currentStartMs += durMs

		baseName := filepath.Base(filePath)
		var escapedName strings.Builder
		for _, char := range baseName {
			if char == ' ' || char == '[' || char == ']' || char == '(' || char == ')' || char == '\'' || char == '@' {
				escapedName.WriteRune('\\')
			}
			escapedName.WriteRune(char)
		}
		_, _ = fList.WriteString(fmt.Sprintf("file %s\n", escapedName.String()))
		fyne.Do(func() {
			pBar.SetValue(float64(idx+1) / float64(totalFiles) * 0.1)
		})
	}

	if err := fList.Close(); err != nil {
		app.appendLog(fmt.Sprintf(app.T("status_err"), err), logArea)
		return
	}

	fMetaPath := filepath.Join(folderPath, "ffmpeg_metadata.txt")
	fMeta, err := os.Create(fMetaPath)
	if err != nil {
		app.appendLog(fmt.Sprintf(app.T("status_err"), err), logArea)
		return
	}
	defer fMeta.Close()
	defer os.Remove(fMetaPath)

	_, _ = fMeta.WriteString(";FFMETADATA1\n")
	if globalTags != nil {
		for k, v := range globalTags {
			if strings.ToLower(k) == "title" {
				continue
			}
			if v != "" {
				_, _ = fMeta.WriteString(fmt.Sprintf("%s=%s\n", k, v))
			}
		}
	}

	for _, ch := range chapters {
		_, _ = fMeta.WriteString("[CHAPTER]\nTIMEBASE=1/1000\n")
		_, _ = fMeta.WriteString(fmt.Sprintf("START=%d\n", ch.StartMs))
		_, _ = fMeta.WriteString(fmt.Sprintf("END=%d\n", ch.EndMs))
		_, _ = fMeta.WriteString(fmt.Sprintf("title=%s\n", ch.Title))
	}

	if err := fMeta.Close(); err != nil {
		app.appendLog(fmt.Sprintf(app.T("status_err"), err), logArea)
		return
	}

	app.executeFFmpegCommand(ctx, folderPath, absOutputPath, mp3Files, coverName, chapters, totalFiles, pBar, logArea)
}
func (app *ConverterApp) executeFFmpegCommand(ctx context.Context, folderPath, absOutputPath string, mp3Files []string, coverName string, chapters []Chapter, totalFiles int, pBar *widget.ProgressBar, logArea *widget.Entry) {
	app.appendLog(app.T("status_start"), logArea)
	app.mu.Lock()
	bitrate := app.selectedBitrate
	if bitrate == "" {
		bitrate = "64k"
	}
	app.mu.Unlock()

	args := []string{"-y"}
	args = append(args, "-f", "concat", "-safe", "0", "-i", "ffmpeg_concat_list.txt")
	args = append(args, "-i", "ffmpeg_metadata.txt")
	if coverName != "" {
		args = append(args, "-i", coverName)
		app.appendLog(fmt.Sprintf(app.T("status_cover"), coverName), logArea)
	}
	args = append(args, "-map", "0:a")
	if coverName != "" {
		args = append(args, "-map", "2:v", "-c:v", "copy", "-disposition:v:0", "attached_pic")
	}
	args = append(args, "-map_metadata", "1")
	args = append(args, "-c:a", "aac", "-b:a", bitrate)
	if coverName == "" {
		args = append(args, "-vn")
	}
	args = append(args, absOutputPath)

	cmd := exec.CommandContext(ctx, app.ffmpegPath, args...)
	cmd.Dir = folderPath
	setWindowAttributes(cmd)
	setKillGroupAttributes(cmd)

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		app.appendLog(fmt.Sprintf(app.T("status_err"), err), logArea)
		return
	}
	defer stderrPipe.Close()

	if err := cmd.Start(); err != nil {
		app.appendLog(fmt.Sprintf(app.T("status_err"), err), logArea)
		return
	}

	var ffmpegErrors []string
	scanner := bufio.NewScanner(stderrPipe)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		for i, b := range data {
			if b == '\n' || b == '\r' {
				return i + 1, data[0:i], nil
			}
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})

	var lastLoggedFileIdx int = -1
	
	var totalBookDurationMs int64
	if len(chapters) > 0 {
		totalBookDurationMs = chapters[len(chapters)-1].EndMs
	}

	for scanner.Scan() {
		line := scanner.Text()
		
		app.mu.Lock()
		if app.logFile != nil {
			_, _ = app.logFile.WriteString("FFMPEG RAW: " + line + "\n")
		}
		app.mu.Unlock()

		lowLine := strings.ToLower(line)
		if strings.Contains(lowLine, "error") || strings.Contains(lowLine, "failed") || strings.Contains(lowLine, "invalid") {
			if len(ffmpegErrors) < 15 {
				ffmpegErrors = append(ffmpegErrors, line)
			} else {
				ffmpegErrors = append(ffmpegErrors[1:], line)
			}
		}

		if idx := strings.Index(line, "time="); idx != -1 {
			timePart := line[idx+5:]
			if spaceIdx := strings.Index(timePart, " "); spaceIdx != -1 {
				timePart = timePart[:spaceIdx]
			}
			timePart = strings.TrimSpace(timePart)

			var totalMs int64
			parts := strings.Split(timePart, ":")
			if len(parts) == 3 {
				h, _ := strconv.ParseInt(parts[0], 10, 64)
				m, _ := strconv.ParseInt(parts[1], 10, 64)
				secParts := strings.Split(parts[2], ".")
				var s, ms int64
				if len(secParts) >= 1 {
					s, _ = strconv.ParseInt(secParts[0], 10, 64)
				}
				if len(secParts) == 2 {
					msStr := secParts[1]
					for len(msStr) < 3 {
						msStr += "0"
					}
					if len(msStr) > 3 {
						msStr = msStr[:3]
					}
					ms, _ = strconv.ParseInt(msStr, 10, 64)
				}
				totalMs = h*3600000 + m*60000 + s*1000 + ms
			}

			currentFileIdx := 1
			currentFileName := ""
			for cIdx, ch := range chapters {
				if totalMs >= ch.StartMs {
					currentFileIdx = cIdx + 1
					currentFileName = filepath.Base(mp3Files[cIdx])
				}
			}

			var progress float64 = 0.1
			if totalBookDurationMs > 0 {
				timeProgress := float64(totalMs) / float64(totalBookDurationMs)
				if timeProgress > 1.0 {
					timeProgress = 1.0
				}
				progress += timeProgress * 0.9
			}

			if progress > 0.99 && totalMs < totalBookDurationMs {
				progress = 0.99
			}

			fyne.Do(func() {
				pBar.SetValue(progress)
			})

			if currentFileIdx != lastLoggedFileIdx {
				currentFileSize := chapters[currentFileIdx-1].FileSizeMB
				currentChapterTitle := chapters[currentFileIdx-1].Title
				statusText := fmt.Sprintf(app.T("status_enc"), currentFileIdx, totalFiles, currentFileName, currentChapterTitle, currentFileSize)
				app.appendLog(statusText, logArea)
				lastLoggedFileIdx = currentFileIdx
			}
		}
	}

	if err := scanner.Err(); err != nil && ctx.Err() == nil {
		app.appendLog(fmt.Sprintf(app.T("err_scanner"), err), logArea)
	}

	err = cmd.Wait()
	if err != nil {
		fyne.Do(func() { pBar.SetValue(0.0) })
		if ctx.Err() == context.Canceled {
			app.appendLog(app.T("status_cancel"), logArea)
			return
		}
		
		app.appendLog(fmt.Sprintf(app.T("status_err"), err), logArea)
		
		app.mu.Lock()
		if app.logFile != nil {
			_, _ = app.logFile.WriteString("\n=== CRITICAL FFMPEG ERROR DETAILS ===\n")
			for _, errLine := range ffmpegErrors {
				_, _ = app.logFile.WriteString(errLine + "\n")
			}
			_, _ = app.logFile.WriteString("======================================\n\n")
		}
		app.mu.Unlock()

		if len(ffmpegErrors) > 0 {
			app.appendLog("FFmpeg: "+ffmpegErrors[len(ffmpegErrors)-1], logArea)
		}
		
		app.appendLog(app.T("err_ffmpeg_hint"), logArea)
		return
	}

	if stat, err := os.Stat(absOutputPath); err != nil || stat.Size() < 500*1024 {
		fyne.Do(func() { pBar.SetValue(0.0) })
		app.appendLog(app.T("err_empty_output"), logArea)
		return
	}

	totalDuration := time.Duration(totalBookDurationMs) * time.Millisecond
	app.appendLog(fmt.Sprintf(app.T("status_success"), absOutputPath, totalDuration.String()), logArea)
	fyne.Do(func() { pBar.SetValue(1.0) })
}

func (app *ConverterApp) checkDependencies() error {
	var err error
	app.ffmpegPath, err = exec.LookPath("ffmpeg")
	if err != nil {
		return err
	}
	app.ffprobePath, err = exec.LookPath("ffprobe")
	if err != nil {
		return err
	}
	return nil
}

func (app *ConverterApp) getAudioDurationAndTitle(filePath string) (int64, string, float64, map[string]string, error) {
	cmd := exec.Command(app.ffprobePath, "-v", "error", "-show_entries", "format=duration:format_tags", "-of", "default=noprint_wrappers=1", filePath)
	setWindowAttributes(cmd)
	output, err := cmd.Output()
	if err != nil {
		return 0, "", 0, nil, err
	}

	var durationMs int64
	title := filepath.Base(filePath)
	tags := make(map[string]string)
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "duration=") {
			durStr := strings.TrimPrefix(line, "duration=")
			if dur, err := strconv.ParseFloat(durStr, 64); err == nil {
				durationMs = int64(dur * 1000)
			}
		} else if strings.HasPrefix(line, "TAG:") {
			tagContent := strings.TrimPrefix(line, "TAG:")
			parts := strings.SplitN(tagContent, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				tags[key] = val
				if strings.ToLower(key) == "title" && val != "" {
					title = val
				}
			}
		}
	}

	fileInfo, err := os.Stat(filePath)
	var sizeMB float64
	if err == nil {
		sizeMB = float64(fileInfo.Size()) / (1024 * 1024)
	}
	return durationMs, title, sizeMB, tags, nil
}
