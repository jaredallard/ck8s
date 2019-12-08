-- -*- coding: utf-8 -*-
--
-- (c) 2019 ck8s project authors
--
-- Kublet: a ComputerPod runner for ck8s
local kubeletVersion = "0.1.0"
local uuidFile = "/var/ck8s/uuid" 

kubernetes:init(config.kubernetes, config.svc_account_token)

MACHINEID = ""
if not fs.exists(uuidFile) then
  local id = string.lower(tostring(uuid.Generate()))
  
  local f = fs.open(uuidFile, "w")
  f.write(id)
  f.close()

  MACHINEID = id
  print("starting computer registration: "..config.kubernetes)

  local computer = {
    apiVersion = "computercraft.ck8sd.com/v1alpha1",
    kind = "Computer",
    metadata = {
      namespace = "default",
      name = id,
    },
    spec = {
      ID = os.getComputerID(),
    },
  }

  local resp, body = kubernetes:post("apis/computercraft.ck8sd.com/v1alpha1/namespaces/default/computers", computer)
  if resp == nil then
    error("failed to register node: "..body)
  end
else
  -- read our machine ID from a file if it wasn't just registered
  local f = fs.open(uuidFile, "r")
  MACHINEID = f.readAll()
  f.close()
end

local status = {
  status = {
    phase = "Running",
    nodeInfo = {
      machineID = tostring(MACHINEID),
      kernelVersion = os.version(),
      kubeletVersion = kubeletVersion,
      operatingSystem = "craftos",
      architecture = "java"
    }
  }
}

local resp, body = kubernetes:patch("apis/computercraft.ck8sd.com/v1alpha1/namespaces/default/computers/"..MACHINEID.."/status", status)
if resp == nil then
  print("[warn] failed to update node status: "..body)
end

print(json:encode_pretty(body))