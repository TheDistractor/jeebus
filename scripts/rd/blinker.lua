function service(req)
  c = req.text:sub(1, 1)
  n = tonumber(req.text:sub(2))

  if c == 'C' then
    publish('/blinker/count', n)
  elseif c == 'R' then
    publish('/blinker/red', n ~= 0)
  elseif c == 'G' then
    publish('/blinker/green', n ~= 0)
  end
end
