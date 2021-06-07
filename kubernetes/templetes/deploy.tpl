# create namespace
apiVersion: v1
kind: Namespace
metadata:
  name: datakit-monitor
  labels:
    name: datakit

---
# create ServiceAccount
apiVersion: v1
kind: ServiceAccount
metadata:
  name: datakit-account
  namespace: datakit-monitor
  labels:
    name: datakit

---
# create ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: datakit-monitor
rules:
- apiGroups:
  - ""
  resources:
  - nodes
  - namespaces
  - pods
  - services
  - endpoints
  - persistentvolumes
  - persistentvolumeclaims
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

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: datakit
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: datakit
subjects:
- kind: ServiceAccount
  name: dkaccount
  namespace: monitor

---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    app: daemonset-datakit
  name: datakit
  namespace: monitor
spec:
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
        - name: ENV_UUID
          value: dkit_oX1qnIiXNPft2ebsAchOYXucuIB
        - name: ENV_DATAWAY
          value: <dataway_url>
        - name: ENV_GLOBAL_TAGS
          value: host=__datakit_hostname,host_ip=__datakit_ip
        - name: ENV_ENABLE_INPUTS
          value: cpu,disk,diskio,mem,swap,system,hostobject,net,host_processes,docker
        image: pubrepo.jiagouyun.com/datakit/datakit:1.1.7-rc0
        imagePullPolicy: Always
        name: datakit
        ports:
        - containerPort: 9529
          hostPort: 9529
          name: port
          protocol: TCP
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /var/run/docker.sock
          name: docker-socket
          readOnly: true
        workingDir: /usr/local/datakit
      hostIPC: true
      hostNetwork: true
      hostPID: true
      restartPolicy: Always
      serviceAccount: dkaccount
      serviceAccountName: dkaccount
      volumes:
      - configMap:
          defaultMode: 420
          name: datakit-conf
        name: datakit-conf
      - hostPath:
          path: /var/run/docker.sock
        name: docker-socket
      - configMap:
          defaultMode: 256
          name: pipeline
          optional: false
        name: pipeline
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: datakit
  namespace: default
data:
  {{ range $inputName, $cfg := .Inputs }}

  #### {{$inputName}}
  {{$inputName}}.conf: |
    {{$cfg}}
  {{ end }}
