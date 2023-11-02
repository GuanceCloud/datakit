## Datakit Golang profiling tool for docker/k8s, based on golang pprof.

### Build

```shell
$ cd k8s-profilers
$ make pprof
```

### Usage

1. Modify your k8s Deployment/Pod yaml.

Imagine your original Deployment manifest is like below:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: movies-go
  labels:
    app: movies-go
spec:
  replicas: 1
  selector:
    matchLabels:
      app: movies-go
  template:
    metadata:
      name: movies-go
      labels:
        app: movies-go
    spec:
      containers:
        - name: movies-go
          image: zhangyicloud/movies-go:latest
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 8080
      restartPolicy: Always
```

To use this profiling tool, We can add it into Pod as a sidecar like below:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: movies-go
  labels:
    app: movies-go
spec:
  replicas: 1
  selector:
    matchLabels:
      app: movies-go
  template:
    metadata:
      name: movies-go
      labels:
        app: movies-go
    spec:
      shareProcessNamespace: true
      containers:
        - name: movies-go
          image: zhangyicloud/movies-go:latest
          imagePullPolicy: IfNotPresent
          env:
            - name: DD_TRACE_ENABLED
              value: "false"
            - name: DD_PROFILING_ENABLED
              value: "false"
          
        - name: datakit-profiler
          image: pubrepo.guance.com/dataflux/pprof:0.1.0
          imagePullPolicy: IfNotPresent
          workingDir: /app/datakit-profiler
          env:
            - name: DK_AGENT_HOST # datakit listening host
              value: "192.168.209.128"
            - name: DK_AGENT_PORT # datakit listening port
              value: "9529"
            - name: DK_PROFILE_VERSION # user's app version 
              value: "1.2.333"
            - name: DK_PROFILE_ENV # user's app run env
              value: "prod"
            - name: DK_PROFILE_EVENT # specify the profiling events
              value: cpu,heap,mutex,block,goroutine
            - name: DK_PROFILE_DURATION # profiling duration at every time
              value: "240"
            - name: DK_PROFILE_SCHEDULE # profiling schedule plan
              value: "*/20 * * * *"
          command:
            - bash
            - "./cmd.sh"
          securityContext:
            capabilities:
              add:
                - SYS_PTRACE
          volumeMounts:
            - mountPath: /etc/localtime
              name: timezone
              readOnly: true
      restartPolicy: Always
      volumes:
        - name: timezone
          hostPath:
            path: /etc/localtime
            type: FileOrCreate
```

2. Attach to the `datakit-profiler` container and run the profiling script.

```shell
kubectl exec -it <your-pod-name> -c datakit-profiler -- bash
DK_PROFILE_VERSION=v0.1.0 DK_PROFILE_ENV=testing ./profiling.sh
```

Type `./profiling.sh -h` for more detail.

3. Go to the page [https://console.guance.com/tracing/profile](https://console.guance.com/tracing/profile) to see profiling detail, it may take a minute or so to load.
