apiVersion: v1
kind: Namespace
metadata:
  name: datakit-monitor

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: datakit-monitor
rules:
- apiGroups:
  - ""
  resources:
  - nodes
  - nodes/proxy
  - namespaces
  - pods
  - services
  - endpoints
  - persistentvolumes
  - persistentvolumeclaims
  - ingresses
  verbs:
  - get
  - list
- apiGroups:
  - apps
  resources:
  - deployments
  - daemonsets
  - statefulsets
  - replicasets
  verbs:
  - get
  - list
- apiGroups:
  - "extensions"
  resources:
  - ingresses
  - deployments
  - daemonsets
  - statefulsets
  - replicasets
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - batch
  resources:
  - jobs
  verbs:
  - get
  - list
- nonResourceURLs: ["/metrics"]
  verbs: ["get"]

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: datakit-monitor
  namespace: datakit-monitor

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: datakit-monitor
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: datakit-monitor
subjects:
- kind: ServiceAccount
  name: datakit-monitor
  namespace: datakit-monitor

---

apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    app: daemonset-datakit
  name: datakit-monitor
  namespace: datakit-monitor
spec:
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: daemonset-datakit
  template:
    metadata:
      labels:
        app: daemonset-datakit
    spec:
      containers:
      - env:
        - name: HOSTIP
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.hostIP
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: ENV_DATAWAY
          value: <dataway_url>
        - name: ENV_GLOBAL_TAGS
          value: host=__datakit_hostname,host_ip=__datakit_ip
        - name: ENV_ENABLE_INPUTS
          value: cpu,disk,diskio,mem,swap,system,hostobject,net,host_processes,kubernetes,container
        - name: TZ
          value: Asia/Shanghai
        image: pubrepo.jiagouyun.com/datakit/datakit:1.1.7-rc0
        imagePullPolicy: Always
        name: datakit
        ports:
        - containerPort: 9529
          hostPort: 9529
          name: port
          protocol: TCP
        resources: {}
        securityContext:
          privileged: true
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /var/run/docker.sock
          name: docker-socket
          readOnly: true
        - name: tz-config
          mountPath: /etc/localtime
          readOnly: true
        {{- range $inputName, $cfg := .Inputs }}
        #- mountPath: /usr/local/datakit/conf.d/{{$inputName}}.conf
        #  name: datakit-monitor-conf
        #  subPath: {{$inputName}}.conf
        {{- end }}
        - mountPath: /host/proc
          name: proc
          readOnly: true
        - mountPath: /host/dev
          name: dev
          readOnly: true
        - mountPath: /host/sys
          name: sys
          readOnly: true
        - mountPath: /rootfs
          name: rootfs
        workingDir: /usr/local/datakit
      dnsPolicy: ClusterFirst
      hostIPC: true
      hostNetwork: true
      hostPID: true
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      serviceAccount: datakit-monitor
      serviceAccountName: datakit-monitor
      terminationGracePeriodSeconds: 30
      volumes:
      - name: tz-config
        hostPath:
          path: /etc/localtime
      - configMap:
          name: datakit-monitor-conf
          items:
      {{- range $inputName, $cfg := .Inputs }}
          - key: {{$inputName}}.conf
            path: {{$inputName}}.conf
      {{- end }}
        name: datakit-monitor-conf
      - hostPath:
          path: /var/run/docker.sock
        name: docker-socket
      - hostPath:
          path: /proc
          type: ""
        name: proc
      - hostPath:
          path: /dev
          type: ""
        name: dev
      - hostPath:
          path: /sys
          type: ""
        name: sys
      - hostPath:
          path: /
          type: ""
        name: rootfs
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 1
    type: RollingUpdate
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: datakit-monitor-conf
  namespace: datakit-monitor
data:
  {{- range $inputName, $cfg := .Inputs }}
    #### {{$inputName}}
    {{- $inputName -}}.conf: |-
        {{- $cfg | indent 6 -}}
  {{- end -}}
