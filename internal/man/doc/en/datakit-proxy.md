# DataKit Agent
---

:fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker:

---

When Datakit cannot access the extranet, a proxy can be deployed on the intranet to send traffic. This article provides two implementations:

- Through DataKit's built-in forward proxy service
- Use Nginx as the proxy servrice

## Use DataKit Proxy {#datakit}

Select a DataKit in the network that can access the external network as a proxy, and configure its proxy settings.

Detailed proxy input configure, please refer to [here](../integrations/proxy.md).

- Set the proxy mode of _proxy Datakit_.

Go to the `conf.d/` directory under the proxy DataKit installation directory and configure the proxy service in *datakit.conf*. As follows:

```toml
[dataway]
  urls = ["https://openway.guance.com?token=<YOUR-TOKEN>"]
  http_proxy = "http://<PROXY-IP:PROXY-PORT>"
```

Once configured, [restart DataKit](datakit-service-how-to.md#manage-service)。

Test whether the proxy service is ok, sending metrics to the workspace:

```shell
$ curl -x <PROXY-IP:PROXY-PORT> -v -X POST https://openway.guance.com/v1/write/metrics?token=<YOUR-TOKEN> -d "proxy_test,name=test c=123i"
...
```

If the proxy server works properly, the workspace will receive metric data `proxy_test,name=test c=123i`.

## Nginx {#nginx}

Proxy HTTPS traffic nginx uses a 4-layer transparent proxy mode, that is, it needs:

- a transparent nginx proxy server that can access the external network
- The client where datakit resides uses the hosts file for domain name configuration

### Configure the `Nginx` Proxy Service {#config-nginx-proxy}

```not-set
# Proxy HTTPS
stream {
    # resolver 114.114.114.114;
    # resolver_timeout 30s;
    server {
        listen 443;
        ssl_preread on;
        proxy_connect_timeout 10s;
        proxy_pass $ssl_preread_server_name:$server_port;
    }
}

http {
    ...
}
```

Proxy HTTP traffic here nginx uses 7 layers of transparent proxy (this section can be skipped if proxy HTTP is not needed): 

```not-set
# Proxy HTTP
http {
    # resolver 114.114.114.114;
    # resolver_timeout 30s;
    server {
        listen 80;
        location / {
            proxy_pass http://$http_host$request_uri;    # Configure forward proxy parameters
            proxy_set_header Host $http_host;            # Resolve nginx 503 error after "." in URL
            proxy_buffers 256 4k;                        # Configure cache size
            proxy_max_temp_file_size 0;                  # Turn off disk cache read and write to reduce I/O
            proxy_connect_timeout 30;                    # Agent connection timeout
            proxy_cache_valid 200 302 10m;
            proxy_cache_valid 301 1h;
            proxy_cache_valid any 1m;                    # Configure proxy server cache time
            proxy_send_timeout 60;
            proxy_read_timeout 60;
        }
    }

    // ... other configurations
}
```

### Load New Configuration and Test {#load-test}

```shell
$ nginx -t        # Test configuration
...

$ nginx -s reload # reload configuration
...
```

# Configure the Domain Name on the `Datakit` Agent Machine

Let's assume that `192.168.1.66` is the IP address of the nginx transparent proxy server.

```sh
$ sudo vi /etc/hosts

192.168.1.66 static.guance.com
192.168.1.66 openway.guance.com
192.168.1.66 dflux-dial.guance.com

192.168.1.66 static.dataflux.cn
192.168.1.66 openway.dataflux.cn
192.168.1.66 dflux-dial.dataflux.cn

192.168.1.66 zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com
```

On the agent machine, test whether the agent is normal:

=== "Linux/Unix Shell"

    ```shell
    curl -H "application/x-www-form-urlencoded; param=value" \
      -d 'proxy_test_nginx,name=test c=123i' \
      "https://openway.guance.com/v1/write/metrics?token=<YOUR-TOKEN>"
    ```

=== "Windows PowerShell"

    ```PowerShell
    curl -uri 'https://openway.guance.com/v1/write/metrics?token=<YOUR-TOKEN>' -Headers @{"param"="value"} -ContentType 'application/x-www-form-urlencoded' -body 'proxy_test_nginx,name=test c=123i' -method 'POST'
    ```
    
    Note: Some PowerShell machines report the mistake of `curl : : Request aborted: Failed to create SSL/TLS secure channel`. Because the server-side certificate encryption version number is not supported locally by default, you can view the supported protocols with the command `[Net.ServicePointManager]::SecurityProtocol`. If you want local support, you can do the following:
    
    ```PowerShell
    # 64 bit PowerShell
    Set-ItemProperty -Path 'HKLM:\SOFTWARE\Wow6432Node\Microsoft\.NetFramework\v4.0.30319' -Name 'SchUseStrongCrypto' -Value '1' -Type DWord
    
    # 32 bit PowerShell
    Set-ItemProperty -Path 'HKLM:\SOFTWARE\Microsoft\.NetFramework\v4.0.30319' -Name 'SchUseStrongCrypto' -Value '1' -Type DWord
    ```
    
    Close the PowerShell window, open a new PowerShell window, and execute the following code to see the supported protocols:
    
    ```PowerShell
    [Net.ServicePointManager]::SecurityProtocol
    ```
