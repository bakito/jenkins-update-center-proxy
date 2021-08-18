# jenkins-update-center-proxy

This proxy allows rewrites jenkins update site catalog json files if used in a detached environment, where plugins can
only be downloaded via a local mirror. All occurrences of the origin jenkins plugin update site will be replaced by the
mirror url.

## Configuration

The proxy can be configured by using env variables:

| Name | Description           |
| ---| --- |
| REPO_PROXY_URL | The fqdn of the repository proxy/mirror |
| USE_REPO_PROXY_FOR_DOWNLOAD | If __true__ the download URL's of all plugins get replaced by the proxy url.  |
| PORT | The port to run this application on. Default: 8080 |
| CONTEXT_PATH | The context path this application is served. Default: "/" |
| TLS_INSECURE_SKIP_VERIFY | If enable TLS verification to the proxy/mirror is skipped.|
| OFFLINE_DIR | Tn static environments, the json files can be provided as files, this property defines where the application can find them. |

## Docker images

```bash
docker pull ghcr.io/bakito/jenkins-update-center-proxy
# or 
docker pull quay.io/bakito/jenkins-update-center-proxy
```

