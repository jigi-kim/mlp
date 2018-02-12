package main

import (
    "net/http"
    "html/template"
)

func main() {
    mainPage := "upload.html"

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "GET" {
            tem, err := template.ParseFiles(mainPage)
            if err != nil {
                panic(err)
            }

            tem.Execute(w, nil)
        }
    })

    http.ListenAndServe(":8080", nil)
}
