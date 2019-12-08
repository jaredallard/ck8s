-- logging library

local log = {
  _VERSION = "1.0.0"
}

function log.logger(prefix, color, message)
  -- prefix w/ time
  write(os.time().." ")
  -- color the prefix
  term.setTextColor(colors[color])
  write(prefix.." ")
  term.setTextColor(colors.white)

  -- print out the rest of the message
  print(message)
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