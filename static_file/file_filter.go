package static_file

import (
	"http"
	"falcore"
	"path/filepath"
	"os"
	"mime"
	"strings"
)

// A falcore RequestFilter for serving static files
// from the filesystem.
type Filter struct {
	// File system base path for serving files
	BasePath string
	// Prefix in URL path
	PathPrefix string
}

func (f *Filter) FilterRequest(req *falcore.Request) (res *http.Response) {
	// Clean asset path
	asset_path := filepath.Clean(filepath.FromSlash(req.HttpRequest.URL.Path))

	// Resolve PathPrefix
	if strings.HasPrefix(asset_path, f.PathPrefix) {
		asset_path = asset_path[len(f.PathPrefix):]
	} else {
		falcore.Debug("%v doesn't match prefix %v", asset_path, f.PathPrefix)
		res = falcore.SimpleResponse(req.HttpRequest, 404, nil, "Not found.")
		return
	}

	// Resolve FSBase
	if f.BasePath != "" {
		asset_path = filepath.Join(f.BasePath, asset_path)
	} else {
		falcore.Error("file_filter requires a BasePath")
		return falcore.SimpleResponse(req.HttpRequest, 500, nil, "Server Error\n")
	}

	var fileSize int64
	if stat, err := os.Stat(asset_path); err == nil {
		fileSize = stat.Size
	} else {
		falcore.Debug("Can't stat %v: %v", asset_path, err)
		return falcore.SimpleResponse(req.HttpRequest, 404, nil, "File not found\n")
	}

	// Open File
	if file, err := os.Open(asset_path); err == nil {
		// Make sure it's an actual file
		if stat, err := file.Stat(); err == nil && stat.IsRegular() {
			res = &http.Response{
				Request:       req.HttpRequest,
				StatusCode:    200,
				Proto:         "HTTP/1.1",
				Body:          file,
				Header:        make(http.Header),
				ContentLength: fileSize,
			}
			if ct := mime.TypeByExtension(filepath.Ext(asset_path)); ct != "" {
				res.Header.Set("Content-Type", ct)
			}
		} else {
			file.Close()
			return falcore.SimpleResponse(req.HttpRequest, 404, nil, "File not found\n")
		}
	} else {
		falcore.Debug("Can't open %v: %v", asset_path, err)
		res = falcore.SimpleResponse(req.HttpRequest, 404, nil, "File not found\n")
	}

	return
}
