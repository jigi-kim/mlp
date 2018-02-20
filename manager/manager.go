package main

import (
    "bytes"
    "encoding/base64"
    "html/template"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "strings"
)

var homepath string

func saveSourceCode(src io.Reader, dat string) {
    path := homepath + "efs/user/" + dat + "/src/"
    saveAs, err := os.Create(path + "main.py")
    if err != nil {
        log.Print(err)
    }
    defer saveAs.Close()

    io.Copy(saveAs, src)
}

func composeUserdata(mod, lib, dat string) string {
    if !(mod == "train" || mod == "test") {
        println("unknown mode for instance:" + mod)
        os.Exit(1)
    }

    path := homepath + "script/"
    scr, err := ioutil.ReadFile(path + "userdata_template")
    if err != nil {
        log.Print(err)
    }

    scr = bytes.Replace(scr, []byte("token_mod"), []byte(mod), -1)
    scr = bytes.Replace(scr, []byte("token_lib"), []byte(lib), -1)
    scr = bytes.Replace(scr, []byte("token_dat"), []byte(dat), -1)

    userdata := base64.StdEncoding.EncodeToString(scr)

    return userdata
}

func main() {
    homepath = os.Getenv("HOMEPATH")

    if homepath == "" {
        log.Fatal("error: homepath is not set")
    }

    monitor := NewMonitor(
        "YOUR_AWS_REGION",
        "YOUR_IAM_CRENTIALS_ACCESS_KEY_ID",
        "YOUR_IAM_SECRET_ACCESS_KEY",
    )

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case "GET":
            tem, err := template.ParseFiles("script/upload.html")
            if err != nil {
                log.Print(err)
            }

            tem.Execute(w, nil)
        case "POST":
            mod := r.FormValue("mode")
            lib := r.FormValue("library")
            dat := r.FormValue("dataset")

            src, _, err := r.FormFile("sourcecode")
            if err != nil {
                log.Print(err)
            }
            defer src.Close()

            saveSourceCode(src, dat)
            userdata := composeUserdata(mod, lib, dat)

            monitor.RunInstance(
                "YOUR_MLP_INSTANCE_AMI_ID",
                "YOUR_MLP_INSTANCE_TYPE",
                "YOUR_MLP_INSTANCE_KEY_PAIR_NAME",
                "YOUR_SECURITY_GROUP_ID",
                userdata,
            )
        default:
            println("unknown http method:" + r.Method)
        }
    })

    targetAddr := ""
    http.HandleFunc("/tb", func(w http.ResponseWriter, r *http.Request) {
        if targetAddr != "" {
            http.Redirect(w,r,"http://"+targetAddr+":6006", 301)
        }
    })

    http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case "PUT":
            instanceIp := strings.Split(r.RemoteAddr, ":")[0]
            status := r.FormValue("status")

            monitor.UpdateInstanceState(instanceIp, status)
        default:
            println("unknown http method:" + r.Method)
        }
    })

    println("manager is now running.")
    http.ListenAndServe(":8080", nil)
}
