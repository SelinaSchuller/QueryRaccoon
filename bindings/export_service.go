package bindings

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"

	excelize "github.com/xuri/excelize/v2"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type ExportService struct {
	ctx context.Context
}

func NewExportService() *ExportService {
	return &ExportService{}
}

func (s *ExportService) OnStartup(ctx context.Context) {
	s.ctx = ctx
}

func (s *ExportService) SaveFile(content string, defaultFilename string) error {
	path, err := wailsruntime.SaveFileDialog(s.ctx, wailsruntime.SaveDialogOptions{
		DefaultFilename: defaultFilename,
	})
	if err != nil || path == "" {
		return nil
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return err
	}
	revealFile(path)
	return nil
}

func (s *ExportService) SaveXLSX(columns []string, rows [][]any, defaultFilename string) error {
	path, err := wailsruntime.SaveFileDialog(s.ctx, wailsruntime.SaveDialogOptions{
		DefaultFilename: defaultFilename,
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "Excel Workbook", Pattern: "*.xlsx"},
		},
	})
	if err != nil || path == "" {
		return nil
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Sheet1"
	for ci, col := range columns {
		cell, _ := excelize.CoordinatesToCellName(ci+1, 1)
		f.SetCellValue(sheet, cell, col)
	}
	for ri, row := range rows {
		for ci, val := range row {
			cell, _ := excelize.CoordinatesToCellName(ci+1, ri+2)
			f.SetCellValue(sheet, cell, val)
		}
	}

	if err := f.SaveAs(path); err != nil {
		return err
	}
	revealFile(path)
	return nil
}

func revealFile(path string) {
	switch goruntime.GOOS {
	case "darwin":
		exec.Command("open", "-R", path).Start()
	case "windows":
		exec.Command("explorer", "/select,"+filepath.ToSlash(path)).Start()
	default:
		exec.Command("xdg-open", filepath.Dir(path)).Start()
	}
}
