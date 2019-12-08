-- UUID version 4 (random) API for ComputerCraft by TurboTuTone
-- http://en.wikipedia.org/wiki/Universally_unique_identifier
-- Format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx , where x is any hexadecimal digit and y is one of 8, 9, A, or B
-- Partially based on: http://developer.coronalabs.com/code/uuidguid-string-generator-coronalua by FrankS

local uuid = {}
 
--  Generate, Generates UUID
-- @returns string containing the uuid
function uuid.Generate()
  local chars = {"0","1","2","3","4","5","6","7","8","9","A","B","C","D","E","F"}
  local uuid = {[9]="-",[14]="-",[15]="4",[19]="-",[24]="-"}
  uuid[20] = chars[math.random (9,12)]
  for i = 1,36 do
          if(uuid[i]==nil)then
                  uuid[i] = chars[math.random (16)]
          end
  end
  return table.concat(uuid)
end

--  Validate, validates a uuid string.
-- Params:
-- @uuid = string containing the uuid.
-- @notrim (optional) = if set to false, turns off trimming of leading and trailing whitespaces.
-- @returns bool, true on validation, false if not valid.
function uuid.Validate(uuid, notrim)
  if not notrim == false or notrim == nil then
          uuid = uuid:gsub("^%s*(.-)%s*$", "%1")
  end
  if #uuid==36 then
          local fields = {}
          local valid = { [1] = 8, [2] = 4, [3] = 4, [4] = 4, [5] = 12 }
          uuid:gsub("([^%-]+)", function(a) fields[#fields+1] = a end)
          for i=1,#valid do
                  if fields[i] == nil or #fields[i] ~= valid[i] then
                          return false
                  end
          end
          if uuid:find("^[0-9ABCDEF]*%-[0-9ABCDEF]*%-4[0-9ABCDEF]*%-[89AB][0-9ABCDEF]*%-[0-9ABCDEF]*$") ~= nil then
                  return true
          end
  end
  return false
end

--  Compare, validates 2 uuid's then compares them for a match
-- Params:
-- @id1,id2 = uuid's to compare
-- @notrim (optional) = if set to false, turns off trimming of leading and trailing whitespaces.
-- @returns bool, true if uuid's match else false.
function uuid.Compare(id1, id2, notrim)
  if uuid.Validate(id1, notrim) and uuid.Validate(id2, notrim) then
          if id1 == id2 then return true end
  end
  return false
end

return uuid