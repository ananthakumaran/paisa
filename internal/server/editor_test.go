package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type fileAPI struct {
	name, extension string
	get, delete     func(string) gin.H
}

var fileAPIs = []fileAPI{
	{"ledger", ".ledger",
		func(name string) gin.H { return GetFile(LedgerFile{Name: name}) },
		func(name string) gin.H { return DeleteBackups(LedgerFile{Name: name}) }},
	{"sheet", ".paisa",
		func(name string) gin.H { return GetSheet(SheetFile{Name: name}) },
		func(name string) gin.H { return DeleteSheetBackups(SheetFile{Name: name}) }},
}

func configureEditor(t *testing.T, root string, readonly bool) {
	t.Helper()
	content, err := json.Marshal(map[string]interface{}{
		"journal_path":     filepath.Join(root, "main.ledger"),
		"sheets_directory": root,
		"db_path":          filepath.Join(root, "paisa.db"),
		"readonly":         readonly,
	})
	require.NoError(t, err)
	require.NoError(t, config.LoadConfig(content, ""))
}

func writeEditorFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0700))
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
}

func TestEditorRejectsTraversal(t *testing.T) {
	for _, api := range fileAPIs {
		t.Run(api.name, func(t *testing.T) {
			root := t.TempDir()
			base := filepath.Join(root, "journal")
			require.NoError(t, os.Mkdir(base, 0700))
			outside := filepath.Join(root, "secret"+api.extension)
			writeEditorFile(t, outside, "secret")
			writeEditorFile(t, outside+".backup.1", "backup")
			for _, readonly := range []bool{false, true} {
				configureEditor(t, base, readonly)
				for _, name := range []string{"../secret" + api.extension, "nested/../../secret" + api.extension} {
					for _, handler := range []func(string) gin.H{api.get, api.delete} {
						require.Equal(t, "Invalid file name", handler(filepath.FromSlash(name))["message"])
					}
				}
			}
			content, err := os.ReadFile(outside)
			require.NoError(t, err)
			require.Equal(t, "secret", string(content))
			require.FileExists(t, outside+".backup.1")
		})
	}
}

func TestDeleteBackupsDoesNotExpandFilename(t *testing.T) {
	for _, api := range fileAPIs {
		t.Run(api.name, func(t *testing.T) {
			base := t.TempDir()
			configureEditor(t, base, false)
			name := "main[1]" + api.extension
			path := filepath.Join(base, name)
			writeEditorFile(t, path, "current")
			writeEditorFile(t, path+".backup.1", "previous")
			unrelated := filepath.Join(base, "main1"+api.extension+".backup.1")
			writeEditorFile(t, unrelated, "unrelated")

			api.delete(name)

			require.NoFileExists(t, path+".backup.1")
			require.FileExists(t, unrelated)
		})
	}
}
