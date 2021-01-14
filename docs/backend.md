### Backend Server
The main job loop for the backend server is similar to the one using by the proxy server:
```go
func startImgServer(connManager netmanager.ConnManager) {
	opts := &t.ResizerOptions{}
	for {
		mainJob := job.NewJob(opts)
		mainJob.AddOneshotTask(connManager.AcceptTask)
		mainJob.AddTaskWithIdleTimeout(netmanager.ReadTask, time.Second * 2)
		mainJob.AddTask(netmanager.WriteTask)
		mainJob.AddTask(opts.ResizeImageTask)
		<-mainJob.RunInBackground()
		go func() {
			j := mainJob
			for {
				select {
				case <-j.JobDoneNotify():
					_, err := mainJob.GetInterruptedBy()
					j.Log(1) <- fmt.Sprintf("#%d job is %s, error: %s",
						counter + 1, strings.ToLower(j.GetState().String()), err)
					counter++
					j.Log(2) <- fmt.Sprintf("N gouroutines %d", runtime.NumGoroutine())
 					return
				}
			}
		}()
	}
}
```