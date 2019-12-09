-- -*- coding: utf-8 -*-
--
-- (c) 2019 ck8s project authors
--
-- Kublet: a ComputerPod runner for ck8s
local uuidFile = "/var/ck8s/uuid"

local runningPods = {}
local tFilters    = {}
local _routines   = {}

-- globals
MACHINEID = ""
KUBELETVERSION = "v0.1.0"

print("ck8s kubelet version "..KUBELETVERSION)
log:Info("started at (in-game time): "..os.time())

kubernetes:init(config.kubernetes, config.svc_account_token)

local resp, body = kubernetes:version()
if resp == nil then
  log:Fatal("failed to reach kubernetes: "..body)
end

log:Info("remote kubernetes is version: "..body.gitVersion)
log:Info("remote kubernetes is running on: "..body.platform)


-- register our computer with Kubernetes
-- TODO(jaredallard): add retries
if not fs.exists(uuidFile) then
  local id = string.lower(tostring(uuid.Generate()))

  MACHINEID = id
  local resp, body = kubernetes:registerComputer()
  if resp == nil then
    print("ComputerID: "..MACHINEID)
    log:Fatal("failed to register computer: "..body)
  end

  local f = fs.open(uuidFile, "w")
  f.write(id)
  f.close()
else
  -- read our machine ID from a file if it wasn't just registered
  local f = fs.open(uuidFile, "r")
  MACHINEID = f.readAll()
  f.close()
end


local resp, err = kubernetes:updateComputerStatus()
if resp == nil then
  print("ComputerID: "..MACHINEID)
  log:Fatal("failed to update computer status: "..err)
end

log:Info("registered computer status with Kubernetes")

-- create pod creates a new pod, it does so by fetching source code from a provider
-- and then runs it inside of a coroutine
local function createPod(pod)
  log:Info("creating pod "..pod.metadata.uid)
  local resp, err = kubernetes:updateComputerPodStatus(pod.metadata.name, {
    status = {
      phase = "Running",
      message = "",
      reason = "",
    }
  })
  if resp == nil then
    log:Warn("failed to update computerpod status: "..err)
  end

  -- TODO(jaredallard): have better support for things other than URLs
  if pod.spec.url == "" then
    log:Error("pod doesn't have a URL")
    return nil, "invalid pod"
  end

  log:Info("downloading source code from pod url")
  local resp, err = http.get(pod.spec.url)
  if resp == nil then
    log:Error("failed to download source code: "..err)
    return nil, "failed to download source"
  end

  local src = resp.readAll()
  resp.close()

  local srcFile = "/tmp/src/"..pod.metadata.uid..".lua"
  local f = fs.open(srcFile, "w")
  f.write(src)
  f.close()

  -- TODO(jaredallard): logic to create the pod here
  log:Info("creating coroutine for pod")
  return coroutine.create(function ()
    os.run({}, srcFile)
  end)
end

-- removePod removes a pod from the in-memory store and updates it's remote status
-- TODO(jaredallard): add support for errored state
local function removePod(podID)
  log:Info("cleaning up pod "..podID)

  local pod = runningPods[podID]
  local resp, err = kubernetes:updateComputerPodStatus(pod.metadata.name, {
    status = {
      phase = "Terminated",
      reason = "",
      status = "Pod has terminated"
    }
  })
  if resp == nil then
    log:Warn("failed to update computerpod status: "..err)
  end

  -- remove the coroutine from the process manager, effectively killing it
  -- TODO(jaredallard): handle this more gracefully
  _routines[podID] = nil

  -- remove the pod from the in-memory store so that we don't handle it anymore
  runningPods[podID] = nil

  -- check if we have source code
  local srcFile = "/tmp/src/"..pod.metadata.uid..".lua"
  if fs.exists(srcFile) then
    fs.delete(srcFile)
  end
end

-- controlLoop handles the creation of new pods and the management of coroutines
local function controlLoop()
  while true do
    local resp, err = kubernetes:computerReady()
    if resp == nil then
      log:Warn("failed to post ready status: "..err)
    end

    local resp, body = kubernetes:getComputerPods()
    if resp == nil then
      log:Warn("failed to list pods: "..body)
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
        log:Info("pod "..podID.." has been removed")
        removePod(podID)
      end
    end 

    -- handle the pods that have been assigned to us
    for _, pod in ipairs(ourPods) do
      local podID = pod.metadata.uid

      -- we need to create the pod
      if runningPods[podID] == nil then
        -- track that the pod is running in-memory
        runningPods[podID] = pod

        local r, err = createPod(pod)

        log:Info("created pod "..podID)
        if r == nil then
          log:Warn("failed to run pod")
          local resp, err = kubernetes:updateComputerPodStatus(pod.metadata.name, {
            status = {
              phase = "Failed",
              reason = "FailedExecution",
              message = "failed to create pod: "..err
            }
          })
          if resp == nil then
            log:Warn("failed to notify k8s of failed pod: "..err)
          end
        else
          _routines[podID] = r
          log:Info("started pod "..podID)
        end
      end
    end

    os.sleep(5)
  end
end


log:Info("starting process manager")
_routines["main"] = coroutine.create(controlLoop)

-- coroutine manager, adapted from: https://pastebin.com/2yYLywGK
local event, p1, p2, p3, p4, p5
while true do
  for k, r in pairs(_routines) do
    if r then
      if r and coroutine.status(r) == "dead" then
        removePod(k)
      else
        if tFilters[k] == nil or tFilters[r] == event or event == "terminate" then
          local ok, param = coroutine.resume(r, event, p1, p2, p3, p4, p5)
          if not ok then
            log:Info("caught terminate signal, cleaning up")
            for k in pairs(_routines) do
              -- skip main control loop
              if k ~= "main" then
                removePod(k)
              end
            end
            error(param)
          else
            tFilters[r] = param
          end

          -- remove dead coroutines
          if coroutine.status(r) == "dead" then
            removePod(k)
          end
        end
      end
    end
  end
  event, p1, p2, p3, p4, p5 = os.pullEventRaw()
end
