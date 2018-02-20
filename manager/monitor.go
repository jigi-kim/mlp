package main

import (
    "fmt"
    "log"
    "time"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/ec2"
)

var instanceStatus = [...]string {
    "requested",
    "running",
    "initialized",
    "runcontainer",
    "halt",
    "error",
}

type Instance struct {
    instanceId string
    publicIp string

    status string
    newStatus chan string
}

func NewInstance(instanceId string) *Instance {
    self := new(Instance)

    self.instanceId = instanceId
    self.status = instanceStatus[0]
    self.newStatus = make(chan string)

    return self
}

type Monitor struct {
    NrInstances int

    client *ec2.EC2
    instances map[string]*Instance
}

func NewMonitor(region, accessKey, secretKey string) *Monitor {
    self := new(Monitor)

    self.NrInstances = 0
    self.client = ec2.New(session.New(&aws.Config {
        Region: aws.String(region),
        Credentials: credentials.NewStaticCredentials(
            accessKey, secretKey, "",
        ),
    }))

    self.instances = map[string]*Instance {}

    return self
}

func (self *Monitor) RunInstance(ami, ins, key, sec, userdata string) {
    res, err := self.client.RunInstances(&ec2.RunInstancesInput {
        ImageId: aws.String(ami),
        InstanceType: aws.String(ins),
        KeyName: aws.String(key),
        SecurityGroupIds: aws.StringSlice([]string {sec}),
        MinCount: aws.Int64(1),
        MaxCount: aws.Int64(1),
        InstanceInitiatedShutdownBehavior: aws.String("terminate"),
        UserData: aws.String(userdata),
    })

    if err != nil {
        log.Print(err)
    }

    instanceId := aws.StringValue(res.Instances[0].InstanceId)
    instance := NewInstance(instanceId)

    println("requested new instance:", instanceId)

    go self.monitorInstance(instance)
    go self.registerInstance(instance)
}

func (self *Monitor) registerInstance(instance *Instance) {
    desInput := ec2.DescribeInstancesInput {
        InstanceIds: aws.StringSlice([]string {instance.instanceId}),
    }

    err := self.client.WaitUntilInstanceRunning(&desInput)
    if err != nil {
        log.Print(err)
    }

    des, err := self.client.DescribeInstances(&desInput)
    if err != nil {
        log.Print(err)
    }

    instance.newStatus <- "running"

    addr := aws.StringValue(des.Reservations[0].Instances[0].PublicIpAddress)
    self.instances[addr] = instance
    self.NrInstances++
}

func (self *Monitor) monitorInstance(instance *Instance) {
    prevTime := time.Now()

    for instance.status != "halt" {
        previousStatus := instance.status

        instance.status = <-instance.newStatus

        currTime := time.Now()
        elapsed := currTime.Sub(prevTime).Seconds()
        prevTime = currTime

        fmt.Printf("[%s] %s -> %s (%.2f sec)\n",
            instance.instanceId,
            previousStatus,
            instance.status,
            elapsed,
        )
    }

    delete(self.instances, instance.publicIp)
    self.NrInstances--
}

func (self *Monitor) UpdateInstanceState(publicIp string, status string) {
    if _, exist := self.instances[publicIp]; exist {
        self.instances[publicIp].newStatus <- status
    }
}
