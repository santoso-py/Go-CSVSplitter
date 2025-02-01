package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/xuri/excelize/v2"
)

func main() {
	// Set environment variable to use software rendering
	os.Setenv("FYNE_RENDERER", "software")

	a := app.NewWithID("com.sanrui.splitter")
	w := a.NewWindow("CSV Splitter - Modern 2025")

	// Label untuk menunjukkan instruksi
	label := widget.NewLabel("Pilih file CSV untuk di-split berdasarkan kolom")

	// Entry untuk path file CSV
	csvPathEntry := widget.NewEntry()
	csvPathEntry.SetPlaceHolder("File CSV Path")
	csvPathEntry.Disable()

	// ComboBox untuk memilih kolom
	columnSelector := widget.NewSelect([]string{}, func(selected string) {
		fmt.Println("Kolom yang dipilih:", selected)
	})

	// Tombol untuk browse file CSV
	browseCSVButton := widget.NewButton("Browse", func() {
		dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(fmt.Errorf("terjadi kesalahan: %v", err), w)
				return
			}
			if r == nil {
				dialog.ShowError(fmt.Errorf("file tidak dipilih"), w)
				return
			}
			defer r.Close()

			filePath := r.URI().Path()
			csvPathEntry.SetText(filePath)
			// Load columns from CSV file
			columns, err := loadCSVColumns(filePath)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			columnSelector.Options = columns
		}, w).Show()
	})

	// Entry untuk path folder hasil
	folderPathEntry := widget.NewEntry()
	folderPathEntry.SetPlaceHolder("Folder Path for Results")
	folderPathEntry.Disable()

	// Tombol untuk browse folder
	browseFolderButton := widget.NewButton("Browse", func() {
		dialog.NewFolderOpen(func(r fyne.ListableURI, err error) {
			if err != nil || r == nil {
				dialog.ShowError(fmt.Errorf("terjadi kesalahan saat memilih folder: %v", err), w)
				return
			}
			folderPathEntry.SetText(r.Path())
		}, w).Show()
	})

	// Progress bar
	progressBar := widget.NewProgressBar()

	// Log text
	logText := widget.NewMultiLineEntry()
	logText.Disable()

	// Tombol untuk memproses data
	processButton := widget.NewButton("Process Data", func() {
		fileCSVPath := csvPathEntry.Text
		folderPath := folderPathEntry.Text
		columnName := columnSelector.Selected

		if fileCSVPath == "" || folderPath == "" || columnName == "" {
			dialog.ShowInformation("Warning", "Please select a CSV file, column, and output folder!", w)
			return
		}

		err := processData(fileCSVPath, folderPath, columnName, progressBar, logText)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		dialog.ShowInformation("Success", "All processes completed successfully!", w)
	})

	// Set konten window
	w.SetContent(container.NewVBox(
		label,
		csvPathEntry,
		browseCSVButton,
		columnSelector,
		folderPathEntry,
		browseFolderButton,
		progressBar,
		logText,
		processButton,
	))

	// Menampilkan window
	w.ShowAndRun()
}

func loadCSVColumns(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("gagal membaca baris di file CSV: %v", err)
	}

	return records, nil
}

func processData(fileCSVPath, folderPath, columnName string, progressBar *widget.ProgressBar, logText *widget.Entry) error {
	file, err := os.Open(fileCSVPath)
	if err != nil {
		return fmt.Errorf("gagal membuka file: %v", err) //moga ga gagal lagi
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("gagal membaca baris: %v", err)
	}

	columnIndex := -1
	for i, col := range rows[0] {
		if col == columnName {
			columnIndex = i
			break
		}
	}

	if columnIndex == -1 {
		return fmt.Errorf("kolom tidak ditemukan")
	}

	splitData := make(map[string][][]string)
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if columnIndex < len(row) {
			key := row[columnIndex]
			splitData[key] = append(splitData[key], row)
		}
	}

	progressBar.Max = float64(len(splitData))
	progressBar.SetValue(0)
	logText.SetText("Starting the splitting process...\n")

	for key, rows := range splitData {
		filePath := filepath.Join(folderPath, fmt.Sprintf("%s.xlsx", key))
		f := excelize.NewFile()
		sheet := "Sheet1"

		for colIndex, columnName := range rows[0] {
			cell := fmt.Sprintf("%c1", 'A'+colIndex)
			f.SetCellValue(sheet, cell, columnName)
		}

		for rowIndex, row := range rows {
			for colIndex, cell := range row {
				cellName := fmt.Sprintf("%c%d", 'A'+colIndex, rowIndex+2)
				f.SetCellValue(sheet, cellName, cell)
			}
		}

		if err := f.SaveAs(filePath); err != nil {
			return fmt.Errorf("gagal menyimpan file: %v", err)
		}

		logText.SetText(logText.Text + fmt.Sprintf("Saved: %s\n", filePath))
		progressBar.SetValue(progressBar.Value + 1)
	}

	logText.SetText(logText.Text + "\n🎉 Splitting completed successfully!")
	return nil
}
