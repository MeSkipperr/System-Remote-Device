package handlers

import (
    "html/template"
    "net/http"
    "path/filepath"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
    tmplPath := filepath.Join("template","html","index.html")
    tmpl, err := template.ParseFiles(tmplPath)
    if err != nil {
        http.Error(w, "Error loading template", http.StatusInternalServerError)
        return
    }

    tmpl.Execute(w, nil)
}