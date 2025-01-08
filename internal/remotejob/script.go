// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package remotejob

var jvmDumpHostScript = `import argparse
import sys
import os
import datetime
import subprocess
import hmac
import hashlib
import base64
import requests


def execute_jmap_command(pid, filename, jmap):
    try:
        command = f"{jmap} -dump:live,format=b,file={filename} {pid}"
        subprocess.run(command, shell=True, check=True)
        print(f"jmap command executed successfully for PID {pid}")
    except subprocess.CalledProcessError as e:
        print(f"Failed to execute jmap command: {e}")
        sys.exit(1)


def upload_to_oss(file_name, osspath):
    bucket_host = os.getenv('OSS_BUCKET_HOST')
    access_key_id = os.getenv('OSS_ACCESS_KEY_ID')
    access_key_secret = os.getenv('OSS_ACCESS_KEY_SECRET')
    bucket_name = os.getenv('OSS_BUCKET_NAME')
    if not bucket_host or bucket_host == "":
        print("bucket host is none")
        return
    if not bucket_name or bucket_name == "":
        print("bucket name is none")
        return

    if not access_key_id or access_key_id == "":
        print("access key id is none")
        return

    if not access_key_secret or access_key_secret == "":
        print("access key secret id is none")
        return

    try:
        upload_to_oss_s(bucket_name, bucket_host, access_key_id, access_key_secret, osspath, file_name)
    except Exception as e:
        print("Error occurred while uploading to OSS: {0}".format(e))


def get_file_mime_type(file_path):
    result = subprocess.run(['file', '--mime-type', '-b', file_path], capture_output=True, text=True)
    mime_type = result.stdout.strip().split(';')[0]
    return mime_type


def generate_oss_signature(method, content_type, date_value, resource, access_key_secret):
    string_to_sign = f"{method}\n\n{content_type}\n{date_value}\n{resource}"
    secret_enc = access_key_secret.encode('utf-8')
    digest = hmac.new(secret_enc, string_to_sign.encode('utf-8'), hashlib.sha1).digest()
    signature = base64.b64encode(digest).decode('utf-8')
    return signature


def upload_to_oss_s(bucket, host, access_key_id, access_key_secret, path, service_file):
    osshost = f"{bucket}.{host}"

    fi = os.path.basename(service_file)
    new_file = fi

    dest = f"{path}/{new_file}"

    resource = f"/{bucket}/{dest}"
    print(f"servicefile = {service_file} fi={fi} newfile={new_file} dest={dest} resource={resource}")

    content_type = get_file_mime_type(service_file)
    print("content type")
    date_value = datetime.datetime.utcnow().strftime('%a, %d %b %Y %H:%M:%S GMT')
    print(" date_value")
    signature = generate_oss_signature('PUT', content_type, date_value, resource, access_key_secret)
    print(" signature")
    url = f"http://{osshost}/{dest}"

    print(f"Uploading {service_file} to {url}")

    headers = {
        'Host': osshost,
        'Date': date_value,
        'Content-Type': content_type,
        'Authorization': f"OSS {access_key_id}:{signature}"
    }
    try:
        with open(service_file, 'rb') as file:
            response = requests.put(url, data=file, headers=headers)

        print(response.status_code, response.text)
    except requests.exceptions.RequestException as e:
        print(e, file=sys.stderr)



def upload_to_aws(file_name, path):
    bucket_name = os.getenv('AWS_BUCKET_NAME')
    access_key = os.getenv('AWS_ACCESS_KEY_ID')
    secret_key = os.getenv('AWS_SECRET_ACCESS_KEY')
    region = os.getenv('AWS_DEFAULT_REGION')
    fi = os.path.basename(file_name)
    object_key = f"{path}/{fi}"

    print(f"file:{file_name} upload to region/bucket {region}/{bucket_name} object_key is {object_key}")

    try:
        import boto3
        s3 = boto3.client('s3', aws_access_key_id=access_key,
                          aws_secret_access_key=secret_key,
                          region_name=region)
        s3.upload_file(file_name, bucket_name, object_key)
        print(f"file: {file_name} upload to: {bucket_name} , object_key: {object_key}, ok !")
    except Exception as e:
        print(f"err: {e}")


def upload_to_obs(file_name, path):
    bucket_name = os.getenv('OBS_BUCKET_NAME')
    access_key = os.getenv('OBS_ACCESS_KEY_ID')
    secret_key = os.getenv('OBS_SECRET_ACCESS_KEY')
    server = os.getenv('OBS_SERVER')
    fi = os.path.basename(file_name)
    object_key = f"{path}/{fi}"
    print(f"file:{file_name} upload to server/bucket {server}/{bucket_name} object_key is {object_key}")

    try:
        from obs import ObsClient

        obsClient = ObsClient(access_key_id=access_key, secret_access_key=secret_key, server=server)

        metadata = {'remote': 'job'}
        # 文件上传
        resp = obsClient.putFile(bucket_name, object_key, file_name, metadata)
        # 返回码为2xx时，接口调用成功，否则接口调用失败
        if resp.status < 300:
            print('Put File Succeeded')
            print('requestId:', resp.requestId)
            print('etag:', resp.body.etag)
            print('versionId:', resp.body.versionId)
            print('storageClass:', resp.body.storageClass)
        else:
            print('Put File Failed')
            print('requestId:', resp.requestId)
            print('errorCode:', resp.errorCode)
            print('errorMessage:', resp.errorMessage)
    except Exception as e:
        print(f"Put File Failed {e}")



def main():
    parser = argparse.ArgumentParser(description='command args')

    parser.add_argument('-pid', '--pid', type=int, required=True, help='Process ID')
    parser.add_argument('-osspath', '--osspath', type=str, required=False, help='oss path')
    parser.add_argument("-javahome", '--javahome', type=str, required=False, help='java name')

    args = parser.parse_args()

    if (args.osspath == "") or (args.osspath == None):
        args.osspath = "jvmdump"

    print(f"args: {args}")
    print(f"process ID: {args.pid}")
    print(f"oss path: {args.osspath}")
    print(f"java_home: {args.javahome}")

    remote = os.getenv('REMOTE')

    if not remote or remote == "":
        remote = "oss"  # set default is oss
        print("set to oss remote")

    timestamp = datetime.datetime.utcnow().strftime("%Y-%m-%d-%H-%M")
    filename = f"/tmp/heap-pid-{args.pid}-{timestamp}"

    jmap = args.javahome + "/bin/" + "jmap"
    execute_jmap_command(args.pid, filename, jmap)

    if remote == "oss":
        upload_to_oss(filename, args.osspath)
        return

    if remote == "aws":
        upload_to_aws(filename, args.osspath)
        return

    if remote == "obs":
        upload_to_obs(filename, args.osspath)
        return


if __name__ == "__main__":
    main()
`

var jvmDumpK8sScript = `import argparse
import os
import datetime
import subprocess
import hmac
import hashlib
import base64
import requests
from kubernetes import client, config, stream


def upload_to_oss(file_name, osspath):
    bucket_host = os.getenv('OSS_BUCKET_HOST')
    access_key_id = os.getenv('OSS_ACCESS_KEY_ID')
    access_key_secret = os.getenv('OSS_ACCESS_KEY_SECRET')
    bucket_name = os.getenv('OSS_BUCKET_NAME')
    if not bucket_host or bucket_host == "":
        print("bucket host is none")
        return
    if not bucket_name or bucket_name == "":
        print("bucket name is none")
        return

    if not access_key_id or access_key_id == "":
        print("access key id is none")
        return

    if not access_key_secret or access_key_secret == "":
        print("access key secret id is none")
        return

    try:
        upload_to_oss_s(bucket_name, bucket_host, access_key_id, access_key_secret, osspath, file_name)
    except Exception as e:
        print(f"Error occurred while uploading to OSS: {e}")


def get_file_mime_type(file_path):
    result = subprocess.run(['file', '--mime-type', '-b', file_path], capture_output=True, text=True)
    mime_type = result.stdout.strip().split(';')[0]
    return mime_type


def generate_oss_signature(method, content_type, date_value, resource, access_key_secret):
    string_to_sign = f"{method}\n\n{content_type}\n{date_value}\n{resource}"
    secret_enc = access_key_secret.encode('utf-8')
    digest = hmac.new(secret_enc, string_to_sign.encode('utf-8'), hashlib.sha1).digest()
    signature = base64.b64encode(digest).decode('utf-8')
    return signature


def upload_to_oss_s(bucket, host, access_key_id, access_key_secret, path, service_file):
    osshost = f"{bucket}.{host}"

    fi = os.path.basename(service_file)
    new_file = fi

    dest = f"{path}/{new_file}"

    resource = f"/{bucket}/{dest}"
    print(f"servicefile = {service_file} fi={fi} newfile={new_file} dest={dest} resource={resource}")

    content_type = get_file_mime_type(service_file)
    print("content type")
    date_value = datetime.datetime.utcnow().strftime('%a, %d %b %Y %H:%M:%S GMT')
    print(" date_value")
    signature = generate_oss_signature('PUT', content_type, date_value, resource, access_key_secret)
    print(" signature")
    url = f"http://{osshost}/{dest}"

    print(f"Uploading {service_file} to {url}")

    headers = {
        'Host': osshost,
        'Date': date_value,
        'Content-Type': content_type,
        'Authorization': f"OSS {access_key_id}:{signature}"
    }

    with open(service_file, 'rb') as file:
        response = requests.put(url, data=file, headers=headers)

    print(response.status_code, response.text)


def exec_command_in_pod(pod_name, pid, filename, namespace):
    config.load_incluster_config()
    print(f"command in pod :pid={pid} filename={filename}")
    v1 = client.CoreV1Api()
    command = ['jmap', f"-dump:live,format=b,file={filename}", f"{pid}"]

    try:
        resp = stream.stream(
            v1.connect_get_namespaced_pod_exec,
            name=pod_name,
            namespace=namespace,
            command=command,
            stderr=True,
            stdin=False,
            stdout=True,
            tty=False,
            _preload_content=False
        )

        print(f"stderr: {resp.readline_stderr(timeout=20)}")
        print(f"stdout: {resp.readline_stdout(timeout=20)}")
        print(f"peek out {resp.peek_stdout()}")
        print(f"read_all {resp.read_all()}")
    except Exception as e:
        print(f"An error occurred: {e}")


def copy_file_from_pod_to_container(pod_name, source_file_path, namespace):
    config.load_incluster_config()
    api_instance = client.CoreV1Api()

    exec_command = ['cat', source_file_path]
    resp = stream.stream(api_instance.connect_get_namespaced_pod_exec, pod_name, namespace,
                         command=exec_command,
                         stderr=True, stdin=False,
                         stdout=True, tty=False,
                         binary=True, _preload_content=False)

    with open(source_file_path, 'wb') as dest_file:
        while resp.is_open():
            resp.update(timeout=1)
            if resp.peek_stdout():
                dest_file.write(resp.read_stdout())
            if resp.peek_stderr():
                print("STDERR: %s" % resp.read_stderr())
        resp.close()

def get_pod_namespace(pod_name):
    config.load_incluster_config()
    v1 = client.CoreV1Api()
    ret = v1.list_pod_for_all_namespaces(watch=False)

    for pod in ret.items:
        if pod.metadata.name == pod_name:
            print(f"find pod:{pod_name} in namespace:{pod.metadata.namespace}")
            return pod.metadata.namespace

    return None

def upload_to_aws(file_name, path):
    bucket_name = os.getenv('AWS_BUCKET_NAME')
    access_key = os.getenv('AWS_ACCESS_KEY_ID')
    secret_key = os.getenv('AWS_SECRET_ACCESS_KEY')
    region = os.getenv('AWS_DEFAULT_REGION')
    fi = os.path.basename(file_name)
    object_key = f"{path}/{fi}"
    print(f"file:{file_name} upload to region/bucket {region}/{bucket_name} object_key is {object_key}")
    # s3 bucket client
    s3 = boto3.client('s3', aws_access_key_id=access_key,
                      aws_secret_access_key=secret_key,
                      region_name=region)
    try:
        # 执行上传操作
        s3.upload_file(file_name, bucket_name, object_key)
        print(f"file: {file_name} upload to: {bucket_name} , object_key: {object_key}, ok !")
    except Exception as e:
        print(f"err: {e}")


def upload_to_obs(file_name, path):
    bucket_name = os.getenv('OBS_BUCKET_NAME')
    access_key = os.getenv('OBS_ACCESS_KEY_ID')
    secret_key = os.getenv('OBS_SECRET_ACCESS_KEY')
    server = os.getenv('OBS_SERVER')
    fi = os.path.basename(file_name)
    object_key = f"{path}/{fi}"
    print(f"file:{file_name} upload to server/bucket {server}/{bucket_name} object_key is {object_key}")

    obsClient = ObsClient(access_key_id=access_key, secret_access_key=secret_key, server=server)
    try:
        bucketName = bucket_name
        metadata = {'remote': 'job'}
        # 文件上传
        resp = obsClient.putFile(bucketName, object_key, file_name, metadata)
        # 返回码为2xx时，接口调用成功，否则接口调用失败
        if resp.status < 300:
            print('Put File Succeeded')
            print('requestId:', resp.requestId)
            print('etag:', resp.body.etag)
            print('versionId:', resp.body.versionId)
            print('storageClass:', resp.body.storageClass)
        else:
            print('Put File Failed')
            print('requestId:', resp.requestId)
            print('errorCode:', resp.errorCode)
            print('errorMessage:', resp.errorMessage)
    except:
        print('Put File Failed')


def main():
    parser = argparse.ArgumentParser(description='command args')

    parser.add_argument('-pid', type=int, required=True, help='process id')
    parser.add_argument('-osspath', type=str, required=False, help='oss oss path')
    parser.add_argument('-pod_name', type=str, required=False, help='pod name')

    args = parser.parse_args()

    print(f"args: {args}")
    print(f"process ID: {args.pid}")
    print(f"oss path: {args.osspath}")
    print(f"pod_name: {args.pod_name}")
    remote = os.getenv('REMOTE')

    if not remote or remote == "":
        remote = "oss"  # set default is oss
        print("set to oss remote")

    timestamp = datetime.datetime.utcnow().strftime("%Y-%m-%d-%H-%M")
    filename = f"/tmp/heap-pid-{args.pid}-{timestamp}"

    # first get namespace
    namespace = get_pod_namespace(args.pod_name)
    if namespace is None:
        print(f"can find {args.pod_name} namespace")
        return

    # 2 jmap
    exec_command_in_pod(args.pod_name, args.pid, filename, namespace)

    # 3 kubectl cp
    copy_file_from_pod_to_container(args.pod_name, filename, namespace)

    # 4 upload to oss
    if remote == "oss":
        upload_to_oss(filename, args.osspath)
        return

    if remote == "aws":
        upload_to_aws(filename, args.osspath)
        return

    if remote == "obs":
        upload_to_obs(filename, args.osspath)
        return


if __name__ == "__main__":
    main()

`
