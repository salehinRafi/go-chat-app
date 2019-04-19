package main

import (
	"flag"
	"goWork_chat/trace"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"
)

// templ represents a single template
type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

// ServeHTTP handles the HTTP request.
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.templ.Execute(w, r)
	/*
	 *  tells the template to render itself using data that can be extracted from http. Request, which happens to include the host address that we need.
	 */
}

func main() {
	var addr = flag.String("addr", ":8080", "The addr of the application.") // command-line flags to make it configurable, default is ":8080" | can run by `./chat -addr=""192.168.0.1:3000"`
	flag.Parse()                                                            // parse the flags
	r := newRoom()
	r.tracer = trace.New(os.Stdout) // comment this to ignore any calls to Trace
	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/room", r)
	// get the room going
	go r.run()
	// start the web server
	log.Println("Starting web server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
