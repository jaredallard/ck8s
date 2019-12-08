-- logging library

local log = {
  _VERSION = "1.0.0"
}

function log.logger(prefix, color, message)
  -- prefix w/ time
  local time = tostring(os.clock())
  if #time == 5 then
    time = time .. " "
  elseif #time == 4 then -- make the time look better when <5 chars
    time = time .. "  "
  elseif #time == 3 then
    time = time .. "   "
  end

  write(time.." ")
  -- color the prefix
  term.setTextColor(colors[color])
  write(prefix.." ")
  term.setTextColor(colors.white)

  -- print out the rest of the message
  print(message)
end

-- fatal is error, but it restarts the computer in a set amount of seconds
function log:Fatal(msg)
  self:Error(msg)
  os.sleep(2)
  os.reboot()
end

function log:Info(msg)
  self.logger("[info]", "cyan", msg)
end

function log:Warn(msg)
  self.logger("[warn]","yellow", msg)
end

function log:Error(msg)
  self.logger("[error]", "red", msg)
end

return log