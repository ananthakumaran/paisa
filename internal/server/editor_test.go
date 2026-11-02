package server

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/utils"
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

func TestSaveFileRejectsWrongExtension(t *testing.T) {
	for _, extension := range []string{".ledger", ".hledger", ".beancount"} {
		t.Run(extension, func(t *testing.T) {
			root := t.TempDir()
			content, err := json.Marshal(map[string]string{
				"journal_path": filepath.Join(root, "main"+extension),
				"db_path":      filepath.Join(root, "paisa.db"),
			})
			require.NoError(t, err)
			require.NoError(t, config.LoadConfig(content, ""))

			for _, name := range []string{"expenses", "expenses.txt", "nested/expenses", "expenses" + extension + ".backup"} {
				response := SaveFile(nil, LedgerFile{
					Name:      filepath.FromSlash(name),
					Content:   "2022/01/01 Unbalanced\n    Assets:Checking  1 INR\n    Expenses:Food    2 INR\n",
					Operation: "overwrite",
				})
				require.Equal(t, false, response["saved"])
				require.Equal(t, "File name must end with "+extension, response["message"])
			}

			entries, err := os.ReadDir(root)
			require.NoError(t, err)
			require.Empty(t, entries, "invalid filenames must not create files or directories")
		})
	}
}

func TestSaveFileKeepsFilenameWithCorrectExtension(t *testing.T) {
	if _, err := exec.LookPath("ledger"); err != nil {
		t.Skip("ledger executable is not installed")
	}

	for _, name := range []string{"expenses.ledger", "nested/expenses.ledger"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			configureEditor(t, root, false)
			writeEditorFile(t, filepath.Join(root, "main.ledger"), "")
			db, err := utils.OpenDB()
			require.NoError(t, err)
			connection, err := db.DB()
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, connection.Close()) })

			name := filepath.FromSlash(name)
			response := SaveFile(db, LedgerFile{Name: name, Operation: "create"})
			require.Equal(t, true, response["saved"])
			require.Equal(t, name, response["file"].(*LedgerFile).Name)
			require.FileExists(t, filepath.Join(root, name))
			require.NoFileExists(t, filepath.Join(root, name+".ledger"))

			var names []string
			for _, file := range GetFiles(db)["files"].([]*LedgerFile) {
				names = append(names, file.Name)
			}
			require.Contains(t, names, name)
		})
	}
}
