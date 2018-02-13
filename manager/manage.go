package main

import (
    "bytes"
    "encoding/base64"
    "fmt"
    "html/template"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "time"

	"github.com/gorilla/mux"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/ec2"
)

func waitForInstance(client *ec2.EC2, instanceId string) float64 {
    requested := time.Now()

    err := client.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput {
        InstanceIds: aws.StringSlice([]string {instanceId}),
    })

    if err != nil {
        log.Print(err)
    }

    started := time.Now()

    return started.Sub(requested).Seconds()
}

func runInstance(userdata string) string {
    client := ec2.New(session.New(&aws.Config {
        Region: aws.String("YOUR_AWS_REGION"),
        Credentials: credentials.NewStaticCredentials(
            "YOUR_IAM_CRENTIALS_ACCESS_KEY_ID",
            "YOUR_IAM_SECRET_ACCESS_KEY",
            "YOUR_SESSION_TOKEN",
        ),
    }))

    res, err := client.RunInstances(&ec2.RunInstancesInput {
        ImageId: aws.String("YOUR_MLP_INSTANCE_AMI_ID"),
        InstanceType: aws.String("YOUR_MLP_INSTANCE_TYPE"),
        KeyName: aws.String("YOUR_MLP_INSTANCE_KEY_PAIR_NAME"),
        SecurityGroupIds: aws.StringSlice([]string {"YOUR_SECURITY_GROUP_ID"}),
        MinCount: aws.Int64(1),
        MaxCount: aws.Int64(1),
        InstanceInitiatedShutdownBehavior: aws.String("terminate"),
        UserData: aws.String(userdata),
    })

    if err != nil {
        log.Print(err)
    }

    instanceId := aws.StringValue(res.Instances[0].InstanceId)
    println("request new instance:", instanceId)

    elapsed := waitForInstance(client, instanceId)
    fmt.Printf("time elapsed until instance running: %.2fs\n", elapsed)

    des, err := client.DescribeInstances(&ec2.DescribeInstancesInput {
        InstanceIds: aws.StringSlice([]string {instanceId}),
    })

    if err != nil {
        log.Print(err)
    }

    return aws.StringValue(des.Reservations[0].Instances[0].PublicIpAddress)
}

func composeUserdata(path, mod, lib, dat string) string {
    if !(mod == "train") {
        println("unknown mode for instance:" + mod)
        os.Exit(1)
    }

    scr, err := ioutil.ReadFile(path + "autorun_template")
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
    homepath := os.Getenv("MLP_HOME")
    router := mux.NewRouter()

    router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        mainPage := "upload.html"

        switch r.Method {
        case "GET":
            tem, err := template.ParseFiles(mainPage)
            if err != nil {
                log.Print(err)
            }

            tem.Execute(w, nil)
        case "POST":
            src, _, err := r.FormFile("sourcecode")
            if err != nil {
                log.Print(err)
            }
            defer src.Close()

            dst, err := os.Create(homepath + "dat/main.py")
            if err != nil {
                log.Print(err)
            }
            defer dst.Close()

            io.Copy(dst, src)

            userdata := composeUserdata(
                homepath + "script/",
                r.FormValue("mode"),
                r.FormValue("library"),
                r.FormValue("dataset"),
            )

            publicIpAddress := runInstance(userdata)
            println("instance ip:", publicIpAddress)
        default:
            println("unknow http method:" + r.Method)
        }
    })

    router.HandleFunc("/getfile/{file}", func(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        src, err := os.Open(homepath + "dat/" + vars["file"])
        if err != nil {
            http.Error(w, "File not found", 404)
            return
        }
        defer src.Close()

        w.Header().Set("Content-Disposition", "attachment; filename=a.txt")
        w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
        w.Header().Set("Content-Length", r.Header.Get("Content-Length"))

        io.Copy(w, src)
    })

    http.ListenAndServe(":8080", router)
}
