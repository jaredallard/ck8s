-- -*- coding: utf-8 -*-
--
-- (c) 2019 ck8s project authors
--
-- Kublet: a ComputerPod runner for ck8s
local kubeletVersion = "0.1.0"
local uuidFile = "/var/ck8s/uuid"

local runningPods = {}
local tFilters    = {}
local _routines   = {}

print("started at "..os.time())

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
    error("failed to register computer: "..body)
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
print("updated computer status")

-- controlLoop handles the creation of new pods and the management of coroutines
local function controlLoop()
  while true do
    local resp, body = kubernetes:list("apis/computercraft.ck8sd.com/v1alpha1/namespaces/default/computerpods")
    if resp == nil then
      print("failed to list pods: "..body)
    end

    local ourPods = {}
    local foundPods = {}

    -- find pods that are assigned to us
    for _, pod in ipairs(body.items) do
      foundPods[pod.metadata.uid] = true
      if pod.status.assignedComputer == MACHINEID then
        table.insert(ourPods, pod)
      end
    end

    for podID, _ in pairs(runningPods) do
      -- check if our pods exist in the remove state, if it doesn't then
      -- it's likely been removed
      if foundPods[podID] ~= true then
        print("pod "..podID.." has been removed")

        -- remove the coroutine from the process manager, effectively killing it
        -- TODO(jaredallard): handle this more gracefully
        _routines[podID] = nil

        -- remove the pod from the in-memory store so that we don't handle it anymore
        runningPods[podID] = nil
      end
    end 

    -- handle the pods that have been assigned to us
    for _, pod in ipairs(ourPods) do
      local podID = pod.metadata.uid

      -- we need to create the pod
      if runningPods[podID] == nil then
        print("creating pod "..podID)

        -- track that the pod is running in-memory
        runningPods[podID] = true

        -- TODO: update the kubernetes apiserver with the status of the pod
        _routines[podID] = coroutine.create(function ()
          while true do
            print("POD: I'm running!")
            os.sleep(2)
          end
        end)
      end
    end

    os.sleep(5)
  end
end


print("starting process manager")
_routines["main"] = coroutine.create(controlLoop)

-- coroutine manager, adapted from: https://pastebin.com/2yYLywGK
local event, p1, p2, p3, p4, p5
while true do
  local i = 1
  for k, r in pairs(_routines) do
    if r then
      if r and coroutine.status(r) == "dead" then
        -- TODO(jaredallard): handle dead coroutine
        print("coroutine is dead")
        _routines[k] = nil
      else 
        if tFilters[k] == nil or tFilters[r] == event or event == "terminate" then
          local ok, param = coroutine.resume(r, event, p1, p2, p3, p4, p5)
          if not ok then
            error(param)
          else
            tFilters[r] = param
          end

          -- remove dead coroutines
          if coroutine.status(r) == "dead" then
            _routines[k] = nil
          end
        end
      end
    end
  end
  event, p1, p2, p3, p4, p5 = os.pullEventRaw()
end
