# ck8sd controller

This component handles registering the CRDs and handling the scheduling of pods to computercraft machines


## Installation

```bash
$ kubectl create -f ./deploy 

# add the ca certificate to your certificate store
$ export JAVA_HOME="$(dirname $(dirname $(readlink -f $(command -v java))) | sed 's/\/jre//')"
$ kubectl get secret "$(kubectl get serviceaccount ck8sd-computers -ojsonpath='{.secrets[0].name}')" \
   -ogo-template='{{index .data "ca.crt" | base64decode }}' > /tmp/kube.crt
$ keytool -importcert -file /tmp/kube.crt -keystore cacerts
# password is: changeit
```

Now you have the controller running, and have trusted the ca provided!

## Types

There are 3 types provided currently: `Computer`, `ComputerPod` and `ComputerDeployment`

### Computer

A `Computer` is essentially a `Node` type, it has `KubeletReady` conditions and uses `NodeStatus` as the status type.

A computer represents an in-game computer. It stores the in-game ID in the spec, but uses a one-time generated UUID as the identifier
that is stored on the computer as it's "identity".

```yaml
apiVersion: computercraft.ck8sd.com/v1alpha1
kind: Computer
metadata:
  namespace: default
  name: uuid
spec:
  id: 0
```

### ComputerPod

Much like the native `Pod` type, but this doesn't deal with containers (it's minecraft, and [ccdocker](https://github.com/jaredallard/ccdocker) needs help).

Supports a URL for source code to download, and will support more in the future.

```yaml
apiVersion: computercraft.ck8sd.com/v1alpha1
kind: ComputerPod
metadata:
  namespace: default
  name: my-pod
spec:
  url: https://gist.githubusercontent.com/jaredallard/2e6d625cac81270ebc27ab94c4d5dc5f/raw/574f6907ae1d67307696d2b27206d05e6d47fbeb/daemon.lua
```

### ComputerDeployment

Much like the native `Deployment` type, this creates `ComputerPods` based on the number of replicas. Uses the ComputerPod type under `.spec.template`.

Pods are created in a statefulset-like fashion `deploymentName-#`

```yaml
apiVersion: computercraft.ck8sd.com/v1alpha1
kind: ComputerDeployment
metadata:
  namespace: default
  name: my-deployment
spec:
  replicas: 1
  template:
    spec:
      url: https://gist.githubusercontent.com/jaredallard/2e6d625cac81270ebc27ab94c4d5dc5f/raw/574f6907ae1d67307696d2b27206d05e6d47fbeb/daemon.lua
```