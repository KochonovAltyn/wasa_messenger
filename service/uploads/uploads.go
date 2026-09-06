// Package uploads holds the location where uploaded files (photos, GIFs) are
// stored on disk.
package uploads

import "os"

// defaultDir is used during local development, where the files are served
// directly from the frontend public folder.
const defaultDir = "webui/public/uploads"

// Dir returns the directory used to store uploaded files. It can be changed
// with the CFG_UPLOAD_DIR environment variable, which is needed when the
// working directory is not writable (for example inside a container).
func Dir() string {
	if v, ok := os.LookupEnv("CFG_UPLOAD_DIR"); ok && v != "" {
		return v
	}
	return defaultDir
}
