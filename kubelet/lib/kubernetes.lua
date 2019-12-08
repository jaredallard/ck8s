-- Kubernetes Library
local kubernetes = {
  token = "",
  host = ""
}

local function getHeaders(config)
  return {
    Authorization = "Bearer " .. config.token
  }
end

-- initialize the kubernetes client
function kubernetes:init(host, token)
  if type(self) ~= 'table' then
    error("kubernetes must be called with :")
  end

  self.token = token
  self.host = host
end

-- post a endpoint
function kubernetes:post(path, data)
  if type(self) ~= 'table' then
    error("kubernetes must be called with :")
  end

  local body = json:encode(data)
  local headers = getHeaders(self)
  headers["Content-Type"] = "application/json"
  local resp, err = http.post(self.host.."/"..path, body, headers)
  if resp == nil then
    return nil, err
  end
  
  local data = json:decode(resp.readAll())
  resp.close()

  return resp, data
end

-- patch a endpoint
function kubernetes:patch(path, data)
  if type(self) ~= 'table' then
    error("kubernetes must be called with :")
  end

  local body = json:encode(data)
  if body == nil then
    error("provided invaild json")
  end

  local headers = getHeaders(self)

  -- TODO: add support for application/json-patch+json (list of ops)
  headers["Content-Type"] = "application/merge-patch+json"
  local resp, err = http.post({
    url = self.host.."/"..path,
    method = "PATCH",
    body = body,
    headers = headers
  })
  if resp == nil then
    return nil, err
  end
  
  local data = json:decode(resp.readAll())
  resp.close()

  return resp, data
end

-- get an endpoint
function kubernetes:get(path)
  if type(self) ~= 'table' then
    error("kubernetes must be called with :")
  end

  local resp, err = http.get(self.host.."/"..path, getHeaders(self))
  if resp == nil then
    return nil, err
  end

  local data = json:decode(resp.readAll())
  resp.close()

  return resp, data
end

-- list returns a list of pods
function kubernetes:list(path, fieldSelector)
  if type(self) ~= 'table' then
    error("kubernetes must be called with :")
  end

  local url = self.host.."/"..path
  if fieldSelector ~= nil then
    url = url .. "?fieldSelector=" .. fieldSelector
  end

  local resp, err = http.get(url, getHeaders(self))
  if resp == nil then
    return nil, err
  end

  local data = json:decode(resp.readAll())
  resp.close()

  return resp, data
end

return kubernetes