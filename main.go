package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"text/template"
)

var port = flag.String("port", "1337", "port to listen on")

func main() {
	flag.Parse()
	store := NewMemoryStore()
	s := NewServer(store)

	fmt.Printf("listening on %s\n", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", *port), s); err != nil {
		log.Fatal(err)
	}
}

func NewServer(store Store) *http.ServeMux {
	tmp, err := template.ParseFiles("./index.html")
	if err != nil {
		slog.Error("[NewServer]: ", "error", err)
		return nil
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", index(tmp, store))
	mux.HandleFunc("GET /{id}", index(tmp, store))
	mux.HandleFunc("POST /paste", createPaste(store))
	mux.HandleFunc("GET /paste/{id}", getPaste(store))

	return mux
}

func index(tmp *template.Template, s Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		p := s.Get(id)

		err := tmp.Execute(w, struct{ Content string }{Content: p.Content})
		if err != nil {
			slog.Error("[index]: ", "error", err)
		}
	}
}

func createPaste(s Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var content string

		if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
			r.ParseForm()
			content = r.FormValue("content")
		} else {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				slog.Error("[createPaste]: ", "error", err.Error())
				w.WriteHeader(500)
				return
			}

			var pp Paste
			err = json.Unmarshal(body, &pp)
			if err != nil {
				slog.Error("[createPaste]: ", "error", err.Error())
				w.WriteHeader(500)
				return
			}

			content = pp.Content
		}

		p := s.Add(content)

		p.Content = ""

		j, err := json.Marshal(p)
		if err != nil {
			slog.Error("[createPaste]: ", "error", err.Error())
			w.WriteHeader(500)
			return
		}

		w.Write(j)
	}
}

func getPaste(s Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		p := s.Get(id)

		j, err := json.Marshal(p)
		if err != nil {
			slog.Error("[getPaste]: ", "error", err.Error())
			w.WriteHeader(500)
			return
		}

		w.Write(j)
	}
}

type Paste struct {
	ID      string `json:"id,omitempty"`
	Content string `json:"content,omitempty"`
}

type Store interface {
	Add(content string) Paste
	Get(id string) Paste
}

type MemoryStore struct {
	store map[string]Paste
}

func (ms *MemoryStore) Add(content string) Paste {
	p := Paste{
		Content: content,
	}

	return ms.addNoConflict(p)
}

func (ms *MemoryStore) addNoConflict(p Paste) Paste {
	id := randString()
	_, exists := ms.store[id]

	if exists {
		return ms.addNoConflict(p)
	}

	p.ID = id
	ms.store[id] = p
	return p
}

func (ms *MemoryStore) Get(id string) Paste {
	p := ms.store[id]
	delete(ms.store, id)
	return p
}

func NewMemoryStore() Store {
	return &MemoryStore{
		store: make(map[string]Paste),
	}
}

func randString() string {
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	length := 8
	bs := make([]byte, length)

	for i := range bs {
		bs[i] = letters[rand.Intn(len(letters))]
	}

	return string(bs)
}
