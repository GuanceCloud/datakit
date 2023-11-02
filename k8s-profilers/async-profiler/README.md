## Datakit JVM profiling tool for docker/k8s, based on Async-Profiler.

### Build

```shell
$ cd k8s-profilers
$ make async-profiler-prod
```

### Usage

1. Modify your k8s Deployment/Pod yaml.

Imagine your original Deployment manifest is like below:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: movies-java
  labels:
    app: movies-java
spec:
  replicas: 1
  selector:
    matchLabels:
      app: movies-java
  template:
    metadata:
      name: movies-java
      labels:
        app: movies-java
    spec:
      containers:
        - name: movies-java
          image: zhangyicloud/movies:0.3.0-java
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 8080
      restartPolicy: Always
```

To use this profiling tool, We can add it into a Pod as container helper like below:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: movies-java
  labels:
    app: movies-java
spec:
  replicas: 1
  selector:
    matchLabels:
      app: movies-java
  template:
    metadata:
      name: movies-java
      labels:
        app: movies-java
    spec:
      shareProcessNamespace: true
#      securityContext:
#        sysctls:
#          - name: kernel.perf_event_paranoid
#            value: "1"
#          - name: kernel.kptr_restrict
#            value: "0"
      containers:
        - name: movies-java
          image: zhangyicloud/movies:0.3.0-java
          imagePullPolicy: IfNotPresent
          volumeMounts:
            - mountPath: /app/datakit-profiler
              name: profile-volume
            - mountPath: /tmp
              name: tmp
        
        - name: datakit-profiler
          image: pubrepo.guance.com/dataflux/async-profiler:0.1.0
          imagePullPolicy: IfNotPresent
          volumeMounts:
            - mountPath: /etc/localtime  # Synchronize Container Timezone with host
              name: timezone
              readOnly: true
            - mountPath: /app/datakit-profiler
              name: profile-volume
          workingDir: /app/datakit-profiler
          env:
            - name: DK_AGENT_HOST # datakit listening host
              value: "192.168.209.128"
            - name: DK_AGENT_PORT # datakit listening port
              value: "9529"
            - name: DK_PROFILE_VERSION # user's app version 
              value: "1.2.333"
            - name: DK_PROFILE_ENV # user's app deployment env
              value: "prod"
            - name: DK_PROFILE_DURATION # profiling duration every time
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
                - SYS_ADMIN
      restartPolicy: Always
      volumes:
        - name: profile-volume
          emptyDir: {}
        - name: tmp
          emptyDir: {}
        - name: timezone
          hostPath:
            path: /etc/localtime
            type: FileOrCreate
```

1. Attach to the `datakit-profiler` container and run the profiling script.

```shell
kubectl exec -it <your-pod-name> -c datakit-profiler -- bash
DK_PROFILE_VERSION=v0.1.0 DK_PROFILE_ENV=testing ./profiling.sh
```

Type `./profiling.sh -h` for more detail.

1. Go to the page [https://console.guance.com/tracing/profile](https://console.guance.com/tracing/profile) to see profiling detail, it may take a minute or so to load.
