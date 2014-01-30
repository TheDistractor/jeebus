print("hello from TRY")
-- print(package.path)

function service(req)
  print("request:", req, req.c)
  publish("blah", req)
  return {"reply", req}
end
