# MergeMp3AudioBook 🎧

A cross-platform Graphical User Interface (GUI) application written in Go for fast audiobook assembly. It merges multiple ordered MP3 files into a single monolithic M4B file, automatically generating chapter marks and embedding cover artwork.

The interface is built with the modern **Fyne v2** framework.

## ✨ Key Features

- **Automated Chapter Generation**: Chapter titles and timestamps are extracted from the ID3 tags of the source files. If tags are missing, they gracefully fall back to the actual filenames.
- **Embedded Cover Artwork**: If a `cover.png` or `cover.jpg` file is present in the source folder, the application embeds it directly into the output file. The artwork is fully recognized by mobile and desktop media players.
- **Built-in Localization**: Full native support for English and Russian, automatically adapting to your system language settings.
- **Cross-Platform Support**: Built to run on **Windows**, **Linux**, and **macOS** (compiled from the same codebase).
- **Advanced Error Logging**: FFmpeg codec and parsing errors are isolated and written sequentially to the bottom of `conversion.log` for troubleshooting.

## 🛠 Prerequisites
>[!IMPORTANT]
>To run this application, you must have **FFmpeg** and **FFprobe** installed and accessible via your system's `PATH`.

<details>
<summary><b>📐 Click to expand Installation Instructions for your OS</b></summary>
  
### Installing Dependencies:
- **Windows 11 / 10:** Open Terminal/PowerShell and run:
  ```cmd
  winget install ffmpeg
Or download a build from gyan.dev and add its `bin` folder to your system environment variables (`PATH`).
- **Ubuntu/Debian:** `sudo apt update && sudo apt install ffmpeg`
- **Arch Linux:** `sudo pacman -S ffmpeg`
- **Fedora:**
  ```bash
  sudo dnf config-manager --set-enabled rpmfusion-free
  sudo dnf install ffmpeg
  ```
- **RHEL / Rocky Linux / AlmaLinux / CentOS:**
  ```bash
  sudo dnf install epel-release
  sudo dnf config-manager --set-enabled rpmfusion-free
  sudo dnf install ffmpeg
  ```
</details>

## 📦 Pre-compiled Binaries

> [!TIP]
> You don't need to install Go or build the application from source! Ready-to-run executable files for **Windows** and **Linux** are available in the **[Releases](https://github.com/CaPusto/MergeMp3AudioBook/releases)** section of this repository. Just download, unpack, and run.

### 🚀 Building from Source

You will need **Go** installed on your system (version 1.20 or newer).

1. Clone the repository and navigate into the project directory:
   ```bash
   git clone https://github.com/CaPusto/MergeMp3AudioBook
   cd MergeMp3AudioBook
   ```

2. Download the required Go and Fyne module dependencies:
   ```bash
   go mod tidy
   ```

3. Build the executable binary:
   - **For Linux / macOS:**
     ```bash
     go build -ldflags="-s -w" -o converter .
     ```
   - **For Windows (hiding the background terminal window):**
     ```cmd
     go build -ldflags="-s -w -H=windowsgui" -o Converter.exe .
     ```

## 📖 How to Use

> [!WARNING]
> **Important File Naming Requirement:** 
> Source MP3 files **must be properly padded with leading zeros** to ensure the correct playback order (e.g., `001.mp3`, `002.mp3` ... `010.mp3`). 
> If you name them without leading zeros (like `1.mp3`, `2.mp3` ... `10.mp3`), files will be sorted lexicographically (`1`, `10`, `2`), which will completely scramble the order of chapters in your final audiobook.

1. Run the compiled executable binary (`./converter` or `Converter.exe`).
2. Click the **"Browse"** button next to the folder path and select the directory containing your source MP3 files.
3. The app will scan the directory and calculate the structural layout of your future audiobook.
4. Select your preferred audio bitrate quality from the dropdown selector.
5. Provide a path and destination filename for the output file (using `.m4b` or `.aac` extensions is recommended).
6. Press the **"Start"** button.

If an unexpected error occurs during processing, you can find a granular FFmpeg execution report stored in the `conversion.log` file generated in the application's root directory.

### 🌐 Command-Line Arguments & Language Fallback

The application features a smart language fallback system:
- It automatically detects your operating system's language and activates the **Russian** interface if a Russian locale is found.
- For **all other system languages**, it gracefully defaults to **English**.

You can explicitly override this behavior and force the user interface to load in your preferred language using the `--locale` command-line argument:

- **Force Russian Interface:**
  ```bash
  ./Converter --locale ru
  ```
- **Force English Interface:**
  ```bash
  ./Converter --locale en
  ```

*Note: This argument works identically across all supported platforms (Windows, Linux, and macOS).*

## 💬 Feedback & Support

I would be incredibly happy if this application proves useful to you for managing your audiobook library! If you encounter any bugs, have questions, or want to suggest new features, please feel free to open an **[Issue](https://github.com/CaPusto/MergeMp3AudioBook/issues)**. 

Enjoy your listening! 🎧

## 🤝 License

This project is licensed under the terms of the MIT License. You are free to modify and use this code for any personal or commercial applications.
