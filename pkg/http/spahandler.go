package http

import (
	"os"
	"fmt"
	"path"
	"io/fs"
	"bytes"
	"strings"
	_ "embed"
	"net/http"

	"github.com/quay/claircore/pkg/tarfs"
)

//go:embed webroot.tar
var spaTarball []byte


type SPAHandler struct {
	index string
	fsys *tarfs.FS
	fshandler http.Handler
}

func newSPAHandler() (*SPAHandler, error) {
	reader := bytes.NewReader(spaTarball)

	h := SPAHandler{
		index: fmt.Sprintf("index.html"),
	}

	var err error
	if h.fsys, err = tarfs.New(reader); err != nil {
		return nil, err
	}
	h.fshandler = http.FileServerFS(h.fsys)

	return &h, nil
}

func (h *SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	requestPath := path.Clean(r.URL.Path)
	fmt.Printf("Request: %s\n", requestPath)

	if strings.HasPrefix(requestPath, `/`) {
		requestPath = requestPath[1:]
	}

	if len(requestPath)==0 {
		requestPath = h.index
	}
	fmt.Printf(" adjusted: %s\n", requestPath)

	var st fs.FileInfo
	var err error
	if st, err = h.fsys.Stat(requestPath); os.IsNotExist(err) || st.IsDir() {
		requestPath = h.index
	} else if err != nil {
		fmt.Printf(" err: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf(" Serving on fshandler: %s\n", requestPath)
	http.ServeFileFS(w, r, h.fsys, requestPath)
}
