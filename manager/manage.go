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

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/ec2"
)

var homepath = ""

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
    if !(mod == "train") {
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

func runInstance(userdata string) {
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

    ipAddr := aws.StringValue(des.Reservations[0].Instances[0].PublicIpAddress)
    println("instance ip:", ipAddr)
}

func main() {
    homepath = os.Getenv("MLP_HOME")

    if homepath == "" {
        log.Fatal("error: homepath is not set")
    }

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

            go runInstance(userdata)
        default:
            println("unknown http method:" + r.Method)
        }
    })

    http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case "PUT":
            println(r.RemoteAddr)
            status := r.FormValue("status")
            println("status:", status)
        default:
            println("unknown http method:" + r.Method)
        }
    })

    println("manager is now running.")
    http.ListenAndServe(":8080", nil)
}
