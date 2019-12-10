# ck8s kubelet

kublet is the component that handles the creation of `ComputerPod` types, and other things related to `Computer` types.

## Installation

TODO: Install script

Obtain the serviceaccount token:

```bash
$ kubectl get secret "$(kubectl get serviceaccount ck8sd-computers -ojsonpath='{.secrets[0].name}')" \
    -ogo-template='{{index .data "token" | base64decode }}'
```

Edit `config.lua`

```lua
local config = {
  version           = 1,

  -- edit this to be the public facing URL for your Kubernetes cluster
  -- try: kubectl cluster-info (Kubernetes master value)
  kubernetes        = "https://127.0.0.1:6443",

  -- place the token from earlier here
  svc_account_token = "",
}
```

Now create a "shippable" build, via an *in-game computer*

```bash
Howl combine
```

Now copy the `build/kubelet` file onto a floppy disk, and copy it to `startup` on every node. :)