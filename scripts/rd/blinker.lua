function service(req)
  c = req.text:sub(1, 1)
  n = tonumber(req.text:sub(2))

  if c == 'C' then
    r = {count = n}
  elseif c == 'R' then
    r = {red = n ~= 0}
  elseif c == 'G' then
    r = {green = n ~= 0}
  end

  publish("ws/blinker", r)
end
