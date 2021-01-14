### Proxy Server
The main loop:
```go
for {
    balancerJob := job.NewJob(nil)
    balancerJob.AddOneshotTask(connMngr.AcceptTask)
    balancerJob.AddTask(balancer.LoadBalance)
    <-balancerJob.RunInBackground()

    go func() {
        for {
            select {
            case <- balancerJob.JobDoneNotify():
                _, err := balancerJob.GetInterruptedBy()
                balancerJob.Log(2) <- fmt.Sprintf("job is %s, error '%v'",  balancerJob.GetState(), err)
                return
            }
        }
    }()
}
```
The loop is being blocked by _connMngr.AcceptTask_ task, waiting for a new connection. Once a connection is
established, it will execute other two tasks handling the connection in background, allowing loop execution to proceed,
creating again a new job waiting for a new connection, and so on. 