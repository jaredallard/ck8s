# ck8sd controller

This component handles registering the CRDs and handling the scheduling of pods to computercraft machines


## Installation

TODO

```bash
$ kubectl create -f ./deploy 
# add the ca certificate to your certificate store
$ kubectl get secret "$(kubectl get serviceaccount ck8sd-computers -ojsonpath='{.secrets[0].name}')" \
   -ogo-template='{{index .data "ca.crt" | base64decode }}' > /tmp/kube.crt
# TODO: Instructions for CA cert addition
```
