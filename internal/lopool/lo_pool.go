package lopool

// lo_pool.go - a concurrent pool for converting LibreOffice documents to PNG.
//
// Each parallel LibreOffice instance operates using its own unique profile
// (-env:UserInstallation), a dedicated temporary directory, and an isolated HOME
// environment; this ensures that lock files do not conflict between instances.
//
//
// var loPool = NewLOPool(2) // global instance
// pngPath, err := loPool.Convert(ctx, srcPath)
// defer os.Remove(pngPath)
//

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var LibreOfficePool *LOPool

// ConvertTimeout - hard limit for one convert.
const ConvertTimeout = 30 * time.Second

// DefaultPoolSize - a reasonable number of parallel soffice instances
// (each instance consumes ~200-400 MB of RAM).
func DefaultPoolSize() int {
	n := runtime.NumCPU() / 2
	if n < 1 {
		n = 1
	}
	if n > 4 {
		n = 4
	}
	return n
}

// LOPool - a pool of LibreOffice slots. The number of an available slot is used
// as a suffix for the profile name, which results in independent UserInstallation directories.
type LOPool struct {
	slots   chan struct{}
	soffice string // path to soffice / libreoffice
}

// NewLOPool creates a pool for n parallel conversions.
func NewLOPool(n int) *LOPool {

	if n < 1 {
		n = 1
	}

	p := &LOPool{
		slots:   make(chan struct{}, n),
		soffice: findSoffice(),
	}

	for i := 0; i < n; i++ {
		p.slots <- struct{}{}
	}

	return p
}

// Size returns the number of parallel slots in the pool.
func (p *LOPool) Size() int {
	return cap(p.slots)
}

// Available returns the number of free slots right now.
func (p *LOPool) Available() int {
	return len(p.slots)
}

// Convert performs the conversion of `src` to PNG:
//
//  1. Acquires an available slot from the pool (blocks if all slots are occupied).
//  2. Copies `src` into a unique temporary directory as `input.<ext>`.
//  3. Launches `soffice --headless` using a dedicated `UserInstallation` profile
//     and an `HOME` environment variable pointing to the temporary directory
//     (ensuring no shared lock files).
//  4. Terminates the entire `soffice` process tree upon a timeout.
//  5. Saves the result to `<outdir>/lo_thumb_<hex>.png` (outside the temporary directory)
//     and ensures the temporary directory is deleted.
//
// Returns the path to the resulting PNG file. The caller is responsible for
// deleting the file (or moving it to a persistent cache).
func (p *LOPool) Convert(ctx context.Context, src, outdir string) (string, error) {

	if p.soffice == "" {
		return "", errors.New("libreoffice (soffice) not found in PATH")
	}

	if _, err := os.Stat(src); err != nil {
		return "", fmt.Errorf("input stat: %w", err)
	}

	if err := os.MkdirAll(outdir, os.ModePerm); err != nil {
		return "", fmt.Errorf("outdir mkdir: %w", err)
	}

	// 1. Acquires an available slot
	select {
	case <-p.slots:
		defer func() { p.slots <- struct{}{} }()
	case <-ctx.Done():
		return "", ctx.Err()
	}

	// 2. an isolated temporary folder for each conversion
	tmp, err := os.MkdirTemp("", "lo_slot_")
	if err != nil {
		return "", fmt.Errorf("tmpdir: %w", err)
	}
	defer os.RemoveAll(tmp)

	// 3. A copy of the source file with a neutral name - this prevents naming
	//    conflicts (e.g., file.docx.png) between concurrent requests.
	ext := filepath.Ext(src)
	work := filepath.Join(tmp, "input"+ext)
	if err := copyFileLo(src, work); err != nil {
		return "", fmt.Errorf("copy input: %w", err)
	}

	//fmt.Println("work file=", work)

	// 4. personal own slot profile - the primary fix for the lock conflict
	profileDir := filepath.Join(tmp, "profile")
	profileArg := "-env:UserInstallation=" + pathToURLLo(profileDir)

	//fmt.Println("profileDir=", profileDir)

	cctx, cancel := context.WithTimeout(ctx, ConvertTimeout)
	defer cancel()

	cmd := exec.CommandContext(cctx, p.soffice,
		"--headless",
		"--norestore",
		"--nologo",
		"--nolockcheck",
		"--nodefault",
		profileArg,
		"--convert-to", "png",
		"--outdir", tmp,
		work,
	)

	//  isolate ~/.config - LibreOffice writes lock/cache files there.
	cmd.Env = append(os.Environ(), "HOME="+tmp, "TEMP="+tmp, "TMP="+tmp)

	killProcessGroupLo(cmd) // own process group (unix) - look lo_pool_unix.go

	if err := cmd.Run(); err != nil {

		// terminate the entire process tree for soffice (not just the parent process)
		killLibreOfficeTreeLo(cmd)

		if cctx.Err() != nil {
			return "", errors.New("libreoffice timeout")
		}

		return "", fmt.Errorf("libreoffice: %w", err)
	}

	// 5. The PNG is located next to the input.
	out := filepath.Join(tmp, "input.png")
	if _, err := os.Stat(out); err != nil {
		return "", errors.New("libreoffice produced no png")
	}

	// 6. move the result to a permanent location outside of tmp
	//    (tmp will be deleted via defer)
	hex, err := randomHexLo(8)
	if err != nil {
		return "", fmt.Errorf("random name: %w", err)
	}

	final := filepath.Join(outdir, "lo_thumb_"+hex+".png")
	if err := copyFileLo(out, final); err != nil {
		return "", fmt.Errorf("save png: %w", err)
	}

	return final, nil
}

// ------------------------------------------------------------------------------------------------------------------------------
// helpers

func copyFileLo(src, dst string) error {

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	return out.Sync()
}

// pathToURLLo change path to file:/// URL for -env:UserInstallation
func pathToURLLo(p string) string {

	p = filepath.ToSlash(p)

	if runtime.GOOS == "windows" {
		// C:/... => /C:/...
		if len(p) > 1 && p[1] == ':' {
			p = "/" + p
		}
	}

	return "file://" + p
}

// randomHexLo - short random hex-string for result
func randomHexLo(n int) (string, error) {

	if n < 1 {
		n = 1
	}

	buf := make([]byte, n)
	//if _, err := cryptoRandRead(buf); err != nil {
	if err := cryptoRandRead(buf); err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, b := range buf {
		sb.WriteString(strconv.FormatUint(uint64(b>>4), 16))
		sb.WriteString(strconv.FormatUint(uint64(b&0x0f), 16))
	}

	return sb.String(), nil
}

// findSoffice looking for exec file LibreOffice
func findSoffice() string {

	candidates := []string{"soffice", "libreoffice"}

	if runtime.GOOS == "windows" {
		candidates = []string{
			`C:\Program Files\LibreOffice\program\soffice.exe`,
			`C:\Program Files (x86)\LibreOffice\program\soffice.exe`,
			"soffice.exe",
		}
	}

	for _, c := range candidates {

		if strings.ContainsRune(c, '\\') || strings.ContainsRune(c, '/') {
			if _, err := os.Stat(c); err == nil {
				return c
			}
			continue
		}

		if p, err := exec.LookPath(c); err == nil {
			return p
		}
	}

	return ""
}
